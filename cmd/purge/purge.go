package purge

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
	Use:   "purge remote:path",
	Short: `Remove the path and all of its contents.`,
	Long: `
Remove the path and all of its contents.  Note that this does not obey
include/exclude filters - everything will be removed.  Use the ` + "`delete`" + `
command if you want to selectively delete files. To delete empty directories only,
use command ` + "`rmdir`" + ` or ` + "`rmdirs`" + `.

**Important**: Since this can cause data loss, test first with the
` + "`--dry-run` or the `--interactive`/`-i`" + ` flag.
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fdst := cmd.NewFsDir(args)
		cmd.Run(true, false, command, func() error {
			return operations.Purge(context.Background(), fdst, "")
		})
	},
}
