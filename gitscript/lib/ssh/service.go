package ssh

import (
	"net"

	"github.com/gliderlabs/comlab/pkg/com"
	"github.com/gliderlabs/comlab/pkg/log"
	"github.com/gliderlabs/ssh"
)

func NewServer() *ssh.Server {
	server := &ssh.Server{}
	server.SetOption(ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		handlers := com.Enabled(new(AuthHandler), nil)
		if len(handlers) == 0 {
			log.Info("no auth handler")
			return false
		}
		return handlers[0].(AuthHandler).HandleAuth(ctx, key)
	}))
	server.Handle(func(sess ssh.Session) {
		for _, com := range com.Enabled(new(SessionHandler), nil) {
			com.(SessionHandler).HandleSSH(sess)
		}
	})
	return server
}

func (c *Component) Stop() {
	if c.listener != nil {
		c.listener.Close()
	}
}

func (c *Component) Serve() {
	server := NewServer()
	server.SetOption(ssh.HostKeyFile(com.GetString("hostkey_pem")))
	var err error
	c.listener, err = net.Listen("tcp", com.GetString("listen_addr"))
	if err != nil {
		panic(err)
	}
	log.Info("listening on", com.GetString("listen_addr"))
	server.Serve(c.listener)
}
