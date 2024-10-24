package tajarin

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/tajarin/pkg/tcp"
	"go.uber.org/zap"
)

type producerCfg struct {
	maxNodes      int64
	listenAddress string
}

func NewProducerCmd(io commands.IO, logger *zap.Logger) *commands.Command {
	cfg := &producerCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "producer",
			ShortUsage: "producer [flags]",
			ShortHelp:  "spin up a Tajarin producer",
			LongHelp:   "Starts a process waiting for validator subscribe to generate genesis file and config",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execStart(ctx, cfg, logger)
		},
	)
}

func (c *producerCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(
		&c.maxNodes,
		"max-nodes",
		0,
		"maximum number of nodes that can be syncronized",
	)

	fs.StringVar(
		&c.listenAddress,
		"listen-address",
		tcp.DefaultListenAddress,
		fmt.Sprintf("listening address of node synchronizer [deault: %s]", tcp.DefaultListenAddress),
	)
}

func execStart(ctx context.Context, c *producerCfg, logger *zap.Logger) error {
	producer := NewTajarinProducer(c.maxNodes, c.listenAddress)
	producer.ListenAndWait(ctx, logger)

	return nil
}
