// kvmtools/disks.go
// last modification: Feb 17 2026
package kvmtools

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	//"path/filepath"
	"strings"
)

type domXML struct {
	Devices struct {
		Disks []struct {
			Device string `xml:"device,attr"` // "disk" or "cdrom"
			Source struct {
				File string `xml:"file,attr"`
			} `xml:"source"`
		} `xml:"disk"`
	} `xml:"devices"`
}

// GetDiskPathsFromXML reads the paths from the libvirt XML file.
func GetDiskPathsFromXML(xmlPath string) ([]string, error) {
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, err
	}
	var d domXML
	if err := xml.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	var out []string
	for _, d := range d.Devices.Disks {
		if d.Device == "disk" && d.Source.File != "" { // only disks, no iso!!
			out = append(out, d.Source.File)
		}
	}
	return out, nil
}

// Fallback method via `virsh domblklist --details`
func GetDiskPathsViaVirsh(vmName string) ([]string, error) {
    // 1️⃣ Aufruf von virsh
    out, err := exec.Command("virsh", "domblklist", vmName, "--details").
        CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("virsh domblklist failed: %w – %s", err, out)
    }

    // 2️⃣ Zeilenweise durchgehen
    var paths []string
    scanner := bufio.NewScanner(bytes.NewReader(out))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        // Header‑Zeile und leere Zeilen überspringen
        if line == "" || strings.HasPrefix(line, "Target") || strings.HasPrefix(line, "---") {
            continue
        }

        // 3️⃣ Felder splitten (Whitespace‑basiert)
        fields := strings.Fields(line)
        // Erwartetes Layout: Target Device Type Source
        // Beispiel: vda disk block /var/lib/libvirt/images/my-vm-system.qcow2
        if len(fields) < 4 {
            // malformed – ignorieren, aber nicht fatal
            continue
        }

        device := fields[1] // zweites Feld = "disk" oder "cdrom"
        source := fields[3] // viertes Feld = Pfad

        if device == "disk" && source != "" && source != "-" {
            paths = append(paths, source)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error scanning virsh output: %w", err)
    }

    return paths, nil
}