package server

import (
	"github.com/gliderlabs/comlab/pkg/com"
)

func init() {
	com.Register("server", &Component{},
		com.Option("workdir", "local/tmp", "work directory"))
}

type Component struct{}
