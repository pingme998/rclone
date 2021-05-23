// +build linux darwin,amd64

package mount2

import (
	"testing"

	"github.com/pingme998/rclone/vfs/vfstest"
)

func TestMount(t *testing.T) {
	vfstest.RunTests(t, false, mount)
}
