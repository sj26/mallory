package mallory

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"sync"
)

//
type SSH struct {
	// global config file
	Cfg *Config
	// connect URL
	URL *url.URL
	// SSH client
	Client *ssh.Client
	// SSH client config
	CliCfg *ssh.ClientConfig
	// direct fetcher
	Direct *Direct
	// atomic Dial
	mutex sync.RWMutex
}

// Create and initialize
func NewSSH(c *Config) (self *SSH, err error) {
	self = &SSH{
		Cfg:    c,
		CliCfg: &ssh.ClientConfig{},
	}
	// e.g.  ssh://user:passwd@192.168.1.1:1122
	self.URL, err = url.Parse(c.RemoteServer)
	if err != nil {
		return
	}

	if self.URL.User != nil {
		self.CliCfg.User = self.URL.User.Username()
	} else {
		u, err := user.Current()
		if err != nil {
			return self, err
		}
		// u.Name is the full name, should not be used
		self.CliCfg.User = u.Username
	}

	// 1) try agent first
	for {
		sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			break
		}
		agent := agent.NewClient(sock)
		signers, err := agent.Signers()
		if err != nil {
			break
		}
		self.CliCfg.Auth = append(self.CliCfg.Auth, ssh.PublicKeys(signers...))
	}
	// 2) try RSA keyring
	for {
		id_rsa := c.PrivateKey
		pem, err := ioutil.ReadFile(id_rsa)
		if err != nil {
			L.Printf("ReadFile %s failed:%s\n", c.PrivateKey, err)
			break
		}
		signer, err := ssh.ParsePrivateKey(pem)
		if err != nil {
			L.Printf("ParsePrivateKey %s failed:%s\n", c.PrivateKey, err)
			break
		}
		self.CliCfg.Auth = append(self.CliCfg.Auth, ssh.PublicKeys(signer))
		// stop !!
		break
	}
	// 3) try password
	for {
		if self.URL.User == nil {
			break
		}
		if pass, ok := self.URL.User.Password(); ok {
			self.CliCfg.Auth = append(self.CliCfg.Auth, ssh.Password(pass))
		}
		// stop here!!
		break
	}

	if len(self.CliCfg.Auth) == 0 {
		//TODO: keyboard intercative
		err = errors.New("Invalid auth method, please add password or generate ssh keys")
		return
	}

	self.Client, err = ssh.Dial("tcp", self.URL.Host, self.CliCfg)
	if err != nil {
		return
	}

	dial := func(network, addr string) (c net.Conn, err error) {
		// FIXME: unexported net.errClosing
		errClosing := errors.New("use of closed network connection")

		for i := 0; i < 3; i++ {
			if err != nil {
				// stop when ssh.Dial failed
				break
			}
			// need read lock, we'll reconnect Cli if is disconnected
			// use read write lock may slow down connection ?
			self.mutex.RLock()
			saveClient := self.Client
			if self.Client != nil {
				c, err = self.Client.Dial(network, addr)
			} else {
				// The reason why both Cli and err are nil is that, the previous round
				// connection is failed, which keeps self.Cli nil.
				err = errClosing
			}
			self.mutex.RUnlock()

			// We want to reconnect the network when disconnected.
			if err != nil && err.Error() == errClosing.Error() {
				// we may change the Cli, need write lock
				self.mutex.Lock()
				if saveClient == self.Client {
					if self.Client != nil {
						self.Client.Close()
					}
					self.Client, err = ssh.Dial("tcp", self.URL.Host, self.CliCfg)
				}
				self.mutex.Unlock()
				continue
			}
			// do not reconnect when no error or other errors
			break
		}
		return
	}

	self.Direct = &Direct{
		Tr: &http.Transport{Dial: dial},
	}
	return
}

func (self *SSH) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.Direct.ServeHTTP(w, r)
}

func (self *SSH) Connect(w http.ResponseWriter, r *http.Request) {
	self.Direct.Connect(w, r)
}
