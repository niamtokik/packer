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
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	name := config.VMName

	if config.DiskImage && !config.UseBackingFile {
		return multistep.ActionContinue
	}
	
	var diskFullPaths, diskSizes []string

	ui.Say("Creating required virtual machine disks")
	// The 'main' or 'default' disk
	diskFullPaths = append(diskFullPaths, filepath.Join(config.OutputDir, name))
	diskSizes = append(diskSizes, fmt.Sprintf("%s", config.DiskSize))
	
	// Additional disks
	if len(config.AdditionalDiskSize) > 0 {
		for i, diskSize := range config.AdditionalDiskSize {
			path := filepath.Join(config.OutputDir, fmt.Sprintf("%s-%d", name, i+1))
			diskFullPaths = append(diskFullPaths, path)
			size := fmt.Sprintf("%s", diskSize)
			diskSizes = append(diskSizes, size)
		}
	}

	for i, diskFullPath := range diskFullPaths {
		log.Printf("[INFO] Creating disk with Path: %s and Size: %s", diskFullPath, diskSizes[i])
		command := []string{
			"create",
			"-s",
			diskSizes[i],
			diskFullPath,
		}

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

