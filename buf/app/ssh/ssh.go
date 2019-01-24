package ssh

import (
	"fmt"
	"io"
	"os"

	"github.com/gliderlabs/exp/buf/app/api"
	"github.com/gliderlabs/exp/buf/app/store"
	"github.com/gliderlabs/exp/buf/lib/multibuf"
	"github.com/gliderlabs/ssh"
)

func (c *Component) HandleSSH(sess ssh.Session) {
	args := sess.Command()
	if len(args) == 0 {
		return
	}
	op, err := api.ParseBufferOperation(sess.User(), args)
	if err != nil {
		panic(err)
	}
	switch op.Command[0] {
	case "append":
		buf, err := store.Selected().FetchOrCreate(op.BufferID)
		if err != nil {
			panic(err)
		}
		io.Copy(buf, sess)
		buf.Close()
	case "append-more":
		buf, err := store.Selected().FetchOrCreate(op.BufferID)
		if err != nil {
			panic(err)
		}
		io.Copy(buf, sess)
	case "flush":
		buf, err := store.Selected().FetchByID(op.BufferID)
		if err != nil {
			panic(err)
		}
		if buf != nil {
			io.Copy(multibuf.Flusher{sess}, multibuf.WriterTo{buf})
		}
	case "peek":
		buf, err := store.Selected().FetchByID(op.BufferID)
		if err != nil {
			panic(err)
		}
		if buf != nil {
			io.Copy(multibuf.Peeker{sess}, multibuf.WriterTo{buf})
		}
	case "flush-forever":
		buf, err := store.Selected().FetchOrCreate(op.BufferID)
		if err != nil {
			panic(err)
		}
		io.Copy(multibuf.Flusher{sess}, multibuf.WriterTo{buf})
	case "peek-forever":
		buf, err := store.Selected().FetchOrCreate(op.BufferID)
		if err != nil {
			panic(err)
		}
		io.Copy(multibuf.Peeker{sess}, multibuf.WriterTo{buf})
	default:
		fmt.Fprintln(sess.Stderr(), "Unrecognized command")
		os.Exit(1)
	}
}

func (c *Component) AuthenticateSSH(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}
