package main

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"strconv"
	"os"
)

// ---------- Datenmodell ----------
type DomainConfig struct {
	Name    string `form:"name"`
	Memory  int    `form:"memory"` // MiB
	VCPU    int    `form:"vcpu"`
	Disk    string `form:"disk"`    // Pfad zum Disk‑Image
	Network string `form:"network"` // z. B. "default"
}

// ---------- libvirt‑XML ----------
type Domain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`

	Name   string   `xml:"name"`
	Memory Memory   `xml:"memory"`
	VCpus  Vcpu     `xml:"vcpu"`
	OS     OS       `xml:"os"`
	Devices Devices `xml:"devices"`
}

type Memory struct {
	Unit string `xml:"unit,attr"`
	Size int    `xml:",chardata"` // KiB
}

type Vcpu struct {
	Placement string `xml:"placement,attr,omitempty"`
	Count     int    `xml:",chardata"`
}

type OS struct {
	Type OSType `xml:"type"`
}

type OSType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Value   string `xml:",chardata"`
}

type Devices struct {
	Disk    *Disk    `xml:"disk,omitempty"`
	Network *Network `xml:"interface,omitempty"`
	Graphics *Graphics `xml:"graphics,omitempty"`
}

type Disk struct {
	Type   string   `xml:"type,attr"`
	Device string   `xml:"device,attr"`
	Driver Driver   `xml:"driver"`
	Source Source   `xml:"source"`
	Target Target   `xml:"target"`
}

type Driver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Source struct {
	File string `xml:"file,attr,omitempty"`
}

type Target struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr,omitempty"`
}

type Network struct {
	Type   string   `xml:"type,attr"`
	Source NetSrc   `xml:"source"`
	Model  NetModel `xml:"model"`
}

type NetSrc struct {
	Network string `xml:"network,attr,omitempty"`
	Bridge  string `xml:"bridge,attr,omitempty"`
}

type NetModel struct {
	Type string `xml:"type,attr"`
}

type Graphics struct {
	Type string `xml:"type,attr"`
	Port string `xml:"port,attr"`
}

// ---------- HTML‑Templates ----------
var tmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="de">
<head>
	<meta charset="UTF-8">
	<title>KVM‑Config‑Generator</title>
	<style>
		body{font-family:sans-serif;max-width:800px;margin:auto;padding:1rem;}
		label{display:block;margin-top:0.5rem;}
		input, textarea{width:100%;padding:0.4rem;}
	</style>
</head>
<body>
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
</body>
</html>
`))

// ---------- Handler ----------
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

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
	// Zahlen konvertieren, bei Fehlern Default‑Werte setzen
	if m, err := strconv.Atoi(r.FormValue("memory")); err == nil {
		cfg.Memory = m
	}
	if v, err := strconv.Atoi(r.FormValue("vcpu")); err == nil {
		cfg.VCPU = v
	}

	xmlStr, err := generateLibvirtXML(cfg)
	if err != nil {
		http.Error(w, "XML konnte nicht erzeugt werden: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Zwischenspeichern, damit der Download‑Link funktioniert
	_ = os.WriteFile("./tmp/"+cfg.Name+".xml", []byte(xmlStr), 0644)

	// Seite erneut rendern, jetzt mit Ergebnis
	tmpl.Execute(w, map[string]string{
		"XML":  xmlStr,
		"Name": cfg.Name,
	})
}

// Download‑Handler
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Kein Name angegeben", http.StatusBadRequest)
		return
	}
	path := "./tmp/" + name + ".xml"
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Datei nicht gefunden", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename="+name+".xml")
	w.Write(data)
}

// ---------- XML‑Generator ----------
func generateLibvirtXML(c DomainConfig) (string, error) {
	d := Domain{
		Type: "kvm",
		Name: c.Name,
		Memory: Memory{
			Unit: "KiB",
			Size: c.Memory * 1024, // MiB → KiB
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
	// Optional: VNC‑Zugang
	d.Devices.Graphics = &Graphics{Type: "vnc", Port: "-1"}

	out, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	// Header hinzufügen
	return xml.Header + string(out), nil
}

// ---------- main ----------
func main() {
	// Einfacher Ordner für temporäre XML‑Dateien
	_ = os.MkdirAll("./tmp", 0755)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/download", downloadHandler)

	// Auf Port 8080 lauschen (lokal oder hinter Reverse‑Proxy)
	println("Starte Server unter http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}