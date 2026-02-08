// model/model.go
// last modification: February 08 2026
package model

import (
	"fmt"
	"path/filepath"
	"strings"

	// internal
	"configurator/internal/config"
)

/* --------------------
	[Modul: config] Load DomainConfig
-------------------- */

type DomainConfig struct {
	Name    string // VM‑Name (z. B. "arch‑01")
	ISOPath string // Path to ISO file (can be empty > later prompt)
	MemMiB int // RAM in MiB
	VCPU   int // Anzahl vCPUs

	//A VM can have any number of disks.
	//The first disk will typically be the system drive.
	Disks []DiskSpec

	// network & other hardware
	Network    string // bridge | nat | none
	NestedVirt string // vmx | smx 
	Graphics   string // spice | vnc | none
	Sound      string // ac97 | ich6 | ich9
	FileSystem string // virtiofs | 9p | none
	BootOrder  string // e.g. "cdrom,hd"
	Firmware   string // BIOS | EFI (optional, not yet active)
}

type DiskSpec struct {
	Name    string
	Path    string
	SizeGiB int
	Bus     string
}

// [Modul: config] Load Distro‑Defaults
func EffectiveDiskPath(d config.VMConfig, global struct {
	DiskPath string
	DiskSize int
}) string {
	if d.DiskPath != "" {
		return d.DiskPath
	}
	return global.DiskPath
}

func EffectiveDiskSize(d config.VMConfig, global struct {
	DiskPath string
	DiskSize int
}) int {
	if d.DiskSize != 0 {
		return d.DiskSize
	}
	return global.DiskSize
}

// buildDiskArg create string for --disk
func BuildDiskArgs(disks []DiskSpec, vmName string) []string {
	var args []string
	for _, d := range disks {
		// Determine base path
		base := strings.TrimSpace(d.Path)

		// If only one directory is specified → <dir>/<vmName>-<disk.Name>.qcow2
		if !strings.Contains(filepath.Base(base), ".") {
			// no file name > we build a unique name
			file := fmt.Sprintf("%s-%s.qcow2", vmName, d.Name)
			base = filepath.Join(base, file)
		} else if !strings.HasSuffix(base, ".qcow2") {
			base += ".qcow2"
		}

		// Assemble options
		opts := []string{
			fmt.Sprintf("path=%s", base),
			"format=qcow2",
		}
		if d.SizeGiB > 0 {
			opts = append([]string{fmt.Sprintf("size=%d", d.SizeGiB)}, opts...)
		}
		if d.Bus != "" {
			opts = append(opts, fmt.Sprintf("bus=%s", d.Bus))
		}

		// The finished argument (e.g. "size=20,path=/var/vms/system.qcow2,format=qcow2,bus=virtio")
		args = append(args, strings.Join(opts, ","))
	}
	return args
}

// Helper: return the *first* Disk (System‑Disk) of a VM
func (c *DomainConfig) PrimaryDisk() *DiskSpec {
	if len(c.Disks) == 0 {
		return nil // no disc defined yet
	}
	return &c.Disks[0] // the first element is considered the “system disk”
}
// EOF