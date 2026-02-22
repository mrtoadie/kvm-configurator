// kvmtools/rename.go
// last modified: Feb 22 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"configurator/internal/style"
	"configurator/internal/utils"
)

func RenameVM(r *bufio.Reader, oldName, xmlDir string) error {
	// Existence check (virsh dominfo)
	if _, err := exec.Command("virsh", "dominfo", oldName).CombinedOutput(); err != nil {
		return fmt.Errorf("VM %q nicht gefunden (virsh dominfo): %w", oldName, err)
	}

	// ask for new name
	newName, err := utils.Prompt(r, os.Stdout,
		style.Colourise(fmt.Sprintf("New name for VM %q: ", oldName), style.ColorYellow))
	if err != nil {
		return fmt.Errorf("Entry failed: %w", err)
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("New name cannot be empty")
	}
	if newName == oldName {
		return fmt.Errorf("New name is identical to the old one – nothing to do")
	}

	// virsh domrename
	renameCmd := exec.Command("virsh", "domrename", oldName, newName)
	if out, err := renameCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("virsh domrename failed: %w – %s", err, string(out))
	}
	style.Successf("Successfully renamed VM %s to %s", oldName, newName)

	// rename XML file (if existing) // BUGS!!
	oldXML := filepath.Join(xmlDir, oldName+".xml")
	newXML := filepath.Join(xmlDir, newName+".xml")
	if _, err := os.Stat(oldXML); err == nil {
		if err := os.Rename(oldXML, newXML); err != nil {
			style.RedError("XML file could not be renamed", oldXML, err)
		} else {
			style.Info("XML file renamed", fmt.Sprintf("%s → %s", oldXML, newXML))
		}
	} else {
		style.RedError("XML file not found (ignore)", oldXML, nil)
	}

	// rename disk
	if paths, err := GetDiskPathsViaVirsh(newName); err == nil && len(paths) > 0 {
		oldDisk := paths[0] // only system disk will be renamed (yet)
		dir := filepath.Dir(oldDisk)
		ext := filepath.Ext(oldDisk)
		newDisk := filepath.Join(dir, newName+ext)

		if err := os.Rename(oldDisk, newDisk); err != nil {
			style.RedError("Disk file could not be renamed", oldDisk, err)
		} else {
			style.Info("Disk file renamed", fmt.Sprintf("%s → %s", oldDisk, newDisk))
		}
	}

	return nil
}
