// Package mallory implements a simple http proxy support direct and GAE remote fetcher
package mallory

import (
	"net/http"
	"sync"
)

type Server struct {
	// config file
	Cfg *Config
	// direct fetcher
	Direct *Direct
	// ssh fetcher, to connect remote proxy server
	SSH *SSH
	// for serve http
	mutex sync.RWMutex
}

// Create and intialize
func NewServer(c *Config) (self *Server, err error) {
	ssh, err := NewSSH(c)
	if err != nil {
		return
	}

	self = &Server{
		Cfg:          c,
		Direct:       NewDirect(),
		SSH:          ssh,
	}
	return
}

// HTTP proxy accepts requests with following two types:
//  - CONNECT
//    Generally, this method is used when the client want to connect server with HTTPS.
//    In fact, the client can do anything he want in this CONNECT way...
//    The request is something like:
//      CONNECT www.google.com:443 HTTP/1.1
//    Only has the host and port information, and the proxy should not do anything with
//    the underlying data. What the proxy can do is just exchange data between client and server.
//    After accepting this, the proxy should response
//      HTTP/1.1 200 OK
//    to the client if the connection to the remote server is established.
//    Then client and server start to exchange data...
//
//  - non-CONNECT, such as GET, POST, ...
//    In this case, the proxy should redo the method to the remote server.
//    All of these methods should have the absolute URL that contains the host information.
//    A GET request looks like:
//      GET weibo.com/justmao945/.... HTTP/1.1
//    which is different from the normal http request:
//      GET /justmao945/... HTTP/1.1
//    Because we can be sure that all of them are http request, we can only redo the request
//    to the remote server and copy the reponse to client.
//
func (self *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	L.Printf("%s %s %s\n", r.Method, r.RequestURI, r.Proto)

	if r.Method == "CONNECT" {
		self.SSH.Connect(w, r)
	} else if r.URL.IsAbs() {
		// This is an error if is not empty on Client
		r.RequestURI = ""
		// If no Accept-Encoding header exists, Transport will add the headers it can accept
		// and would wrap the response body with the relevant reader.
		r.Header.Del("Accept-Encoding")
		// curl can add that, see
		// http://homepage.ntlworld.com/jonathan.deboynepollard/FGA/web-proxy-connection-header.html
		r.Header.Del("Proxy-Connection")
		// Connection is single hop Header:
		// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
		// 14.10 Connection
		//   The Connection general-header field allows the sender to specify
		//   options that are desired for that particular connection and MUST NOT
		//   be communicated by proxies over further connections.
		r.Header.Del("Connection")
		self.SSH.ServeHTTP(w, r)
	} else {
		L.Printf("%s is not a full URL path\n", r.RequestURI)
	}
}
