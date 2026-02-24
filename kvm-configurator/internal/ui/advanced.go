// ui/advanced.go
// last modified: Feb 24 2026
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"configurator/internal/model"
	"configurator/internal/style"
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

// editNested promts for nested virtualsiation driver
func editNested(r *bufio.Reader, cfg *model.DomainConfig) {
	raw, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Nested-Virtualisation (vmx for Intel, smx for AMD): "))
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Err("Failed to read virtualisation input:"), err)
		return
	}

	if raw == "" {
		return
	}

	v := strings.TrimSpace(strings.ToLower(raw))
	allowed := map[string]bool{"vmx": true, "smx": true}

	if !allowed[v] {
		fmt.Fprintln(os.Stderr, style.Err("Invalid virtualisation – allowed values: vmx, smx (you entered: %s)\n"), v)
		return
	}

	cfg.NestedVirt = v
	style.Success("Nested-Virtualisation", v, "")
}

// editBoot promts for the vm boot order
func editBoot(r *bufio.Reader, cfg *model.DomainConfig) {
	const bootKey = "boot"
	raw, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Boot order (comma-separated, e.g. hd,cdrom,network): "))
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Err("Failed to read boot order input:"), err)
		return
	}

	if raw == "" {
		return
	}

	v := strings.TrimSpace(strings.ToLower(raw))
	allowed := map[string]bool{"hd": true, "cdrom": true, "network": true}
	if !allowed[v] {
		fmt.Fprintln(os.Stderr, style.Err("Invalid boot order – allowed values: hd, cdrom, network (you entered: %s)\n"), v)
		return
	}

	cfg.BootOrder = v
	style.Success("Boot order", v, "")
}

// editGraphics prompts for the graphics driver
func editGraphics(r *bufio.Reader, cfg *model.DomainConfig) {
	raw, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Graphics (spice (default) or vnc): "))
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Err("Failed to read graphics input:"), err)
		return
	}

	if raw == "" {
		return
	}

	v := strings.TrimSpace(strings.ToLower(raw))
	allowed := map[string]bool{"spice": true, "vnc": true}

	if !allowed[v] {
		fmt.Fprintln(os.Stderr, style.Err("Invalid graphics driver – allowed values: spice, vnc (you entered: %s)\n"), v)
		return
	}

	cfg.Graphics = v
	style.Success("Graphics", v, "")
}

// editSound promts for sound driver
func editSound(r *bufio.Reader, cfg *model.DomainConfig) {
	raw, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Sound (none, ac97, ich6, ich9 (default)): "))
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Err("Failed to read sound input:"), err)
		return
	}

	if raw == "" {
		return
	}

	v := strings.TrimSpace(strings.ToLower(raw))
	allowed := map[string]bool{"none": true, "ac97": true, "ich6": true, "ich9": true}

	if !allowed[v] {
		fmt.Fprintln(os.Stderr, style.Err("Invalid sound driver – allowed values: none, ac97, ich6, ich9 (you entered: %s)\n"), v)
		return
	}

	cfg.Sound = v
	style.Success("Sound", v, "")
}

// editFilesystem promts for mounting fielsystems
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
