// internal/ui/error.go
// last modification: January 26 2026
package ui

import (
	"fmt"
	"os"
	"errors"
)

// ---------------------------------------------------------------------
// 1️⃣  Öffentliche Fehler‑Variablen (einmal definiert, überall nutzbar)
// ---------------------------------------------------------------------

var (
	// Allgemeine Konfigurations‑Probleme
	ErrConfigMissing   = errors.New("Configuration file not found")
	ErrConfigInvalid   = errors.New("Konfigurationsdatei ungültig oder fehlerhaft")
	ErrOSListMissing   = errors.New("OS‑Liste (oslist.yaml) fehlt")
	ErrOSListInvalid   = errors.New("OS‑Liste konnte nicht geparst werden")
	ErrWorkDirInvalid  = errors.New("Cannot resolve work directory")

	// Prerequisite‑Probleme
	ErrVirtInstallMissing = errors.New("„virt‑install“ nicht im $PATH")
	ErrVirshMissing       = errors.New("„virsh“ nicht im $PATH")

	// VM‑Erstellungs‑Probleme
	ErrISONotFound      = errors.New("ISO‑Datei nicht erreichbar")
	ErrDiskCreationFail = errors.New("Disk‑Argument konnte nicht gebaut werden")
	ErrVirtInstallFail  = errors.New("virt‑install schlug fehl")
	ErrVirshDefineFail  = errors.New("virsh define fehlgeschlagen")
)

// ---------------------------------------------------------------------
// 2️⃣  Hilfsfunktion: Fehler mit Kontext formatieren & ausgeben
// ---------------------------------------------------------------------

// Report gibt einen hübsch eingefärbten Fehler aus.
// * ctx* – kurzer Kontext‑Text (z. B. „VM‑Erstellung“)
// * err* – der eigentliche Fehler (kann einer der vordefinierten sein)
func Report(ctx string, err error) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf("❌  %s: %v", ctx, err)
	// Wir nutzen bereits Colourise, um den Text rot zu färben.
	fmt.Fprintln(os.Stderr, Colourise(msg, Red))
}

// ---------------------------------------------------------------------
// 3️⃣  Convenience‑Wrapper für Must / Warn, die das Registry‑File nutzen
// ---------------------------------------------------------------------

// MustFatal ist ein Shortcut für ui.Must, aber mit vordefinierten
// Fehlermeldungen aus diesem File.  Wenn *err* nil ist, passiert nichts.
// Sonst wird das Programm mit rotem Kreuz beendet.
func MustFatal(err error, ctx string) {
	if err != nil {
		Report(ctx, err)
		os.Exit(1)
	}
}

// WarnSoft ist ein Shortcut für ui.Warn, ebenfalls mit Registry‑Fehlern.
func WarnSoft(err error, ctx string) {
	if err != nil {
		Report(ctx, err)
	}
}
/* --------------------
	Uniform error message + exit
	example: ui.Fatal(err, "Error loading file-config")
-------------------- */
func Fatal(err error, ctx string) {
	Report(ctx, err) // farbiges Ausgeben
	os.Exit(1)       // sofortiger Abbruch
}

