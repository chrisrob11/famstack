package cmds

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// VersionCommand returns the version command configuration
func VersionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Show FamStack version",
		Action:  showVersionMain,
	}
}

// showVersionMain displays the current version (main command)
func showVersionMain(c *cli.Context) error {
	version := Version
	fmt.Printf("FamStack version: %s\n", version)
	return nil
}
