// kvmtools/diskops.go
// last modification: Feb 16 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"configurator/internal/ui"
	"configurator/internal/utils"
)

// ------------------------------------------------------------
// Hilfsfunktion: Pfad einer Disk aus dem XML holen
// ------------------------------------------------------------
func diskPathFromXML(vmName, xmlDir string) (string, error) {
	xmlPath := filepath.Join(xmlDir, vmName+".xml")
	paths, err := GetDiskPathsFromXML(xmlPath) // bereits in kvmtools/disks.go definiert
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("keine Disk‑Einträge im XML gefunden")
	}
	// Wir nehmen einfach die erste (System‑)Disk – das reicht für die Demo‑Ops
	return paths[0], nil
}

// ------------------------------------------------------------
// 1️⃣ Resize
// ------------------------------------------------------------
func ResizeDisk(r *bufio.Reader, vmName, xmlDir string) error {
	imgPath, err := diskPathFromXML(vmName, xmlDir)
	if err != nil {
		return err
	}

	sizeStr, _ := ui.ReadLine(r, utils.Colourise("Neue Größe (GiB, positiv): ", utils.ColorYellow))
	newSize, err := strconv.Atoi(sizeStr)
	if err != nil || newSize <= 0 {
		return fmt.Errorf("bitte eine positive Ganzzahl eingeben")
	}

	spinner := utils.SpinnerProgress("Resize läuft …")
	defer spinner.Stop()

	cmd := exec.Command("qemu-img", "resize", imgPath, fmt.Sprintf("+%dG", newSize))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize fehlgeschlagen: %v – %s", err, string(out))
	}
	utils.Successf("Disk %s um %d GiB vergrößert", filepath.Base(imgPath), newSize)
	return nil
}

// ------------------------------------------------------------
// 2️⃣ Convert
// ------------------------------------------------------------
func ConvertDisk(r *bufio.Reader, vmName, xmlDir string) error {
	srcPath, err := diskPathFromXML(vmName, xmlDir)
	if err != nil {
		return err
	}

	fmt.Println(utils.Colourise("\nZiel‑Formate:", utils.ColorBlue))
	fmt.Println("[1] qcow2   (Standard, komprimiert)")
	fmt.Println("[2] raw     (uncompressed, schnell)")
	fmt.Println("[3] vdi     (VirtualBox‑Kompatibel)")

	choice, _ := ui.ReadLine(r, utils.Colourise("Format wählen: ", utils.ColorYellow))
	var tgtFmt string
	switch choice {
	case "1":
		tgtFmt = "qcow2"
	case "2":
		tgtFmt = "raw"
	case "3":
		tgtFmt = "vdi"
	default:
		return fmt.Errorf("unbekanntes Format")
	}

	ext := "." + tgtFmt
	newPath := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ext

	spinner := utils.SpinnerProgress("Conversion läuft …")
	defer spinner.Stop()

	cmd := exec.Command("qemu-img", "convert", "-O", tgtFmt, srcPath, newPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion fehlgeschlagen: %v – %s", err, string(out))
	}
	utils.Successf("Disk %s nach %s konvertiert", filepath.Base(srcPath), tgtFmt)

	// Optional: XML‑Eintrag aktualisieren – das ist ein kleiner Bonus:
	updateXMLPath(vmName, xmlDir, srcPath, newPath)

	return nil
}

// ------------------------------------------------------------
// 3️⃣ Repair
// ------------------------------------------------------------
func RepairDisk(r *bufio.Reader, vmName, xmlDir string) error {
	imgPath, err := diskPathFromXML(vmName, xmlDir)
	if err != nil {
		return err
	}

	// 1️⃣ Prüfen
	checkCmd := exec.Command("qemu-img", "check", imgPath)
	out, err := checkCmd.CombinedOutput()
	if err == nil {
		fmt.Println(utils.Colourise("\nDisk ist intakt – kein Eingriff nötig.", utils.ColorGreen))
		fmt.Printf("%s\n", string(out))
		return nil
	}

	// 2️⃣ Reparieren (amend ist das einfachste Mittel)
	fmt.Println(utils.Colourise("\nInkonsistenz entdeckt – versuche Reparatur …", utils.ColorRed))
	spinner := utils.SpinnerProgress("Repair läuft …")
	defer spinner.Stop()

	repairCmd := exec.Command("qemu-img", "amend", "-f", "qcow2", imgPath)
	repOut, repErr := repairCmd.CombinedOutput()
	if repErr != nil {
		return fmt.Errorf("Reparatur fehlgeschlagen: %v – %s", repErr, string(repOut))
	}
	utils.Successf("Disk %s repariert", filepath.Base(imgPath))
	fmt.Printf("%s\n", string(repOut))
	return nil
}

// ------------------------------------------------------------
// Hilfsfunktion: XML‑Eintrag anpassen (nur für Convert)
// ------------------------------------------------------------
func updateXMLPath(vmName, xmlDir, oldPath, newPath string) {
	xmlFile := filepath.Join(xmlDir, vmName+".xml")
	data, err := os.ReadFile(xmlFile)
	if err != nil {
		return // wir ignorieren Fehler – das XML wird beim nächsten `virsh define` neu gelesen
	}
	updated := strings.ReplaceAll(string(data), oldPath, newPath)
	_ = os.WriteFile(xmlFile, []byte(updated), 0644)
}

// ------------------------------------------------------------
// Sub‑Menu, das du aus `vmmenu.go` aufrufst
// ------------------------------------------------------------
func DiskOpsMenu(r *bufio.Reader, vmName, xmlDir string) error {
	for {
		fmt.Println(utils.BoxCenter(55, []string{"=== DISK‑OPERATIONS für " + vmName + " ==="}))
		fmt.Println(utils.Box(55, []string{
			"[1] Resize  (Größe ändern)",
			"[2] Convert (Format wechseln)",
			"[3] Repair  (Image prüfen)",
			"[0] Zurück",
		}))

		choice, _ := ui.ReadLine(r, utils.Colourise("\nAuswahl: ", utils.ColorYellow))
		switch choice {
		case "1":
			return ResizeDisk(r, vmName, xmlDir)
		case "2":
			return ConvertDisk(r, vmName, xmlDir)
		case "3":
			return RepairDisk(r, vmName, xmlDir)
		case "0", "":
			return nil // zurück zum VM‑Menu
		default:
			fmt.Println(utils.Colourise("Ungültige Auswahl!", utils.ColorRed))
		}
	}
}