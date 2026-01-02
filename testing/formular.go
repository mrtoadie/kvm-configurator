package main

import (
	"bufio"
	"fmt"
	"os"
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

/* -------------------------------------------------
   Hilfsfunktion: Eingabe mit Tab‑Writer formatieren
   ------------------------------------------------- */
func promptForm(cfg *DomainConfig) {
	in := bufio.NewReader(os.Stdin)

	// Tab‑Writer initialisieren (min‑Breite, Tab‑Breite, Padding, PadChar, Flags)
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	for {
		// ---- Formular‑Überschrift ----
		fmt.Fprintln(w, "\n=== VM‑Konfiguration ===\t")
		// ---- Jede Zeile: Label\tAktueller Wert\tPrompt ----
		fmt.Fprintf(w, "[1] Name:\t%s\t[Enter für unverändert]\n", cfg.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t[Enter für unverändert]\n", cfg.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\t[Enter für unverändert]\n", cfg.VCPU)
		fmt.Fprintf(w, "[4] Disk‑Pfad:\t%s\t[Enter für unverändert]\n", cfg.Disk)
		fmt.Fprintf(w, "[5] Netzwerk:\t%s\t[Enter für unverändert]\n", cfg.Network)
		w.Flush() // alles ausgeben, bevor wir nach Eingaben fragen

		// ---- Eingaben nacheinander einholen ----
		fmt.Print("\nBitte Feldnamen eingeben (z. B. Name, RAM, vCPU, Disk, Network) oder leer zum Beenden: ")
		fieldRaw, _ := in.ReadString('\n')
		field := strings.TrimSpace(strings.ToLower(fieldRaw))
		if field == "" {
			break // fertig
		}

		switch field {
		case "1":
			fmt.Print(">> Neuer Name: ")
			val, _ := in.ReadString('\n')
			cfg.Name = strings.TrimSpace(val)

		case "2", "mem", "memory":
			fmt.Print("RAM in MiB: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.MemMiB = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "3":
			fmt.Print("vCPU‑Anzahl: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.VCPU = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "4":
			fmt.Print("Disk‑Pfad (leer = keine Disk): ")
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

/* -------------------------------------------------
   Hauptfunktion – nutzt das Formular‑Prompt
   ------------------------------------------------- */
func main() {
	// Start‑Konfiguration (kann leer sein)
	cfg := DomainConfig{
		Name:    "my‑guest",
		MemMiB:  1024,
		VCPU:    2,
		Disk:    "",
		Network: "default",
	}

	promptForm(&cfg)

	// Ergebnis ausgeben (hier einfach als Tabelle)
	fmt.Println("\n=== Endgültige Konfiguration ===")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
	fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
	fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
	fmt.Fprintf(w, "Disk‑Pfad:\t%s\n", cfg.Disk)
	fmt.Fprintf(w, "Netzwerk:\t%s\n", cfg.Network)
	w.Flush()
}