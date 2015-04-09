# Mallory

Simple HTTP/HTTPS proxy using SSH.

## How

 * `go get github.com/sj26/mallory/cmd/mallory`
 * `$GOPATH/bin/mallory -addr "127.0.0.1:8080" -remote "ssh://example.com:22"`
 * `curl --proxy http://127.0.0.1:8080 http://example.com`

You can also export some environment variables to use the proxy by default:

 * `export HTTP_PROXY=http://127.0.0.1:8080 HTTPS_PROXY=http://127.0.0.1:8080`

It works when configured in a browser, too.

## Configuration

Mallory will try to authenticate using your ssh agent, an ssh private key file (it must be unencrypted), or a username/password provided in the remote address (`ssh://me:secret@example.com:22`).

The following flags are available:

 * `addr="127.0.0.1:1315"`: mallory server address
 * `remote=""`: remote server address
 * `ssh-key="$HOME/.ssh/id_rsa"`: ssh private key file

