package openbsd

import (
	"context"
	"fmt"
	// "log"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	// _ := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName

	var diskFullPaths, diskSizes []string

	ui.Say("Creating required virtual machine disks")
	diskFullPaths = append(diskFullPaths, filepath.Join("config.OutputDir", name))
	diskSizes = append(diskSizes, fmt.Sprintf("%s", config.DiskSize))

	return multistep.ActionContinue
}

