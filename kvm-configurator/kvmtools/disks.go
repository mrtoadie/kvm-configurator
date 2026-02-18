// kvmtools/disks.go
// last modification: Feb 18 2026
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

// fallback method via `virsh domblklist --details`
func GetDiskPathsViaVirsh(vmName string) ([]string, error) {
    // call from virsh
    out, err := exec.Command("virsh", "domblklist", vmName, "--details").
        CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("virsh domblklist failed: %w – %s", err, out)
    }

    // go through line by line
    var paths []string
    scanner := bufio.NewScanner(bytes.NewReader(out))
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        // skip header line and empty lines
        if line == "" || strings.HasPrefix(line, "Target") || strings.HasPrefix(line, "---") {
            continue
        }

        // split fields (whitespace based)
        fields := strings.Fields(line)
        // expected Layout: Target Device Type Source
        // example: vda disk block /var/lib/libvirt/images/my-vm-system.qcow2
        if len(fields) < 4 {
            // malformed – ignore, but not fatal
            continue
        }

        device := fields[1] // second field = "disk" or "cdrom"
        source := fields[3] // fourth field = path

        if device == "disk" && source != "" && source != "-" {
            paths = append(paths, source)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error scanning virsh output: %w", err)
    }

    return paths, nil
}