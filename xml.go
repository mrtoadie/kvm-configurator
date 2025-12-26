package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

// Root element
type Domain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`

	Name   string   `xml:"name"`
	Memory Memory   `xml:"memory"`
	VCPU   VCPU     `xml:"vcpu"`
	OS     OS       `xml:"os"`
	Devices Devices `xml:"devices"`
}

// Sub‑elements
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
	Disk      Disk      `xml:"disk"`
	Interface Interface `xml:"interface"`
	Graphics  Graphics  `xml:"graphics"`
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
	File string `xml:"file,attr,omitempty"` // for file‑backed disks
	// network, hostdev, etc. can be added later
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
	// bridge, dev, etc. can be added later
}

type Graphics struct {
	Type    string `xml:"type,attr"`
	AutoPort string `xml:"autoport,attr,omitempty"`
}

func main() {
	d := Domain{
		Type: "kvm",
		Name: "my-guest",
		Memory: Memory{
			Unit:  "MiB",
			Value: 1024,
		},
		VCPU: VCPU{
			Placement: "static",
			Value:     2,
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
			Graphics: Graphics{
				Type:    "spice",
				AutoPort: "yes",
			},
		},
	}

	output, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal XML: %v\n", err)
		os.Exit(1)
	}

	// Add the XML header – libvirt expects it.
	fmt.Println(xml.Header + string(output))
}