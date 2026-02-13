// internal/constants.go
// last modification: February 13 2026
package config

import (
	"os"
	"path/filepath"
)

const (
    CmdVirtInstall  = "virt-install"
    CmdVirsh        = "virsh"
    ConfigDir       = ".config/kvm-configurator"
    FileConfig      = "oslist.yaml"
)
//var PrereqCommands = []string{CmdVirtInstall, CmdVirsh}

func ConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// if cannot resolve $HOME, fall back to the current working directory
		// the subsequent Exists() check will surface the problem.
		return FileConfig
	}
	return filepath.Join(home, ConfigDir, FileConfig)
}