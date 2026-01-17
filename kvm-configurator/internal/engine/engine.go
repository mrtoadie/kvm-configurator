// engine/engine.go
// last modification: January 17 2026
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
	"configurator/internal/ui"
)

/* --------------------
	CreateVM receives a fully‑filled DomainConfig, the os‑variant string
	and the absolute path to the ISO file.
-------------------- */
func CreateVM(cfg model.DomainConfig, variant, isoPath string, fp *config.FilePaths) error {
	// Check if the ISO file exists
	if _, err := os.Stat(isoPath); err != nil {
		return fmt.Errorf("\x1b[31mISO not accessible: %w\x1b[0m", err)
	}

	// create Disk‑Argument
	diskArg, haveRealDisk := model.BuildDiskArg(cfg)

	// CPU‑Argument (Nested Virtualisation)
	cpuBase := "host-passthrough" // always included
	cpuArg := cpuBase
	if strings.TrimSpace(cfg.NestedVirt) != "" {
		cpuArg = fmt.Sprintf("%s,+%s", cpuBase, cfg.NestedVirt)
	}

	// boot-argument (boot order)
	bootArg := ""
  if strings.TrimSpace(cfg.BootOrder) != "" {
    // virt‑install --boot cdrom,network
    bootArg = fmt.Sprintf("--boot %s", cfg.BootOrder)
  }

// -------------------------------------------------
// ❻ Boot‑Order aus cfg übernehmen, wenn angegeben
// -------------------------------------------------
//bootArg := "hd" // Standardfallback
/*
if strings.TrimSpace(cfg.BootOrder) != "" {
    // Nur erlaubte Keywords zulassen (einfacher Filter)
    allowed := map[string]bool{
        "cdrom": true, "hd": true, "network": true,
    }
    parts := strings.Split(cfg.BootOrder, ",")
    var clean []string
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if allowed[p] {
            clean = append(clean, p)
        }
    }
    if len(clean) > 0 {
        bootArg = strings.Join(clean, ",")
    }
}*/

	if haveRealDisk {
		fmt.Println(ui.Colourise(
    fmt.Sprintf("Using custom disk: %s", diskArg),
    ui.Yellow,
		))
	} else {
		fmt.Println(ui.Colourise("\x1b[34mNo custom disk – passing '--disk none'\x1b[0m", ui.Yellow))
	}

	// Arguments for virt‑install
	args := []string{
		"--name", cfg.Name,
		"--memory", strconv.Itoa(cfg.MemMiB),
		"--vcpus", strconv.Itoa(cfg.VCPU),
		"--cpu", cpuArg,
		"--os-variant", variant,
		"--disk", diskArg,
		"--cdrom", isoPath,
		//"--boot", "hd",
		"--boot", bootArg,
		"--print-xml",
	}
	// Debug output
	fmt.Print(args)
	// SimpelProgress
	stop := make(chan struct{})
	ui.SimpleProgress("\x1b[34mRunning virt-install:", stop)
	
	// run virt-install
	cmd := exec.Command("virt-install", args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("\x1b[31mvirt-install failed: %w – %s\x1b[0m", err, errOut.String())
	}
	// stopp progessbar
	close(stop)
	
	// Ensure that only one domain block is present.
	raw := out.Bytes()
	xmlStr := string(raw)

	// Find the 'first' </domain> tag – discard everything after it
	firstEndIdx := strings.Index(xmlStr, "</domain>")
	if firstEndIdx == -1 {
		return fmt.Errorf(ui.Colourise("Failed to locate closing </domain> tag in virt-install output", ui.Red))
	}
	// +len("</domain>") includes the tag itself
	cleanXMLStr := xmlStr[:firstEndIdx+len("</domain>")]
	cleanXML := []byte(cleanXMLStr)

	// xml path from config
	xmlDir := strings.TrimSpace(fp.Filepaths.XmlDir) 
	if xmlDir == "" {
		// fallback to current dir
		xmlDir = "."
	}
	xmlFileName := cfg.Name + ".xml"
	xmlFullPath := filepath.Join(xmlDir, xmlFileName)
	
	// Save XML
	if err := os.WriteFile(xmlFullPath, cleanXML, 0644); err != nil {
			return fmt.Errorf("\x1b[31mcould not write XML: %w\x1b[0m", err)
	}
	abs, _ := filepath.Abs(xmlFullPath)
	fmt.Printf("\n\x1b[32mXML definition saved under: %s\n\x1b[0m", abs)

	// Define the new VM >> libvirt
	if err := exec.Command("virsh", "define", xmlFullPath).Run(); err != nil {
		return fmt.Errorf("\x1b[31mvirsh define failed: %w\x1b[0m", err)
	}
	fmt.Println(ui.Colourise("\nVM successfully registered with libvirt/qemu (not yet started).", ui.Green))
	return nil
}
// EOF