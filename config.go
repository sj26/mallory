package mallory

// Memory representation for mallory.json
type Config struct {
	// private file file
	PrivateKey string `json:"id_rsa"`
	// local addr to listen and serve, default is 127.0.0.1:1315
	LocalServer string `json:"local"`
	// remote addr to connect, e.g. ssh://user@linode.my:22
	RemoteServer string `json:"remote"`
}
