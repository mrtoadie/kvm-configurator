// file: web.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

/* -------------------------------------------------
   Datenmodelle für die HTTP‑API (identisch zu DomainConfig)
   ------------------------------------------------- */
type apiDomainConfig struct {
	Name     string `json:"name"`      // Distribution‑Name (z. B. "Ubuntu 22.04")
	MemMiB   int    `json:"memMiB"`    // RAM in MiB
	VCPU     int    `json:"vcpu"`      // Anzahl vCPUs
	Disk     string `json:"disk"`      // Pfad zum Speicherort
	Disksize int    `json:"disksize"`  // Größe in GB
	Network  string `json:"network"`   // Netzwerk‑Name
}

/* -------------------------------------------------
   Hilfsfunktion: API‑Struct → interne DomainConfig
   ------------------------------------------------- */
func toDomainConfig(a apiDomainConfig) DomainConfig {
	return DomainConfig{
		Name:     a.Name,
		MemMiB:   a.MemMiB,
		VCPU:     a.VCPU,
		Disk:     a.Disk,
		Disksize: a.Disksize,
		Network:  a.Network,
	}
}

/* -------------------------------------------------
   /api/oslist  –  GET
   Liefert die in oslist.yaml definierten Distributionen
   ------------------------------------------------- */
func osListHandler(w http.ResponseWriter, r *http.Request) {
	type distroInfo struct {
		Name     string `json:"name"`
		CPU      int    `json:"cpu"`
		RAM      int    `json:"ram"`
		Disksize int    `json:"disksize"`
	}
	out := make([]distroInfo, len(osList))
	for i, d := range osList {
		out[i] = distroInfo{
			Name:     d.Name,
			CPU:      d.CPU,
			RAM:      d.RAM,
			Disksize: d.Disksize,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

/* -------------------------------------------------
   /api/create  –  POST
   Erwartet ein JSON‑Body mit allen VM‑Parametern und legt die VM an.
   ------------------------------------------------- */
func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Nur POST erlaubt", http.StatusMethodNotAllowed)
		return
	}
	var apiCfg apiDomainConfig
	if err := json.NewDecoder(r.Body).Decode(&apiCfg); err != nil {
		http.Error(w, "Ungültiges JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Variante (os‑variant) anhand des Namens bestimmen
	variant, ok := variantByName[apiCfg.Name]
	if !ok {
		http.Error(w, "Unbekannte Distribution: "+apiCfg.Name, http.StatusBadRequest)
		return
	}
	cfg := toDomainConfig(apiCfg)

	// Aufruf der bereits vorhandenen createVM‑Logik
	if err := createVM(cfg, variant); err != nil {
		http.Error(w, "Fehler beim Anlegen der VM: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "VM erfolgreich angelegt")
}

/* -------------------------------------------------
   Server‑Start‑Funktion
   ------------------------------------------------- */
func startWebServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oslist", osListHandler)   // GET → Liste der OS
	mux.HandleFunc("/api/create", createVMHandler) // POST → VM anlegen

	addr := ":8080"
	fmt.Printf("🚀 Web‑UI läuft unter http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}

/* -------------------------------------------------
   Dummy‑Funktion wird jetzt zum Server‑Starter
   ------------------------------------------------- */
func dummyTest() {
	// Server im Hintergrund starten, damit das Hauptmenü weiterläuft
	go startWebServer()
	fmt.Println("✅ Web‑Server gestartet – besuche http://localhost:8080")
}