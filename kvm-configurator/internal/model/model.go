// model/model.go
// last modification: February 03 2026
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
	Name, DiskPath, Network, ISOPath string
	NestedVirt, BootOrder string
	Graphics, Sound, FileSystem string
	MemMiB, VCPU, DiskSize int
}

/* --------------------
	[Modul: config] Load Distro‑Defaults
-------------------- */
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

/* --------------------
	buildDiskArg create string for --disk
-------------------- */
func BuildDiskArg(cfg DomainConfig) (arg string, ok bool) {
	// no input > no disk-arg
	if strings.TrimSpace(cfg.DiskPath) == "" {
		return "", false
	}

	// normalise base path
	base := strings.TrimSpace(cfg.DiskPath)

	// check if needs to attach '.qcow2'
	if strings.Contains(filepath.Base(base), ".") {
		if !strings.HasSuffix(base, ".qcow2") {
			base += ".qcow2"
		}
	} else {
		// Only directory specified > Append file <VM name>.qcow2
		base = filepath.Join(base, cfg.Name+".qcow2")
	}

	// put them together
	opts := []string{
		fmt.Sprintf("path=%s", base),
		"format=qcow2",
	}
	// add disk size if > 0
	if cfg.DiskSize > 0 {
		opts = append([]string{fmt.Sprintf("size=%d", cfg.DiskSize)}, opts...)
	}

	return strings.Join(opts, ","), true
}
// EOF

