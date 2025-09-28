package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"famstack/internal/cmds"
)

func main() {
	app := &cli.App{
		Name:    "famstack",
		Usage:   "Family task management system",
		Version: cmds.Version,
		Commands: []*cli.Command{
			cmds.StartCommand(),
			cmds.EncryptionCommand(),
			cmds.UserCommand(),
			cmds.UpdateCommand(),
			cmds.VersionCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
