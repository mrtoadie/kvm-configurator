// ui/advanced.go
// last modified: Feb 22 2026
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"configurator/internal/model"
	"configurator/internal/utils"
)

// Helper: validate comma‑separated list against an allow‑list.
func validateList(input string, allowed map[string]bool) bool {
	for _, part := range strings.Split(input, ",") {
		if !allowed[strings.TrimSpace(part)] {
			return false
		}
	}
	return true
}

// Individual field editors (tiny, testable functions).
func editNested(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		utils.Colourise(">> Nested-Virtualisation (vmx for Intel, smx for AMD): ", utils.ColorBlue)); err == nil && v != "" {
		cfg.NestedVirt = v
		utils.Success("Nested-Virtualisation", v, "")
	}
}

func editBoot(r *bufio.Reader, cfg *model.DomainConfig) {
	const bootKey = "boot"
	if v, err := utils.Prompt(r, os.Stdout,
		utils.Colourise(">> Boot order (comma-separated, e.g. hd,cdrom,network): ", utils.ColorBlue)); err == nil && v != "" {

		allowed := map[string]bool{"hd": true, "cdrom": true, "network": true}
		if !validateList(v, allowed) {
			utils.RedError("Invalid boot order", v, nil)
			return
		}
		cfg.BootOrder = v
		utils.Success("Boot order", v, "")
	}
}

func editGraphics(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		utils.Colourise(">> Graphics (spice (default) or vnc): ", utils.ColorBlue)); err == nil && v != "" {
		cfg.Graphics = v
		utils.Success("Graphics", v, "")
	}
}

func editSound(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		utils.Colourise(">> Sound (none, ac97, ich6, ich9 (default)): ", utils.ColorBlue)); err == nil && v != "" {
		cfg.Sound = v
		utils.Success("Sound", v, "")
	}
}

func editFilesystem(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		utils.Colourise(">> Filesystem / Mount (/src/dir,/guest/dir): ", utils.ColorBlue)); err == nil && v != "" {
		cfg.FileSystem = v
		utils.Success("Filesystem", v, "")
	}
}

// print menu
func printAdvancedMenu(cfg *model.DomainConfig) {
	lines := utils.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Parameter\tCurrent")
		fmt.Fprintln(w, "---------\t-------")
		fmt.Fprintf(w, "[a]\tNested-Virtualisation\t%s\n", cfg.NestedVirt)
		fmt.Fprintf(w, "[b]\tBoot-Order\t%s\n", cfg.BootOrder)
		fmt.Fprintf(w, "[c]\tGraphics\t%s\n", cfg.Graphics)
		fmt.Fprintf(w, "[d]\tSound\t%s\n", cfg.Sound)
		fmt.Fprintf(w, "[e]\tFilesystem\t%s\n", cfg.FileSystem)
		fmt.Fprintln(w, "[0]\tBack to main menu")
	})
	fmt.Println(utils.Box(60, lines))
}

// Main dispatcher
func editAdvanced(r *bufio.Reader, cfg *model.DomainConfig) {
	const (
		optNested = "a"
		optBoot   = "b"
		optGraf   = "c"
		optSound  = "d"
		optFS     = "e"
		optBack   = "0"
	)

	handlers := map[string]func(){
		optNested: func() { editNested(r, cfg) },
		optBoot:   func() { editBoot(r, cfg) },
		optGraf:   func() { editGraphics(r, cfg) },
		optSound:  func() { editSound(r, cfg) },
		optFS:     func() { editFilesystem(r, cfg) },
	}

	for {
		fmt.Println(utils.BoxCenter(51, []string{"=== ADVANCED PARAMETERS ==="}))
		printAdvancedMenu(cfg)

		choice, err := utils.Prompt(r, os.Stdout,
			utils.Colourise("\nSelect an option (or press Enter to go back): ", utils.ColorYellow))
		if err != nil {
			if err == io.EOF {
				return // user hit Ctrl‑D → graceful exit
			}
			utils.RedError("Read error", "", err)
			continue
		}
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "" || choice == optBack {
			return
		}
		if h, ok := handlers[choice]; ok {
			h()
		} else {
			utils.RedError("Invalid selection", choice, nil)
		}
	}
}