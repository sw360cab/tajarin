package tajarin

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap"
)

type subscriberCfg struct {
	JsonTajarinRequest
	listenAddress string
}

func NewSubscriberCmd(io commands.IO, logger *zap.Logger) *commands.Command {
	cfg := &subscriberCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "subscriber",
			ShortUsage: "subscriber [flags]",
			ShortHelp:  "spin up a Tajarin subscriber",
			LongHelp:   "Starts a process sending basic information about validator and waiting for a full configuration",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execSubscribe(cfg, logger)
		},
	)
}

func (c *subscriberCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.Name,
		"name",
		"",
		"name of the validator to add",
	)

	fs.StringVar(
		&c.Address,
		"address",
		"",
		"address of the validator to add",
	)

	fs.StringVar(
		&c.Name,
		"pub-key",
		"",
		"public key of the validator to add",
	)

	fs.StringVar(
		&c.Name,
		"p2p-address",
		"",
		"p2p node ID derived from the private key of the validator to add",
	)

	fs.StringVar(
		&c.Name,
		"p2p-host",
		"",
		"p2p node host address of the validator to add",
	)

	fs.StringVar(
		&c.Name,
		"p2p-port",
		DefaultP2PPort,
		fmt.Sprintf("(opt.) p2p node port of the validator to add [default:%s]", DefaultP2PPort),
	)

	fs.StringVar(
		&c.listenAddress,
		"listen-address",
		DefaultListenAddress,
		fmt.Sprintf("listening address of node synchronizer [deault: %s]", DefaultListenAddress),
	)
}

func execSubscribe(c *subscriberCfg, logger *zap.Logger) error {
	ts := TajarinSubscriber{}
	ts.Subscribe(JsonTajarinRequest(*&c.JsonTajarinRequest), c.listenAddress, logger)
	return nil
}
