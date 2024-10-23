package tajarin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"go.uber.org/zap"
)

const (
	ConfigPersistentPeers = "persistent_peers"
	DefaultListenAddress  = "0.0.0.0:8546"
	DefaultP2PPort        = "26656"
)

type generateCfg struct {
	outputPath        string
	chainID           string
	genesisTime       int64
	blockMaxTxBytes   int64
	blockMaxDataBytes int64
	blockMaxGas       int64
	blockTimeIota     int64
}

type validatorAddCfg struct {
	power       int64
	name        string
	pubKey      string
	address     string
	genesisPath string
}

type JsonTajarinRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	PubKey  string `json:"pubKey" validate:"required"`
	P2PKey  string `json:"p2pKey" validate:"required"`
	P2PHost string `json:"p2pHost" validate:"required"`
	P2PPort string `json:"p2pPort,omitempty"`
}

func (jt *JsonTajarinRequest) toP2PEndpointString() string {
	return fmt.Sprintf("%s@%s:%s", jt.P2PKey, jt.P2PHost, jt.P2PPort)
}

type JsonTajarinResponse struct {
	Genesis json.RawMessage   `json:"genesis"`
	Config  map[string]string `json:"config"`
}

type ConfigValue []string

type TCPListener struct {
	maxSubs           int64
	addr              string
	genesisPath       string
	logger            *zap.Logger
	openConnections   []net.Conn
	validatorsGenesis []validatorAddCfg
	configItems       map[string]ConfigValue
}

func NewTCPListener(logger *zap.Logger, addr string, maxSubs int64) *TCPListener {
	return &TCPListener{
		maxSubs:           maxSubs,
		addr:              addr,
		genesisPath:       "genesis.json",
		logger:            logger,
		validatorsGenesis: []validatorAddCfg{},
		configItems:       map[string]ConfigValue{},
	}
}

// Serve serves the JSON-RPC server
func (tl *TCPListener) Serve(ctx context.Context) error {
	var sem sync.Mutex
	var subscribers int64

	if tl.maxSubs <= 0 {
		tl.logger.Sugar().Fatalf("Insufficient Number of Subribers: %d", tl.maxSubs)
	}

	listener, err := net.Listen("tcp", tl.addr)
	if err != nil {
		return err
	}

	// Close the listener when the application closes.
	defer listener.Close()
	defer tl.logger.Info("TCP server shut down")
	defer func() { // close opened connections
		one := make([]byte, 1)
		for _, conn := range tl.openConnections {
			if r, _ := conn.Read(one); r == 0 {
				conn.Close()
			}
		}
	}()
	tl.logger.Info(
		"TCP server started",
		zap.String("address", listener.Addr().String()),
	)

	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			tl.logger.Fatal(err.Error())
		}

		if subscribers > tl.maxSubs {
			conn.Close()
			break
		}
		sem.Lock()
		subscribers += 1
		tl.logger.Info("New connection added")
		// Handle connections in the same goroutine.
		// using go routines will imply a channel
		err = tl.handleRequest(conn)
		if err != nil {
			tl.logger.Info("Incoming request failed")
			conn.Close()
			subscribers -= 1
		}
		// add connetion
		tl.openConnections = append(tl.openConnections, conn)
		sem.Unlock()
		if subscribers == tl.maxSubs {
			break
		}
	}

	tl.logger.Info("Response ready to be created")
	execGenerate(&generateCfg{
		outputPath: tl.genesisPath,
	})
	for _, validatorCfg := range tl.validatorsGenesis {
		execValidatorAdd(&validatorCfg)
	}

	jtResp := JsonTajarinResponse{}
	jsonFile, err := os.OpenFile(tl.genesisPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	// Read the file into a byte array
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	// Unmarshal the byte array into the struct
	err = json.Unmarshal(byteValue, &jtResp.Genesis)
	if err != nil {
		return err
	}
	// configs in response
	tempMap := map[string]string{}
	for configKey, configValues := range tl.configItems {
		tempMap[configKey] = strings.Join(configValues, ",")
	}
	jtResp.Config = tempMap

	// marshal the struct back to a JSON string
	marshaledJSON, err := json.Marshal(jtResp)
	if err != nil {
		return err
	}

	// notify all the connection
	for _, conn := range tl.openConnections {
		conn.Write(marshaledJSON)
		conn.Close()
	}

	return nil
}

// Handles incoming requests.
func (tl *TCPListener) handleRequest(conn net.Conn) error {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	readLen, err := conn.Read(buf)
	if err != nil {
		return err
	}

	var jsonTajarin JsonTajarinRequest
	// Parse the JSON data
	err = json.Unmarshal(buf[:readLen], &jsonTajarin)
	if err != nil {
		return err
	}

	// add validator cfg
	tl.validatorsGenesis = append(tl.validatorsGenesis, validatorAddCfg{
		name:        jsonTajarin.Name,
		address:     jsonTajarin.Address,
		pubKey:      jsonTajarin.PubKey,
		power:       1,
		genesisPath: tl.genesisPath,
	})

	// add general config
	addItemToMap(tl.configItems, ConfigPersistentPeers, jsonTajarin.toP2PEndpointString())
	return nil
}

func addItemToMap(myMap map[string]ConfigValue, key string, item string) {
	// Check if the key already exists in the map
	if _, exists := myMap[key]; exists {
		// If it exists, append the new item to the existing list
		myMap[key] = append(myMap[key], item)
	} else {
		// If the key doesn't exist, create a new list with the item
		myMap[key] = []string{item}
	}
}

func execValidatorAdd(cfg *validatorAddCfg) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check the validator address
	address, err := crypto.AddressFromString(cfg.address)
	if err != nil {
		return fmt.Errorf("invalid validator address, %w", err)
	}

	// Check the voting power
	if cfg.power < 1 {
		return err
	}

	// Check the name
	if cfg.name == "" {
		return err
	}

	// Check the public key
	pubKey, err := crypto.PubKeyFromBech32(cfg.pubKey)
	if err != nil {
		return fmt.Errorf("invalid validator public key, %w", err)
	}

	// Check the public key matches the address
	if pubKey.Address() != address {
		return err
	}

	validator := types.GenesisValidator{
		Address: address,
		PubKey:  pubKey,
		Power:   cfg.power,
		Name:    cfg.name,
	}

	// Check if the validator exists
	for _, genesisValidator := range genesis.Validators {
		// There is no need to check if the public keys match
		// since the address is derived from it, and the derivation
		// is checked already
		if validator.Address == genesisValidator.Address {
			return err
		}
	}

	// Add the validator
	genesis.Validators = append(genesis.Validators, validator)

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	fmt.Printf(
		"Validator with address %s added to genesis file\n",
		cfg.address,
	)

	return nil
}

func execGenerate(cfg *generateCfg) error {
	// Start with the default configuration
	genesis := getDefaultGenesis()

	// Set the genesis time
	if cfg.genesisTime > 0 {
		genesis.GenesisTime = time.Unix(cfg.genesisTime, 0)
	}

	// Set the chain ID
	if cfg.chainID != "" {
		genesis.ChainID = cfg.chainID
	}

	// Set the max tx bytes
	if cfg.blockMaxTxBytes > 0 {
		genesis.ConsensusParams.Block.MaxTxBytes = cfg.blockMaxTxBytes
	}

	// Set the max data bytes
	if cfg.blockMaxDataBytes > 0 {
		genesis.ConsensusParams.Block.MaxDataBytes = cfg.blockMaxDataBytes
	}

	// Set the max block gas
	if cfg.blockMaxGas > 0 {
		genesis.ConsensusParams.Block.MaxGas = cfg.blockMaxGas
	}

	// Set the block time IOTA
	if cfg.blockTimeIota > 0 {
		genesis.ConsensusParams.Block.TimeIotaMS = cfg.blockTimeIota
	}

	// Validate the genesis
	if validateErr := genesis.ValidateAndComplete(); validateErr != nil {
		return fmt.Errorf("unable to validate genesis, %w", validateErr)
	}

	// Save the genesis file to disk
	if saveErr := genesis.SaveAs(cfg.outputPath); saveErr != nil {
		return fmt.Errorf("unable to save genesis, %w", saveErr)
	}

	fmt.Printf("Genesis successfully generated at %s\n", cfg.outputPath)
	return nil
}

// getDefaultGenesis returns the default genesis config
func getDefaultGenesis() *types.GenesisDoc {
	return &types.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         "dev",
		ConsensusParams: types.DefaultConsensusParams(),
	}
}
