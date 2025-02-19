package test

import (
	"github.com/pingme998/rclone/cmd"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(Command)
}

// Command definition for cobra
var Command = &cobra.Command{
	Use:   "test <subcommand>",
	Short: `Run a test command`,
	Long: `Rclone test is used to run test commands.

Select which test comand you want with the subcommand, eg

    rclone test memory remote:

Each subcommand has its own options which you can see in their help.

**NB** Be careful running these commands, they may do strange things
so reading their documentation first is recommended.
`,
}
