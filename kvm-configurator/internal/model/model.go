// model/model.go
// last modification: January 18 2026
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
	Name, Disk, Network, ISOPath, NestedVirt, BootOrder, Graphics, Sound, FileSystem string
	MemMiB, VCPU, DiskSize int
}

/* --------------------
	[Modul: config] Load Distro‑Defaults
-------------------- */
func EffectiveDiskPath(d config.Distro, global struct {
	DiskPath string
	DiskSize int
}) string {
	if d.DiskPath != "" {
		return d.DiskPath
	}
	return global.DiskPath
}

func EffectiveDiskSize(d config.Distro, global struct {
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
func BuildDiskArg(cfg DomainConfig) (string, bool) {
	if strings.TrimSpace(cfg.Disk) == "" {
		return "", false
	}
	p := strings.TrimSpace(cfg.Disk)

	if !strings.HasSuffix(p, ".qcow2") && !strings.Contains(filepath.Base(p), ".") {
		p = filepath.Join(p, cfg.Name+".qcow2")
	} else if !strings.HasSuffix(p, ".qcow2") {
		p = p + ".qcow2"
	}
	opts := []string{
		fmt.Sprintf("path=%s", p),
		"format=qcow2",
	}
	if cfg.DiskSize > 0 {
		opts = append([]string{fmt.Sprintf("size=%d", cfg.DiskSize)}, opts...)
	}
	return strings.Join(opts, ","), true
}
// EOF