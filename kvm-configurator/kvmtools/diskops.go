// kvmtools/diskops.go
// last modification: Feb 22 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	// internal
	"configurator/internal/utils"
)

func getRealDiskPath(vmName string) (string, error) {
	paths, err := GetDiskPathsViaVirsh(vmName)
	if err != nil {
		return "", fmt.Errorf("virsh query failed for VM %s: %w", vmName, err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("no disk entries found for VM %s (virsh returns empty list)", vmName)
	}
	return paths[0], nil // first entry = system disk
}

// disk resize
func ResizeDisk(r *bufio.Reader, vmName string) error {
	imgPath, err := getRealDiskPath(vmName)
	if err != nil {
		return err
	}

	sizeStr, _ := utils.Prompt(r, os.Stdout,
		utils.Colourise("New size (e.g. 10 to add 10 GiB to the disk): ", utils.ColorYellow))
	newSize, err := strconv.Atoi(sizeStr)
	if err != nil || newSize <= 0 {
		return fmt.Errorf("please enter a positive integer (e.g. 10 to add 10 GiB to the disk)")
	}
	
	/// test output
	fmt.Println("\n" + imgPath + vmName + "\n")
	//

	spinner := utils.SpinnerProgress("Resize is running …")
	defer spinner.Stop()

	cmd := exec.Command("qemu-img", "resize", imgPath, fmt.Sprintf("+%dG", newSize))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize failed: %v – %s", err, string(out))
	}
	utils.Successf("Disk %s increased by %dGiB", filepath.Base(imgPath), newSize)
	return nil
}

// convert
func ConvertDisk(r *bufio.Reader, vmName string) error {
	srcPath, err := getRealDiskPath(vmName)
	if err != nil {
		return err
	}

	fmt.Println(utils.Colourise("\nTarget formats:", utils.ColorBlue))
	fmt.Println("[1] qcow2   (Standard, compressed)")
	fmt.Println("[2] raw     (uncompressed, fast)")
	fmt.Println("[3] vdi     (VirtualBox-Compatible)")

	choice, _ := utils.Prompt(r, os.Stdout,
		utils.Colourise("Select format: ", utils.ColorYellow))

	var tgtFmt string
	switch choice {
	case "1":
		tgtFmt = "qcow2"
	case "2":
		tgtFmt = "raw"
	case "3":
		tgtFmt = "vdi"
	default:
		return fmt.Errorf("unknown format")
	}

	ext := "." + tgtFmt
	newPath := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ext

	spinner := utils.SpinnerProgress("Conversion is underway …")
	defer spinner.Stop()

	cmd := exec.Command("qemu-img", "convert", "-O", tgtFmt, srcPath, newPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("conversion failed: %v – %s", err, string(out))
	}
	utils.Successf("Converted disk %s to %s", filepath.Base(srcPath), tgtFmt)

	// update XML entry (e.g. boot disk from qcow to vdi)
	updateXMLPath(vmName, newPath, srcPath)

	return nil
}

// repair
func RepairDisk(r *bufio.Reader, vmName string) error {
	imgPath, err := getRealDiskPath(vmName)
	if err != nil {
		return err
	}

	// check
	checkCmd := exec.Command("qemu-img", "check", imgPath)
	out, err := checkCmd.CombinedOutput()
	if err == nil {
		fmt.Println(utils.Colourise("\nDisk is intact - no intervention required.", utils.ColorGreen))
		fmt.Printf("%s\n", string(out))
		return nil
	}

	// repair (amend is the easiest way)
	fmt.Println(utils.Colourise("\nInconsistency detected -attempt repair …", utils.ColorRed))
	spinner := utils.SpinnerProgress("Repair is running …")
	defer spinner.Stop()

	repairCmd := exec.Command("qemu-img", "amend", "-f", "qcow2", imgPath)
	repOut, repErr := repairCmd.CombinedOutput()
	if repErr != nil {
		return fmt.Errorf("Repair failed: %v – %s", repErr, string(repOut))
	}
	utils.Successf("Disk %s repaired", filepath.Base(imgPath))
	fmt.Printf("%s\n", string(repOut))
	return nil
}

// helper function: Adjust XML entry (only for Convert)
func updateXMLPath(vmName, newPath, oldPath string) {
	// assuming the current working directory is the
	// project root and the XML directory is located there
	possibleDirs := []string{
		"./xml",
		".", // fallback: aktuelle Directory
	}
	var xmlPath string
	for _, d := range possibleDirs {
		tmp := filepath.Join(d, vmName+".xml")
		if _, err := os.Stat(tmp); err == nil {
			xmlPath = tmp
			break
		}
	}
	if xmlPath == "" {
		utils.RedError("XML update failed – file not found", vmName, nil)
		return
	}

	data, err := os.ReadFile(xmlPath)
	if err != nil {
		utils.RedError("XML update failed (read)", xmlPath, err)
		return
	}
	updated := strings.ReplaceAll(string(data), oldPath, newPath)
	if err := os.WriteFile(xmlPath, []byte(updated), 0644); err != nil {
		utils.RedError("XML update failed (write)", xmlPath, err)
	}
}

// Sub-menu that call up from "vmmenu.go"
func DiskOpsMenu(r *bufio.Reader, vmName string) error {
	for {
		fmt.Println(utils.BoxCenter(55,
			[]string{"=== DISK-OPERATIONS FOR " + vmName + " ==="}))
		fmt.Println(utils.Box(55, []string{
			"[1] Resize",
			"[2] Convert (change file format)",
			"[3] Repair (check image)",
			"[0] Back",
		}))

		choice, _ := utils.Prompt(r, os.Stdout, 
			utils.Colourise("\nSelection: ", utils.ColorYellow))

		switch choice {
		case "1":
			return ResizeDisk(r, vmName)
		case "2":
			return ConvertDisk(r, vmName)
		case "3":
			return RepairDisk(r, vmName)
		case "0", "":
			return nil
		default:
			fmt.Println(utils.Colourise("Invalid selection!", utils.ColorRed))
		}
	}
}
// EOF
