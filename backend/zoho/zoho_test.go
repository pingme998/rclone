// Test Zoho filesystem interface
package zoho_test

import (
	"testing"

	"github.com/pingme998/rclone/backend/zoho"
	"github.com/pingme998/rclone/fstest/fstests"
)

// TestIntegration runs integration tests against the remote
func TestIntegration(t *testing.T) {
	fstests.Run(t, &fstests.Opt{
		RemoteName: "TestZoho:",
		NilObject:  (*zoho.Object)(nil),
	})
}
