package openbsd

import (
	"fmt"
	"os"
	"log"
)

const BuilderId = "transcend.openbsd"

type Artifact struct {
	dir   string
	f     []string
	state map[string]interface{}
}

func (*Artifact) BuilderId() string {
	log.Println("BuilderId")
	return BuilderId
}

func (a *Artifact) Files() []string {
	log.Println("Files")
	return a.f
}

func (*Artifact) Id() string {
	log.Println("Id")
	return "VM"
}

func (a *Artifact) String() string {
	log.Println("String")
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	log.Println("State")
	return a.state[name]
}

func (a *Artifact) Destroy() error {
	log.Println("Destroy")
	return os.RemoveAll(a.dir)
}
