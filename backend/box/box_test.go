// Test Box filesystem interface
package box_test

import (
	"testing"

	"github.com/pingme998/rclone/backend/box"
	"github.com/pingme998/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestBox:",
		NilObject:  (*box.Object)(nil),
	})
}
