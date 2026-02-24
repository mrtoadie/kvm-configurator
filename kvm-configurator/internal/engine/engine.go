// engine/engine.go
// last modified: Feb 22 2026
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
	"configurator/internal/style"
)

/*
CreateVM receives a fully‑filled DomainConfig, the os‑variant string
and the absolute path to the ISO file
*/
func CreateVM(cfg model.DomainConfig, variant, isoPath, xmlDir string) error {
	// build disk arguments (multiple!)
	diskArgs := model.BuildDiskArgs(cfg.Disks, cfg.Name)

	// CPU arguments
	cpuBase := "host-passthrough"
	cpuArg := cpuBase
	if strings.TrimSpace(cfg.NestedVirt) != "" {
		cpuArg = fmt.Sprintf("%s,+%s", cpuBase, cfg.NestedVirt)
	}

	// base arguments
	args := []string{
		"--name", cfg.Name,
		"--memory", strconv.Itoa(cfg.MemMiB),
		"--vcpus", strconv.Itoa(cfg.VCPU),
		"--cpu", cpuArg,
		"--os-variant", variant,
		"--cdrom", cfg.ISOPath,
		//"--boot", "hd,cdrom",
		"--boot", cfg.BootOrder,
		//"--boot", "network,cdrom,hd",
		"--graphics", cfg.Graphics,
		"--sound", cfg.Sound,
		"--filesystem", cfg.FileSystem,
		"--print-xml",
		//"--print-step",
	}

	// append all disk arguments (each “--disk” + the argument)
	for _, da := range diskArgs {
		args = append(args, "--disk", da)
	}

	// progress-spinner
	spinner := style.SpinnerProgress("\x1b[34mCreation of the VM " + cfg.Name + " is in progress")
	defer spinner.Stop()

	cmd := exec.Command(config.CmdVirtInstall, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		style.RedError("virt-install failed: ", ">", err)
		return err
	}

	// Ensure that only one domain block is present
	raw := out.Bytes()
	xmlStr := string(raw)

	// Find the 'first' </domain> tag – discard everything after it
	firstEndIdx := strings.Index(xmlStr, "</domain>")
	if firstEndIdx == -1 {
		return fmt.Errorf(style.Colourise("Failed to locate closing </domain> tag in virt-install output", style.ColRed))
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

	// save XML
	if err := os.WriteFile(xmlFullPath, cleanXML, 0644); err != nil {
		//return fmt.Errorf("\x1b[31mcould not write XML: %w\x1b[0m", err)
		style.RedError("Could not write XML", xmlFileName, err)
		//os.Exit(1)
	} else {
		abs, _ := filepath.Abs(xmlFullPath)
		if err != nil {
			abs = xmlFullPath // fallback
		}
		style.Successf("\n\nXML definition saved under: %s", abs)
	}

	// define the new VM >> libvirt
	if err := exec.Command("virsh", "define", xmlFullPath).Run(); err != nil {
		//return fmt.Errorf("\x1b[31mvirsh define failed: %w\x1b[0m", err)
		style.RedError("virsh define failed: %w", ">", err)
		return err

	}
	style.Successf("VM successfully registered with libvirt/qemu (not yet started).")
	return nil
}
