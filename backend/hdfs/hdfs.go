// +build !plan9

package hdfs

import (
	"path"
	"strings"

	"github.com/pingme998/rclone/fs"
	"github.com/pingme998/rclone/fs/config"
	"github.com/pingme998/rclone/lib/encoder"
)

func init() {
	fsi := &fs.RegInfo{
		Name:        "hdfs",
		Description: "Hadoop distributed file system",
		NewFs:       NewFs,
		Options: []fs.Option{{
			Name:     "namenode",
			Help:     "hadoop name node and port",
			Required: true,
			Examples: []fs.OptionExample{{
				Value: "namenode:8020",
				Help:  "Connect to host namenode at port 8020",
			}},
		}, {
			Name:     "username",
			Help:     "hadoop user name",
			Required: false,
			Examples: []fs.OptionExample{{
				Value: "root",
				Help:  "Connect to hdfs as root",
			}},
		}, {
			Name: "service_principal_name",
			Help: `Kerberos service principal name for the namenode

Enables KERBEROS authentication. Specifies the Service Principal Name
(<SERVICE>/<FQDN>) for the namenode.`,
			Required: false,
			Examples: []fs.OptionExample{{
				Value: "hdfs/namenode.hadoop.docker",
				Help:  "Namenode running as service 'hdfs' with FQDN 'namenode.hadoop.docker'.",
			}},
			Advanced: true,
		}, {
			Name: "data_transfer_protection",
			Help: `Kerberos data transfer protection: authentication|integrity|privacy

Specifies whether or not authentication, data signature integrity
checks, and wire encryption is required when communicating the the
datanodes. Possible values are 'authentication', 'integrity' and
'privacy'. Used only with KERBEROS enabled.`,
			Required: false,
			Examples: []fs.OptionExample{{
				Value: "privacy",
				Help:  "Ensure authentication, integrity and encryption enabled.",
			}},
			Advanced: true,
		}, {
			Name:     config.ConfigEncoding,
			Help:     config.ConfigEncodingHelp,
			Advanced: true,
			Default:  (encoder.Display | encoder.EncodeInvalidUtf8 | encoder.EncodeColon),
		}},
	}
	fs.Register(fsi)
}

// Options for this backend
type Options struct {
	Namenode               string               `config:"namenode"`
	Username               string               `config:"username"`
	ServicePrincipalName   string               `config:"service_principal_name"`
	DataTransferProtection string               `config:"data_transfer_protection"`
	Enc                    encoder.MultiEncoder `config:"encoding"`
}

// xPath make correct file path with leading '/'
func xPath(root string, tail string) string {
	if !strings.HasPrefix(root, "/") {
		root = "/" + root
	}
	return path.Join(root, tail)
}
