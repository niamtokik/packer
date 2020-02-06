package openbsd

import (
	"context"
	// "errors"
	"fmt"
	// "log"
	// "os"
	"os/exec"
	// "path/filepath"
	"regexp"
	// "runtime"
	"strings"
	// "time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	// "github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/common/shutdowncommand"
	"github.com/hashicorp/packer/helper/communicator"
	// "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	// "github.com/hashicorp/packer/template/interpolate"
)



type Builder struct {
	config Config
	runner multistep.Runner
}

type Config struct {
	common.PackerConfig            `mapstructure:",squash"`
	common.HTTPConfig              `mapstructure:",squash"`
	common.ISOConfig               `mapstructure:",squash"`
	shutdowncommand.ShutdownConfig `mapstructure:",squash"`
	Comm                           communicator.Config `mapstructure:",squash"`
	common.FloppyConfig            `mapstructure:",squash"`		

	ISOSkipCache bool `mapstructure:"iso_skip_cache" required:"false"`
	VmctlBinary string `mapstructure:"vmctl_binary" required:"false"`
	DiskSize string `mapstructure:"disk_size" required:"false"`
	VMName string `mapstructure:"vm_name" required:"false"`
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	return b.config.FlatMapstructure().HCL2Spec()
}

func (b *Builder) Prepare(raw ...interface{}) ([]string, []string, error) {
	var errs *packer.MultiError
	warnings := make([]string, 0)
	
	if b.config.DiskSize == "" || b.config.DiskSize == "0" {
		b.config.DiskSize = "1024M"
	} else {
		re := regexp.MustCompile(`^[\d]+(b|k|m|g|t){0,1}$`)
		matched := re.MatchString(strings.ToLower(b.config.DiskSize))
		if !matched {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Invalid disk size."))
		} else {
			re = regexp.MustCompile(`^[\d]+$`)
			matched = re.MatchString(strings.ToLower(b.config.DiskSize))
			if matched {
				b.config.DiskSize = fmt.Sprintf("%sM", b.config.DiskSize)
			}
		}
	}
	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	_, err := b.newDriver(b.config.VmctlBinary)
	
	if err != nil {
		return nil, fmt.Errorf("Failed creating Qemu driver: %s", err)
	}

	return nil, nil
}

func (b *Builder) newDriver(vmctlBinary string) (Driver, error) {
	vmctlPath, err := exec.LookPath(vmctlBinary)
	if err != nil {
		return nil, err
	}

	driver := &VmctlDriver{
		VmctlPath: vmctlPath,
	}
	
	return driver, nil
}
