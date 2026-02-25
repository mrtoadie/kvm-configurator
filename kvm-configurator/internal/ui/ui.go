// ui/ui.go
// last modified: Feb 27 2026
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	// internal packages
	"configurator/internal/config"
	"configurator/internal/model"
	"configurator/internal/style"
	"configurator/internal/utils"
)

// PUBLIC API (functions used by the rest of the program)

// NewEditor creates the interactive “customise VM” editor
func NewEditor(r *bufio.Reader, w io.Writer,
	cfg *model.DomainConfig, defaultDisk, isoDir string) *Editor {
	return &Editor{
		in:          r,
		out:         w,
		cfg:         cfg,
		defaultDisk: defaultDisk,
		isoDir:      isoDir,
	}
}

// SelectDistro lets the user pick a distribution from the YAML list
func SelectDistro(r *bufio.Reader, list []config.VMConfig) (config.VMConfig, error) {
	return selectDistroImpl(r, list)
}

// SelectISO lets the user pick an ISO file from a directory
func SelectISO(r *bufio.Reader, workDir string) (string, error) {
	return selectISOImpl(r, workDir)
}

// ShowSummary prints a final overview before the VM is created
func ShowSummary(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
	showSummaryImpl(r, cfg, isoPath)
}

// EDITOR (the “CUSTOMIZE VM” loop)
type Editor struct {
	in          *bufio.Reader
	out         io.Writer
	cfg         *model.DomainConfig
	defaultDisk string
	isoDir      string
}

// Run executes the interactive editor
func (e *Editor) Run() {
	for {
		e.drawMenu()
		choice, _ := utils.Prompt(e.in, e.out,
			style.PromptMsg("\nSelect or press Enter to continue: "))
		if choice == "" {
			break
		}
		switch choice {
		case "1":
			e.editName()
		case "2":
			e.editRAM()
		case "3":
			e.editVCPU()
		case "4":
			e.editDiskPath()
		case "5":
			e.editDiskSize()
		case "6":
			e.addExtraDisk()
		case "7":
			e.selectISO()
		case "8":
			e.editNetwork()
		case "0":
			editAdvanced(e.in, e.cfg) // advanced submenu
		default:
			fmt.Fprintln(e.out, style.Err("Invalid selection!"))
		}
	}
}

// custom vm edit actions
func (e *Editor) editName() {
	if v, _ := utils.Prompt(e.in, e.out, style.PromptMsg(">> New Name: ")); v != "" {
		e.cfg.Name = v
	}
}

func (e *Editor) editRAM() {
	ans, _ := utils.Ask(e.in, e.out, style.PromptMsg("RAM (MiB)"), fmt.Sprintf("%d", e.cfg.MemMiB))
	if ans != "" {
		if i, err := utils.MustInt(ans); err == nil {
			e.cfg.MemMiB = i
		}
	}
}

func (e *Editor) editVCPU() {
	ans, _ := utils.Ask(e.in, e.out, style.PromptMsg("vCPU"), fmt.Sprintf("%d", e.cfg.VCPU))
	if ans != "" {
		if i, err := utils.MustInt(ans); err == nil {
			e.cfg.VCPU = i
		}
	}
}

// Disk‑Path – uses the default path if nothing is entered
func (e *Editor) editDiskPath() {
	def := e.defaultDisk
	if primary := e.cfg.PrimaryDisk(); primary != nil && primary.Path != "" {
		def = primary.Path
	}
	ans, _ := utils.Ask(e.in, e.out, style.PromptMsg("Disk path"), def)
	if ans == "" {
		ans = def
	}
	if primary := e.cfg.PrimaryDisk(); primary != nil {
		primary.Path = os.ExpandEnv(ans)
	} else {
		e.cfg.Disks = append(e.cfg.Disks, model.DiskSpec{
			Name: "system",
			Path: os.ExpandEnv(ans),
		})
	}
	fmt.Fprintf(e.out, "\x1b[32mDisk will be stored at: %s\x1b[0m\n",
		e.cfg.PrimaryDisk().Path)
}

// Disk‑Size – only for the first (system) disk
func (e *Editor) editDiskSize() {
	ans, _ := utils.Ask(e.in, e.out, style.PromptMsg("Disk size (GB)"), "")
	if ans == "" {
		return
	}
	if i, err := utils.MustInt(ans); err == nil {
		if primary := e.cfg.PrimaryDisk(); primary != nil {
			primary.SizeGiB = i
		} else {
			e.cfg.Disks = append(e.cfg.Disks, model.DiskSpec{
				Name:    "system",
				SizeGiB: i,
			})
		}
	}
}

// Add a secondary disk – delegated to PromptAddDisk
func (e *Editor) addExtraDisk() {
	if err := PromptAddDisk(e.in, e.cfg, e.defaultDisk); err != nil {
		style.RedError("Add Disk failed", "", err)
	}
}

// ISO selection – uses the existing SelectISO helper
func (e *Editor) selectISO() {
	isoPath, err := SelectISO(e.in, e.isoDir)
	if err != nil {
		fmt.Fprintf(e.out, "\x1b[31mISO selection failed: %v\x1b[0m\n", err)
		return
	}
	e.cfg.ISOPath = isoPath
	fmt.Fprintf(e.out, "\x1b[32mSelected ISO: %s\x1b[0m\n", isoPath)
}

// Network
func (e *Editor) editNetwork() {
	if v, _ := utils.Prompt(e.in, e.out, style.PromptMsg(">> Network (none or default): ")); true {
		e.cfg.Network = v
	}
}

// UI rendering for the editor
func (e *Editor) drawMenu() {
	isoFile := filepath.Base(e.cfg.ISOPath)

	fmt.Println(style.BoxCenter(51, []string{"CUSTOMIZE VM"}))
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintf(w, "[1] Name:\t%s\n", e.cfg.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\n", e.cfg.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\n", e.cfg.VCPU)
		if primary := e.cfg.PrimaryDisk(); primary != nil {
			fmt.Fprintf(w, "[4] Disk-Path:\t%s\n", primary.Path)
			fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\n", primary.SizeGiB)
		} else {
			fmt.Fprintf(w, "[4] Disk-Path:\t<none>\n")
			fmt.Fprintf(w, "[5] Disk-Size (GB):\t<none>\n")
		}
		fmt.Fprintln(w, "[6] Add more disks")
		fmt.Fprintf(w, "[7] ISO:\t%s\n", isoFile)
		fmt.Fprintf(w, "[8] Network:\t%s\n", e.cfg.Network)
		fmt.Fprintln(w, "[0] Advanced Parameters")
	})
	fmt.Println(style.Box(51, lines))
}

// ADVANCED PARAMETERS SUB‑MENU
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
		fmt.Println(style.BoxCenter(51, []string{"ADVANCED PARAMETERS"}))
		printAdvancedMenu(cfg)

		choice, err := utils.Prompt(r, os.Stdout,
			style.PromptMsg("\nSelect an option (or press Enter to go back): "))
		if err != nil {
			if err == io.EOF {
				return // Ctrl‑D
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

// sub‑functions for each advanced option
func editNested(r *bufio.Reader, cfg *model.DomainConfig) {
	editEnum(r, cfg, &cfg.NestedVirt,
		">> Nested-Virtualisation (vmx for Intel, smx for AMD): ",
		"Nested-Virtualisation",
		map[string]bool{"vmx": true, "smx": true})
}

func editBoot(r *bufio.Reader, cfg *model.DomainConfig) {
	editEnum(r, cfg, &cfg.BootOrder,
		">> Boot order (comma-separated, e.g. hd,cdrom,network): ",
		"Boot order",
		map[string]bool{"hd": true, "cdrom": true, "network": true})
}

func editGraphics(r *bufio.Reader, cfg *model.DomainConfig) {
	editEnum(r, cfg, &cfg.Graphics,
		">> Graphics (spice (default) or vnc): ",
		"Graphics",
		map[string]bool{"spice": true, "vnc": true})
}

func editSound(r *bufio.Reader, cfg *model.DomainConfig) {
	editEnum(r, cfg, &cfg.Sound,
		">> Sound (none, ac97, ich6, ich9 (default)): ",
		"Sound",
		map[string]bool{"none": true, "ac97": true, "ich6": true, "ich9": true})
}

func editFilesystem(r *bufio.Reader, cfg *model.DomainConfig) {
	if v, err := utils.Prompt(r, os.Stdout,
		style.Hint(">> Filesystem / Mount (/src/dir,/guest/dir): ")); err == nil && v != "" {
		cfg.FileSystem = v
		style.Success("Filesystem", v, "")
	}
}

// advanced menu screen
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
	fmt.Println(style.Box(51, lines))
}

// DISK HELPERS (add‑disk)

// PromptAddDisk asks the user for a secondary disk and appends it to cfg
func PromptAddDisk(r *bufio.Reader, cfg *model.DomainConfig, defaultDiskPath string) error {
	fmt.Println(style.BoxCenter(55, []string{"=== ADD DISK ==="}))

	// name
	name, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Disk name (e.g. swap, data, backup): "))
	if name == "" {
		return nil
	}

	// path – suggest either the primary‑disk path or the global default
	var suggestedPath string
	if primary := cfg.PrimaryDisk(); primary != nil && primary.Path != "" {
		suggestedPath = primary.Path
	} else {
		suggestedPath = defaultDiskPath
	}
	prompt := fmt.Sprintf("Disk path (default: %s): ", suggestedPath)
	path, _ := utils.Prompt(r, os.Stdout, style.Colourise(prompt, style.ColYellow))
	if path == "" {
		path = suggestedPath
	}

	// size
	sizeStr, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Size in GiB (0 = default): "))
	size, _ := strconv.Atoi(sizeStr)

	// optional bus
	bus, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Bus (virtio|scsi|sata|usb, default virtio): "))
	if bus == "" {
		bus = "virtio"
	}

	// attach
	cfg.Disks = append(cfg.Disks, model.DiskSpec{
		Name:    name,
		Path:    path,
		SizeGiB: size,
		Bus:     bus,
	})
	style.Success("Disk", name, "added")
	return nil
}

// DISTRIBUTION & ISO SELECTION
func selectDistroImpl(r *bufio.Reader, list []config.VMConfig) (config.VMConfig, error) {
	fmt.Println(style.BoxCenter(51, []string{"Select an operating system"}))

	// sort case‑insensitively
	sorted := append([]config.VMConfig(nil), list...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	// table
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "No.\tName\tCPU\tRAM (MiB)\tDisk (GB)")
		fmt.Fprintln(w, "---\t----\t---\t---------\t---------")
		for i, d := range sorted {
			fmt.Fprintf(w, "%2d\t%s\t%d\t%d\t%d\n",
				i+1, d.Name, d.CPU, d.RAM, d.DiskSize)
		}
	})
	fmt.Print(style.Box(51, lines))

	// prompt
	ans, err := utils.Prompt(r, os.Stdout,
		style.PromptMsg("\nPlease enter a number (or press ENTER for default Arch Linux): "))
	if err != nil {
		return config.VMConfig{}, err
	}
	// default = first entry
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

func selectISOImpl(r *bufio.Reader, workDir string) (string, error) {
	files, err := utils.ListFiles(workDir)
	if err != nil {
		return "", fmt.Errorf("listing files in %s failed: %w", workDir, err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no files found in %s", workDir)
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})

	fmt.Println(style.BoxCenter(60, []string{"Select ISO"}))
	choice, err := utils.PromptSelection(bufio.NewReader(os.Stdin), os.Stdout, files)
	if err != nil {
		return "", err
	}
	if choice == utils.CancelChoice {
		return "", fmt.Errorf("selection aborted")
	}
	selected := files[choice-1]
	abs, _ := filepath.Abs(selected)
	return abs, nil
}

// SUMMARY DISPLAY
func showSummaryImpl(r *bufio.Reader, cfg *model.DomainConfig, isoPath string) {
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
