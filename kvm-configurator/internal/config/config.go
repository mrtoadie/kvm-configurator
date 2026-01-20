// config/config.go
// last modification: January 18 2026
package config

import (
	"fmt"
	"os"
	// external
	"gopkg.in/yaml.v3"
)

// ---------- OS‑Liste ----------
type Distro struct {
	Name       	string `yaml:"name"`
	ID         	string `yaml:"id"`
	CPU        	int    `yaml:"cpu"`
	RAM        	int    `yaml:"ram"`
	DiskSize   	int    `yaml:"disksize"`
	DiskPath   	string `yaml:"diskpath"`
	ISOPath   	string `yaml:"input_dir"`
	NestedVirt 	string `yaml:"nvirt"`
	Network    	string `yaml:"network"`   // bridge | nat | none
  Graphics   	string `yaml:"graphics"`  // spice | vnc | none
	Sound			 	string `yaml:"sound"`
	FileSystem 	string `yaml:"filesystem"`
  //BootOrder  []int  `yaml:"boot_order,omitempty"`// [1,2] (disk, cdrom)
	BootOrder		string `yaml:"bootorder"`
}

type OSRoot struct {
	Defaults struct {
		DiskPath string `yaml:"diskpath"`
		DiskSize int    `yaml:"disksize"`
	} `yaml:"defaults"`
	OSList []Distro `yaml:"oslist"`
}

type AdvancedFeaturesRoot struct {
	AdvancedFeatures struct {
		StartInit bool `yaml:"start_init"`
	} `yaml:"advanced_features"`
	AdvFeatures []AdvancedFeaturesRoot `yaml:"oslist"`
}

/* --------------------
	LoadOSList
-------------------- */
func LoadOSList(path string) (list []Distro, defaults struct {
	DiskPath string
	DiskSize int
}, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, defaults, err
	}
	var root OSRoot
	if err = yaml.Unmarshal(b, &root); err != nil {
		return nil, defaults, err
	}
	defaults.DiskPath = root.Defaults.DiskPath
	defaults.DiskSize = root.Defaults.DiskSize
	return root.OSList, defaults, nil
}

/* --------------------
	Path config
-------------------- */
type FilePaths struct {
	Filepaths struct {
		InputDir string `yaml:"input_dir"`
		XmlDir   string `yaml:"xml_dir"` 
	} `yaml:"filepaths"`
}

/* --------------------
	LoadFilePaths – read only block „filepaths“
-------------------- */
func LoadFilePaths(path string) (*FilePaths, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fp FilePaths
	if err = yaml.Unmarshal(data, &fp); err != nil {
		return nil, err
	}
	return &fp, nil
}

/* --------------------
	ResolveWorkDir – returns the directory to be scanned
-------------------- */
func ResolveWorkDir(fp *FilePaths) (string, error) {
    if fp.Filepaths.InputDir != "" {
        return fp.Filepaths.InputDir, nil
    }
    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("cannot determine working directory: %w", err)
    }
    return cwd, nil
}
// EOF