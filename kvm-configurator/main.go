// Version 1.0.6
// Autor: 	MrToadie
// GitHub: 	https://github.com/mrtoadie/
// Repo: 		https://github.com/mrtoadie/kvm-configurator
// Lisence: MIT
// last modification: Feb 03 2026
package main

import (
	"bufio"
	"fmt"
	"os"
	"errors"
	// internal
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/model"
	"configurator/internal/ui"
	"configurator/kvmtools"
)

/* --------------------
	MAIN
-------------------- */
func main() {
	// [Modul: prereqs] validates if (virt‑install, virsh) is installed
	if err := config.EnsureAll("virt-install", "virsh"); err != nil {
			ui.RedError("virt-install not found", "verfiy $PATH", err)
			os.Exit(1)
	}
	// for debug only
	//ui.Success("✅ Prereqs OK", "virt-install & virsh FOUND!", "")

	// [Modul: prereqs] check if config file exists or invalid
	ok, err := config.Exists()
  if err != nil {
    ui.RedError("Configuration file invalid or corrupt", "", err)
		os.Exit(1)
  }
  if ok {
		// program starts		
  } else {
		ui.RedError("File does not exist", "verfiy $PATH", err)
  }
	
	// [Modul: config] loads File‑Config (input_dir)
	fp, err := config.LoadFilePaths("oslist.yaml")
	if errors.Is(err, os.ErrNotExist) {
    ui.RedError("Configuration file not found", ">", err)
    os.Exit(1)
	}

	workDir, err := config.ResolveWorkDir(fp)
	if errors.Is(err, os.ErrNotExist) {
		ui.RedError("Cannot resolve work directory", "verify $PATH", err)
    os.Exit(1)
	}
	
	// [Modul: config] loading global Defaults
	osList, defaults, err := config.LoadOSList("oslist.yaml")
	if errors.Is(err, os.ErrNotExist) {
		ui.RedError("Configuration file not found", ">", err)
    os.Exit(1)
	}

	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(ui.Colourise("\n=== MAIN MENU ===", ui.Blue))
		fmt.Println("[1] New VM")
		fmt.Println("[2] KVM-Tools")
		fmt.Println("[0] Exit")
		fmt.Print(ui.Colourise("Selection: ", ui.Yellow))

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Print(ui.Colourise("\nPlease enter a valid number.", ui.Red))
			ui.SimpleError("BLA","",err, ui.Blue)
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
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ui.Red, err, ui.Reset)
			}
		case 2:
			kvmtools.Start(r)
		default:
			fmt.Println(ui.Colourise("\nInvalid selection!", ui.Red))
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
	variant := variantByName[distro.Name]

	// Disk‑Path‑Default from selectet distro
	defaultDiskPath := distro.DiskPath
  if defaultDiskPath == "" {
    defaultDiskPath = defs.DiskPath
  }

	// create basic config from default vaules
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
		ui.Success("VM", cfg.Name, "successfully build!")
	}
	return nil
}
// EOF
