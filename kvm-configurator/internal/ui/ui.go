// ui/ui.go
package ui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"path/filepath" 
	"strconv"
	"strings"
	"text/tabwriter"
	// internal
	"configurator/internal/fileutils"
	"configurator/internal/config"
	"configurator/internal/model"
)

/* --------------------
	waiting for input
-------------------- */
func readLine(r *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	s, err := r.ReadString('\n')
	return strings.TrimSpace(s), err
}

/* --------------------
	Loading distro list from yaml
-------------------- */
func PromptSelectDistro(r *bufio.Reader, list []config.Distro) (config.Distro, error) {
	fmt.Println("\n\x1b[34m=== Select an operating system ===\x1b[0m")
	sorted := append([]config.Distro(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "No.\tName\tCPU\tRAM (MiB)\tDisk (GB)")
	for i, d := range sorted {
		fmt.Fprintf(w, "%2d\t%s\t%d\t%d\t%d\n",
			i+1, d.Name, d.CPU, d.RAM, d.DiskSize)
	}
	w.Flush()

	line, err := readLine(r, "\nPlease enter a number (or press ENTER for default Arch Linux): ")
	if err != nil {
		return config.Distro{}, err
	}
	idx := 1 // default = Arch
	if line != "" {
		if i, e := strconv.Atoi(line); e == nil && i >= 1 && i <= len(sorted) {
			idx = i
		} else {
			return config.Distro{}, fmt.Errorf("\x1b[31mInvalid selection\x1b[0m")
		}
	}
	return sorted[idx-1], nil
}

/* --------------------
	[Modul: internal/ui/ui.go] Select ISO
	PromptSelectISO – selects an ISO file from the specified directory
	The return value is the 'absolute path' to the file (for virt‑install)
-------------------- */
func PromptSelectISO(r *bufio.Reader, workDir string, maxLines int) (string, error) {
	// workDir is directory from filepaths.input_dir
	files, err := fileutils.ListFiles(workDir)
	if err != nil {
		return "", fmt.Errorf("listing files in %s failed: %w", workDir, err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files found in %s", workDir)
	}
	// sort iso list by name
	sort.Slice(files, func(i, j int) bool {
  	return strings.ToLower(files[i]) < strings.ToLower(files[j])
  })

	// Show menu for selection
	choice, err := fileutils.PromptSelection(files)
	if err != nil {
		return "", err
	}
	if choice == 0 {
		return "", fmt.Errorf("selection aborted")
	}
	selected := files[choice-1]

	// Return the 'absolute path' so that virt‑install can find it reliably
	abs, _ := filepath.Abs(selected)
	return abs, nil
}

/* --------------------
	Form – allows changes to the fields
-------------------- */
func PromptEditDomainConfig(r *bufio.Reader, cfg *model.DomainConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for {
		fmt.Fprintln(w, "\n\x1b[34m=== VM-Config ===\t\x1b[0m")
		fmt.Fprintf(w, "[1] Name:\t%s\t[default]\n", cfg.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t[default]\n", cfg.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\t[default]\n", cfg.VCPU)
		fmt.Fprintf(w, "[4] Disk-Path:\t%s\t[Enter path for virtual hdd]\n", cfg.Disk)
		fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\t[default]\n", cfg.DiskSize)
		fmt.Fprintf(w, "[6] Network:\t%s\t[default]\n", cfg.Network)
		fmt.Fprintf(w, "[7] Advanced Parameters [optional]")
		w.Flush()

		f, _ := readLine(r, "\nSelect or press Enter to continue: ")
		if f == "" {
			break
		}
		switch f {
		case "1":
			if v, _ := readLine(r, ">> New Name: "); v != "" {
				cfg.Name = v
			}
		case "2":
			if v, _ := readLine(r, ">> RAM (MiB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.MemMiB = i
				}
			}
		case "3":
			if v, _ := readLine(r, ">> vCPU: "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.VCPU = i
				}
			}
		case "4":
			if v, _ := readLine(r, ">> Disk path (empty = no disk): "); true {
				cfg.Disk = v
			}
		case "5":
			if v, _ := readLine(r, ">> Disksize (GB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.DiskSize = i
				}
			}
		case "6":
			if v, _ := readLine(r, ">> Network (comma-separated): "); true {
				cfg.Network = v
			}
		case "7":
			editAdvanced(r, cfg)
		default:
			fmt.Println("\x1b[31mInvalid input!\x1b[0m")
		}
	}
}

/* --------------------
	Form – Advanced Parameters
-------------------- */
func editAdvanced(r *bufio.Reader, cfg *model.DomainConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	for {
		fmt.Fprintln(w, "\n\x1b[34m=== Advanced Parameters ===\t\x1b[0m")
		fmt.Fprintln(w, "[a] Nested-Virtualisation\t[default]")
		fmt.Fprintln(w, "[b] Boot-Order\t[default] 'not implemented yet'")
		fmt.Fprintln(w, "[0] Back to main menu")
		w.Flush()

		choice, _ := readLine(r, "\nSelect an option (or press Enter to go back): ")
		if choice == "" || strings.EqualFold(choice, "0") {
			return
		}
		switch strings.ToLower(choice) {
		case "a":
			if v, _ := readLine(r, ">> Nested-Virtualisation (vmx for Intel or smx for AMD): "); v != "" {
				cfg.NestedVirt = v
				fmt.Println("Nested-Virtualisation is set to\x1b[32m", v)
			}
		case "b": // bug - disk is no parameter
			prompt := ">> Boot-Order (comma-separated, e.g. cdrom,disk,network): "
      if v, _ := readLine(r, prompt); v != "" {
        // clean spaces and all lower case
        cleaned := strings.ReplaceAll(strings.ToLower(v), " ", "")
        cfg.BootOrder = cleaned
        fmt.Printf("Boot-Order set to \x1b[32m%s\x1b[0m\n", cleaned)
      } else {
        // no input = use default
        cfg.BootOrder = "cdrom,disk,network"
        fmt.Println("Boot-Order left unchanged – using default.")
      }
		default:
			fmt.Println("\x1b[31mInvalid input!\x1b[0m")
		}
	}
}

/* --------------------
	Summary table
-------------------- */
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprintln(w, "\n\x1b[34m=== VM-SUMMARY ===\x1b[0m")
	fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
	fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
	fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
	fmt.Fprintf(w, "Disk-Path:\t%s\n", cfg.Disk)
	fmt.Fprintf(w, "Disk-Size (GB):\t%d\n", cfg.DiskSize)
	fmt.Fprintf(w, "Network:\t%s\n", cfg.Network)
	fmt.Fprintf(w, "Nested-Virtualisation:\t%s\n", cfg.NestedVirt)
	fmt.Fprintf(w, "ISO-File:\t%s\n", isoPath)
	fmt.Fprintf(w, "Boot-Order:\t%s\n", cfg.BootOrder)
	w.Flush()

	fmt.Print("\nPress ENTER to create VM … ")
	_, _ = r.ReadString('\n')
}
// EOF