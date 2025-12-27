package main

import (
	"fmt"
)

func main() {
	for {
		// ----- Hauptmenü -----
		fmt.Println("\n=== Hauptmenü ===")
		fmt.Println("1) Menü A")
		fmt.Println("2) Menü B")
		fmt.Println("3) Menü C")
		fmt.Println("4) Menü D")
		fmt.Println("0) Beenden")
		var choice int
		fmt.Print("Auswahl: ")
		if _, err := fmt.Scanln(&choice); err != nil {
			fmt.Println("Bitte eine Zahl eingeben.")
			continue
		}

		switch choice {
		case 0:
			fmt.Println("Tschüss!")
			return
		case 1:
			subMenu("A")
		case 2:
			subMenu("B")
		case 3:
			subMenu("C")
		case 4:
			subMenu("D")
		default:
			fmt.Println("Ungültige Auswahl.")
		}
	}
}

// ----- Untermenü (identisch für A‑D) -----
func subMenu(label string) {
	for {
		fmt.Printf("\n--- Menü %s ---\n", label)
		fmt.Printf("1) Untermenü %s‑1\n", label)
		fmt.Printf("2) Untermenü %s‑2\n", label)
		fmt.Println("0) Zurück")
		var sub int
		fmt.Print("Auswahl: ")
		if _, err := fmt.Scanln(&sub); err != nil {
			fmt.Println("Bitte eine Zahl eingeben.")
			continue
		}
		switch sub {
		case 0:
			return // zurück zum Hauptmenü
		case 1:
			fmt.Printf("▶️  Aktion %s‑1 ausgeführt\n", label)
		case 2:
			fmt.Printf("▶️  Aktion %s‑2 ausgeführt\n", label)
		default:
			fmt.Println("Ungültige Auswahl.")
		}
	}
}