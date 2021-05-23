// Test Sugarsync filesystem interface
package sugarsync_test

import (
	"testing"

	"github.com/pingme998/rclone/backend/sugarsync"
	"github.com/pingme998/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestSugarSync:Test",
		NilObject:  (*sugarsync.Object)(nil),
	})
}
