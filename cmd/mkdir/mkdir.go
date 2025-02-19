package mkdir

import (
	"context"
	"strings"

	"github.com/pingme998/rclone/cmd"
	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/operations"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
}

var commandDefinition = &cobra.Command{
	Use:   "mkdir remote:path",
	Short: `Make the path if it doesn't already exist.`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fdst := cmd.NewFsDir(args)
		if !fdst.Features().CanHaveEmptyDirectories && strings.Contains(fdst.Root(), "/") {
			fs.Logf(fdst, "Warning: running mkdir on a remote which can't have empty directories does nothing")
		}
		cmd.Run(true, false, command, func() error {
			return operations.Mkdir(context.Background(), fdst, "")
		})
	},
}
