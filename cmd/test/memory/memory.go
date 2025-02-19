package memory

import (
	"context"
	"runtime"
	"sync"

	"github.com/pingme998/rclone/cmd"
	"github.com/pingme998/rclone/cmd/test"
	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/operations"
	"github.com/spf13/cobra"
)

func init() {
	test.Command.AddCommand(commandDefinition)
}

var commandDefinition = &cobra.Command{
	Use:   "memory remote:path",
	Short: `Load all the objects at remote:path into memory and report memory stats.`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 1, command, args)
		fsrc := cmd.NewFsSrc(args)
		cmd.Run(false, false, command, func() error {
			ctx := context.Background()
			objects, _, err := operations.Count(ctx, fsrc)
			if err != nil {
				return err
			}
			objs := make([]fs.Object, 0, objects)
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			var mu sync.Mutex
			err = operations.ListFn(ctx, fsrc, func(o fs.Object) {
				mu.Lock()
				objs = append(objs, o)
				mu.Unlock()
			})
			if err != nil {
				return err
			}
			runtime.GC()
			runtime.ReadMemStats(&after)
			usedMemory := after.Alloc - before.Alloc
			fs.Logf(nil, "%d objects took %d bytes, %.1f bytes/object", len(objs), usedMemory, float64(usedMemory)/float64(len(objs)))
			fs.Logf(nil, "System memory changed from %d to %d bytes a change of %d bytes", before.Sys, after.Sys, after.Sys-before.Sys)
			return nil
		})
	},
}
