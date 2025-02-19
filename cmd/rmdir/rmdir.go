package rmdir

import (
	"context"

	"github.com/pingme998/rclone/cmd"
	"github.com/pingme998/rclone/fs/operations"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
}

var commandDefinition = &cobra.Command{
	Use:   "rmdir remote:path",
	Short: `Remove the empty directory at path.`,
	Long: `
This removes empty directory given by path. Will not remove the path if it
has any objects in it, not even empty subdirectories. Use
command ` + "`rmdirs`" + ` (or ` + "`delete`" + ` with option ` + "`--rmdirs`" + `)
to do that.

To delete a path and any objects in it, use ` + "`purge`" + ` command.
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fdst := cmd.NewFsDir(args)
		cmd.Run(true, false, command, func() error {
			return operations.Rmdir(context.Background(), fdst, "")
		})
	},
}
