// kvmtools/diskops.go
// last modification: Feb 17 2026
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

func debugShowPaths(vmName string) {
    fmt.Println(utils.Colourise("\n🔍 Debug‑Info für VM: "+vmName, utils.ColorCyan))

    // 1️⃣ XML‑Pfad
    xmlPath := filepath.Join(vmName+".xml")
    fmt.Println("  • Erwarteter XML‑Pfad:", xmlPath)

    if data, err := os.ReadFile(xmlPath); err == nil {
        fmt.Println("  • XML‑Datei gefunden, Größe:", len(data), "Bytes")
        // optional: ein kurzer Ausschnitt
        preview := strings.SplitN(string(data), "\n", 5)
        fmt.Println("    └─ Vorschau:", strings.Join(preview, " | "))
    } else {
        fmt.Println("  • XML‑Datei NICHT gefunden:", err)
    }

    // 2️⃣ Virsh‑Abfrage
    fmt.Println("  • Virsh‑Abfrage (domblklist)…")
    if paths, err := GetDiskPathsViaVirsh(vmName); err == nil && len(paths) > 0 {
        fmt.Println("    └─ Virsh liefert", len(paths), "Disk‑Pfad(e):")
        for i, p := range paths {
            fmt.Printf("        [%d] %s\n", i+1, p)
        }
    } else {
        fmt.Println("    └─ Virsh liefert KEINE Disk‑Einträge:", err)
    }
}

func getRealDiskPath(vmName string) (string, error) {
	paths, err := GetDiskPathsViaVirsh(vmName)
	if err != nil {
		return "", fmt.Errorf("virsh‑Abfrage fehlgeschlagen für VM %s: %w", vmName, err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("keine Disk‑Einträge für VM %s gefunden (virsh liefert leere Liste)", vmName)
	}
	return paths[0], nil // erster Eintrag = System‑Disk
}


// ------------------------------------------------------------
// Hilfsfunktion: Pfad einer Disk aus dem XML holen
// (optional – wird momentan nicht benutzt, aber kann nützlich sein)
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
// Hilfsfunktion: Pfad einer Disk via libvirt (virsh) holen
// ------------------------------------------------------------
func diskPathFromVirsh(vmName string) (string, error) {
	paths, err := GetDiskPathsViaVirsh(vmName)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("keine Disk‑Einträge für VM %s gefunden", vmName)
	}
	// Wir gehen davon aus, dass das erste Element die System‑Disk ist.
	return paths[0], nil
}

// ------------------------------------------------------------
// 1️⃣ Resize
// ------------------------------------------------------------
func ResizeDisk(r *bufio.Reader, vmName string) error {
	imgPath, err := getRealDiskPath(vmName)
	if err != nil {
		return err
	}

	sizeStr, _ := ui.ReadLine(r,
		utils.Colourise("Neue Größe (GiB, positiv): ", utils.ColorYellow))
	newSize, err := strconv.Atoi(sizeStr)
	if err != nil || newSize <= 0 {
		return fmt.Errorf("bitte eine positive Ganzzahl eingeben")
	}
	
	
	fmt.Println(imgPath + vmName)


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
func ConvertDisk(r *bufio.Reader, vmName string) error {
	srcPath, err := getRealDiskPath(vmName)
	if err != nil {
		return err
	}

	fmt.Println(utils.Colourise("\nZiel‑Formate:", utils.ColorBlue))
	fmt.Println("[1] qcow2   (Standard, komprimiert)")
	fmt.Println("[2] raw     (uncompressed, schnell)")
	fmt.Println("[3] vdi     (VirtualBox‑Kompatibel)")

	choice, _ := ui.ReadLine(r,
		utils.Colourise("Format wählen: ", utils.ColorYellow))

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

	// XML‑Eintrag aktualisieren (optional, aber nice)
	updateXMLPath(vmName, newPath, srcPath) // wir schreiben gleich um (siehe unten)

	return nil
}

// ------------------------------------------------------------
// 3️⃣ Repair
// ------------------------------------------------------------
func RepairDisk(r *bufio.Reader, vmName string) error {
	// **Pfad ermitteln – live via virsh**
	imgPath, err := diskPathFromVirsh(vmName)
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
/* --------------------------------------------------------------
   Hilfsfunktion: XML‑Eintrag anpassen (nur für Convert)
   -------------------------------------------------------------- */
func updateXMLPath(vmName, newPath, oldPath string) {
	// Wir gehen davon aus, dass das XML im selben Verzeichnis liegt,
	// in dem `engine.CreateVM` die Datei abgelegt hat.
	// Der Pfad ist also: <xmlDir>/<vmName>.xml – wir suchen das Verzeichnis
	// dynamisch, weil wir das `xmlDir` nicht mehr übergeben.
	// Wir gehen davon aus, dass das aktuelle Arbeitsverzeichnis das
	// Projekt‑Root ist und das XML‑Verzeichnis dort liegt (z. B. "./xml").
	possibleDirs := []string{
		"./xml",
		".", // fallback: aktuelle Directory
	}
	var xmlPath string
	for _, d := range possibleDirs {
		tmp := filepath.Join(d, vmName+".xml")
		if _, err := os.Stat(tmp); err == nil {
			xmlPath = tmp
			break
		}
	}
	if xmlPath == "" {
		// Wenn wir das XML nicht finden, geben wir nur einen Hinweis aus.
		utils.RedError("XML‑Update fehlgeschlagen – Datei nicht gefunden", vmName, nil)
		return
	}

	data, err := os.ReadFile(xmlPath)
	if err != nil {
		utils.RedError("XML‑Update fehlgeschlagen (Lesen)", xmlPath, err)
		return
	}
	updated := strings.ReplaceAll(string(data), oldPath, newPath)
	if err := os.WriteFile(xmlPath, []byte(updated), 0644); err != nil {
		utils.RedError("XML‑Update fehlgeschlagen (Schreiben)", xmlPath, err)
	}
}

// ------------------------------------------------------------
// Sub‑Menu, das du aus `vmmenu.go` aufrufst
// ------------------------------------------------------------

func DiskOpsMenu(r *bufio.Reader, vmName string) error {
	for {
		fmt.Println(utils.BoxCenter(55,
			[]string{"=== DISK‑OPERATIONS für " + vmName + " ==="}))
		fmt.Println(utils.Box(55, []string{
			"[1] Resize  (Größe ändern)",
			"[2] Convert (Format wechseln)",
			"[3] Repair  (Image prüfen)",
			"[0] Zurück",
		}))

		choice, _ := ui.ReadLine(r,
			utils.Colourise("\nAuswahl: ", utils.ColorYellow))

		switch choice {
		case "1":
			return ResizeDisk(r, vmName)
		case "2":
			return ConvertDisk(r, vmName)
		case "3":
			return RepairDisk(r, vmName)
		case "0", "":
			return nil // zurück zum VM‑Menu
		default:
			fmt.Println(utils.Colourise("Ungültige Auswahl!", utils.ColorRed))
		}
	}
}

