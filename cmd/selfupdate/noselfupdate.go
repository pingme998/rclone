// +build noselfupdate

package selfupdate

import (
	"github.com/pingme998/rclone/lib/buildinfo"
)

func init() {
	buildinfo.Tags = append(buildinfo.Tags, "noselfupdate")
}
