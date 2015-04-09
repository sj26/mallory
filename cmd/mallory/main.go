package main

import (
	"flag"
	m "github.com/justmao945/mallory"
	"net/http"
	"os"
)

var L = m.L

func main() {
	c := &m.Config{}
	flag.StringVar(&c.LocalServer, "addr", "127.0.0.1:1315", "mallory server address")
	flag.StringVar(&c.RemoteServer, "remote", "", "remote server address")
	flag.StringVar(&c.PrivateKey, "ssh-key", "$HOME/.ssh/id_rsa", "ssh private key file")
	flag.Parse()

	c.PrivateKey = os.ExpandEnv(c.PrivateKey)

	L.Printf("Starting...\n")
	srv, err := m.NewServer(c)
	if err != nil {
		L.Fatalln(err)
	}

	L.Printf("Listen and serve HTTP proxy on %s\n", c.LocalServer)
	L.Printf("Remote SSH server: %s\n", c.RemoteServer)
	L.Fatalln(http.ListenAndServe(c.LocalServer, srv))
}
