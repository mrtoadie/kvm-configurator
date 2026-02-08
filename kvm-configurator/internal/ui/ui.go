// ui/ui.go
// last modification: February 08 2026
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"text/tabwriter"
	// internal
	"configurator/internal/config"
	"configurator/internal/fileutils"
	"configurator/internal/model"
	"configurator/internal/utils"
)

// waiting for input
func ReadLine(r *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	s, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", io.EOF // clean exit (e.g. CTRL+D)
		}
		return "", err
	}
	return strings.TrimSpace(s), nil
}

// Loading distro list from yaml
func PromptSelectDistro(r *bufio.Reader, list []config.VMConfig) (config.VMConfig, error) {
	//fmt.Println(utils.Colourise("\n=== Select an operating system ===", utils.ColorBlue))
	fmt.Println(utils.BoxCenter(51, []string{"Select an operating system"}))

	sorted := append([]config.VMConfig(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	lines := utils.TableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "No.\tName\tCPU\tRAM (MiB)\tDisk (GB)")
		fmt.Fprintln(w, "---\t----\t---\t---------\t---------")
		for i, d := range sorted {
			fmt.Fprintf(w, "%2d\t%s\t%d\t%d\t%d\n",
				i+1, d.Name, d.CPU, d.RAM, d.DiskSize)
		}
		w.Flush()
	})
	fmt.Print(utils.Box(51, lines))

	line, err := ReadLine(r, utils.Colourise("\nPlease enter a number (or press ENTER for default Arch Linux): ", utils.ColorYellow))
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
		fmt.Println(utils.BoxCenter(51, []string{"=== VM-CONFIG ==="}))
		// Convert menu lines into line slices with Tabwriter
		lines := utils.TableToLines(func(w *tabwriter.Writer) {
			fmt.Fprintf(w, "[1] Name:\t%s\n", cfg.Name)
			fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t\n", cfg.MemMiB)
			fmt.Fprintf(w, "[3] vCPU:\t%d\n", cfg.VCPU)
			// show first disk
			if primary := cfg.PrimaryDisk(); primary != nil {
				fmt.Fprintf(w, "[4] Disk-Path:\t%s\n", primary.Path)
				fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\n", primary.SizeGiB)
			} else {
				fmt.Fprintf(w, "[4] Disk-Path:\t<none>\n")
				fmt.Fprintf(w, "[5] Disk-Size (GB):\t<none>\n")
			}
			fmt.Fprintf(w, "[6] ISO:\t%s\n", isoFile) //cfg.ISOPath
			fmt.Fprintf(w, "[7] Network:\t%s\n", cfg.Network)
			fmt.Fprintf(w, "[8] Advanced Parameters\n")
			fmt.Fprintf(w, "[9] Add disks\n")
		})
		// Build a box around it and spend it
		fmt.Println(utils.Box(51, lines))

		f, _ := ReadLine(r, utils.Colourise("\nSelect or press Enter to continue: ", utils.ColorYellow))
		if f == "" {
			break
		}

		switch f {
		case "1":
			if v, _ := ReadLine(r, ">> New Name: "); v != "" {
				cfg.Name = v
			}
		case "2":
			if v, _ := ReadLine(r, ">> RAM (MiB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.MemMiB = i
				}
			}
		case "3":
			if v, _ := ReadLine(r, ">> vCPU: "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					cfg.VCPU = i
				}
			}
		case "4":
			// Change disk path (we edit the *first*disk)
			prompt := fmt.Sprintf(">> Disk path (default: %s): ", defaultDiskPath)
			if v, _ := ReadLine(r, prompt); v != "" {
				// If no disc exists yet, we create one
				if primary := cfg.PrimaryDisk(); primary != nil {
					primary.Path = os.ExpandEnv(v)
				} else {
					// Create a new system disk
					cfg.Disks = append(cfg.Disks, model.DiskSpec{
						Name: "system",
						Path: os.ExpandEnv(v),
						// Size and Bus remain empty -can be set later
					})
				}
			} else {
				// Empty input → use default
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
			// Change disk size (only for the first disk)
			if v, _ := ReadLine(r, ">> Disk size (GB): "); v != "" {
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
		case "7":
			if v, _ := ReadLine(r, ">> Network (none or default): "); true {
				cfg.Network = v
			}
		case "8":
			editAdvanced(r, cfg)
		case "6":
			isoPath, err := PromptSelectISO(r, isoWorkDir)
			if err != nil {
				fmt.Printf("\x1b[31mISO selection failed: %v\x1b[0m\n", err)
				continue
			}
			cfg.ISOPath = isoPath
			fmt.Printf("\x1b[32mSelected ISO: %s\x1b[0m\n", isoPath)
		case "9":
			if err := PromptAddDisk(r, cfg, defaultDiskPath); err != nil {
				utils.RedError("Add Disk failed", "", err)
			}
		}
	}
}

// Form – Advanced Parameters
func editAdvanced(r *bufio.Reader, cfg *model.DomainConfig) {
	fmt.Println(utils.BoxCenter(51, []string{"=== ADVANCED PARAMETERS ==="}))
	for {
		lines := utils.TableToLines(func(w *tabwriter.Writer) {
			fmt.Fprintln(w, "Parameter\t Default\t Set")
			fmt.Fprintln(w, "---------\t -------\t ---")
			fmt.Fprintln(w, "[a] Nested-Virtualisation\t", cfg.NestedVirt)
			fmt.Fprintln(w, "[b] Boot-Order\t", cfg.BootOrder)
			fmt.Fprintln(w, "[c] Graphics\t", cfg.Graphics)
			fmt.Fprintln(w, "[d] Sound\t", cfg.Sound)
			fmt.Fprintln(w, "[e] Filesystem\t", cfg.FileSystem)
			fmt.Fprintln(w, "[0] Back to main menu")
		})
		fmt.Println(utils.Box(60, lines))

		choice, _ := ReadLine(r, utils.Colourise("\nSelect an option (or press Enter to go back): ", utils.ColorYellow))
		if choice == "" || strings.EqualFold(choice, "0") {
			return
		}
		switch strings.ToLower(choice) {
		case "a":
			if v, _ := ReadLine(r, utils.Colourise(">> Nested-Virtualisation (vmx for Intel or smx for AMD): ", utils.ColorBlue)); v != "" {
				cfg.NestedVirt = v
				fmt.Println("Nested-Virtualisation is set to\x1b[32m", v)
			}
		case "b": // bug - nothing happend
			if v, _ := ReadLine(r, ">> Boot order: "); v != "" {
				cfg.BootOrder = v
				fmt.Println("Boot order is set to", v)
			}
		case "c":
			if v, _ := ReadLine(r, utils.Colourise(">> Graphics (spice (default) or vnc): ", utils.ColorBlue)); v != "" {
				cfg.Graphics = v
				fmt.Println(utils.Colourise("Graphics is set to", utils.ColorBlue), v)
			}
		case "d":
			if v, _ := ReadLine(r, utils.Colourise(">> Sound (none, ac97, ich6 or ich9 (default)): ", utils.ColorBlue)); v != "" {
				cfg.Sound = v
				fmt.Println(utils.Colourise("Sound is set to", utils.ColorBlue), v)
			}
		case "e":
			if v, _ := ReadLine(r, utils.Colourise(">> Filesystem / Mount (/my/source/dir,/dir/in/guest): ", utils.ColorBlue)); v != "" {
				cfg.FileSystem = v
				fmt.Println(utils.Colourise("Filesystem / Mount is set to", utils.ColorBlue), v)
			}
		default:
			//fmt.Println(Colourise("Invalid input!", Red))
			//WarnSoft(ErrSelection, "")
		}
	}
}

// Summary table before vm creation
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	// trim filepath and filename to only display filename
	isoFile := filepath.Base(cfg.ISOPath)
	//isoName := strings.TrimSuffix(isoFile, filepath.Ext(isoFile))

	fmt.Println(utils.BoxCenter(51, []string{"=== VM-SUMMARY ==="}))
	lines := utils.TableToLines(func(w *tabwriter.Writer) {
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

	fmt.Println(utils.Box(60, lines))

	fmt.Print(utils.Colourise("\nPress ENTER to create VM … ", utils.ColorYellow))
	_, _ = r.ReadString('\n')
}

// EOF
