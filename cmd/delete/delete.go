package delete

import (
	"context"
	"strings"

	"github.com/pingme998/rclone/cmd"
	"github.com/pingme998/rclone/fs/config/flags"
	"github.com/pingme998/rclone/fs/operations"
	"github.com/spf13/cobra"
)

var (
	rmdirs = false
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
	cmdFlags := commandDefinition.Flags()
	flags.BoolVarP(cmdFlags, &rmdirs, "rmdirs", "", rmdirs, "rmdirs removes empty directories but leaves root intact")
}

var commandDefinition = &cobra.Command{
	Use:   "delete remote:path",
	Short: `Remove the files in path.`,
	// Warning! "|" will be replaced by backticks below
	Long: strings.ReplaceAll(`
Remove the files in path.  Unlike |purge| it obeys include/exclude
filters so can be used to selectively delete files.

|rclone delete| only deletes files but leaves the directory structure
alone. If you want to delete a directory and all of its contents use
the |purge| command.

If you supply the |--rmdirs| flag, it will remove all empty directories along with it.
You can also use the separate command |rmdir| or |rmdirs| to
delete empty directories only.

For example, to delete all files bigger than 100 MiB, you may first want to
check what would be deleted (use either):

    rclone --min-size 100M lsl remote:path
    rclone --dry-run --min-size 100M delete remote:path

Then proceed with the actual delete:

    rclone --min-size 100M delete remote:path

That reads "delete everything with a minimum size of 100 MiB", hence
delete all files bigger than 100 MiB.

**Important**: Since this can cause data loss, test first with the
|--dry-run| or the |--interactive|/|-i| flag.
`, "|", "`"),
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fsrc := cmd.NewFsSrc(args)
		cmd.Run(true, false, command, func() error {
			if err := operations.Delete(context.Background(), fsrc); err != nil {
				return err
			}
			if rmdirs {
				fdst := cmd.NewFsDir(args)
				return operations.Rmdirs(context.Background(), fdst, "", true)
			}
			return nil
		})
	},
}
