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
	ISO 		string
}

// mapping os variants
var osVariant = map[string]string{
	"Arch Linux": "archlinux",
	"Ubuntu":     "ubuntu20.04",
}

// chooseVariant checks if the distro is known for virt-install identifier
func chooseVariant(distro string) (string, error) {
	if v, ok := osVariant[distro]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown distro %q", distro)
}

// ask asks a question and reads the input from stdin
func ask(prompt string, r *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	text, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// buildVirtInstallCmd generates the argument list for virt‑install
func buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant string, diskSizeGB int) []string {
	return []string{
		"--name", name,
		"--memory", ram,
		"--vcpus", cpus,
		"--os-variant", variant,
		"--disk", fmt.Sprintf("path=%s,size=%d,format=qcow2", diskPath, diskSizeGB),
		"--cdrom", iso,
		"--boot", "hd",
		"--print-xml",
		// ... more options ...
	}
}

// writeXML saves XML file
func writeXML(xmlData []byte, filename string) error {
	return os.WriteFile(filename, xmlData, 0644)
}

// defineVM runs "virsh define <xml>"
func defineVM(xmlPath string) error {
	cmd := exec.Command("virsh", "define", xmlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// renderTable prints a nice tab‑writer table
func renderTable(pairs map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for k, v := range pairs {
		fmt.Fprintf(w, "%s:\t%s\n", k, v)
	}
	w.Flush()
}

// Prompt form (interactive editing)
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

		fmt.Print("\nSelect or press Enter to continue: ")
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
				fmt.Println("Invalid number – value remains unchanged.")
			}

		case "3":
			fmt.Print("vCPU-Anzahl: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.VCPU = v
			} else {
				fmt.Println("Invalid number – value remains unchanged.")
			}

		case "4":
			fmt.Print("Disk path (empty = no disk): ")
			val, _ := in.ReadString('\n')
			cfg.Disk = strings.TrimSpace(val)

		case "5":
			fmt.Print("Network (comma‑separated): ")
			val, _ := in.ReadString('\n')
			cfg.Network = strings.TrimSpace(val)

		default:
			fmt.Println("Invalid input!")
		}
	}
}

// createVMFromConfig builds the XML via virt‑install and registers it with libvirt
func createVMFromConfig(cfg DomainConfig) error {
	r := bufio.NewReader(os.Stdin)

	diskSizeStr, err := ask("Disk size in GB (e.g., 20): ", r)
	if err != nil {
		return err
	}
	diskSizeGB, err := strconv.Atoi(diskSizeStr)
	if err != nil {
		return fmt.Errorf("invalid disk size: %w", err)
	}

	iso, err := ask("Path to the installation ISO: ", r)
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

	// call virt‑install
	name := cfg.Name
	ram := strconv.Itoa(cfg.MemMiB)
	cpus := strconv.Itoa(cfg.VCPU)

	// default virtual‑disk path
	diskPath := fmt.Sprintf("/run/media/toadie/vm/QEMU/%s.qcow2", name)

	// (optional) you can overwrite the iso path entered above – here we keep the user input
	//iso = "/run/media/toadie/data/ISOs/ubuntu-24.04.3-desktop-amd64.iso"

	cmdArgs := buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant, diskSizeGB)
	virtCmd := exec.Command("virt-install", cmdArgs...)

	// Capture stdout (the XML) and stderr (error messages)
	var outBuf, errBuf bytes.Buffer
	virtCmd.Stdout = &outBuf
	virtCmd.Stderr = &errBuf

	if err := virtCmd.Run(); err != nil {
		return fmt.Errorf("virt‑install (XML‑Erzeugung) fehlgeschlagen: %w – %s", err, errBuf.String())
	}

	rawXML := outBuf.Bytes()

	// optional: quick sanity check – print size / snippet
	fmt.Printf("XML erzeugt (%d Bytes).\n", len(rawXML))
	if len(rawXML) > 0 {
		fmt.Printf("Erster Teil des XML:\n%s\n", string(rawXML[:200]))
	}

	// --- write XML to file ----------------------------------------------------
	xmlFile := fmt.Sprintf("%s.xml", name)
	if err := writeXML(rawXML, xmlFile); err != nil {
		return fmt.Errorf("konnte XML nicht speichern: %w", err)
	}
	absPath, _ := filepath.Abs(xmlFile)
	fmt.Printf("XML‑Definition gespeichert unter: %s\n", absPath)

	// --- register VM with libvirt ---------------------------------------------
	if err := defineVM(xmlFile); err != nil {
		return fmt.Errorf("virsh define fehlgeschlagen: %w", err)
	}
	fmt.Println("VM erfolgreich bei libvirt/qemu registriert (noch nicht gestartet).")
	return nil
}

// ---------------------------------------------------------------------------
// MAIN
// ---------------------------------------------------------------------------
func main() {
	// default configuration
	cfg := DomainConfig{
		Name:    "new-machine",
		MemMiB:  1024,
		VCPU:    2,
		Disk:    "",
		Network: "default",
		ISO:		 "test",
	}

	// interactive editing
	promptForm(&cfg)

	// show a short summary
	fmt.Println("\n=== Summary ===")
	renderTable(map[string]string{
		"Name":      cfg.Name,
		"RAM (MiB)": strconv.Itoa(cfg.MemMiB),
		"vCPU":      strconv.Itoa(cfg.VCPU),
		"Disk-Path": cfg.Disk,
		"Network":   cfg.Network,
		"ISO":			 cfg.ISO,
	})
	fmt.Println("\n===============")

	// create the VM (virt‑install → XML → virsh define)
	if err := createVMFromConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating the VM:", err)
		os.Exit(1)
	}
}