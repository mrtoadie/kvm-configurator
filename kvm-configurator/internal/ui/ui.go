// ui/ui.go
// last modified: Feb 22 2026
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
	"configurator/internal/utils"
)

// Loading distro list from yaml
func PromptSelectDistro(r *bufio.Reader, list []config.VMConfig) (config.VMConfig, error) {
	fmt.Println(utils.BoxCenter(51, []string{"Select an operating system"}))

	sorted := append([]config.VMConfig(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	lines := utils.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "No.\tName\tCPU\tRAM (MiB)\tDisk (GB)")
		fmt.Fprintln(w, "---\t----\t---\t---------\t---------")
		for i, d := range sorted {
			fmt.Fprintf(w, "%2d\t%s\t%d\t%d\t%d\n",
				i+1, d.Name, d.CPU, d.RAM, d.DiskSize)
		}
		w.Flush()
	})
	fmt.Print(utils.Box(51, lines))

	line, err := utils.Prompt(r, os.Stdout, utils.Colourise("\nPlease enter a number (or press ENTER for default Arch Linux): ", utils.ColorYellow))
	if err != nil {
		return config.VMConfig{}, err
	}
	idx := 1 // default = Arch
	if line != "" {
		if i, e := strconv.Atoi(line); e == nil && i >= 1 && i <= len(sorted) {
			idx = i
		} else {
			return config.VMConfig{}, fmt.Errorf(utils.Colourise("Invalid selection", utils.ColorRed))
		}
	}
	return sorted[idx-1], nil
}

/*
PromptSelectISO – selects an ISO file from the specified directory
The return value is the 'absolute path' to the file (for virt‑install)
*/
func PromptSelectISO(r *bufio.Reader, workDir string) (string, error) {
	// workDir is directory from filepaths.isopath
	files, err := utils.ListFiles(workDir)
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
	choice, err := utils.PromptSelection(bufio.NewReader(os.Stdin), os.Stdout, files)
	if err != nil {
		return "", err
	}
	if choice == 0 {
		return "", fmt.Errorf(utils.Colourise("selection aborted", utils.ColorYellow))
	}
	selected := files[choice-1]
	// Return the 'absolute path' so that virt‑install can find it reliably
	abs, _ := filepath.Abs(selected)
	return abs, nil
}

// Form – allows changes to the fields
func PromptEditDomainConfig(r *bufio.Reader, cfg *model.DomainConfig, defaultDiskPath string, isoWorkDir string) {
	for {
		isoFile := filepath.Base(cfg.ISOPath)
		fmt.Println(utils.BoxCenter(51, []string{"VM-CONFIG"}))
		// Convert menu lines into line slices with Tabwriter
		lines := utils.MustTableToLines(func(w *tabwriter.Writer) {
			fmt.Fprintf(w, "[1] Name:\t%s\n", cfg.Name)
			fmt.Fprintf(w, "[2] RAM (MiB):\t%d\n", cfg.MemMiB)
			fmt.Fprintf(w, "[3] vCPU:\t%d\n", cfg.VCPU)
			// show first disk
			if primary := cfg.PrimaryDisk(); primary != nil {
				fmt.Fprintf(w, "[4] Disk-Path:\t%s\n", primary.Path)
				fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\n", primary.SizeGiB)
			} else {
				fmt.Fprintf(w, "[4] Disk-Path:\t<none>\n")
				fmt.Fprintf(w, "[5] Disk-Size (GB):\t<none>\n")
			}
			fmt.Fprintf(w, "[6] Add more disks\n")
			fmt.Fprintf(w, "[7] ISO:\t%s\n", isoFile)
			fmt.Fprintf(w, "[8] Network:\t%s\n", cfg.Network)
			fmt.Fprintf(w, "[0] Advanced Parameters\n")
		})
		// Build a box around it and spend it
		fmt.Println(utils.Box(51, lines))

		f, _ := utils.Prompt(r, os.Stdout, utils.Colourise("\nSelect or press Enter to continue: ", utils.ColorYellow))
		if f == "" {
			break
		}

		switch f {
		case "1":
			if v, _ := utils.Prompt(r, os.Stdout, ">> New Name: "); v != "" {
				cfg.Name = v
			}
		case "2":
			if v, _ := utils.Prompt(r, os.Stdout, ">> RAM (MiB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.MemMiB = i
				}
			}
		case "3":
			if v, _ := utils.Prompt(r, os.Stdout, ">> vCPU: "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.VCPU = i
				}
			}
		case "4":
			// Change disk path (we edit the *first*disk)
			prompt := fmt.Sprintf(">> Disk path (default: %s): ", defaultDiskPath)
			if v, _ := utils.Prompt(r, os.Stdout, prompt); v != "" {
				// If no disc exists yet, we create one
				if primary := cfg.PrimaryDisk(); primary != nil {
					primary.Path = os.ExpandEnv(v)
				} else {
					// create a new system disk
					cfg.Disks = append(cfg.Disks, model.DiskSpec{
						Name: "system",
						Path: os.ExpandEnv(v),
						// Size and Bus remain empty -can be set later
					})
				}
			} else {
				// empty input > use default
				if primary := cfg.PrimaryDisk(); primary != nil {
					primary.Path = os.ExpandEnv(defaultDiskPath)
				} else {
					cfg.Disks = append(cfg.Disks, model.DiskSpec{
						Name: "system",
						Path: os.ExpandEnv(defaultDiskPath),
					})
				}
			}
			if primary := cfg.PrimaryDisk(); primary != nil {
				fmt.Printf("\x1b[32mDisk will be stored at: %s\x1b[0m\n", primary.Path)
			}

		case "5":
			// change disk size (only for the first disk)
			if v, _ := utils.Prompt(r, os.Stdout, ">> Disk size (GB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					if primary := cfg.PrimaryDisk(); primary != nil {
						primary.SizeGiB = i
					} else {
						// no disc yet > create a new one
						cfg.Disks = append(cfg.Disks, model.DiskSpec{
							Name:    "system",
							SizeGiB: i,
						})
					}
				}
			}
		case "6":
			if err := PromptAddDisk(r, cfg, defaultDiskPath); err != nil {
				utils.RedError("Add Disk failed", "", err)
			}
		case "7":
			isoPath, err := PromptSelectISO(r, isoWorkDir)
			if err != nil {
				fmt.Printf("\x1b[31mISO selection failed: %v\x1b[0m\n", err)
				continue
			}
			cfg.ISOPath = isoPath
			fmt.Printf("\x1b[32mSelected ISO: %s\x1b[0m\n", isoPath)
		case "8":
			if v, _ := utils.Prompt(r, os.Stdout, ">> Network (none or default): "); true {
				cfg.Network = v
			}
		case "0":
			editAdvanced(r, cfg)
		}
	}
}

// Summary table before vm creation
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	// trim filepath and filename to only display filename
	isoFile := filepath.Base(cfg.ISOPath)
	//isoName := strings.TrimSuffix(isoFile, filepath.Ext(isoFile))

	fmt.Println(utils.BoxCenter(51, []string{"VM-SUMMARY"}))
	lines := utils.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
		fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
		fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
		// is disk primary or addition
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

	fmt.Println(utils.Box(51, lines))

	_, _ = utils.Prompt(r, os.Stdout,
		utils.Colourise("\nPress ENTER to create VM … ", utils.ColorYellow))
}
