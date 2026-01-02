package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
)

type DomainConfig struct {
	Name    string
	MemMiB  int
	VCPU    int
	Disk    string
	Network string
}

// mapping os variants
var osVariant = map[string]string{
	"Arch Linux": "archlinux",
	"Ubuntu":     "ubuntu20.04",
}

// chooseVariant liefert den virt‑install‑Identifier für eine bekannte Distro.
func chooseVariant(distro string) (string, error) {
	if v, ok := osVariant[distro]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown distro %q", distro)
}

// ask stellt eine Frage und liest die Eingabe von stdin.
func ask(prompt string, r *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	text, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// buildVirtInstallCmd erzeugt die Argumentliste für virt‑install.
func buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant string, diskSizeGB int) []string {
	return []string{
		"--name", name,
		"--memory", ram,
		"--vcpus", cpus,
		"--os-variant", variant,
		"--disk", fmt.Sprintf("path=%s,size=%d,format=qcow2", diskPath, diskSizeGB),
		"--cdrom", iso,
		"--boot", "hd",
		// ... more options ...
	}
}

// writeXML speichert das XML in eine Datei.
func writeXML(xmlData []byte, filename string) error {
	return os.WriteFile(filename, xmlData, 0644)
}

// defineVM führt `virsh define <xml>` aus.
func defineVM(xmlPath string) error {
	cmd := exec.Command("virsh", "define", xmlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Render Tab‑Writer
func renderTable(pairs map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for k, v := range pairs {
		fmt.Fprintf(w, "%s:\t%s\n", k, v)
	}
	w.Flush()
}

/* -------------------------------------------------
   Prompt‑Formular (interaktive Bearbeitung)
   ------------------------------------------------- */
func promptForm(cfg *DomainConfig) {
	in := bufio.NewReader(os.Stdin)

	// Tab‑Writer Formular
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	for {
		fmt.Fprintln(w, "\n=== VM-Config ===\t")
		fmt.Fprintf(w, "[1] Name:\t%s\t[default]\n", cfg.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t[default]\n", cfg.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\t[default]\n", cfg.VCPU)
		fmt.Fprintf(w, "[4] Disk-Path:\t%s\t[Enter path for virtual hdd]\n", cfg.Disk)
		fmt.Fprintf(w, "[5] Network:\t%s\t[default]\n", cfg.Network)
		w.Flush()

		fmt.Print("\nSelect or enter to continue: ")
		fieldRaw, _ := in.ReadString('\n')
		field := strings.TrimSpace(strings.ToLower(fieldRaw))
		if field == "" {
			break
		}

		switch field {
		case "1":
			fmt.Print(">> New Name: ")
			val, _ := in.ReadString('\n')
			cfg.Name = strings.TrimSpace(val)

		case "2":
			fmt.Print("RAM in MiB: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.MemMiB = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "3":
			fmt.Print("vCPU-Anzahl: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.VCPU = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "4":
			fmt.Print("Disk-Path (leer = keine Disk): ")
			val, _ := in.ReadString('\n')
			cfg.Disk = strings.TrimSpace(val)

		case "5":
			fmt.Print("Netzwerk (kommagetrennt): ")
			val, _ := in.ReadString('\n')
			cfg.Network = strings.TrimSpace(val)

		default:
			fmt.Println("Unbekanntes Feld – bitte Name, RAM, vCPU, Disk oder Network eingeben.")
		}
	}
}

/* -------------------------------------------------
   VM‑Erstellung aus einer DomainConfig
   ------------------------------------------------- */
func createVMFromConfig(cfg DomainConfig) error {
	r := bufio.NewReader(os.Stdin)

	// ---------- 1️⃣ Werte, die nicht im Config stehen ----------
	// Disk‑Größe (GB)
	diskSizeStr, err := ask("Disk‑Größe in GB (z. B. 20): ", r)
	if err != nil {
		return err
	}
	diskSizeGB, err := strconv.Atoi(diskSizeStr)
	if err != nil {
		return fmt.Errorf("ungültige Disk‑Größe: %w", err)
	}

	// Pfad zur ISO
	iso, err := ask("Pfad zur Installations‑ISO: ", r)
	if err != nil {
		return err
	}

	// Distro (Ubuntu / Arch Linux)
	distro, err := ask("Distro (Ubuntu / Arch Linux): ", r)
	if err != nil {
		return err
	}
	variant, err := chooseVariant(distro)
	if err != nil {
		return err
	}

	// virt‑install aufrufen ----------
	name := cfg.Name
	ram := strconv.Itoa(cfg.MemMiB) // virt‑install erwartet RAM als String (MB)
	cpus := strconv.Itoa(cfg.VCPU)

	// Zielpfad für das Disk‑Image
	diskPath := fmt.Sprintf("/run/media/toadie/vm/QEMU/%s.qcow2", name)
iso = "/run/media/toadie/data/ISOs/ubuntu-24.04.3-desktop-amd64.iso"
	cmdArgs := buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant, diskSizeGB)
	virtCmd := exec.Command("virt-install", cmdArgs...)
	var out bytes.Buffer
	//virtCmd.Stdout = &out
	//virtCmd.Stderr = os.Stderr

	if err := virtCmd.Run(); err != nil {
		return fmt.Errorf("virt-install (XML‑Erzeugung) fehlgeschlagen: %w", err)
	}
	rawXML := out.Bytes()

	// save xml
	xmlFile := fmt.Sprintf("%s.xml", name)
	if err := writeXML(rawXML, xmlFile); err != nil {
		return fmt.Errorf("konnte XML nicht speichern: %w", err)
	}
	absPath, _ := filepath.Abs(xmlFile)
	fmt.Printf("XML‑Definition gespeichert unter: %s\n", absPath)

	// define vm
	if err := defineVM(xmlFile); err != nil {
		return fmt.Errorf("virsh define fehlgeschlagen: %w", err)
	}
	fmt.Println("VM erfolgreich bei libvirt/qemu registriert (noch nicht gestartet).")
	return nil
}
// MAIN
func main() {
	// default config
	cfg := DomainConfig{
		Name:    "my‑guest",
		MemMiB:  1024,
		VCPU:    2,
		Disk:    "",
		Network: "default",
	}

	// interactive input
	promptForm(&cfg)

	// config overview
	fmt.Println("\n=== Endgültige Konfiguration ===")
	renderTable(map[string]string{
		"Name":      cfg.Name,
		"RAM (MiB)": strconv.Itoa(cfg.MemMiB),
		"vCPU":      strconv.Itoa(cfg.VCPU),
		"Disk‑Pfad": cfg.Disk,
		"Netzwerk":  cfg.Network,
	})

	// VM anlegen (virsh/virt‑install)
	if err := createVMFromConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler beim Anlegen der VM:", err)
		os.Exit(1)
	}
}