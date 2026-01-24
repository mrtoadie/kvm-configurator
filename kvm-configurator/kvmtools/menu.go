// kvmtools/menu.go
// last modification: January 24 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	// interne UI‑Hilfen (Farbgebung)
	"configurator/internal/ui"
)

// Start zeigt das Menü, liest die Auswahl und delegiert an exec.Run.
// Der Aufrufer (z. B. main.go) muss nur einen *bufio.Reader* übergeben.
func Start(r *bufio.Reader) {
	for {
		printMenu()
		fmt.Print(ui.Colourise("\nBitte wählen (oder q zum Zurück): ", ui.Yellow))

		choiceRaw, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, ui.Colourise("Eingabefehler – bitte erneut versuchen.", ui.Red))
			continue
		}
		choice := strings.TrimSpace(choiceRaw)

		// Rückkehr zum Hauptmenü
		if choice == "q" || choice == "quit" {
			return
		}

		// error output
		if err := Run(choice); err != nil {
			fmt.Fprintln(os.Stderr, ui.Colourise("❌ "+err.Error(), ui.Red))
		} else {
			fmt.Println(ui.Colourise("✅ Befehl erfolgreich ausgeführt.", ui.Green))
		}
	}
}

// printMenu erzeugt die tabellarische Ausgabe.
func printMenu() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	menuItems := []struct {
		Key   string
		Desc  string
	}{
		{"1", "Alle VMs auflisten (virsh list --all)"},
		{"2", "Laufende VMs auflisten (virsh list)"},
		{"3", "VM neu starten"},
		{"4", "VM herunterfahren"},
		{"5", "VM zerstören"},
		{"6", "Host-System-Infos (nodeinfo)"},
		{"7", "start"},
		{"q", "Zurück zum Hauptmenü"},
	}

	fmt.Fprintln(w, ui.Colourise("\n=== KVM‑TOOLS ===", ui.Blue))
	fmt.Fprintln(w, "Auswahl\tBeschreibung")
	for _, it := range menuItems {
		fmt.Fprintf(w, "%s\t%s\n", it.Key, it.Desc)
	}
	w.Flush()
}