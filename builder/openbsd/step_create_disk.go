package openbsd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	name := []string{config.VMName, "qcow2"}
	name_qcow := strings.Join(name, ".")

	var diskFullPaths, diskSizes []string

	ui.Say("Creating required virtual machine disks")
	
	// The 'main' or 'default' disk
	
	diskFullPaths = append(diskFullPaths, filepath.Join(config.OutputDir, name_qcow))
	diskSizes = append(diskSizes, fmt.Sprintf("%s", config.DiskSize))

	for i, diskFullPath := range diskFullPaths {
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskFullPath, diskSizes[i])
		command := []string{
			"create",
			"-s",
			diskSizes[i],
			diskFullPath,
		}

		command = append(command,
			diskFullPath,
			diskSizes[i])
		
		if err := driver.Vmctl(command...); err != nil {
			err := fmt.Errorf("Error creating hard drive: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("vmctl_disk_paths", diskFullPaths)
	
	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {}
