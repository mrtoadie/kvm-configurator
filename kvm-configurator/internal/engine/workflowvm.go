// engine/workflowvm.go
// last modification: Feb 05 2026
package engine

import (
	"fmt"
	"bufio"
	"configurator/internal/config"
	"configurator/internal/ui"
	"configurator/internal/utils"
	"configurator/internal/model"
)
/* --------------------
	Workflow "New VM"
-------------------- */
func RunNewVMWorkflow(
	r *bufio.Reader,
	osList []config.VMConfig,
	defs struct {
		DiskPath string
		DiskSize int
	},
	variantByName map[string]string,
	isoWorkDir string,
	fp *config.FilePaths, //load xml_dir
) error {

	// choosing distribution
	distro, err := ui.PromptSelectDistro(r, osList)
	if err != nil {
		return fmt.Errorf("\x1b[31mOS selection failed: %w\x1b[0m", err)
	}
	// validating
	variant, ok := variantByName[distro.Name]
	if !ok {
		return fmt.Errorf("no varriant found for distro %q", distro.Name)
	}

	// Disk‑Path‑Default from selected distro
	defaultDiskPath := distro.DiskPath
  if defaultDiskPath == "" {
    defaultDiskPath = defs.DiskPath
  }

	// create basic config from default values
	cfg := model.DomainConfig{
		Name:       distro.Name,
		MemMiB:     distro.RAM,
		VCPU:       distro.CPU,
		DiskSize:   model.EffectiveDiskSize(distro, defs),
		ISOPath:		distro.ISOPath,
		Network: 		distro.Network,
		NestedVirt: distro.NestedVirt,
		Graphics: 	distro.Graphics,
		Sound:			distro.Sound,
		FileSystem: distro.FileSystem,
		BootOrder: 	distro.BootOrder,
	}

	// Optional Edit Menu for last edits
	ui.PromptEditDomainConfig(r, &cfg, defaultDiskPath, isoWorkDir)

	// Summary
	ui.ShowSummary(r, &cfg, cfg.ISOPath)

	// Create VM
	if err := CreateVM(cfg, variant, cfg.ISOPath, fp); err != nil {
		//return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
		utils.RedError("VM creation failed", cfg.Name, err)
		// WIP
		//return ui.Fatal(ui.ErrVMCreationFail, "%w")
	}	else {
		utils.Success("VM", cfg.Name, "successfully built!")
	}
	return nil
}