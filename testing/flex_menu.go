package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const tmplDir = "../template/templates" // relativer Pfad zum XML‑Ordner

func main() {
	for {
		fmt.Println("\n=== Hauptmenü ===")
		fmt.Println("1) Vorhandene XML‑Dateien anzeigen")
		fmt.Println("2) Neue XML‑Datei aus Template erzeugen")
		fmt.Println("0) Beenden")
		choice := readInt("Auswahl: ", 0, 2)

		switch choice {
		case 0:
			fmt.Println("Tschüss!")
			return
		case 1:
			showExistingXML()
		case 2:
			createFromTemplate()
		}
	}
}

/* -------------------------------------------------
   Hilfsfunktion: sichere Ganzzahl‑Eingabe
   ------------------------------------------------- */
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

/* -------------------------------------------------
   1) Vorhandene XML‑Dateien auflisten & anzeigen
   ------------------------------------------------- */
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

	// Untermenü – Datei auswählen
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

/* -------------------------------------------------
   2) Neue XML aus einem vorhandenen Template erzeugen
   ------------------------------------------------- */
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

/* -------------------------------------------------
   Kleine Hilfs‑Wrapper
   ------------------------------------------------- */
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