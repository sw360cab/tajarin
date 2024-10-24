package tcp

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gnolang/tajarin/pkg/gnoutils"
	tajson "github.com/gnolang/tajarin/pkg/json"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	ConfigPersistentPeers = "persistent_peers"
	DefaultListenAddress  = "0.0.0.0:8546"
	DefaultP2PPort        = "26656"
)

type TCPListener struct {
	maxSubs           int64
	addr              string
	genesisPath       string
	logger            *zap.Logger
	openConnections   []net.Conn
	validatorsGenesis []gnoutils.ValidatorAddCfg
	configItems       map[string]gnoutils.ConfigValue
}

func NewTCPListener(logger *zap.Logger, addr string, maxSubs int64) *TCPListener {
	return &TCPListener{
		maxSubs:           maxSubs,
		addr:              addr,
		genesisPath:       "genesis.json",
		logger:            logger,
		validatorsGenesis: []gnoutils.ValidatorAddCfg{},
		configItems:       map[string]gnoutils.ConfigValue{},
	}
}

// Serve serves the JSON-RPC server
func (tl *TCPListener) Serve(ctx context.Context) error {
	var sem sync.Mutex
	var subscribers int64
	var jtResp tajson.JsonTajarinResponse

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

		sem.Lock()
		if subscribers > tl.maxSubs {
			conn.Close()
			break
		}
		subscribers += 1
		tl.logger.Info("New connection added")
		// Handle connections in the same goroutine.
		// Note: Using go routines will imply a channel
		err = tl.handleRequest(conn)
		if err != nil {
			tl.logger.Sugar().Error("Incoming request failed", err)
			jtResp = tajson.JsonTajarinResponse{}
			jtResp.Error = err.Error()
			tl.writeJsonAndCloseConnection(conn, jtResp)
			subscribers -= 1
		}
		// add connetion
		tl.openConnections = append(tl.openConnections, conn)
		if subscribers == tl.maxSubs {
			break
		}
		sem.Unlock()
	}

	tl.logger.Sugar().Infof("The required number of validators (%d) has subscribed. Creating a Response...", tl.maxSubs)
	// Generate Genesis File
	err = gnoutils.ExecGenerateGenesis(&gnoutils.GenerateCfg{
		OutputPath: tl.genesisPath,
	})
	if err != nil {
		return err
	}
	tl.logger.Sugar().Infof("Genesis successfully generated at %s\n", tl.genesisPath)

	// Add Validators to Genesis
	for _, validatorCfg := range tl.validatorsGenesis {
		err = gnoutils.ExecValidatorAdd(&validatorCfg)
		if err != nil {
			break
		}
		tl.logger.Sugar().Infof("Validator with address %s added to genesis file\n", *&validatorCfg.Address)
	}
	if err != nil {
		return err
	}

	// Generating response
	jtResp = tajson.JsonTajarinResponse{}

	// Add Genesis to response
	var genesisJson = json.RawMessage{}
	err = tl.marshallGenesisJson(&genesisJson)
	if err != nil {
		return err
	}
	jtResp.Genesis = genesisJson

	// Add Config to response
	finalConfigMap := map[string]string{}
	for configKey, configValues := range tl.configItems {
		finalConfigMap[configKey] = strings.Join(configValues, ",")
	}
	jtResp.Config = finalConfigMap

	// Marshal the struct back to a JSON string
	marshaledJSON, err := json.Marshal(jtResp)
	if err != nil {
		return err
	}

	// Notify all the connections
	for _, conn := range tl.openConnections {
		conn.Write(marshaledJSON)
		conn.Close()
	}

	return nil
}

// Write Json and Close connection
func (tl *TCPListener) writeJsonAndCloseConnection(conn net.Conn, jtResp tajson.JsonTajarinResponse) error {
	marshaledJSON, err := json.Marshal(jtResp)
	if err != nil {
		return err
	}
	conn.Write(marshaledJSON)
	conn.Close()
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

	var jsonTajarin tajson.JsonTajarinRequest
	// Parse the JSON data
	err = json.Unmarshal(buf[:readLen], &jsonTajarin)
	if err != nil {
		return err
	}

	err = validator.New().Struct(jsonTajarin)
	if err != nil {
		return err
	}

	// add validator cfg
	tl.validatorsGenesis = append(tl.validatorsGenesis, gnoutils.ValidatorAddCfg{
		Name:        jsonTajarin.Name,
		Address:     jsonTajarin.Address,
		PubKey:      jsonTajarin.PubKey,
		Power:       1,
		GenesisPath: tl.genesisPath,
	})

	// add general config
	gnoutils.AddItemToMap(tl.configItems, ConfigPersistentPeers, jsonTajarin.ToP2PEndpointString())
	return nil
}

// Marshall Genesis file
func (tl *TCPListener) marshallGenesisJson(genesisJson *json.RawMessage) error {
	// Add genesis file into a Json Object
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
	err = json.Unmarshal(byteValue, &genesisJson)
	if err != nil {
		return err
	}
	return nil
}
