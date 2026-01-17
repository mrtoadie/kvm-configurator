// Version 1.0
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
)

func main() {
	// [Modul: prereqs] validates if (virt‑install, virsh) is installed
	if err := prereq.EnsureAll("virt-install", "virsh"); err != nil {
		prereq.FatalIfMissing(err)
	}

	// [Modul: config] loads File‑Config (input_dir, max_lines)
	fp, err := config.LoadFilePaths("oslist.yaml")
	if err != nil {
		log.Fatalf("\x1b[31mError loading file-config: %v\x1b[0m", err)
	}
	workDir, err := config.ResolveWorkDir(fp)
	if err != nil {
		log.Fatalf("\x1b[31mCannot resolve work directory: %v\x1b[0m", err)
	}

	// [Modul: config] loading globale Defaults
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
		fmt.Println("[0] Exit")
		fmt.Print("Selection: ")

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Println(ui.Colourise("Please enter a valid number.", ui.Red))
			continue
		}
		switch sel {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			// workDir & maxLines from File‑Config
			if err := runNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				fp.Filepaths.MaxLines,
				fp, // complete config (idk if good or not)
			); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[31mError: %v\x1b[0m\n", err)
			}
		//case 2:
		default:
			fmt.Println(ui.Colourise("Invalid selection!", ui.Red))
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
	maxLines int,
	fp *config.FilePaths, //load xml_dir
) error {

	// choosing distribution
	distro, err := ui.PromptSelectDistro(r, osList)
	if err != nil {
		return fmt.Errorf("\x1b[31mOS selection failed: %w\x1b[0m", err)
	}
	variant := variantByName[distro.Name]

// Disk‑Path‑Default aus der gewählten Distro holen
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
		//Disk:       model.EffectiveDiskPath(distro, defs),
		Network:    "default",
		NestedVirt: distro.NestedVirt,
	}

	// Optional Edit Menu for last edits
	ui.PromptEditDomainConfig(r, &cfg, defaultDiskPath)

	// Select ISO (uses the directory from the YAML)
	isoPath, err := ui.PromptSelectISO(r, isoWorkDir, maxLines)
	if err != nil {
		return fmt.Errorf("\x1b[31mISO selection failed: %w\x1b[0m", err)
	}

	// Summary
	ui.ShowSummary(r, &cfg, isoPath)

	// Create VM
	if err := engine.CreateVM(cfg, variant, isoPath, fp); err != nil {
		return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
	}
	return nil
}
// EOF
