package openbsd

import (
	"context"
	"fmt"
	// "log"
	// "path/filepath"
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
	ui := state.Get("ui").(packer.Ui)
	ui.Say("test")
	return multistep.ActionContinue
}


func (s *stepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if err := driver.Stop(); err != nil {
		ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
	}
}
