package main

import (
	"encoding/xml"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/* -------------------------------------------------------------
   --------------------------- Datenmodell -----------------------
   ------------------------------------------------------------- */
type DomainConfig struct {
	Name    string `form:"name"`
	Memory  int    `form:"memory"` // MiB
	VCPU    int    `form:"vcpu"`
	Disk    string `form:"disk"`    // Pfad zum Disk‑Image
	Network string `form:"network"` // z. B. "default"
	OS      string `form:"os"`      // Dropdown‑Wert
}

/* -------------------------------------------------------------
   --------------------------- libvirt‑XML ----------------------
   ------------------------------------------------------------- */
type Domain struct {
	XMLName    xml.Name `xml:"domain"`
	XmlnsQemu  string   `xml:"xmlns:qemu,attr,omitempty"` // optional, für qemu‑Optionen
	Type       string   `xml:"type,attr"`

	Name       string   `xml:"name"`
	Memory     Memory   `xml:"memory"`
	CurrentMem *Memory  `xml:"currentMemory,omitempty"`
	VCpus      Vcpu     `xml:"vcpu"`
	OS         OS       `xml:"os"`
	Features   Features `xml:"features"` // muss vor clock & devices stehen
	Clock      Clock    `xml:"clock"`
	OnPoweroff string   `xml:"on_poweroff"`
	OnReboot   string   `xml:"on_reboot"`
	OnCrash    string   `xml:"on_crash"`
	Devices    Devices  `xml:"devices"`
}

/* ----- Grundelemente ------------------------------------------------- */
type Memory struct {
	Unit string `xml:"unit,attr"`
	Size int    `xml:",chardata"` // KiB
}

type Vcpu struct {
	Placement string `xml:"placement,attr,omitempty"`
	Count     int    `xml:",chardata"`
}

/* ----- OS ----------------------------------------------------------- */
type OS struct {
	Type OSType `xml:"type"` // <type arch="…" machine="…">hvm</type>
	Boot *Boot   `xml:"boot,omitempty"`
}

type OSType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Value   string `xml:",chardata"` // meist "hvm"
}

type Boot struct {
	Dev string `xml:"dev,attr"` // z. B. "hd"
}

/* ----- Features ------------------------------------------------------ */
type Features struct {
	ACPI *Empty `xml:"acpi,omitempty"`
	APIC *Empty `xml:"apic,omitempty"`
	PAE  *Empty `xml:"pae,omitempty"`
}

/* Leerer Marker – erzeugt ein leeres Tag <acpi/> usw. */
type Empty struct{}

/* ----- Clock ---------------------------------------------------------- */
type Clock struct {
	Offset string `xml:"offset,attr"` // "utc"
}

/* ----- Devices -------------------------------------------------------- */
type Devices struct {
	Disk     *Disk     `xml:"disk,omitempty"`
	Network  *Network  `xml:"interface,omitempty"`
	Graphics *Graphics `xml:"graphics,omitempty"`

	Input   []Input   `xml:"input,omitempty"`   // Maus + Tastatur
	Console *Console  `xml:"console,omitempty"` // virsh console
}

/* ----- Disk ---------------------------------------------------------- */
type Disk struct {
	Type   string `xml:"type,attr"`   // file
	Device string `xml:"device,attr"` // disk
	Driver Driver `xml:"driver"`
	Source Source `xml:"source"`
	Target Target `xml:"target"`
}

type Driver struct {
	Name string `xml:"name,attr"` // qemu
	Type string `xml:"type,attr"` // qcow2 / raw …
}

type Source struct {
	File string `xml:"file,attr,omitempty"`
}

type Target struct {
	Dev string `xml:"dev,attr"`          // vda, vdb …
	Bus string `xml:"bus,attr,omitempty"` // virtio, ide …
}

/* ----- Network ------------------------------------------------------- */
type Network struct {
	Type   string  `xml:"type,attr"` // network
	Source NetSrc  `xml:"source"`
	Model  NetModel `xml:"model"`
}

type NetSrc struct {
	Network string `xml:"network,attr,omitempty"`
	Bridge  string `xml:"bridge,attr,omitempty"`
}

type NetModel struct {
	Type string `xml:"type,attr"` // virtio, e1000 …
}

/* ----- Graphics ------------------------------------------------------ */
type Graphics struct {
	Type     string `xml:"type,attr"`     // vnc / spice
	Port     string `xml:"port,attr"`     // -1 → auto
	AutoPort string `xml:"autoport,attr"` // yes
	Listen   string `xml:"listen,attr"`   // 0.0.0.0 (optional, aber praktisch)
}

/* ----- Input & Console ---------------------------------------------- */
type Input struct {
	Type string `xml:"type,attr"` // mouse / keyboard
	Bus  string `xml:"bus,attr"`  // ps2 / usb
}

type Console struct {
	Type   string        `xml:"type,attr"` // pty / tcp
	Target ConsoleTarget `xml:"target"`
}

type ConsoleTarget struct {
	Type string `xml:"type,attr"` // serial / virtio
	Port string `xml:"port,attr"` // 0, 1 …
}

/* -------------------------------------------------------------
   --------------------------- HTML‑Template ----------------------
   ------------------------------------------------------------- */
var tmpl = template.Must(template.New("page").Parse(`
<!DOCTYPE html>
<html lang="de">
<head>
	<meta charset="UTF-8">
	<title>KVM‑Config‑Generator</title>
	<style>
		body{font-family:sans-serif;max-width:800px;margin:auto;padding:1rem;}
		label{display:block;margin-top:0.5rem;}
		input, select, textarea{width:100%;padding:0.4rem;}
	</style>
</head>
<body>
<h1>KVM‑Konfiguration erstellen</h1>
<form method="POST" action="/create">
	<label>Name: <input name="name" required></label>

	<label>OS:
		<select name="os" required>
			<option value="">Bitte wählen …</option>
			<option value="ubuntu">Ubuntu (Linux)</option>
			<option value="debian">Debian (Linux)</option>
			<option value="centos">CentOS (Linux)</option>
			<option value="fedora">Fedora (Linux)</option>
			<option value="windows">Windows 10/11</option>
		</select>
	</label>

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

/* -------------------------------------------------------------
   --------------------------- Helpers ----------------------------
   ------------------------------------------------------------- */
// Entfernt gefährliche Zeichen aus Dateinamen (z. B. "../")
func sanitizeFileName(name string) string {
	name = filepath.Base(name)               // nur letzte Komponente
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "_")
	if name == "" {
		name = "unnamed"
	}
	return name
}

/* -------------------------------------------------------------
   --------------------------- HTTP‑Handler ----------------------
   ------------------------------------------------------------- */
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("template execute error: %v", err)
	}
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
		OS:      r.FormValue("os"),
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

	safeName := sanitizeFileName(cfg.Name)
	_ = os.WriteFile("./tmp/"+safeName+".xml", []byte(xmlStr), 0644)

	if err := tmpl.Execute(w, map[string]string{
		"XML":  xmlStr,
		"Name": safeName,
	}); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

// Download‑Handler
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	name := sanitizeFileName(r.URL.Query().Get("name"))
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

/* -------------------------------------------------------------
   --------------------------- OS‑Mapping ------------------------
   ------------------------------------------------------------- */
func osToOSType(os string) OSType {
	switch os {
	case "ubuntu", "debian", "centos", "fedora":
		return OSType{
			Arch:    "x86_64",
			Machine: "pc-i440fx-2.9",
			Value:   "hvm",
		}
	case "windows":
		// Q35 ist für Windows‑VMs besser geeignet
		return OSType{
			Arch:    "x86_64",
			Machine: "pc-q35-2.9",
			Value:   "hvm",
		}
	default:
		// Fallback – gleiche Werte wie für Linux‑Guests
		return OSType{
			Arch:    "x86_64",
			Machine: "pc-i440fx-2.9",
			Value:   "hvm",
		}
	}
}

/* -------------------------------------------------------------
   --------------------------- XML‑Generator --------------------
   ------------------------------------------------------------- */
func generateLibvirtXML(c DomainConfig) (string, error) {
	// Basis‑Domain‑Objekt
	d := Domain{
		XmlnsQemu: "http://libvirt.org/schemas/domain/qemu/1.0", // optional, aber korrekt
		Type:      "kvm",
		Name:      c.Name,
		Memory: Memory{
			Unit: "KiB",
			Size: c.Memory * 1024, // MiB → KiB
		},
		CurrentMem: &Memory{
			Unit: "KiB",
			Size: c.Memory * 1024,
		},
		VCpus: Vcpu{
			Placement: "static",
			Count:     c.VCPU,
		},
		OS: OS{
			Type: osToOSType(c.OS),
			Boot: &Boot{Dev: "hd"},
		},
		Features: Features{
			ACPI: &Empty{},
			APIC: &Empty{},
			PAE:  &Empty{},
		},
		Clock:      Clock{Offset: "utc"},
		OnPoweroff: "destroy",
		OnReboot:   "restart",
		OnCrash:    "destroy",
		Devices:    Devices{},
	}

	/* ---------- Disk (falls angegeben) ---------- */
	if strings.TrimSpace(c.Disk) != "" {
		d.Devices.Disk = &Disk{
			Type:   "file",
			Device: "disk",
			Driver: Driver{Name: "qemu", Type: "qcow2"},
			Source: Source{File: c.Disk},
			Target: Target{Dev: "vda", Bus: "virtio"},
		}
	}

	/* ---------- Netzwerk (falls angegeben) ---------- */
	if strings.TrimSpace(c.Network) != "" {
		d.Devices.Network = &Network{
			Type: "network",
			Source: NetSrc{
				Network: c.Network,
			},
			Model: NetModel{Type: "virtio"},
		}
	}

	/* ---------- Grafik (VNC) ---------- */
	d.Devices.Graphics = &Graphics{
		Type:     "vnc",
		Port:     "-1",   // -1 → automatischer freier Port
		AutoPort: "yes",  // zwingt libvirt, einen Port zu wählen
		Listen:   "0.0.0.0", // erreichbar von außen (optional)
	}

	/* ---------- Eingabegeräte (Maus + Tastatur) ---------- */
	d.Devices.Input = []Input{
		{Type: "mouse", Bus: "ps2"},
		{Type: "keyboard", Bus: "ps2"},
	}

	/* ---------- Console (virsh) ---------- */
	d.Devices.Console = &Console{
		Type: "pty",
		Target: ConsoleTarget{
			Type: "serial",
			Port: "0",
		},
	}

	// XML serialisieren (schön eingerückt)
	out, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}

/* -------------------------------------------------------------
   ------------------------------ main ---------------------------
   ------------------------------------------------------------- */
func main() {
	// Verzeichnis für temporäre XML‑Dateien anlegen
	if err := os.MkdirAll("./tmp", 0755); err != nil {
		log.Fatalf("Kann ./tmp nicht anlegen: %v", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/download", downloadHandler)

	log.Println("Starte Server unter http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server beendet: %v", err)
	}
}