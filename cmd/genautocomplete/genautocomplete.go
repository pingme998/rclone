package genautocomplete

import (
	"github.com/pingme998/rclone/cmd"
	"github.com/spf13/cobra"
)

func init() {
	cmd.Root.AddCommand(completionDefinition)
}

var completionDefinition = &cobra.Command{
	Use:   "genautocomplete [shell]",
	Short: `Output completion script for a given shell.`,
	Long: `
Generates a shell completion script for rclone.
Run with --help to list the supported shells.
`,
}
