// +build linux freebsd

package mount

import (
	"runtime"
	"testing"

	"github.com/pingme998/rclone/vfs/vfstest"
)

func TestMount(t *testing.T) {
	if runtime.NumCPU() <= 2 {
		t.Skip("FIXME skipping mount tests as they lock up on <= 2 CPUs - See: https://github.com/pingme998/rclone/issues/3154")
	}
	vfstest.RunTests(t, false, mount)
}
