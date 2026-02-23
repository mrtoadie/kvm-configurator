// ui/ui.go
// last modified: Feb 23 2026
package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	// internal
	"configurator/internal/config"
	"configurator/internal/model"
	"configurator/internal/style"
	"configurator/internal/utils"
)

// Selecting the operating system from the YAML list
func SelectDistro(r *bufio.Reader, list []config.VMConfig) (config.VMConfig, error) {
	fmt.Println(style.BoxCenter(51, []string{"Select an operating system"}))

	//Sort – case‑insensitive
	sorted := append([]config.VMConfig(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	// Create table
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "No.\tName\tCPU\tRAM (MiB)\tDisk (GB)")
		fmt.Fprintln(w, "---\t----\t---\t---------\t---------")
		for i, d := range sorted {
			fmt.Fprintf(w, "%2d\t%s\t%d\t%d\t%d\n",
				i+1, d.Name, d.CPU, d.RAM, d.DiskSize)
		}
	})
	fmt.Print(style.Box(51, lines))

	// Prompt – we use the central prompt helper from utils
	ans, err := utils.Prompt(r, os.Stdout,
		style.PromptMsg("\nPlease enter a number (or press ENTER for default Arch Linux): "))
	if err != nil {
		return config.VMConfig{}, err
	}
	// Default selection = first element (Arch)
	idx := 1
	if ans != "" {
		if i, e := strconv.Atoi(ans); e == nil && i >= 1 && i <= len(sorted) {
			idx = i
		} else {
			return config.VMConfig{}, fmt.Errorf(style.Err("Invalid selection"))
		}
	}
	return sorted[idx-1], nil
}

// -------------------------------------------------------------------
// ISO selection (files from a directory)
// -------------------------------------------------------------------
func SelectISO(r *bufio.Reader, workDir string) (string, error) {
	// Get all regular files
	files, err := utils.ListFiles(workDir)
	if err != nil {
		return "", fmt.Errorf("listing files in %s failed: %w", workDir, err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files found in %s", workDir)
	}
	// Sort alphabetically
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})

	// The user can decide – we use what is already there
	// PromptSelection‑Helper that outputs a numbered list.
	choice, err := utils.PromptSelection(bufio.NewReader(os.Stdin), os.Stdout, files)
	if err != nil {
		return "", err
	}
	if choice == utils.CancelChoice {
		return "", fmt.Errorf(style.PromptMsg("selection aborted"))
	}
	selected := files[choice-1]

	// Return as an absolute path so that virt‑install can reliably find it
	abs, _ := filepath.Abs(selected)
	return abs, nil
}

// summary
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	isoFile := filepath.Base(cfg.ISOPath)

	fmt.Println(style.BoxCenter(51, []string{"VM-SUMMARY"}))
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
		fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
		fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)

		if primary := cfg.PrimaryDisk(); primary != nil {
			fmt.Fprintf(w, "Disk-Path:\t%s\n", primary.Path)
			fmt.Fprintf(w, "Disk-Size (GB):\t%d\n", primary.SizeGiB)
		} else {
			fmt.Fprintf(w, "Disk-Path:\t<none>\n")
			fmt.Fprintf(w, "Disk-Size (GB):\t<none>\n")
		}

		fmt.Fprintf(w, "Network:\t%s\n", cfg.Network)
		fmt.Fprintf(w, "Nested-Virtualisation:\t%s\n", cfg.NestedVirt)
		fmt.Fprintf(w, "ISO-File:\t%s\n", isoFile)
		fmt.Fprintf(w, "Boot-Order:\t%s\n", cfg.BootOrder)
		fmt.Fprintf(w, "Graphic:\t%s\n", cfg.Graphics)
		fmt.Fprintf(w, "Sound:\t%s\n", cfg.Sound)
		fmt.Fprintf(w, "Filesystem:\t%s\n", cfg.FileSystem)
	})
	fmt.Print(style.Box(51, lines))

	_, _ = utils.Prompt(r, os.Stdout,
		style.PromptMsg("\nPress ENTER to create VM … "))
}

// -------------------------------------------------------------------
//Note: The actual edit loop is now in
//`ui/editor.go` (type `Editor`).  If you want to start it from outside:
//
//ed := ui.NewEditor(r, os.Stdout, cfg, defaultDiskPath, isoWorkDir)
//ed.Run()
//
// -------------------------------------------------------------------
