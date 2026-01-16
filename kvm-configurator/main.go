// Version 1.0
/*
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
		fmt.Println("\x1b[34m\n=== MAIN MENU ===\x1b[0m")
		fmt.Println("[1] New VM")
		fmt.Println("[2] Check")
		fmt.Println("[0] Exit")
		fmt.Print("Selection: ")

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Println("\x1b[31mPlease enter a valid number.\x1b[0m")
			continue
		}
		switch sel {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			// workDir & maxLines from File‑Config
			if err := runNewVMWorkflow(r, osList, defaults, variantByName, workDir, fp.Filepaths.MaxLines); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[31mError: %v\x1b[0m\n", err)
			}
		//case 2:

		default:
			fmt.Println("\x1b[31mInvalid selection!\x1b[0m")
		}
	}
}

// ----------------------------------------------------------------
// Workflow "New VM“
// ----------------------------------------------------------------
func runNewVMWorkflow(
	r *bufio.Reader,
	osList []config.Distro,
	defs struct{ DiskPath string; DiskSize int },
	variantByName map[string]string,
	isoWorkDir string,
	maxLines int,
) error {

	// choosing distribution
	distro, err := ui.PromptSelectDistro(r, osList)
	if err != nil {
		return fmt.Errorf("\x1b[31mOS selection failed: %w\x1b[0m", err)
	}
	variant := variantByName[distro.Name]

	// create basic config from default vaules
	cfg := model.DomainConfig{
		Name:       distro.Name,
		MemMiB:     distro.RAM,
		VCPU:       distro.CPU,
		DiskSize:   model.EffectiveDiskSize(distro, defs),
		Disk:       model.EffectiveDiskPath(distro, defs),
		Network:    "default",
		NestedVirt: distro.NestedVirt,
	}

	// Optional Edit Menu
	ui.PromptEditDomainConfig(r, &cfg)

	// Select ISO (uses the directory from the YAML)
	isoPath, err := ui.PromptSelectISO(r, isoWorkDir, maxLines)
	if err != nil {
		return fmt.Errorf("\x1b[31mISO selection failed: %w\x1b[0m", err)
	}

	// Summary
    ui.ShowSummary(r, &cfg, isoPath)

	// Create VM
	if err := engine.CreateVM(cfg, variant, isoPath); err != nil {
		return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
	}
	return nil
}
*/

// Version 1.0
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	// interne Pakete
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/model"
	"configurator/internal/prereq"
	"configurator/internal/ui"
)

func main() {
	// ── Prüfen, ob die nötigen Tools da sind ────────────────────────────────
	if err := prereq.EnsureAll("virt-install", "virsh"); err != nil {
		prereq.FatalIfMissing(err)
	}

	// ── Laden der Dateikonfiguration (input_dir, max_lines, xml_dir) ────────
	fp, err := config.LoadFilePaths("oslist.yaml")
	if err != nil {
		log.Fatalf("\x1b[31mError loading file-config: %v\x1b[0m", err)
	}
	workDir, err := config.ResolveWorkDir(fp)
	if err != nil {
		log.Fatalf("\x1b[31mCannot resolve work directory: %v\x1b[0m", err)
	}

	// ── Laden der OS‑Liste und globaler Defaults ─────────────────────────────
	osList, defaults, err := config.LoadOSList("oslist.yaml")
	if err != nil {
		log.Fatalf("\x1b[31mError loading OS list: %v\x1b[0m", err)
	}
	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	// ── Hauptmenü ─────────────────────────────────────────────────────────────
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\x1b[34m\n=== MAIN MENU ===\x1b[0m")
		fmt.Println("[1] New VM")
		fmt.Println("[2] Check")
		fmt.Println("[0] Exit")
		fmt.Print("Selection: ")

		var sel int
		if _, err := fmt.Scanln(&sel); err != nil {
			fmt.Println("\x1b[31mPlease enter a valid number.\x1b[0m")
			continue
		}
		switch sel {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			// workDir & maxLines kommen aus der File‑Config
			if err := runNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				fp.Filepaths.MaxLines,
				fp, // ← jetzt auch die komplette Config übergeben
			); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[31mError: %v\x1b[0m\n", err)
			}
		//case 2:
		default:
			fmt.Println("\x1b[31mInvalid selection!\x1b[0m")
		}
	}
}

// ---------------------------------------------------------------------
// Workflow "New VM"
// ---------------------------------------------------------------------
func runNewVMWorkflow(
	r *bufio.Reader,
	osList []config.Distro,
	defs struct{ DiskPath string; DiskSize int },
	variantByName map[string]string,
	isoWorkDir string,
	maxLines int,
	fp *config.FilePaths, // ← neuer Parameter, damit wir den xml_dir weiterreichen können
) error {

	// 1️⃣ Distribution auswählen
	distro, err := ui.PromptSelectDistro(r, osList)
	if err != nil {
		return fmt.Errorf("\x1b[31mOS selection failed: %w\x1b[0m", err)
	}
	variant := variantByName[distro.Name]

	// 2️⃣ Basis‑Config aus den Defaults bauen
	cfg := model.DomainConfig{
		Name:       distro.Name,
		MemMiB:     distro.RAM,
		VCPU:       distro.CPU,
		DiskSize:   model.EffectiveDiskSize(distro, defs),
		Disk:       model.EffectiveDiskPath(distro, defs),
		Network:    "default",
		NestedVirt: distro.NestedVirt,
	}

	// 3️⃣ Optionales Edit‑Menu (falls du noch was ändern willst)
	ui.PromptEditDomainConfig(r, &cfg)

	// 4️⃣ ISO auswählen (Verzeichnis kommt aus der YAML)
	isoPath, err := ui.PromptSelectISO(r, isoWorkDir, maxLines)
	if err != nil {
		return fmt.Errorf("\x1b[31mISO selection failed: %w\x1b[0m", err)
	}

	// 5️⃣ Kurze Zusammenfassung zeigen
	ui.ShowSummary(r, &cfg, isoPath)

	// 6️⃣ VM erzeugen – **jetzt mit dem Config‑Pointer**
	if err := engine.CreateVM(cfg, variant, isoPath, fp); err != nil {
		return fmt.Errorf("\x1b[31mVM creation failed: %w\x1b[0m", err)
	}
	return nil
}