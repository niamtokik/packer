package openbsd

import (
	"fmt"
	"os"
)

const BuilderId = "transcend.openbsd"

type Artifact struct {
	dir   string
	f     []string
	state map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (*Artifact) Id() string {
	return "VM"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	return a.state[name]
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
