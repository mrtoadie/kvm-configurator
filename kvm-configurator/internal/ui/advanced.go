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
	"configurator/internal/style"
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
		style.Hint(">> Nested-Virtualisation (vmx for Intel, smx for AMD): ")); err == nil && v != "" {
		cfg.NestedVirt = v
		style.Success("Nested-Virtualisation", v, "")
	}
}

func editBoot(r *bufio.Reader, cfg *model.DomainConfig) {
	const bootKey = "boot"
	if v, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Boot order (comma-separated, e.g. hd,cdrom,network): ")); err == nil && v != "" {

		allowed := map[string]bool{"hd": true, "cdrom": true, "network": true}
		if !validateList(v, allowed) {
			style.RedError("Invalid boot order", v, nil)
			return
		}
		cfg.BootOrder = v
		style.Success("Boot order", v, "")
	}
}

func editGraphics(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Graphics (spice (default) or vnc): ")); err == nil && v != "" {
		cfg.Graphics = v
		style.Success("Graphics", v, "")
	}
}

func editSound(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Sound (none, ac97, ich6, ich9 (default)): ")); err == nil && v != "" {
		cfg.Sound = v
		style.Success("Sound", v, "")
	}
}

func editFilesystem(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Filesystem / Mount (/src/dir,/guest/dir): ")); err == nil && v != "" {
		cfg.FileSystem = v
		style.Success("Filesystem", v, "")
	}
}

// print menu
func printAdvancedMenu(cfg *model.DomainConfig) {
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Parameter\tCurrent")
		fmt.Fprintln(w, "---------\t-------")
		fmt.Fprintf(w, "[a]\tNested-Virtualisation\t%s\n", cfg.NestedVirt)
		fmt.Fprintf(w, "[b]\tBoot-Order\t%s\n", cfg.BootOrder)
		fmt.Fprintf(w, "[c]\tGraphics\t%s\n", cfg.Graphics)
		fmt.Fprintf(w, "[d]\tSound\t%s\n", cfg.Sound)
		fmt.Fprintf(w, "[e]\tFilesystem\t%s\n", cfg.FileSystem)
		fmt.Fprintln(w, "[0]\tBack to main menu")
	})
	fmt.Println(style.Box(60, lines))
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
		fmt.Println(style.BoxCenter(51, []string{"=== ADVANCED PARAMETERS ==="}))
		printAdvancedMenu(cfg)

		choice, err := utils.Prompt(r, os.Stdout,
			style.PromptMsg("\nSelect an option (or press Enter to go back): "))
		if err != nil {
			if err == io.EOF {
				return // user hit Ctrl‑D → graceful exit
			}
			style.RedError("Read error", "", err)
			continue
		}
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "" || choice == optBack {
			return
		}
		if h, ok := handlers[choice]; ok {
			h()
		} else {
			style.RedError("Invalid selection", choice, nil)
		}
	}
}