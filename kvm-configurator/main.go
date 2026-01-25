// Version 1.0.6
// Autor: 	MrToadie
// GitHub: 	https://github.com/mrtoadie/
// Repo: 		https://github.com/mrtoadie/kvm-configurator
// Lisence: MIT
// last modification: January 25 2026
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	// internal
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/model"
	"configurator/internal/prereq"
	"configurator/internal/ui"
	"configurator/kvmtools"
)

func main() {
	// [Modul: prereqs] validates if (virt‑install, virsh) is installed
	if err := prereq.EnsureAll("virt-install", "virsh"); err != nil {
		prereq.FatalIfMissing(err)
	}

	// [Modul: prereqs] check if config file exists
	ok, err := prereq.Exists()
  if err != nil {
    log.Fatalf("Fehler beim Prüfen: %v", err)
  }
  if ok {
		// program starts		
  } else {
    fmt.Println("Datei existiert nicht")
  }
	
	// [Modul: config] loads File‑Config (input_dir)
	fp, err := config.LoadFilePaths("oslist.yaml")
	if err != nil {
		log.Fatalf("\x1b[31mError loading file-config: %v\x1b[0m", err)
	}
	workDir, err := config.ResolveWorkDir(fp)
	if err != nil {
		log.Fatalf("\x1b[31mCannot resolve work directory: %v\x1b[0m", err)
	}

	// [Modul: config] loading global Defaults
	osList, defaults, err := config.LoadOSList("oslist.yaml")
	if err != nil {
		log.Fatalf("\x1b[31mError loading OS list: %v\x1b[0m", err)
	}
	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	// Mainmenu
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(ui.Colourise("\n=== MAIN MENU ===", ui.Blue))
		fmt.Println("[1] New VM")
		fmt.Println("[2] Check")
		//fmt.Println("[3] DiskImage Tools")
		fmt.Println("[3] KVM-Tools")
		fmt.Println("[0] Exit")
		fmt.Print(ui.Colourise("Selection: ", ui.Yellow))

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Print(ui.Colourise("\nPlease enter a valid number.", ui.Red))
			continue
		}
		switch sel {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			// workDir from File‑Config
			if err := runNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				fp,
			); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[31mError: %v\x1b[0m\n", err)
			}
		//case 2:
		//case 3:
			// diskimage tools
		case 3:
			kvmtools.Start(r)
		default:
			fmt.Println(ui.Colourise("\nInvalid selection!", ui.Red))
		}
	}
}

// ---------------------------------------------------------------------
// Workflow "New VM"
// ---------------------------------------------------------------------
func runNewVMWorkflow(
	r *bufio.Reader,
	osList []config.Distro,
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
		//Network:    "default",
		Network: 		distro.Network,
		NestedVirt: distro.NestedVirt,
		Graphics: 	distro.Graphics,
		Sound:			distro.Sound,
		FileSystem: distro.FileSystem,
		BootOrder: 	distro.BootOrder,
	}

	// Optional Edit Menu for last edits
	ui.PromptEditDomainConfig(r, &cfg, defaultDiskPath, isoWorkDir)

	/* OLD PROMT
		Select ISO (uses the directory from the YAML)
	isoPath, err := ui.PromptSelectISO(r, isoWorkDir)
	if err != nil {
		return fmt.Errorf("\x1b[31mISO selection failed: %w\x1b[0m", err)
	}*/

	// Summary
	ui.ShowSummary(r, &cfg, cfg.ISOPath)

	// Create VM
	if err := engine.CreateVM(cfg, variant, cfg.ISOPath, fp); err != nil {
		return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
	}
	return nil
}
// EOF
