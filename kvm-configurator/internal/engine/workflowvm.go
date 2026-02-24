// engine/workflowvm.go
// last modified: Feb 22 2026
package engine

import (
	"bufio"
	"configurator/internal/config"
	"configurator/internal/model"
	"configurator/internal/ui"

	"configurator/internal/style"
	"fmt"
	"os"
)

// Workflow "New VM"
func RunNewVMWorkflow(
	r *bufio.Reader,
	osList []config.VMConfig,
	defs struct {
		DiskPath string
		DiskSize int
	},
	variantByName map[string]string,
	isoWorkDir string, // directory in which the ISOs are located
	isoPath string, // Path to ISO directory (can be empty → cwd fallback)
	xmlDir string, // Destination directory for the libvirt XML file
) error {
	// choosing distribution
	distro, err := ui.SelectDistro(r, osList)
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
		Name:   distro.Name,
		MemMiB: distro.RAM,
		VCPU:   distro.CPU,
		/*
			We create the *system disk*(first element).
			The path can be a directory (later becomes <vm>-system.qcow2)
			or a complete file name
		*/
		Disks: []model.DiskSpec{
			{
				Name:    "system",                              // name of the first disk
				Path:    defaultDiskPath,                       // can be empty → will be filled later
				SizeGiB: model.EffectiveDiskSize(distro, defs), // Size (if defined)
				Bus:     "virtio",                              // Default bus type (can be changed later)
			},
		},
		ISOPath:    distro.ISOPath,
		Network:    distro.Network,
		NestedVirt: distro.NestedVirt,
		Graphics:   distro.Graphics,
		Sound:      distro.Sound,
		FileSystem: distro.FileSystem,
		BootOrder:  distro.BootOrder,
	}

	// Optional Edit Menu for last edits
	//ui.PromptEditDomainConfig(r, &cfg, defaultDiskPath, isoWorkDir)

	// ---------- NEW PART ----------
	// Instead of the old PromptEditDomainConfig function we now spin up the
	// editor object that bundles all the edit‑logic.
	editor := ui.NewEditor(r, os.Stdout, &cfg, defaultDiskPath, isoWorkDir)
	editor.Run()
	// --------------------------------

	// Summary
	ui.ShowSummary(r, &cfg, cfg.ISOPath)

	// Create VM
	if err := CreateVM(cfg, variant, cfg.ISOPath, xmlDir); err != nil {
		//return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
		style.RedError("VM creation failed", cfg.Name, err)
		// WIP
		//return ui.Fatal(ui.ErrVMCreationFail, "%w")
	} else {
		style.Success("VM", cfg.Name, "successfully built!")
	}
	return nil
}
