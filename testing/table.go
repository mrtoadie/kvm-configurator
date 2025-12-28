package main

import (
	"fmt"
	"text/tabwriter"
	"os"
)

type Config struct {
	Name   string
	MemMiB int
	VCPU   int
}

func main() {
	cfg := Config{Name: "guest", MemMiB: 1024, VCPU: 2}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tRAM(MiB)\tvCPU\n")
	fmt.Fprintf(w, "%s\t%d\t%d\n", cfg.Name, cfg.MemMiB, cfg.VCPU)
	w.Flush()

	// Eingabe im "Spreadsheet‑Stil"
	fmt.Print("\nNeue Werte (Name RAM VCPU) – leer für unverändert: ")
	var name string
	var ram, cpu int
	_, err := fmt.Scanf("%s %d %d\n", &name, &ram, &cpu)
	if err != nil {
		// Scanf bricht bei leerer Eingabe ab → wir ignorieren den Fehler
	}
	if name != "" {
		cfg.Name = name
	}
	if ram != 0 {
		cfg.MemMiB = ram
	}
	if cpu != 0 {
		cfg.VCPU = cpu
	}

	// Ergebnis erneut ausgeben
	fmt.Println("\nAktualisierte Konfiguration:")
	w = tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tRAM(MiB)\tvCPU\n")
	fmt.Fprintf(w, "%s\t%d\t%d\n", cfg.Name, cfg.MemMiB, cfg.VCPU)
	w.Flush()
}