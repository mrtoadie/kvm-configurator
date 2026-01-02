package main

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/* -----------------------------------------------------------------
   Konstanten
   ----------------------------------------------------------------- */
const tmplDir = "../template/templates" // <- hier liegt dein Template‑Ordner

/* -----------------------------------------------------------------
   Datenmodell (für das Formular)
   ----------------------------------------------------------------- */
type DomainConfig struct {
	Name    string `form:"name"`
	Memory  int    `form:"memory"` // MiB
	VCPU    int    `form:"vcpu"`
	Disk    string `form:"disk"`
	Network string `form:"network"`
}

/* -----------------------------------------------------------------
   libvirt‑XML‑Strukturen (unverändert)
   ----------------------------------------------------------------- */
type Domain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`

	Name   string   `xml:"name"`
	Memory Memory   `xml:"memory"`
	VCpus  Vcpu     `xml:"vcpu"`
	OS     OS       `xml:"os"`
	Devices Devices `xml:"devices"`
}
type Memory struct{ Unit string `xml:"unit,attr"`; Size int `xml:",chardata"` }
type Vcpu struct{ Placement string `xml:"placement,attr,omitempty"`; Count int `xml:",chardata"` }
type OS struct{ Type OSType `xml:"type"` }
type OSType struct{ Arch, Machine, Value string `xml:",attr,omitempty"` }
type Devices struct {
	Disk     *Disk     `xml:"disk,omitempty"`
	Network  *Network  `xml:"interface,omitempty"`
	Graphics *Graphics `xml:"graphics,omitempty"`
}
type Disk struct {
	Type   string `xml:"type,attr"`
	Device string `xml:"device,attr"`
	Driver Driver `xml:"driver"`
	Source Source `xml:"source"`
	Target Target `xml:"target"`
}
type Driver struct{ Name, Type string `xml:"name,attr" xml:"type,attr"` }
type Source struct{ File string `xml:"file,attr,omitempty"` }
type Target struct{ Dev, Bus string `xml:"dev,attr" xml:"bus,attr,omitempty"` }
type Network struct {
	Type   string   `xml:"type,attr"`
	Source NetSrc   `xml:"source"`
	Model  NetModel `xml:"model"`
}
type NetSrc struct{ Network, Bridge string `xml:"network,attr,omitempty" xml:"bridge,attr,omitempty"` }
type NetModel struct{ Type string `xml:"type,attr"` }
type Graphics struct{ Type, Port string `xml:"type,attr" xml:"port,attr"` }

/* -----------------------------------------------------------------
   HTML‑Templates
   ----------------------------------------------------------------- */
var (
	mainTmpl = template.Must(template.New("main").Parse(`
<!DOCTYPE html>
<html lang="de"><head><meta charset="UTF-8"><title>KVM‑Config‑Generator</title>
<style>
body{font-family:sans-serif;max-width:800px;margin:auto;padding:1rem;}
label{display:block;margin-top:0.5rem;}
input,textarea{width:100%;padding:0.4rem;}
</style></head><body>
<h1>KVM‑Konfiguration erstellen</h1>
<form method="POST" action="/create">
	<label>Name: <input name="name" required></label>
	<label>RAM (MiB): <input name="memory" type="number" min="64" required></label>
	<label>vCPU‑Anzahl: <input name="vcpu" type="number" min="1" required></label>
	<label>Disk‑Pfad (z. B. /var/lib/libvirt/images/vm.qcow2):
		<input name="disk"></label>
	<label>Netzwerk (z. B. default): <input name="network"></label>
	<button type="submit">XML erzeugen</button>
</form>

{{if .XML}}
<h2>Erzeugtes libvirt‑XML</h2>
<pre>{{.XML}}</pre>
<a href="/download?name={{.Name}}">Download XML</a>
{{end}}

<hr>
<h2>Weitere Aktionen</h2>
<ul>
<li><a href="/templates">Alle XML‑Templates anzeigen</a></li>
<li><a href="/files">Alle vorhandenen XML‑Dateien (nicht‑Templates) anzeigen</a></li>
</ul>
</body></html>`))

	templatesListTmpl = template.Must(template.New("list").Parse(`
<!DOCTYPE html>
<html lang="de"><head><meta charset="UTF-8"><title>XML‑Templates</title>
<style>
body{font-family:sans-serif;max-width:800px;margin:auto;padding:1rem;}
a{color:#0066cc;}
pre{background:#f8f8f8;padding:1rem;overflow-x:auto;}
</style></head><body>
<h1>Verfügbare XML‑Templates</h1>
<ul>
{{range .Files}}
<li><a href="/templates/view?file={{.}}">{{.}}</a></li>
{{else}}
<li>Keine Templates gefunden.</li>
{{end}}
</ul>
<p><a href="/">Zurück zum Hauptmenü</a></p>
</body></html>`))

	viewTemplateTmpl = template.Must(template.New("view").Parse(`
<!DOCTYPE html>
<html lang="de"><head><meta charset="UTF-8"><title>Template {{.Name}}</title>
<style>
body{font-family:sans-serif;max-width:800px;margin:auto;padding:1rem;}
pre{background:#f0f0f0;padding:1rem;overflow-x:auto;}
</style></head><body>
<h1>Template: {{.Name}}</h1>
<pre>{{.Content}}</pre>
<p><a href="/templates">Zurück zur Übersicht</a></p>
</body></html>`))
)

/* -----------------------------------------------------------------
   Hilfsfunktion: alle *.xml‑Dateien in einem Verzeichnis holen
   ----------------------------------------------------------------- */
func listXML(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".xml") {
			res = append(res, e.Name())
		}
	}
	return res, nil
}

/* -----------------------------------------------------------------
   Handler: Startseite (Formular + Links)
   ----------------------------------------------------------------- */
func indexHandler(w http.ResponseWriter, r *http.Request) {
	mainTmpl.Execute(w, nil)
}

/* -----------------------------------------------------------------
   Handler: XML aus Formular erzeugen (wie vorher)
   ----------------------------------------------------------------- */
func createHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formular konnte nicht gelesen werden", http.StatusBadRequest)
		return
	}
	cfg := DomainConfig{
		Name:    r.FormValue("name"),
		Disk:    r.FormValue("disk"),
		Network: r.FormValue("network"),
	}
	if m, err := strconv.Atoi(r.FormValue("memory")); err == nil {
		cfg.Memory = m
	}
	if v, err := strconv.Atoi(r.FormValue("vcpu")); err == nil {
		cfg.VCPU = v
	}
	xmlStr, err := generateLibvirtXML(cfg)
	if err != nil {
		http.Error(w, "XML konnte nicht erzeugt werden: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	// Zwischenspeichern für den Download‑Link
	tmpPath := filepath.Join("tmp", cfg.Name+".xml")
	_ = os.WriteFile(tmpPath, []byte(xmlStr), 0644)

	mainTmpl.Execute(w, map[string]string{
		"XML":  xmlStr,
		"Name": cfg.Name,
	})
}

/* -----------------------------------------------------------------
   Handler: Download des erzeugten XML
   ----------------------------------------------------------------- */
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Kein Name angegeben", http.StatusBadRequest)
		return
	}
	path := filepath.Join("tmp", name+".xml")
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Datei nicht gefunden", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition",
		"attachment; filename="+name+".xml")
	w.Write(data)
}

/* -----------------------------------------------------------------
   Neuer Handler: Auflistung aller Templates
   ----------------------------------------------------------------- */
func listTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := listXML(tmplDir)
	if err != nil {
		http.Error(w, "Kann Templates nicht lesen: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	templatesListTmpl.Execute(w, struct{ Files []string }{files})
}

/* -----------------------------------------------------------------
   Neuer Handler: Einzelnen Template‑Inhalt anzeigen
   ----------------------------------------------------------------- */
func viewTemplateHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		http.Error(w, "Keine Datei angegeben", http.StatusBadRequest)
		return
	}
	fullPath := filepath.Join(tmplDir, file)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Datei nicht gefunden", http.StatusNotFound)
		return
	}
	viewTemplateTmpl.Execute(w, struct {
		Name    string
		Content string
	}{Name: file, Content: string(data)})
}

/* -----------------------------------------------------------------
   Optional: Liste aller *nicht‑Template* XML‑Dateien (wie vorher)
   ----------------------------------------------------------------- */
func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Hier könntest du ein zweites Verzeichnis angeben, z. B. "./files"
	// Für das Beispiel zeigen wir einfach dieselben Dateien wie bei den Templates.
	files, err := listXML(tmplDir)
	if err != nil {
		http.Error(w, "Kann Dateien nicht lesen: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	templatesListTmpl.Execute(w, struct{ Files []string }{files})
}

/* -----------------------------------------------------------------
   XML‑Generator (unverändert)
   ----------------------------------------------------------------- */
func generateLibvirtXML(c DomainConfig) (string, error) {
	d := Domain{
		Type: "kvm",
		Name: c.Name,
		Memory: Memory{
			Unit: "KiB",
			Size: c.Memory * 1024,
		},
		VCpus: Vcpu{
			Placement: "static",
			Count:     c.VCPU,
		},
		OS: OS{
			Type: OSType{
				Arch:    "x86_64",
				Machine: "pc-i440fx-2.9",
				Value:   "hvm",
			},
		},
		Devices: Devices{},
	}
	if c.Disk != "" {
		d.Devices.Disk = &Disk{
			Type:   "file",
			Device: "disk",
			Driver: Driver{Name: "qemu", Type: "qcow2"},
			Source: Source{File: c.Disk},
			Target: Target{Dev: "vda", Bus: "virtio"},
		}
	}
	if c.Network != "" {
		d.Devices.Network = &Network{
			Type: "network",
			Source: NetSrc{
				Network: c.Network,
			},
			Model: NetModel{Type: "virtio"},
		}
	}
	d.Devices.Graphics = &Graphics{Type: "vnc", Port: "-1"}

	out, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}

/* -----------------------------------------------------------------
   main()
   ----------------------------------------------------------------- */
func main() {
	// Temporäres Verzeichnis für erzeugte XML‑Dateien anlegen
	if err := os.MkdirAll("tmp", 0755); err != nil {
		panic("Kann tmp‑Verzeichnis nicht anlegen: " + err.Error())
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/download", downloadHandler)

	// ==== neue Routen für Templates ==========================
	http.HandleFunc("/templates", listTemplatesHandler)          // Übersicht
	http.HandleFunc("/templates/view", viewTemplateHandler)     // Einzeldatei
	// ========================================================

	// Optional: wenn du einen separaten Ordner für „normale“ XML‑Dateien hast,
	// kannst du hier einen eigenen Handler registrieren.
	// http.HandleFunc("/files", listFilesHandler)

	println("Server läuft unter http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}