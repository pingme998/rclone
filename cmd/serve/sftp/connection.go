// +build !plan9

package sftp

import (
	"context"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/hash"
	"github.com/pingme998/rclone/vfs"
	"golang.org/x/crypto/ssh"
)

func describeConn(c interface {
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
}) string {
	return fmt.Sprintf("serve sftp %s->%s", c.RemoteAddr(), c.LocalAddr())
}

// Return the exit status of the command
type exitStatus struct {
	RC uint32
}

// The incoming exec command
type execCommand struct {
	Command string
}

var shellUnEscapeRegex = regexp.MustCompile(`\\(.)`)

// Unescape a string that was escaped by rclone
func shellUnEscape(str string) string {
	str = strings.Replace(str, "'\n'", "\n", -1)
	str = shellUnEscapeRegex.ReplaceAllString(str, `$1`)
	return str
}

// Info about the current connection
type conn struct {
	vfs      *vfs.VFS
	handlers sftp.Handlers
	what     string
}

// execCommand implements an extremely limited number of commands to
// interoperate with the rclone sftp backend
func (c *conn) execCommand(ctx context.Context, out io.Writer, command string) (err error) {
	binary, args := command, ""
	space := strings.Index(command, " ")
	if space >= 0 {
		binary = command[:space]
		args = strings.TrimLeft(command[space+1:], " ")
	}
	args = shellUnEscape(args)
	fs.Debugf(c.what, "exec command: binary = %q, args = %q", binary, args)
	switch binary {
	case "df":
		about := c.vfs.Fs().Features().About
		if about == nil {
			return errors.New("df not supported")
		}
		usage, err := about(ctx)
		if err != nil {
			return errors.Wrap(err, "About failed")
		}
		total, used, free := int64(-1), int64(-1), int64(-1)
		if usage.Total != nil {
			total = *usage.Total / 1024
		}
		if usage.Used != nil {
			used = *usage.Used / 1024
		}
		if usage.Free != nil {
			free = *usage.Free / 1024
		}
		perc := int64(0)
		if total > 0 && used >= 0 {
			perc = (100 * used) / total
		}
		_, err = fmt.Fprintf(out, `		Filesystem                   1K-blocks      Used Available Use%% Mounted on
/dev/root %d %d  %d  %d%% /
`, total, used, free, perc)
		if err != nil {
			return errors.Wrap(err, "send output failed")
		}
	case "md5sum", "sha1sum":
		ht := hash.MD5
		if binary == "sha1sum" {
			ht = hash.SHA1
		}
		var hashSum string
		if args == "" {
			// empty hash for no input
			if ht == hash.MD5 {
				hashSum = "d41d8cd98f00b204e9800998ecf8427e"
			} else {
				hashSum = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
			}
			args = "-"
		} else {
			node, err := c.vfs.Stat(args)
			if err != nil {
				return errors.Wrapf(err, "hash failed finding file %q", args)
			}
			if node.IsDir() {
				return errors.New("can't hash directory")
			}
			o, ok := node.DirEntry().(fs.ObjectInfo)
			if !ok {
				return errors.New("unexpected non file")
			}
			hashSum, err = o.Hash(ctx, ht)
			if err != nil {
				return errors.Wrap(err, "hash failed")
			}
		}
		_, err = fmt.Fprintf(out, "%s  %s\n", hashSum, args)
		if err != nil {
			return errors.Wrap(err, "send output failed")
		}
	case "echo":
		// special cases for rclone command detection
		switch args {
		case "'abc' | md5sum":
			if c.vfs.Fs().Hashes().Contains(hash.MD5) {
				_, err = fmt.Fprintf(out, "0bee89b07a248e27c83fc3d5951213c1  -\n")
				if err != nil {
					return errors.Wrap(err, "send output failed")
				}
			} else {
				return errors.New("md5 hash not supported")
			}
		case "'abc' | sha1sum":
			if c.vfs.Fs().Hashes().Contains(hash.SHA1) {
				_, err = fmt.Fprintf(out, "03cfd743661f07975fa2f1220c5194cbaff48451  -\n")
				if err != nil {
					return errors.Wrap(err, "send output failed")
				}
			} else {
				return errors.New("sha1 hash not supported")
			}
		default:
			_, err = fmt.Fprintf(out, "%s\n", args)
			if err != nil {
				return errors.Wrap(err, "send output failed")
			}
		}
	default:
		return errors.Errorf("%q not implemented\n", command)
	}
	return nil
}

// handle a new incoming channel request
func (c *conn) handleChannel(newChannel ssh.NewChannel) {
	fs.Debugf(c.what, "Incoming channel: %s\n", newChannel.ChannelType())
	if newChannel.ChannelType() != "session" {
		err := newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		fs.Debugf(c.what, "Unknown channel type: %s\n", newChannel.ChannelType())
		if err != nil {
			fs.Errorf(c.what, "Failed to reject unknown channel: %v", err)
		}
		return
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		fs.Errorf(c.what, "could not accept channel: %v", err)
		return
	}
	defer func() {
		err := channel.Close()
		if err != nil && err != io.EOF {
			fs.Debugf(c.what, "Failed to close channel: %v", err)
		}
	}()
	fs.Debugf(c.what, "Channel accepted\n")

	isSFTP := make(chan bool, 1)
	var command execCommand

	// Handle out-of-band requests
	go func(in <-chan *ssh.Request) {
		for req := range in {
			fs.Debugf(c.what, "Request: %v\n", req.Type)
			ok := false
			var subSystemIsSFTP bool
			var reply []byte
			switch req.Type {
			case "subsystem":
				fs.Debugf(c.what, "Subsystem: %s\n", req.Payload[4:])
				if string(req.Payload[4:]) == "sftp" {
					ok = true
					subSystemIsSFTP = true
				}
			case "exec":
				err := ssh.Unmarshal(req.Payload, &command)
				if err != nil {
					fs.Errorf(c.what, "ignoring bad exec command: %v", err)
				} else {
					ok = true
					subSystemIsSFTP = false
				}
			}
			fs.Debugf(c.what, " - accepted: %v\n", ok)
			err = req.Reply(ok, reply)
			if err != nil {
				fs.Errorf(c.what, "Failed to Reply to request: %v", err)
				return
			}
			if ok {
				// Wake up main routine after we have responded
				isSFTP <- subSystemIsSFTP
			}
		}
	}(requests)

	// Wait for either subsystem "sftp" or "exec" request
	if <-isSFTP {
		fs.Debugf(c.what, "Starting SFTP server")
		server := sftp.NewRequestServer(channel, c.handlers)
		defer func() {
			err := server.Close()
			if err != nil && err != io.EOF {
				fs.Debugf(c.what, "Failed to close server: %v", err)
			}
		}()
		err = server.Serve()
		if err == io.EOF || err == nil {
			fs.Debugf(c.what, "exited session")
		} else {
			fs.Errorf(c.what, "completed with error: %v", err)
		}
	} else {
		var rc = uint32(0)
		err := c.execCommand(context.TODO(), channel, command.Command)
		if err != nil {
			rc = 1
			_, errPrint := fmt.Fprintf(channel.Stderr(), "%v\n", err)
			if errPrint != nil {
				fs.Errorf(c.what, "Failed to write to stderr: %v", errPrint)
			}
			fs.Debugf(c.what, "command %q failed with error: %v", command.Command, err)
		}
		_, err = channel.SendRequest("exit-status", false, ssh.Marshal(exitStatus{RC: rc}))
		if err != nil {
			fs.Errorf(c.what, "Failed to send exit status: %v", err)
		}
	}
}

// Service the incoming Channel channel in go routine
func (c *conn) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go c.handleChannel(newChannel)
	}
}
