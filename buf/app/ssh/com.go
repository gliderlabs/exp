package ssh

import (
	"github.com/gliderlabs/comlab/pkg/com"
)

func init() {
	com.Register("buf.ssh", &Component{})
}

type Component struct{}
