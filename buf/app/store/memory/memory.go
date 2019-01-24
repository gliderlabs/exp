package memory

import (
	"sync"

	"github.com/gliderlabs/comlab/pkg/com"
	"github.com/gliderlabs/exp/buf/app/api"
	"github.com/gliderlabs/exp/buf/lib/multibuf"
)

func init() {
	com.Register("store.memory", &Component{})
}

type Component struct {
	sync.Mutex
	buffers map[string]*buffer
}

func (c *Component) DaemonInitialize() error {
	c.buffers = make(map[string]*buffer)
	return nil
}

func (c *Component) FetchByID(id api.BufferID) (api.Buffer, error) {
	c.Lock()
	defer c.Unlock()
	buf, exists := c.buffers[id.String()]
	if !exists {
		return nil, nil
	}
	return buf, nil
}

func (c *Component) FetchOrCreate(id api.BufferID) (api.Buffer, error) {
	c.Lock()
	defer c.Unlock()
	buf, exists := c.buffers[id.String()]
	if !exists || (buf.Closed() && buf.Buffered() == 0) {
		buf = newBuffer(id)
		c.buffers[id.String()] = buf
	}
	return buf, nil
}

type buffer struct {
	multibuf.MulticastBuffer
	id api.BufferID
}

func (b *buffer) ID() api.BufferID {
	return b.id
}

func newBuffer(id api.BufferID) *buffer {
	return &buffer{
		MulticastBuffer: multibuf.NewMulticastBuffer(),
		id:              id,
	}
}
