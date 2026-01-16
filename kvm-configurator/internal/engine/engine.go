// engine/engine.go
package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	// internal
	"configurator/internal/model"
	"configurator/internal/config"
)

func startSimpleProgress(msg string, stopChan <-chan struct{}) {
	go func() {
		chars := []rune{'⣾','⣽','⣻','⢿','⡿','⣟','⣯','⣷'}
		//chars := []string{"⣾","⣽","⣻","⢿","⡿","⣟","⣯","⣷"}
		//chars := []rune{'▁','▂','▃','▄','▅','▆','▇','█'}
		i := 0
		for {
			select {
			case <-stopChan:
				fmt.Print("\r")               // Zeile zurücksetzen
				return
			default:
				fmt.Printf("\r%s %c ", msg, chars[i%len(chars)])
				time.Sleep(100 * time.Millisecond)
				i++
			}
	}
}()
}

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
	cpuBase := "host-passthrough"
	cpuArg := cpuBase
	if strings.TrimSpace(cfg.NestedVirt) != "" {
		cpuArg = fmt.Sprintf("%s,+%s", cpuBase, cfg.NestedVirt)
	}

	// Calling virt‑install
	args := []string{
		"--name", cfg.Name,
		"--memory", strconv.Itoa(cfg.MemMiB),
		"--vcpus", strconv.Itoa(cfg.VCPU),
		"--cpu", cpuArg,
		"--os-variant", variant,
		"--disk", diskArg,
		"--cdrom", isoPath,
		"--boot", "hd",
		"--print-xml",
	}
	if haveRealDisk {
		fmt.Println("Using custom disk:", diskArg)
	} else {
		fmt.Println("\x1b[34mNo custom disk – passing '--disk none'\x1b[0m")
	}

	////////////////
	stop := make(chan struct{})
	startSimpleProgress("\x1b[34mRunning virt-install:", stop)
	//////////////
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
		return fmt.Errorf("\x1b[31mFailed to locate closing </domain> tag in virt‑install output\x1b[0m")
	}
	// +len("</domain>") includes the tag itself
	cleanXMLStr := xmlStr[:firstEndIdx+len("</domain>")]
	cleanXML := []byte(cleanXMLStr)

	// Short Preview
	/*
	fmt.Printf("\nXML created (%d Bytes).\n", len(cleanXML))
	if len(cleanXML) > 0 {
		preview := 200
		if len(cleanXML) < preview {
			preview = len(cleanXML)
		}
		fmt.Printf("First lines of the XML:\n%s\n", string(cleanXML[:preview]))
	}
*/
/*
	// Save XML
	xmlFile := cfg.Name + ".xml"
	if err := os.WriteFile(xmlFile, cleanXML, 0644); err != nil {
		return fmt.Errorf("\x1b[31mcould not write XML: %w\x1b[0m", err)
	}
	abs, _ := filepath.Abs(xmlFile)
	fmt.Printf("\x1b[32mXML definition saved under: %s\n\x1b[0m", abs)
*/

	// ---------- Neuer Teil: Pfad aus der Config -----------------------
	xmlDir := strings.TrimSpace(fp.Filepaths.XmlDir) // sollte nie leer sein
	if xmlDir == "" {
		// Sicherheits‑Fallback: aktuelles Arbeitsverzeichnis
		xmlDir = "."
	}
	xmlFileName := cfg.Name + ".xml"
	xmlFullPath := filepath.Join(xmlDir, xmlFileName)
	// -------------------------------------------------------------------
/*
	// Define the new VM >> libvirt
	if err := exec.Command("virsh", "define", xmlFile).Run(); err != nil {
		return fmt.Errorf("\x1b[31mvirsh define failed: %w\x1b[0m", err)
	}
	fmt.Println("\x1b[32mVM successfully registered with libvirt/qemu (not yet started).\x1b[0m")
	return nil
	*/

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
	fmt.Println("\n\x1b[32mVM successfully registered with libvirt/qemu (not yet started).\x1b[0m")
	return nil
	
}