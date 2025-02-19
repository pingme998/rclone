package seafile

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/lib/pacer"
)

const (
	minSleep      = 100 * time.Millisecond
	maxSleep      = 10 * time.Second
	decayConstant = 2 // bigger for slower decay, exponential
)

// Use only one pacer per server URL
var (
	pacers     map[string]*fs.Pacer
	pacerMutex sync.Mutex
)

func init() {
	pacers = make(map[string]*fs.Pacer, 0)
}

// getPacer returns the unique pacer for that remote URL
func getPacer(ctx context.Context, remote string) *fs.Pacer {
	pacerMutex.Lock()
	defer pacerMutex.Unlock()

	remote = parseRemote(remote)
	if existing, found := pacers[remote]; found {
		return existing
	}

	pacers[remote] = fs.NewPacer(
		ctx,
		pacer.NewDefault(
			pacer.MinSleep(minSleep),
			pacer.MaxSleep(maxSleep),
			pacer.DecayConstant(decayConstant),
		),
	)
	return pacers[remote]
}

// parseRemote formats a remote url into "hostname:port"
func parseRemote(remote string) string {
	remoteURL, err := url.Parse(remote)
	if err != nil {
		// Return a default value in the very unlikely event we're not going to parse remote
		fs.Infof(nil, "Cannot parse remote %s", remote)
		return "default"
	}
	host := remoteURL.Hostname()
	port := remoteURL.Port()
	if port == "" {
		if remoteURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return fmt.Sprintf("%s:%s", host, port)
}
