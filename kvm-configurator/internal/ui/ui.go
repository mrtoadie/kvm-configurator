// ui/ui.go
// last modification: January 18 2026
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
	fmt.Println(Colourise("\n=== Select an operating system ===", Blue))
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

	line, err := readLine(r, Colourise("\nPlease enter a number (or press ENTER for default Arch Linux): ", Yellow))
	if err != nil {
		return config.Distro{}, err
	}
	idx := 1 // default = Arch
	if line != "" {
		if i, e := strconv.Atoi(line); e == nil && i >= 1 && i <= len(sorted) {
			idx = i
		} else {
			return config.Distro{}, fmt.Errorf(Colourise("Invalid selection", Red))
		}
	}
	return sorted[idx-1], nil
}

/* --------------------
	PromptSelectISO – selects an ISO file from the specified directory
	The return value is the 'absolute path' to the file (for virt‑install)
-------------------- */
func PromptSelectISO(r *bufio.Reader, workDir string) (string, error) {
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
		return "", fmt.Errorf(Colourise("selection aborted", Yellow))
	}
	selected := files[choice-1]

	// Return the 'absolute path' so that virt‑install can find it reliably
	abs, _ := filepath.Abs(selected)
	return abs, nil
}

/* --------------------
	expand disk path
-------------------- */
func expandPath(p string) string {
    if strings.HasPrefix(p, "~"+string(os.PathSeparator)) {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, p[2:])
    }
    return p
}

/* --------------------
	Form – allows changes to the fields
-------------------- */
func PromptEditDomainConfig(r *bufio.Reader, cfg *model.DomainConfig, defaultDiskPath string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 10, 2, ' ', 0)
	for {
		fmt.Fprintln(w, Colourise("\n=== VM-Config ===\t", Blue))
		fmt.Fprintf(w, "[1] Name:\t%s\t[default]\n", cfg.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t\n", cfg.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\t[default]\n", cfg.VCPU)
		fmt.Fprintf(w, "[4] Disk-Path:\t%s\t[default = no disk path]\n", cfg.Disk)
		fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\t[default]\n", cfg.DiskSize)
		fmt.Fprintf(w, "[6] Network:\t%s\t[default]\n", cfg.Network)
		fmt.Fprintf(w, "[7] Advanced Parameters\n")
		w.Flush()

		f, _ := readLine(r, Colourise("\nSelect or press Enter to continue: ", Yellow))
		if f == "" {
			break
		}
		switch f {
		case "1": // name
			if v, _ := readLine(r, ">> New Name: "); v != "" {
				cfg.Name = v
			}
		case "2": // ram
			if v, _ := readLine(r, ">> RAM (MiB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.MemMiB = i
				}
			}
		case "3": // vcpu
			if v, _ := readLine(r, ">> vCPU: "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.VCPU = i
				}
			}
		case "4":	// disk path
    	prompt := fmt.Sprintf(">> Disk path (default: %s): ", defaultDiskPath)
      if v, _ := readLine(r, prompt); v != "" {
        // user input
				cfg.Disk = expandPath(v)
      } else {
        // no intpu = default path
        cfg.Disk = expandPath(defaultDiskPath)
      }
      fmt.Printf("\x1b[32mDisk will be stored at: %s\x1b[0m\n", cfg.Disk)
		case "5": // disksize
			if v, _ := readLine(r, ">> Disksize (GB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.DiskSize = i
				}
			}
		case "6":
			if v, _ := readLine(r, ">> Network (none or default): "); true {
				cfg.Network = v
			}
		case "7":
			editAdvanced(r, cfg)
		default:
			fmt.Println(Colourise("Invalid input!", Red))
		}
	}
}

/* --------------------
	Form – Advanced Parameters
-------------------- */
func editAdvanced(r *bufio.Reader, cfg *model.DomainConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	for {
		fmt.Fprintln(w, Colourise("\n=== Advanced Parameters ===\t", Blue))
		fmt.Fprintln(w, "Parameter\t Default\t Set")
		fmt.Fprintln(w, "[a] Nested-Virtualisation\t", cfg.NestedVirt)
		fmt.Fprintln(w, "[b] Boot-Order\t", cfg.BootOrder)
		fmt.Fprintln(w, "[c] Graphics\t", cfg.Graphics)
		fmt.Fprintln(w, "[d] Sound\t", cfg.Sound)
		fmt.Fprintln(w, "[e] Filesystem\t", cfg.FileSystem)
		fmt.Fprintln(w, "-------------------------------------")
		fmt.Fprintln(w, "[0] Back to main menu")
		w.Flush()

		choice, _ := readLine(r, Colourise("\nSelect an option (or press Enter to go back): ", Yellow))
		if choice == "" || strings.EqualFold(choice, "0") {
			return
		}
		switch strings.ToLower(choice) {
		case "a":
			if v, _ := readLine(r, Colourise(">> Nested-Virtualisation (vmx for Intel or smx for AMD): ", Blue)); v != "" {
				cfg.NestedVirt = v
				fmt.Println("Nested-Virtualisation is set to\x1b[32m", v)
			}
		case "b": // bug - nothing happend
			if v, _ := readLine(r, ">> Boot order: "); v != "" {
				cfg.BootOrder = v
				fmt.Println("Boot order is set to", v)
			}
		case "c":
			if v, _ := readLine(r, Colourise(">> Graphics (spice (default) or vnc): ", Blue)); v != "" {
				cfg.Graphics = v
				fmt.Println(Colourise("Graphics is set to", Blue), v)
			}
		case "d":
			if v, _ := readLine(r, Colourise(">> Sound (none, ac97, ich6 or ich9 (default)): ", Blue)); v != "" {
				cfg.Sound = v
				fmt.Println(Colourise("Sound is set to", Blue), v)
			}
		case "e":
			if v, _ := readLine(r, Colourise(">> Filesystem / Mount (/my/source/dir,/dir/in/guest): ", Blue)); v != "" {
				cfg.FileSystem = v
				fmt.Println(Colourise("Filesystem / Mount is set to", Blue), v)
			}
		default:
			fmt.Println(Colourise("Invalid input!", Red))
		}
	}
}

/* --------------------
	Summary table
-------------------- */
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprintln(w, Colourise("\n=== VM-SUMMARY ===", Blue))
	fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
	fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
	fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
	fmt.Fprintf(w, "Disk-Path:\t%s\n", cfg.Disk)
	fmt.Fprintf(w, "Disk-Size (GB):\t%d\n", cfg.DiskSize)
	fmt.Fprintf(w, "Network:\t%s\n", cfg.Network)
	fmt.Fprintf(w, "Nested-Virtualisation:\t%s\n", cfg.NestedVirt)
	fmt.Fprintf(w, "ISO-File:\t%s\n", isoPath)
	fmt.Fprintf(w, "Boot-Order:\t%s\n", cfg.BootOrder)
	fmt.Fprintf(w, "Graphic:\t%s\n", cfg.Graphics)
	fmt.Fprintf(w, "Sound:\t%s\n", cfg.Sound)
	fmt.Fprintf(w, "Filesystem:\t%s\n", cfg.FileSystem)
	w.Flush()

	fmt.Print(Colourise("\nPress ENTER to create VM … ", Yellow))
	_, _ = r.ReadString('\n')
}
// EOF