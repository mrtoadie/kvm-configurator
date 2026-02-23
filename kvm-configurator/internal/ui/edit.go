package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"io"
"text/tabwriter"
	"configurator/internal/model"
	"configurator/internal/style"
	"configurator/internal/utils"
)

// -------------------------------------------------------------------
// Editor – kapselt den gesamten „Edit‑Domain‑Config“-Workflow.
// -------------------------------------------------------------------
type Editor struct {
	in          *bufio.Reader   // Eingabe‑Stream (z. B. os.Stdin)
	out         io.Writer       // Ausgabe‑Stream (z. B. os.Stdout)
	cfg         *model.DomainConfig
	defaultDisk string // globaler Default‑Pfad, kommt aus dem Config‑File
	isoDir      string // Verzeichnis, in dem die ISOs liegen
}

// NewEditor ist der Konstruktor – nice und explizit.
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

// -------------------------------------------------------------------
// Run – die eigentliche Schleife, die das alte Switch‑Monster ersetzt.
// -------------------------------------------------------------------
func (e *Editor) Run() {
	for {
		e.drawMenu()
		choice, _ := utils.Prompt(e.in, e.out,
			style.PromptMsg("\nSelect or press Enter to continue: "))
		if choice == "" {
			break // fertig – zurück zum Caller
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
			editAdvanced(e.in, e.cfg) // bleibt unverändert
		default:
			fmt.Fprintln(e.out, style.Err("Invalid selection!"))
		}
	}
}

// -------------------------------------------------------------------
// Einzelne Edit‑Methoden – jeweils nur ein Prompt + ein Update.
// -------------------------------------------------------------------
func (e *Editor) editName() {
	if v, _ := utils.Prompt(e.in, e.out, ">> New Name: "); v != "" {
		e.cfg.Name = v
	}
}

func (e *Editor) editRAM() {
	ans, _ := utils.Ask(e.in, e.out, "RAM (MiB)", fmt.Sprintf("%d", e.cfg.MemMiB))
	if ans != "" {
		if i, err := utils.MustInt(ans); err == nil {
			e.cfg.MemMiB = i
		}
	}
}

func (e *Editor) editVCPU() {
	ans, _ := utils.Ask(e.in, e.out, "vCPU", fmt.Sprintf("%d", e.cfg.VCPU))
	if ans != "" {
		if i, err := utils.MustInt(ans); err == nil {
			e.cfg.VCPU = i
		}
	}
}

// Disk‑Path – nutzt den Default‑Pfad, wenn nichts eingegeben wird.
func (e *Editor) editDiskPath() {
	def := e.defaultDisk
	if primary := e.cfg.PrimaryDisk(); primary != nil && primary.Path != "" {
		def = primary.Path
	}
	ans, _ := utils.Ask(e.in, e.out, "Disk path", def)
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

// Disk‑Size – nur für das erste (System‑)Disk.
func (e *Editor) editDiskSize() {
	ans, _ := utils.Ask(e.in, e.out, "Disk size (GB)", "")
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

// Add a secondary disk – delegiert an das bereits vorhandene PromptAddDisk.
func (e *Editor) addExtraDisk() {
	if err := PromptAddDisk(e.in, e.cfg, e.defaultDisk); err != nil {
		style.RedError("Add Disk failed", "", err)
	}
}

// ISO‑Auswahl – nutzt die bereits existierende SelectISO‑Funktion.
func (e *Editor) selectISO() {
	isoPath, err := SelectISO(e.in, e.isoDir)
	if err != nil {
		fmt.Fprintf(e.out, "\x1b[31mISO selection failed: %v\x1b[0m\n", err)
		return
	}
	e.cfg.ISOPath = isoPath
	fmt.Fprintf(e.out, "\x1b[32mSelected ISO: %s\x1b[0m\n", isoPath)
}

// Netzwerk‑Parameter
func (e *Editor) editNetwork() {
	if v, _ := utils.Prompt(e.in, e.out, ">> Network (none or default): "); true {
		e.cfg.Network = v
	}
}

// -------------------------------------------------------------------
// UI‑Hilfsfunktion: Menü‑Box zeichnen (identisch zu alter Version)
// -------------------------------------------------------------------
func (e *Editor) drawMenu() {
	isoFile := filepath.Base(e.cfg.ISOPath)
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