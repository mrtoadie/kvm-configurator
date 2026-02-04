// internal/config/config.go
// last modification: Feb 04 2026
package config

import (
	"fmt"
	"os"
	"reflect"
	// external
	"gopkg.in/yaml.v3"
)

// VMConfig represents a single operating‑system or guest definition coming from the YAML file
type VMConfig struct {
	Name        string 		`yaml:"name"`				// display name "Arch Linux"
	ID          string 		`yaml:"id"`					// identifier, e.g. "archlinux"
	CPU         int    		`yaml:"cpu"`				// number of vCPUs
	RAM         int    		`yaml:"ram"`				// RAM in MiB
	DiskSize    int    		`yaml:"disksize"`		// disk size in GiB
	DiskPath    string 		`yaml:"diskpath"`		// path disk image
	ISOPath     string 		`yaml:"isopath"`	// path iso image
	NestedVirt  string 		`yaml:"nvirt"`			// vmx (intel), smx (amd)
	Network     string 		`yaml:"network"`   	// bridge | nat | none
	Graphics    string 		`yaml:"graphics"`  	// graphic driver / mode: spice | vnc | none
	Sound       string 		`yaml:"sound"`			// ac97 | ich6 | ich9 (default)
	FileSystem  string 		`yaml:"filesystem"`	// virtiofs | 9p | none
	BootOrder   string 		`yaml:"bootorder"` 	// stored as a string for backward compatibility
	Firmware		string 		`yaml:"firmware"` 	// BIOS | EFI (not working yet)
}

type Defaults struct {
	DiskPath string
	DiskSize int
}
/* ---------------------------------------------------------
   FilePaths – only the “filepaths” block from the config file
--------------------------------------------------------- */
type FilePaths struct {
	Filepaths struct {
		IsoPath string `yaml:"isopath"`
		XmlDir   string `yaml:"xmlpath"`
	} `yaml:"filepaths"`
}
// OSRoot mirrors the top-level structure of the OS list YAML file.
type OSRoot struct {
	Defaults Defaults  `yaml:"defaults"`
	OSList   []VMConfig `yaml:"oslist"`
}
// AdvancedFeaturesRoot mirrors the advanced‑features YAML file (currently unused elsewhere)
type AdvancedFeaturesRoot struct {
	AdvancedFeatures struct {
		StartInit bool `yaml:"start_init"`
	} `yaml:"advanced_features"`
	AdvFeatures []AdvancedFeaturesRoot `yaml:"oslist"`
}

/* ---------------------------------------------------------
   Helper: expand all string fields in a struct using os.ExpandEnv
   (recursively handles nested structs, slices and maps)
--------------------------------------------------------- */
func expandStrings(v interface{}) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return // nothing to do
	}
	val = val.Elem()
	expandValue(val)
}

// ExpandEnvInStruct recursively expands environment variables in all string fields
// of structs, slices, maps, and their nested contents using os.ExpandEnv.
func ExpandEnvInStruct(v any) {
	if v == nil {
		return
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	expandValue(val.Elem())
}

func expandValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			if f.CanSet() {
				expandValue(f)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			expandValue(val.Index(i))
		}
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			if v.Kind() == reflect.String {
				newStr := os.ExpandEnv(v.String())
				val.SetMapIndex(k, reflect.ValueOf(newStr))
			} else {
				expandValue(v)
			}
			// Keys are usually static strings, we leave them untouched.
			_ = k
		}
	case reflect.String:
		val.SetString(os.ExpandEnv(val.String()))
	}
}

/* ---------------------------------------------------------
   LoadOSList – reads the OS list YAML file and expands env vars
--------------------------------------------------------- */
func LoadOSList(path string) ([]VMConfig, Defaults, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, Defaults{}, fmt.Errorf("read OS-list file %q: %w", path, err)
	}

	var root OSRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, Defaults{}, fmt.Errorf("parse OS-list file %q: %w", path, err)
	}

	ExpandEnvInStruct(&root)

	return root.OSList, root.Defaults, nil
}

/* ---------------------------------------------------------
   LoadFilePaths – reads the filepaths block and expands env vars
--------------------------------------------------------- */
func LoadFilePaths(path string) (*FilePaths, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fp FilePaths
	if err = yaml.Unmarshal(data, &fp); err != nil {
		return nil, err
	}
	// Expand ${HOME}, ${USER}, … in the two path strings.
	expandStrings(&fp)
	return &fp, nil
}

/* ---------------------------------------------------------
   ResolveWorkDir – returns the directory that should be scanned
   (prefers the configured InputDir, falls back to the current working directory)
--------------------------------------------------------- */
func ResolveWorkDir(fp *FilePaths) (string, error) {
	if fp.Filepaths.IsoPath != "" {
		return fp.Filepaths.IsoPath, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	return cwd, nil
}
// EOF