// internal/config/config.go
// last modified: Feb 22 2026
package config

import (
	"fmt"
	"os"
	// internal
	"configurator/internal/utils"
	// external
	"gopkg.in/yaml.v3"
)

type FullConfig struct {
	IsoPath  string     // path to iso dir
	XmlDir   string     // path to xml save dir
	Defaults Defaults   // global‑defaults (DiskPath, DiskSize …)
	OSList   []VMConfig // OS‑Liste
}

// VMConfig represents a single operating‑system or guest definition coming from the YAML file
type VMConfig struct {
	Name       string `yaml:"name"`       // display name "Arch Linux"
	ID         string `yaml:"id"`         // identifier, e.g. "archlinux"
	CPU        int    `yaml:"cpu"`        // number of vCPUs
	RAM        int    `yaml:"ram"`        // RAM in MiB
	DiskSize   int    `yaml:"disksize"`   // disk size in GiB
	DiskPath   string `yaml:"diskpath"`   // path disk image
	ISOPath    string `yaml:"isopath"`    // path iso image
	NestedVirt string `yaml:"nvirt"`      // vmx (intel), smx (amd)
	Network    string `yaml:"network"`    // bridge | nat | none
	Graphics   string `yaml:"graphics"`   // graphic driver / mode: spice | vnc | none
	Sound      string `yaml:"sound"`      // ac97 | ich6 | ich9 (default)
	FileSystem string `yaml:"filesystem"` // virtiofs | 9p | none
	BootOrder  string `yaml:"bootorder"`  // stored as a string for backward compatibility
	Firmware   string `yaml:"firmware"`   // BIOS | EFI (not working yet)
}

// global defaults (are overwritten by every VM entry)
type Defaults struct {
	DiskPath string
	DiskSize int
}

// load yaml
func LoadAll(path string) (*FullConfig, error) {
	// read config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config %q: %w", path, err)
	}

	// Unmarshall into a temporary root struct containing both filepaths and oslist
	var raw struct {
		Filepaths struct {
			IsoPath string `yaml:"isopath"`
			XmlDir  string `yaml:"xmlpath"`
		} `yaml:"filepaths"`

		Defaults Defaults   `yaml:"defaults"`
		OSList   []VMConfig `yaml:"oslist"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	utils.ExpandEnvInStruct(&raw)

	// assemble result
	return &FullConfig{
		IsoPath:  raw.Filepaths.IsoPath,
		XmlDir:   raw.Filepaths.XmlDir,
		Defaults: raw.Defaults,
		OSList:   raw.OSList,
	}, nil
}
