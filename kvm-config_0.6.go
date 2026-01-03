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
	"Ubuntu 24.04 LTS":   "ubuntu24.04",
	"Ubuntu 25.10":   		"ubuntu25.10",
	"Debian 12":          "debian12",
	"Debian 13":          "debian13",
	"Fedora 43":          "fedora43",
	"Arch Linux":         "archlinux",
	"openSUSE Leap 16.0": "opensuse16.0",
	"Windows 10":					"win10"
}

// chooseVariant checks if the distro is known for virt-install identifier
func chooseVariant(key string) (string, error) {
	if v, ok := osVariant[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown distro %q", key)
}

// ask asks a question and reads the input from stdin
func selectOSMenu(r *bufio.Reader) (string, error) {
	fmt.Println("\n=== Choosing the operating system ===")
	
	keys := make([]string, 0, len(osVariant))
	for k := range osVariant {
		keys = append(keys, k)
	}

	for i, name := range keys {
		fmt.Printf("  %2d) %s\n", i+1, name)
	}
	fmt.Print("\nPlease enter a number (or press ENTER for default Ubuntu 24.04): ")

	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)

	// ENTER default Ubuntu 24.04
	if line == "" {
		return "Ubuntu 24.04 LTS", nil
	}

	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(keys) {
		return "", fmt.Errorf("Invalid selection")
	}
	return keys[idx-1], nil
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

// Render Tab‑Writer
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
				fmt.Println("Invalid number – value remains unchanged.")
			}
		case "3":
			fmt.Print("vCPU: ")
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
			fmt.Print("Network (comma-separated): ")
			val, _ := in.ReadString('\n')
			cfg.Network = strings.TrimSpace(val)
		default:
			fmt.Println("Invalid input!")
		}
	}
}

// VM‑creation from DomainConfig
func createVMFromConfig(cfg DomainConfig) error {
	r := bufio.NewReader(os.Stdin)

	// ask disk size
	sizeStr, err := ask("Disk size in GB (e.g., 20): ", r)
	if err != nil {
		return err
	}
	diskSizeGB, err := strconv.Atoi(sizeStr)
	if err != nil {
		return fmt.Errorf("Invalid disk size: %w", err)
	}

	// ask for ISO path
	iso, err := ask("Path to the installation ISO: ", r)
	if err != nil {
		return err
	}

	// ask for OS (submenu)
	osChoice, err := selectOSMenu(r)
	if err != nil {
		return fmt.Errorf("Invalid OS selection: %w", err)
	}
	variant, err := chooseVariant(osChoice)
	if err != nil {
		return err
	}

	// build the vm with virt‑install
	name := cfg.Name
	ram := strconv.Itoa(cfg.MemMiB)
	cpus := strconv.Itoa(cfg.VCPU)

	// defaul virtual-disk path
	diskPath := fmt.Sprintf("/run/media/toadie/vm/QEMU/%s.qcow2", name)
	// hard-coded iso path
	//iso = "/run/media/toadie/data/ISOs/clonezilla-live-20251017-questing-amd64.iso"
	// command arguments
	cmdArgs := buildVirtInstallCmd(name, ram, cpus, diskPath, iso, variant, diskSizeGB)
	virtCmd := exec.Command("virt-install", cmdArgs...)

	// XML output
	var outBuf, errBuf bytes.Buffer
	virtCmd.Stdout = &outBuf
	virtCmd.Stderr = &errBuf

	if err := virtCmd.Run(); err != nil {
		return fmt.Errorf("virt-install failed: %w – %s", err, errBuf.String())
	}
	rawXML := outBuf.Bytes()

	// (optional) show first lines of the xml
	fmt.Printf("XML created (%d Bytes).\n", len(rawXML))
	if len(rawXML) > 0 {
		fmt.Printf("First lines of the XML:\n%s\n", string(rawXML[:200]))
	}

	// save XML
	xmlFile := fmt.Sprintf("%s.xml", name)
	if err := writeXML(rawXML, xmlFile); err != nil {
		return fmt.Errorf("could not write XML: %w", err)
	}
	absPath, _ := filepath.Abs(xmlFile)
	fmt.Printf("XML definition saved under: %s\n", absPath)

	// define VM
	if err := defineVM(xmlFile); err != nil {
		return fmt.Errorf("virsh define failed: %w", err)
	}
	fmt.Println("VM successfully registered with libvirt/qemu (not yet started).")
	return nil
}

// MAIN
func main() {
	// Default config
	cfg := DomainConfig{
		Name:    "new-machine",
		MemMiB:  1024,
		VCPU:    2,
		Disk:    "",
		Network: "default",
	}

	// interactive query
	promptForm(&cfg)

	// show summary
	fmt.Println("\n=== SUMMARY ===")
	renderTable(map[string]string{
		"Name":      cfg.Name,
		"RAM (MiB)": strconv.Itoa(cfg.MemMiB),
		"vCPU":      strconv.Itoa(cfg.VCPU),
		"Disk-Path": cfg.Disk,
		"Network":  cfg.Network,
	})

	// VM creation
	if err := createVMFromConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating the VM:", err)
		os.Exit(1)
	}
}