package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/tajarin/tajarin"
	"go.uber.org/zap"
)

func main() {
	ioCommands := commands.NewDefaultIO()
	logger, _ := zap.NewProduction()
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "starts the tajarin synchronizer",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		tajarin.NewProducerCmd(ioCommands, logger),
		tajarin.NewSubscriberCmd(ioCommands, logger),
	)
	cmd.Execute(context.Background(), os.Args[1:])
}
