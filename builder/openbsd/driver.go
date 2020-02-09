package openbsd

import (
	// "bufio"
	// "bytes"
	"fmt"
	// "io"
	"log"
	"os/exec"
	// "regexp"
	// "strings"
	// "sync"
	// "syscall"
	// "time"
	// "unicode"
)

type Driver interface {
	Stop() error
	Vmctl(vmctlArgs ...string) error
}

type VmctlDriver struct {
	VmctlPath string
	vmctlCmd *exec.Cmd
}

func (d *VmctlDriver) Stop() error {
	return nil
}

func (d *VmctlDriver) Vmctl(vmctlArgs ...string) error {
	log.Printf("Executing %s: %#v", d.VmctlPath, vmctlArgs)
	cmd := exec.Command(d.VmctlPath, vmctlArgs...)
	err := cmd.Start()
	
	if err != nil {
		err = fmt.Errorf("Error starting VM: %s", err)
		return err
	}
	return nil
}

func (d *VmctlDriver) Verify() error {
	return nil
}
