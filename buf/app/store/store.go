package store

import (
	"github.com/gliderlabs/comlab/pkg/com"
	"github.com/gliderlabs/exp/buf/app/api"
)

func init() {
	com.Register("store", struct{}{},
		com.Option("backend", "store.memory", "Store backend"))
}

type Store interface {
	FetchByID(id api.BufferID) (api.Buffer, error)
	FetchOrCreate(id api.BufferID) (api.Buffer, error)
}

func Selected() Store {
	backend := com.Select(com.GetString("backend"), new(Store))
	if backend == nil {
		panic("Unable to find selected backend: " + com.GetString("backend"))
	}
	return backend.(Store)
}
