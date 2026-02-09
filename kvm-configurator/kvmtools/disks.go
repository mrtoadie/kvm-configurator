// kvmtools/disks.go
// last modification: Feb 09 2026
package kvmtools

import (
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
			Device string `xml:"device,attr"` // "disk" oder "cdrom"
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
	out, err := exec.Command("virsh", "domblklist", vmName, "--details").
		CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("virsh domblklist failed: %w – %s", err, out)
	}
	var paths []string
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) >= 4 && f[0] != "Target" { //Skip header
			if p := f[3]; p != "" && p != "-" {
				paths = append(paths, p)
			}
		}
	}
	return paths, nil
}