package types

import (
	"path/filepath"
	"strings"
)

type Process struct {
	Pid              int
	PidNamespaceId   int64
	MountNamespaceId int64
	NetNamespaceId   int64

	Cmd          string
	CmdTruncated bool

	Args          []string
	ArgsTruncated bool
}

func (p Process) MatchComm(name string) bool {
	filename := p.Comm()
	if len(filename) > 15 {
		filename = filename[:15]
	}
	return name == filename
}

func (p Process) FormatArgs() string {
	s := strings.Join(p.Args, " ")
	if p.ArgsTruncated {
		s += "..."
	}
	return s
}

func (p Process) Comm() string {
	return filepath.Base(p.Cmd)
}
