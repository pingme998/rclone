package fshttp

import (
	"context"
	"net"
	"time"

	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/accounting"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func dialContext(ctx context.Context, network, address string, ci *fs.ConfigInfo) (net.Conn, error) {
	return NewDialer(ctx).DialContext(ctx, network, address)
}

// Dialer structure contains default dialer and timeout, tclass support
type Dialer struct {
	net.Dialer
	timeout time.Duration
	tclass  int
}

// NewDialer creates a Dialer structure with Timeout, Keepalive,
// LocalAddr and DSCP set from rclone flags.
func NewDialer(ctx context.Context) *Dialer {
	ci := fs.GetConfig(ctx)
	dialer := &Dialer{
		Dialer: net.Dialer{
			Timeout:   ci.ConnectTimeout,
			KeepAlive: 30 * time.Second,
		},
		timeout: ci.Timeout,
		tclass:  int(ci.TrafficClass),
	}
	if ci.BindAddr != nil {
		dialer.Dialer.LocalAddr = &net.TCPAddr{IP: ci.BindAddr}
	}
	return dialer
}

// Dial connects to the address on the named network.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext connects to the address on the named network using
// the provided context.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return c, err
	}
	if d.tclass != 0 {
		if addr, ok := c.RemoteAddr().(*net.IPAddr); ok {
			if addr.IP.To16() != nil && addr.IP.To4() == nil {
				err = ipv6.NewConn(c).SetTrafficClass(d.tclass)
			} else {
				err = ipv4.NewConn(c).SetTOS(d.tclass)
			}
			if err != nil {
				return c, err
			}
		}
	}
	return newTimeoutConn(c, d.timeout)
}

// A net.Conn that sets a deadline for every Read or Write operation
type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

// create a timeoutConn using the timeout
func newTimeoutConn(conn net.Conn, timeout time.Duration) (c *timeoutConn, err error) {
	c = &timeoutConn{
		Conn:    conn,
		timeout: timeout,
	}
	err = c.nudgeDeadline()
	return
}

// Nudge the deadline for an idle timeout on by c.timeout if non-zero
func (c *timeoutConn) nudgeDeadline() (err error) {
	if c.timeout == 0 {
		return nil
	}
	when := time.Now().Add(c.timeout)
	return c.Conn.SetDeadline(when)
}

// Read bytes doing idle timeouts
func (c *timeoutConn) Read(b []byte) (n int, err error) {
	// Ideally we would LimitBandwidth(len(b)) here and replace tokens we didn't use
	n, err = c.Conn.Read(b)
	accounting.TokenBucket.LimitBandwidth(accounting.TokenBucketSlotTransportRx, n)
	// Don't nudge if no bytes or an error
	if n == 0 || err != nil {
		return
	}
	// Nudge the deadline on successful Read or Write
	err = c.nudgeDeadline()
	return n, err
}

// Write bytes doing idle timeouts
func (c *timeoutConn) Write(b []byte) (n int, err error) {
	accounting.TokenBucket.LimitBandwidth(accounting.TokenBucketSlotTransportTx, len(b))
	n, err = c.Conn.Write(b)
	// Don't nudge if no bytes or an error
	if n == 0 || err != nil {
		return
	}
	// Nudge the deadline on successful Read or Write
	err = c.nudgeDeadline()
	return n, err
}
