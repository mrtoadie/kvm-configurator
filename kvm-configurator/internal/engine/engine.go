// engine/engine.go
// last modification: Feb 08 2026
package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	// internal
	"configurator/internal/config"
	"configurator/internal/model"
	"configurator/internal/utils"
)

/*
	CreateVM receives a fully‑filled DomainConfig, the os‑variant string
	and the absolute path to the ISO file
*/
func CreateVM(cfg model.DomainConfig, variant, isoPath string, xmlDir string) error {
	// Check if the ISO file exists
	if _, err := os.Stat(isoPath); err != nil {
		//return fmt.Errorf("\x1b[31mISO not accessible: %w\x1b[0m", err)
		utils.RedError("ISO not accessible", isoPath, err)
	}
	// create Disk‑Argument
	diskArg, haveRealDisk := model.BuildDiskArg(cfg)

	// CPU‑Argument (Nested Virtualisation)
	cpuBase := "host-passthrough" // always included
	cpuArg := cpuBase
	if strings.TrimSpace(cfg.NestedVirt) != "" {
		cpuArg = fmt.Sprintf("%s,+%s", cpuBase, cfg.NestedVirt)
	}
	// disk
	if haveRealDisk {
		fmt.Println(utils.Colourise(
    fmt.Sprintf("Using custom disk: %s", diskArg),
    utils.ColorYellow,
		))
	} else {
		fmt.Println(utils.Colourise("\x1b[34mNo custom disk – passing '--disk none'\x1b[0m", utils.ColorYellow))
	}

	// Arguments for virt‑install
	args := []string{
		"--name", cfg.Name,
		"--memory", strconv.Itoa(cfg.MemMiB),
		"--vcpus", strconv.Itoa(cfg.VCPU),
		"--cpu", cpuArg,
		"--os-variant", variant,
		"--disk", diskArg,
		"--cdrom", cfg.ISOPath,
		"--boot", "hd,cdrom",
		//"--boot", cfg.BootOrder,
		"--graphics", cfg.Graphics,
		"--sound", cfg.Sound,
		"--filesystem", cfg.FileSystem,
		"--print-xml",
	}

	// Debug output
	//fmt.Print(args)
	
	// SimpleProgress
	spinner := utils.NewProgress("\x1b[34mRunning virt-install:")
	defer spinner.Stop()

	cmd := exec.Command(config.CmdVirtInstall, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
			utils.RedError("virt-install failed: %w – %s",">", err)
			return err
	}
		
	// Ensure that only one domain block is present
	raw := out.Bytes()
	xmlStr := string(raw)

	// Find the 'first' </domain> tag – discard everything after it
	firstEndIdx := strings.Index(xmlStr, "</domain>")
	if firstEndIdx == -1 {
		return fmt.Errorf(utils.Colourise("Failed to locate closing </domain> tag in virt-install output", utils.ColorRed))
	}
	// +len("</domain>") includes the tag itself
	cleanXMLStr := xmlStr[:firstEndIdx+len("</domain>")]
	cleanXML := []byte(cleanXMLStr)

	// xml path from config
	if xmlDir == "" {
		// fallback to current dir
		xmlDir = "."
	}
	xmlFileName := cfg.Name + ".xml"
	xmlFullPath := filepath.Join(xmlDir, xmlFileName)
	
	// Save XML
	if err := os.WriteFile(xmlFullPath, cleanXML, 0644); err != nil {
			//return fmt.Errorf("\x1b[31mcould not write XML: %w\x1b[0m", err)
			utils.RedError("Could not write XML", xmlFileName, err)
			//os.Exit(1)
	} else {
		abs, _ := filepath.Abs(xmlFullPath)
		utils.Successf("XML definition saved under: %s", abs)
	}

	// Define the new VM >> libvirt
	if err := exec.Command("virsh", "define", xmlFullPath).Run(); err != nil {
		//return fmt.Errorf("\x1b[31mvirsh define failed: %w\x1b[0m", err)
		utils.RedError("virsh define failed: %w", ">", err)
		return  err

	}
	//fmt.Println(ui.Colourise("\nVM successfully registered with libvirt/qemu (not yet started).", ui.Green))
	utils.Successf("VM successfully registered with libvirt/qemu (not yet started).")
	return nil
}
// EOF