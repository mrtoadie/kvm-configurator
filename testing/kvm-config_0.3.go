package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	//"strings"
)


// Daten‑Container (DomainConfig)
type DomainConfig struct {
	Name   string 
	MemMiB int    
	VCPU   int    
	// ...
}

// XML Structure
type Domain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`

	Name    string   `xml:"name"`
	Memory  Memory   `xml:"memory"`
	VCPU    VCPU     `xml:"vcpu"`
	OS      OS       `xml:"os"`
	Devices Devices  `xml:"devices"`
}

type Memory struct {
	Unit  string `xml:"unit,attr,omitempty"`
	Value int    `xml:",chardata"`
}
type VCPU struct {
	Placement string `xml:"placement,attr,omitempty"`
	Value     int    `xml:",chardata"`
}
type OS struct {
	Type OSBootType `xml:"type"`
	Boot BootDevice `xml:"boot"`
}
type OSBootType struct {
	Arch    string `xml:"arch,attr,omitempty"`
	Machine string `xml:"machine,attr,omitempty"`
	Value   string `xml:",chardata"`
}
type BootDevice struct {
	Dev string `xml:"dev,attr"`
}
type Devices struct {
	Disk      Disk      `xml:"disk,omitempty"`
	Interface Interface `xml:"interface,omitempty"`
	Graphics  Graphics  `xml:"graphics"`
}
type Disk struct {
	Type   string `xml:"type,attr"`
	Device string `xml:"device,attr"`
	Driver Driver `xml:"driver"`
	Source Source `xml:"source"`
	Target Target `xml:"target"`
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
type Interface struct {
	Type   string        `xml:"type,attr"`
	Source InterfaceSrc `xml:"source"`
}
type InterfaceSrc struct {
	Network string `xml:"network,attr,omitempty"`
}
type Graphics struct {
	Type     string `xml:"type,attr"`
	AutoPort string `xml:"autoport,attr,omitempty"`
}

// Fill DomainConfig
func xmlQuery() (DomainConfig, error) {
	var cfg DomainConfig

	// Name
	fmt.Print("Maschinen‑Name: ")
	if _, err := fmt.Scanln(&cfg.Name); err != nil {
		return cfg, fmt.Errorf("Name einlesen fehlgeschlagen: %w", err)
	}

	// RAM
	fmt.Print("RAM in MiB (Standard = 1024): ")
	var memStr string
	if _, err := fmt.Scanln(&memStr); err != nil && err.Error() != "unexpected newline" {
		return cfg, fmt.Errorf("RAM einlesen fehlgeschlagen: %w", err)
	}
	if memStr == "" {
		cfg.MemMiB = 1024
	} else {
		m, err := strconv.Atoi(memStr)
		if err != nil {
			return cfg, fmt.Errorf("ungültige RAM‑Zahl: %w", err)
		}
		cfg.MemMiB = m
	}

	// CPUs
	fmt.Print("Anzahl vCPUs (Standard = 1): ")
	var cpuStr string
	if _, err := fmt.Scanln(&cpuStr); err != nil && err.Error() != "unexpected newline" {
		return cfg, fmt.Errorf("vCPU einlesen fehlgeschlagen: %w", err)
	}
	if cpuStr == "" {
		cfg.VCPU = 1 
	} else {
		c, err := strconv.Atoi(cpuStr)
		if err != nil {
			return cfg, fmt.Errorf("ungültige vCPU‑Zahl: %w", err)
		}
		cfg.VCPU = c
	}

	return cfg, nil
}

// Bulid Domain config
func buildDomain(cfg DomainConfig) Domain {
	return Domain{
		Type: "kvm",
		Name: cfg.Name,
		Memory: Memory{
			Unit:  "MiB",
			Value: cfg.MemMiB,
		},
		VCPU: VCPU{
			Placement: "static",
			Value:     cfg.VCPU,
		},
		OS: OS{
			Type: OSBootType{
				Arch:    "x86_64",
				Machine: "pc-q35-6.2",
				Value:   "hvm",
			},
			Boot: BootDevice{Dev: "hd"},
		},
		Devices: Devices{
			Graphics: Graphics{
				Type:     "spice",
				AutoPort: "yes",
			},
			// disk hard‑coded
			Disk: Disk{
				Type:   "file",
				Device: "disk",
				Driver: Driver{Name: "qemu", Type: "qcow2"},
				Source: Source{File: "run/media/toadie/vm/QEMU/my-guest.qcow2"},
				Target: Target{Dev: "vda", Bus: "virtio"},
			},
			Interface: Interface{
				Type: "network",
				Source: InterfaceSrc{
					Network: "default",
				},
			},
		},
	}
}

// main – call xmlQuery > buildDomain > XML output
func main() {
	// user inputs
	cfg, err := xmlQuery()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Eingabefehler:", err)
		os.Exit(1)
	}

	// Domain build
	domain := buildDomain(cfg)

	// XML build
	out, err := xml.MarshalIndent(domain, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "XML‑Serialisierung fehlgeschlagen:", err)
		os.Exit(1)
	}

	// libvirt XML‑Header
	fmt.Println(xml.Header + string(out))
}