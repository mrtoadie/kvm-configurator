// internal/constants.go
// last modified: Feb 14 2026
package config

import (
	"os"
	"path/filepath"

	"fmt"
	"io"
)

const (
    CmdVirtInstall  = "virt-install"
    CmdVirsh        = "virsh"
    ConfigDir       = ".config/kvm-configurator"
    FileConfig      = "oslist.yaml"
		InstalledTemplate = "/usr/share/doc/kvm-configurator/oslist.yaml"
)

func ConfigFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// if cannot resolve $HOME, fall back to the current working directory
		// the subsequent Exists() check will surface the problem.
		return FileConfig
	}
	return filepath.Join(home, ConfigDir, FileConfig)
}

// copy default config to user home
func EnsureConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	targetDir := filepath.Join(home, ".config", "kvm-configurator")
	targetFile := filepath.Join(targetDir, "oslist.yaml")

	// does the file already exist? Then we're done.
	if info, err := os.Stat(targetFile); err == nil && !info.IsDir() {
		// file is there -nothing to do.
		return nil
	}

	// create target directory (if it does not already exist)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", targetDir, err)
	}

	// 3️⃣ Template öffnen
	src, err := os.Open(InstalledTemplate)
	if err != nil {
		return fmt.Errorf("could not open template %s: %w", InstalledTemplate, err)
	}
	defer src.Close()

	// create target file (rw‑r‑r‑r, i.e. 0644)
	dst, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", targetFile, err)
	}
	defer dst.Close()

	// copy
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("Copy failed: %w", err)
	}

	fmt.Printf("Configuration copied to %s.\n", targetFile)
	return nil
}