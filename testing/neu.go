package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Mapping distro
var osVariant = map[string]string{
	"Arch Linux": "archlinux",
	"Ubuntu":     "ubuntu20.04",
}

// chooseVariant returns the virt‑install identifier for a distro.
func chooseVariant(distro string) (string, error) {
	if v, ok := osVariant[distro]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown distro %q", distro)
}

// ask prompts the user and reads a line from stdin.
func ask(prompt string, r *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	text, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// build VM
func buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant string, diskSizeGB int) []string {
	return []string{
		"--name", name,
		"--memory", ram,
		"--vcpus", cpus,
		"--os-variant", variant,
		"--disk", fmt.Sprintf("path=%s,size=%d,format=qcow2", diskPath, diskSizeGB),
		"--cdrom", iso,
		"--boot", "hd",
		// XML only, no automatic start.
		"--print-xml",
		// Optional
		// "--graphics", "none",
	}
}

// writeXML saves the XML definition to a file.
func writeXML(xmlData []byte, filename string) error {
	return os.WriteFile(filename, xmlData, 0644)
}

// defineVM registers the VM with libvirt/qemu using virsh.
func defineVM(xmlPath string) error {
	cmd := exec.Command("virsh", "define", xmlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// createVM 
func createVM() error {
	r := bufio.NewReader(os.Stdin)

	// ----- 1️⃣ Collect parameters from the user -----
	name, err := ask("VM-Name: ", r)
	if err != nil {
		return err
	}
	ram, err := ask("RAM in MB (z. B. 2048): ", r)
	if err != nil {
		return err
	}
	cpus, err := ask("Anzahl vCPUs (z. B. 2): ", r)
	if err != nil {
		return err
	}
	diskSizeStr, err := ask("Disk‑Größe in GB (z. B. 20): ", r)
	if err != nil {
		return err
	}
	diskSizeGB, err := strconv.Atoi(diskSizeStr)
	if err != nil {
		return fmt.Errorf("ungültige Disk‑Größe: %w", err)
	}
	iso, err := ask("Pfad zur Installations‑ISO: ", r)
	if err != nil {
		return err
	}
	distro, err := ask("Distro (Ubuntu / Arch Linux): ", r)
	if err != nil {
		return err
	}
	variant, err := chooseVariant(distro)
	if err != nil {
		return err
	}

	// Pfad, wo das Disk‑Image liegen soll.
	diskPath := fmt.Sprintf("/run/media/toadie/vm/QEMU/%s.qcow2", name)

	// ----- 2️⃣ Generate XML via virt‑install -----
	cmdArgs := buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant, diskSizeGB)
	virtCmd := exec.Command("virt-install", cmdArgs...)
	var out bytes.Buffer
	virtCmd.Stdout = &out
	virtCmd.Stderr = os.Stderr

	if err := virtCmd.Run(); err != nil {
		return fmt.Errorf("virt-install (XML‑Erzeugung) fehlgeschlagen: %w", err)
	}
	rawXML := out.Bytes()

	// ----- 3️⃣ Schön formatieren (optional) -----
	var prettyXML []byte
	prettyXML, err = xml.MarshalIndent(struct {
		XMLName xml.Name `xml:"domain"`
		Content []byte   `xml:",innerxml"`
	}{
		Content: rawXML,
	}, "", "  ")
	if err != nil {
		// Wenn MarshalIndent fehlschlägt, nutzen wir einfach das rohe XML.
		prettyXML = rawXML
	} else {
		// xml.MarshalIndent wraps everything in <domain>…</domain>,
		// das ist nicht nötig – wir entfernen die Hülle wieder.
		prettyXML = rawXML // keep original, because the wrapper would be wrong
	}

	// ----- 4️⃣ Save XML to file -----
	xmlFile := fmt.Sprintf("%s.xml", name)
	if err := writeXML(prettyXML, xmlFile); err != nil {
		return fmt.Errorf("konnte XML nicht speichern: %w", err)
	}
	absPath, _ := filepath.Abs(xmlFile) // Ignoriere den Fehler – Pfad ist meist korrekt
	fmt.Printf("XML‑Definition gespeichert unter: %s\n", absPath)

	// ----- 5️⃣ Register (define) the VM with libvirt/qemu -----
	if err := defineVM(xmlFile); err != nil {
		return fmt.Errorf("virsh define fehlgeschlagen: %w", err)
	}
	fmt.Println("VM erfolgreich bei libvirt/qemu registriert (noch nicht gestartet).")
	return nil
}

func main() {
	if err := createVM(); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler:", err)
		os.Exit(1)
	}
}