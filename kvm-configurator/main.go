// Version 1.0.6
// Autor: 	MrToadie
// GitHub: 	https://github.com/mrtoadie/
// Repo: 		https://github.com/mrtoadie/kvm-configurator
// License: MIT
// last modification: Feb 04 2026
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	// internal
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/model"
	"configurator/internal/ui"
	"configurator/kvmtools"
	//"configurator/internal/logger"
)

/* --------------------
	MAIN
-------------------- */
func main() {
	// [Modul: prereqs] validates if (virt‑install, virsh) is installed
	if err := config.EnsureAll(config.CmdVirtInstall, config.CmdVirsh); err != nil {
			ui.RedError("virt-install not found", "verify $PATH", err)
			os.Exit(1)
	}
	// for debug only
	//ui.Success("✅ Prereqs OK", "virt-install & virsh FOUND!", "")

	// [Modul: prereqs] check if config file exists or invalid
	ok, err := config.Exists()
  if err != nil {
    ui.RedError("Configuration file invalid or corrupt", "", err)
		//os.Exit(1)
  }
  if ok {
		// program starts		
  } else {
		ui.RedError("File does not exist", "verify $PATH", err)
  }
	
	// [Modul: config] loads File‑Config (isopath)
	fp, err := config.LoadFilePaths(config.FileConfig)
	if errors.Is(err, os.ErrNotExist) {
    ui.RedError("Configuration file not found ", ">", err)				
    //os.Exit(1)
	}

	workDir, err := config.ResolveWorkDir(fp)
	if errors.Is(err, os.ErrNotExist) {
		ui.RedError("Cannot resolve work directory", "verify $PATH", err)
    //os.Exit(1)
	}
	
	// [Modul: config] loading global Defaults
	osList, defaults, err := config.LoadOSList(config.FileConfig)
	if errors.Is(err, os.ErrNotExist) {
		ui.RedError("Configuration file not found", ">", err)
    //os.Exit(1)
	}

	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(ui.Colourise("\n=== MAIN MENU ===", ui.ColorBlue))
		fmt.Println("[1] New VM")
		fmt.Println("[2] KVM-Tools")
		fmt.Println("[0] Exit")
		fmt.Print(ui.Colourise("Selection: ", ui.ColorYellow))

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Print(ui.Colourise("\nPlease enter a valid number.", ui.ColorRed))
			ui.SimpleError("BLA","",err, ui.ColorBlue)
			continue
		}
		switch sel {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			if err := runNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				fp,
			); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ui.ColorRed, err, ui.ColorReset)
			}
		case 2:
			kvmtools.Start(r)
		default:
			fmt.Println(ui.Colourise("\nInvalid selection!", ui.ColorRed))
		}
	}
}

/* --------------------
	Workflow "New VM"
-------------------- */
func runNewVMWorkflow(
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
	if err := engine.CreateVM(cfg, variant, cfg.ISOPath, fp); err != nil {
		//return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
		ui.RedError("VM creation failed", cfg.Name, err)
		// WIP
		//return ui.Fatal(ui.ErrVMCreationFail, "%w")
	}	else {
		ui.Success("VM", cfg.Name, "successfully built!")
	}
	return nil
}
// EOF
