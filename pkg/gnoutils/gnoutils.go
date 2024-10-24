package gnoutils

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type GenerateCfg struct {
	OutputPath        string
	chainID           string
	genesisTime       int64
	blockMaxTxBytes   int64
	blockMaxDataBytes int64
	blockMaxGas       int64
	blockTimeIota     int64
}

type ValidatorAddCfg struct {
	Power       int64
	Name        string
	PubKey      string
	Address     string
	GenesisPath string
}

type ConfigValue []string

// getDefaultGenesis returns the default genesis config
func getDefaultGenesis() *types.GenesisDoc {
	return &types.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         "dev",
		ConsensusParams: types.DefaultConsensusParams(),
	}
}

// Add a validator to genesis file
// Referencing github.com/gnolang/gno/gno.land/cmd/gnoland/genesis_generate.go
func ExecGenerateGenesis(cfg *GenerateCfg) error {
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
	if saveErr := genesis.SaveAs(cfg.OutputPath); saveErr != nil {
		return fmt.Errorf("unable to save genesis, %w", saveErr)
	}
	return nil
}

// Add a validator to genesis file
// Referencing github.com/gnolang/gno/gno.land/cmd/gnoland/genesis_validator_add.go
func ExecValidatorAdd(cfg *ValidatorAddCfg) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Check the validator address
	address, err := crypto.AddressFromString(cfg.Address)
	if err != nil {
		return fmt.Errorf("invalid validator address, %w", err)
	}

	// Check the voting power
	if cfg.Power < 1 {
		return err
	}

	// Check the name
	if cfg.Name == "" {
		return err
	}

	// Check the public key
	pubKey, err := crypto.PubKeyFromBech32(cfg.PubKey)
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
		Power:   cfg.Power,
		Name:    cfg.Name,
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
	if err := genesis.SaveAs(cfg.GenesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}
	return nil
}

// Handy method to get/put a list of strings element into a map
func AddItemToMap(refMap map[string]ConfigValue, key string, item string) {
	if _, exists := refMap[key]; exists {
		// If it exists, append the new item to the existing list
		refMap[key] = append(refMap[key], item)
	} else {
		// If the key doesn't exist, create a new list with the item
		refMap[key] = []string{item}
	}
}
