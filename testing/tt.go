package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
)

/* -----------------------------------------------------------------
   Konstanten & globale Variablen
   ----------------------------------------------------------------- */
const tmplDir = "../template/templates" // relativer Pfad zum XML‑Ordner

/* -----------------------------------------------------------------
   Datenstruktur für die VM‑Konfiguration (aus dem zweiten Beispiel)
   ----------------------------------------------------------------- */
type DomainConfig struct {
	Name    string
	MemMiB  int
	VCPU    int
	Disk    string
	Network string
}

/* -----------------------------------------------------------------
   Hilfsfunktionen – Eingabe & Dateisystem
   ----------------------------------------------------------------- */

// sichere Ganzzahl‑Eingabe
func readInt(prompt string, min, max int) int {
	for {
		fmt.Print(prompt)
		var in string
		if _, err := fmt.Scanln(&in); err != nil {
			fmt.Println("Bitte eine Zahl eingeben.")
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(in))
		if err != nil || n < min || n > max {
			fmt.Printf("Zahl zwischen %d und %d erwartet.\n", min, max)
			continue
		}
		return n
	}
}

// sichere String‑Eingabe (nicht‑leer)
func readString(prompt string) string {
	for {
		fmt.Print(prompt)
		var s string
		if _, err := fmt.Scanln(&s); err == nil && s != "" {
			return s
		}
		fmt.Println("Bitte einen nicht‑leeren Wert eingeben.")
	}
}

// liefert alle *.xml‑Dateien im tmplDir (ohne Unterordner)
func listXMLFiles() ([]string, error) {
	entries, err := os.ReadDir(tmplDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".xml") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

/* -----------------------------------------------------------------
   1) Vorhandene XML‑Dateien auflisten & anzeigen
   ----------------------------------------------------------------- */
func showExistingXML() {
	files, err := listXMLFiles()
	if err != nil {
		fmt.Println("Fehler beim Lesen des Ordners:", err)
		return
	}
	if len(files) == 0 {
		fmt.Println("Keine XML‑Dateien gefunden.")
		return
	}

	for {
		fmt.Println("\n--- Vorhandene XML‑Dateien ---")
		for i, f := range files {
			fmt.Printf("%d) %s\n", i+1, f)
		}
		fmt.Println("0) Zurück")
		choice := readInt("Datei wählen: ", 0, len(files))

		if choice == 0 {
			return // zurück zum Hauptmenü
		}
		filename := files[choice-1]
		displayFile(filepath.Join(tmplDir, filename))
	}
}

// gibt den Inhalt einer Datei (mit Header) aus
func displayFile(fullPath string) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		fmt.Println("Konnte Datei nicht lesen:", err)
		return
	}
	fmt.Printf("\n--- Inhalt von %s ---\n%s\n", filepath.Base(fullPath), string(data))
}

/* -----------------------------------------------------------------
   2) Neue XML aus einem vorhandenen Template erzeugen
   ----------------------------------------------------------------- */
func createFromTemplate() {
	templates, err := listXMLFiles()
	if err != nil {
		fmt.Println("Fehler beim Lesen des Ordners:", err)
		return
	}
	if len(templates) == 0 {
		fmt.Println("Keine Templates vorhanden.")
		return
	}

	// Template auswählen
	fmt.Println("\n--- Template wählen ---")
	for i, t := range templates {
		fmt.Printf("%d) %s\n", i+1, t)
	}
	fmt.Println("0) Abbrechen")
	choice := readInt("Auswahl: ", 0, len(templates))
	if choice == 0 {
		return
	}
	templateFile := templates[choice-1]

	// Basis‑Daten vom Nutzer einholen
	name := readString("VM‑Name: ")
	mem := readInt("RAM in MiB (z. B. 2048): ", 1, 1<<30)
	cpu := readInt("vCPU‑Anzahl: ", 1, 64)

	// Template‑Datei einlesen
	base, err := os.ReadFile(filepath.Join(tmplDir, templateFile))
	if err != nil {
		fmt.Println("Template konnte nicht gelesen werden:", err)
		return
	}
	content := string(base)

	// Platzhalter ersetzen – sehr simpel, reicht für Demo
	content = strings.ReplaceAll(content, "{{NAME}}", name)
	content = strings.ReplaceAll(content, "{{MEM}}", strconv.Itoa(mem))
	content = strings.ReplaceAll(content, "{{CPU}}", strconv.Itoa(cpu))

	// Ziel‑Dateiname bestimmen
	outName := fmt.Sprintf("%s-%s.xml", strings.TrimSuffix(templateFile, ".xml"), name)
	outPath := filepath.Join(tmplDir, outName)

	if err := os.WriteFile(outPath, []byte(content), fs.ModePerm); err != nil {
		fmt.Println("Fehler beim Schreiben:", err)
		return
	}
	fmt.Printf("Neue XML‑Datei erstellt: %s\n", outPath)
}

/* -----------------------------------------------------------------
   3) Interaktives Formular zur Bearbeitung einer DomainConfig
   ----------------------------------------------------------------- */
func promptForm(cfg *DomainConfig) {
	in := bufio.NewReader(os.Stdin)
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
		w.Flush()

		fmt.Print("\nFeld wählen (1‑5) oder leer zum Beenden: ")
		fieldRaw, _ := in.ReadString('\n')
		field := strings.TrimSpace(strings.ToLower(fieldRaw))
		if field == "" {
			break // fertig
		}

		switch field {
		case "1", "name":
			fmt.Print(">> Neuer Name: ")
			val, _ := in.ReadString('\n')
			cfg.Name = strings.TrimSpace(val)

		case "2", "ram", "mem", "memory":
			fmt.Print("RAM in MiB: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.MemMiB = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "3", "vcpu":
			fmt.Print("vCPU‑Anzahl: ")
			val, _ := in.ReadString('\n')
			if v, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				cfg.VCPU = v
			} else {
				fmt.Println("Ungültige Zahl – Wert bleibt unverändert.")
			}

		case "4", "disk":
			fmt.Print("Disk‑Pfad (leer = keine Disk): ")
			val, _ := in.ReadString('\n')
			cfg.Disk = strings.TrimSpace(val)

		case "5", "network", "net":
			fmt.Print("Netzwerk (kommagetrennt): ")
			val, _ := in.ReadString('\n')
			cfg.Network = strings.TrimSpace(val)

		default:
			fmt.Println("Unbekanntes Feld – bitte 1‑5 oder Namen eingeben.")
		}
	}
}

/* -----------------------------------------------------------------
   Hauptmenü – vereint alle drei Funktionen
   ----------------------------------------------------------------- */
func main() {
	// Start‑Konfiguration für das Formular (kann später geändert werden)
	cfg := DomainConfig{
		Name:    "my‑guest",
		MemMiB:  1024,
		VCPU:    2,
		Disk:    "",
		Network: "default",
	}

	for {
		fmt.Println("\n=== Hauptmenü ===")
		fmt.Println("1) Vorhandene XML‑Dateien anzeigen")
		fmt.Println("2) Neue XML‑Datei aus Template erzeugen")
		fmt.Println("3) VM‑Konfiguration bearbeiten (Tab‑Formular)")
		fmt.Println("0) Beenden")
		choice := readInt("Auswahl: ", 0, 3)

		switch choice {
		case 0:
			fmt.Println("Tschüss – möge dein Code stets kompilieren!")
			return
		case 1:
			showExistingXML()
		case 2:
			createFromTemplate()
		case 3:
			promptForm(&cfg)
			// Ergebnis hübsch ausgeben
			fmt.Println("\n=== Endgültige Konfiguration ===")
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
			fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
			fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
			fmt.Fprintf(w, "Disk‑Pfad:\t%s\n", cfg.Disk)
			fmt.Fprintf(w, "Netzwerk:\t%s\n", cfg.Network)
			w.Flush()
		}
	}
}