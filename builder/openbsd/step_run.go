package openbsd

import (
	"context"
	"fmt"
	// "log"
	"path/filepath"
	// "strings"

	// "github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	// "github.com/hashicorp/packer/template/interpolate"
)

type stepRun struct {
	BootDrive string
	Message string
}

type vmctlArgsTemplateData struct {
	OutputDir string
	Name string
}

func (s *stepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say(s.Message)

	command, err := getCommandArgs(s.BootDrive, state)
	if err != nil {
		err := fmt.Errorf("Error processing VmctlArgs: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := driver.Vmctl(command...); err != nil {
		err := fmt.Errorf("Error launching VM: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}


func (s *stepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if err := driver.Stop(); err != nil {
		ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
	}
}

func getCommandArgs(bootDrive string, state multistep.StateBag) ([]string, error) {
	config := state.Get("config").(*Config)
	vmName := config.VMName
	imgPath := filepath.Join(config.OutputDir, vmName)
	isoPath := state.Get("iso_path").(string)
	
	defaultArgs := make(map[string]interface{})
	outArgs := make([]string, 0)
	
	defaultArgs["start"] = vmName
	outArgs = append(outArgs, "start")
	outArgs = append(outArgs, "-B")
	outArgs = append(outArgs, "cdrom")
	outArgs = append(outArgs, "-b")
	outArgs = append(outArgs, isoPath)
	outArgs = append(outArgs, "-d")
	outArgs = append(outArgs, imgPath)
	outArgs = append(outArgs, vmName)
	
	return outArgs, nil
}
