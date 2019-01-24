package api

import (
	"fmt"
	"path"
	"strings"

	"github.com/gliderlabs/exp/buf/lib/multibuf"
)

type Mode string

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)

type Buffer interface {
	multibuf.MulticastBuffer
	ID() BufferID
}

type BufferID struct {
	Owner string
	Name  string
}

func (id BufferID) String() string {
	return path.Join(id.Owner, id.Name)
}

type BufferOperation struct {
	User        string
	BufferID    BufferID
	Mode        Mode
	Addressable bool
	ReadFlush   bool
	ReadForever bool
	WriteMore   bool
	WriteTee    bool
}

func ParseBufferOperation(user string, args []string) (*BufferOperation, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("not enough args") // TODO: better error
	}
	op := &BufferOperation{User: user}
	mainArg := args[0]
	if strings.HasPrefix(mainArg, "@") {
		op.Addressable = true
	}
	mainArg = strings.TrimPrefix(mainArg, "@")
	cmdParts := strings.SplitN(mainArg, ":", 2)
	mainArg = cmdParts[0]
	if len(cmdParts) > 1 {
		op.Command = append([]string{cmdParts[1]}, args[1:]...)
	} else {
		for k, v := range CommandAbbreviations {
			if strings.HasSuffix(mainArg, k) {
				if strings.HasSuffix(mainArg, fmt.Sprintf("%s%s", k, k)) {
					op.Command = append([]string{v[1]}, args[1:]...)
				} else {
					op.Command = append([]string{v[0]}, args[1:]...)
				}
			}
		}
		if len(op.Command) == 0 {
			op.Command = append([]string{DefaultCommand}, args[1:]...)
		}
	}
	mainArg = strings.Trim(mainArg, SpecialCharset)
	idParts := strings.SplitN(mainArg, "/", 2)
	if len(idParts) > 1 {
		op.BufferID.Owner = idParts[0]
		op.BufferID.Name = idParts[1]
	} else {
		op.BufferID.Owner = user
		op.BufferID.Name = idParts[0]
	}
	return op, nil
}
