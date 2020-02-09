package openbsd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	// "runtime"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/common/shutdowncommand"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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

	
	bootcommand.BootConfig `mapstructure:",squash"`
	OutputDir string `mapstructure:"output_directory" required:"false"`
	Format string `mapstructure:"format" required:"false"`
	SSHHostPortMin int `mapstructure:"ssh_host_port_min" required:"false"`
	SSHHostPortMax int `mapstructure:"ssh_host_port_max" required:"false"`
	SSHWaitTimeout time.Duration `mapstructure:"ssh_wait_timeout" required:"false"`
	VmctlArgs [][]string `mapstructure:"vmctlargs" required:"false"`	
	ISOSkipCache bool `mapstructure:"iso_skip_cache" required:"false"`
	VmctlBinary string `mapstructure:"vmctl_binary" required:"false"`
	DiskSize string `mapstructure:"disk_size" required:"false"`
	VMName string `mapstructure:"vm_name" required:"false"`
	ctx interpolate.Context
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec {
	log.Println("ConfigSpec")
	return b.config.FlatMapstructure().HCL2Spec()
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	log.Println("Prepare")
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"vmctlargs",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}
	
	var errs *packer.MultiError
	warnings := make([]string, 0)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare(&b.config.ctx)...)
	
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

	if b.config.OutputDir == "" {
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}
	
	if b.config.VmctlBinary == "" {
		b.config.VmctlBinary = "vmctl"
	}
	
	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	if b.config.SSHHostPortMin < 0 {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ssh_host_port_min must be positive"))
	}

	if b.config.SSHWaitTimeout != 0 {
		b.config.Comm.SSHTimeout = b.config.SSHWaitTimeout
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
	}
	
	isoWarnings, isoErrs := b.config.ISOConfig.Prepare(&b.config.ctx)
	warnings = append(warnings, isoWarnings...)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	errs = packer.MultiErrorAppend(errs, b.config.HTTPConfig.Prepare(&b.config.ctx)...)
	if es := b.config.Comm.Prepare(&b.config.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	
	if b.config.VmctlArgs == nil {
		b.config.VmctlArgs = make([][]string, 0)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}
	
	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	log.Println("Run")
	
	driver, err := b.newDriver(b.config.VmctlBinary)	
	if err != nil {
		return nil, fmt.Errorf("Failed creating vmctl driver: %s", err)
	}

	steps := []multistep.Step{}

	if !b.config.ISOSkipCache {
		steps = append(steps, &common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			Extension:    b.config.TargetExtension,
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
			Url:          b.config.ISOUrls,
		},
		)
	} else {
		steps = append(steps, &stepSetISO{
			ResultKey: "iso_path",
			Url:       b.config.ISOUrls,
		},
		)
	}

	steps = append(steps, new(stepPrepareOutputDir),
		new(stepCreateDisk),
	)

	steprun := &stepRun{}
	steps = append(steps,steprun,)
	
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(b.config.OutputDir, visit); err != nil {
		return nil, err
	}
	
	artifact := &Artifact{
		dir:   b.config.OutputDir,
		f:     files,
		state: make(map[string]interface{}),
	}

	artifact.state["diskName"] = b.config.VMName
	diskpaths, ok := state.Get("qemu_disk_paths").([]string)
	if ok {
		artifact.state["diskPaths"] = diskpaths
	}
	artifact.state["diskSize"] = b.config.DiskSize
	
	return artifact, nil
}

func (b *Builder) newDriver(vmctlBinary string) (Driver, error) {
	log.Println("newDriver")
	
	vmctlPath, err := exec.LookPath(vmctlBinary)
	if err != nil {
		return nil, err
	}

	driver := &VmctlDriver{
		VmctlPath: vmctlPath,
	}

	if err := driver.Verify(); err != nil {
		return nil, err
	}
	
	return driver, nil
}
