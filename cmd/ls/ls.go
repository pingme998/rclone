package ls

import (
	"context"
	"os"

	"github.com/pingme998/rclone/cmd"
	"github.com/pingme998/rclone/cmd/ls/lshelp"
	"github.com/pingme998/rclone/fs/operations"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
}

var commandDefinition = &cobra.Command{
	Use:   "ls remote:path",
	Short: `List the objects in the path with size and path.`,
	Long: `
Lists the objects in the source path to standard output in a human
readable format with size and path. Recurses by default.

Eg

    $ rclone ls swift:bucket
        60295 bevajer5jef
        90613 canole
        94467 diwogej7
        37600 fubuwic

` + lshelp.Help,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fsrc := cmd.NewFsSrc(args)
		cmd.Run(false, false, command, func() error {
			return operations.List(context.Background(), fsrc, os.Stdout)
		})
	},
}
