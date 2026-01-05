package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

/* --------------------
	Data structure
-------------------- */
type DomainConfig struct {
	Name     string
	MemMiB   int
	VCPU     int
	Disk     string
	Disksize int
	Network  string
}

type distro struct {
	Name     string `yaml:"name"`
	ID       string `yaml:"id"`
	CPU      int    `yaml:"cpu"`
	RAM      int    `yaml:"ram"`
	Disksize int    `yaml:"disksize"`
}

/* --------------------
	Global vars
-------------------- */
var (
	osList          []distro
	variantByName   map[string]string
)

/* --------------------
	waiting for input
-------------------- */
func readLine(r *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	s, err := r.ReadString('\n')
	return strings.TrimSpace(s), err
}

/* --------------------
	Loading OS-List from yaml
-------------------- */
func loadOSList(p string) error {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("Could not read config: %w", err)
	}
	if err = yaml.Unmarshal(b, &osList); err != nil {
		return fmt.Errorf("YAML could not be parsed: %w", err)
	}
	variantByName = make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}
	return nil
}

/* --------------------
	Print the OS-List
-------------------- */
func chooseDistro(r *bufio.Reader) (distro, error) {
	fmt.Println("\n=== Select an operating system ===")
	sorted := append([]distro(nil), osList...)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})
	for i, d := range sorted {
		fmt.Printf(" %2d) %s  (CPU:%d  RAM:%d MiB  Disk:%d GB)\n",
			i+1, d.Name, d.CPU, d.RAM, d.Disksize)
	}
	line, err := readLine(r, "\nPlease enter a number (or press ENTER for default Arch Linux): ")
	if err != nil {
		return distro{}, err
	}
	idx := 1
	if line != "" {
		if i, e := strconv.Atoi(line); e == nil && i >= 1 && i <= len(sorted) {
			idx = i
		} else {
			return distro{}, fmt.Errorf("Invalid selection")
		}
	}
	return sorted[idx-1], nil
}

/* --------------------
	Form – allows changes to the fields
-------------------- */
func (c *DomainConfig) edit(r *bufio.Reader) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	for {
		fmt.Fprintln(w, "\n=== VM-Config ===\t")
		fmt.Fprintf(w, "[1] Name:\t%s\t[default]\n", c.Name)
		fmt.Fprintf(w, "[2] RAM (MiB):\t%d\t[default]\n", c.MemMiB)
		fmt.Fprintf(w, "[3] vCPU:\t%d\t[default]\n", c.VCPU)
		fmt.Fprintf(w, "[4] Disk-Path:\t%s\t[Enter path for virtual hdd]\n", c.Disk)
		fmt.Fprintf(w, "[5] Disk-Size (GB):\t%d\t[default]\n", c.Disksize)
		fmt.Fprintf(w, "[6] Network:\t%s\t[default]\n", c.Network)
		w.Flush()

		f, _ := readLine(r, "\nSelect or enter to continue: ")
		if f == "" {
			break
		}
		switch f {
		case "1":
			if v, _ := readLine(r, ">> New Name: "); v != "" {
				c.Name = v
			}
		case "2":
			if v, _ := readLine(r, ">> RAM (MiB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					c.MemMiB = i
				}
			}
		case "3":
			if v, _ := readLine(r, ">> vCPU: "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					c.VCPU = i
				}
			}
		case "4":
			if v, _ := readLine(r, ">> Disk path (empty = no disk): "); true {
				c.Disk = v
			}
		case "5":
			if v, _ := readLine(r, ">> Disksize (GB): "); v != "" {
				if i, e := strconv.Atoi(v); e == nil && i > 0 {
					c.Disksize = i
				}
			}
		case "6":
			if v, _ := readLine(r, ">> Network (comma-separated): "); true {
				c.Network = v
			}
		default:
			fmt.Println("Invalid input!")
		}
	}
}

/* --------------------
	Create VM
-------------------- */
func createVM(cfg DomainConfig, variant string) error {
	r := bufio.NewReader(os.Stdin)

	iso, err := readLine(r, "Path to the installation ISO: ")
	if err != nil {
		return err
	}

	// build the vm with virt‑install
	args := []string{
		"--name", cfg.Name,
		"--memory", strconv.Itoa(cfg.MemMiB),
		"--vcpus", strconv.Itoa(cfg.VCPU),
		"--os-variant", variant,
		"--disk", fmt.Sprintf("path=/run/media/toadie/vm/QEMU/%s.qcow2,size=%d,format=qcow2", cfg.Name, cfg.Disksize),
		"--cdrom", iso,
		"--boot", "hd",
		"--print-xml",
	}
	cmd := exec.Command("virt-install", args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("virt-install failed: %w – %s", err, errOut.String())
	}

	// XML output
	xml := out.Bytes()
	fmt.Printf("XML created (%d Bytes).\n", len(xml))
	if len(xml) > 0 {
		fmt.Printf("First lines of the XML:\n%s\n", string(xml[:200]))
	}
	// save XML
	xmlFile := cfg.Name + ".xml"
	if err = os.WriteFile(xmlFile, xml, 0644); err != nil {
		return fmt.Errorf("could not write XML: %w", err)
	}
	abs, _ := filepath.Abs(xmlFile)
	fmt.Printf("XML definition saved under: %s\n", abs)

	// define VM
	if err = exec.Command("virsh", "define", xmlFile).Run(); err != nil {
		return fmt.Errorf("virsh define failed: %w", err)
	}
	fmt.Println("VM successfully registered with libvirt/qemu (not yet started).")
	return nil
}

/* --------------------
	Main workflow
-------------------- */
func runNewVMWorkflow() {
	if err := loadOSList("oslist.yaml"); err != nil {
		log.Fatalf("Error loading OS list: %v", err)
	}
	r := bufio.NewReader(os.Stdin)

	// choosing distribution
	d, err := chooseDistro(r)
	if err != nil {
		log.Fatalf("OS selection failed: %v", err)
	}
	variant := variantByName[d.Name]

	// create basic config from default vaules
	cfg := DomainConfig{
		Name:     d.Name,
		MemMiB:   d.RAM,
		VCPU:     d.CPU,
		Disksize: d.Disksize,
		Disk:     "",
		Network:  "default",
	}

	// manual editing
	cfg.edit(r)

	// VM summary
	fmt.Println("\n=== SUMMARY ===")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%s\n", cfg.Name)
	fmt.Fprintf(w, "RAM (MiB):\t%d\n", cfg.MemMiB)
	fmt.Fprintf(w, "vCPU:\t%d\n", cfg.VCPU)
	fmt.Fprintf(w, "Disk-Path:\t%s\n", cfg.Disk)
	fmt.Fprintf(w, "Disksize:\t%s\n", cfg.Disksize)
	fmt.Fprintf(w, "Network:\t%s\n", cfg.Network)
	w.Flush()

	// create vm
	if err = createVM(cfg, variant); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler beim Erzeugen der VM:", err)
		os.Exit(1)
	}
}

/* --------------------
	dummy function for later
-------------------- */
func dummyTest() { fmt.Println("Yay!") }

/* --------------------
	Mainmenu
-------------------- */
func main() {
	for {
		fmt.Println("\n=== MAINMENU ===")
		fmt.Println("[1] New VM")
		fmt.Println("[2] Dummy function")
		fmt.Println("[0] Exit")
		fmt.Print("Selection: ")

		var choice int
		if _, err := fmt.Scanln(&choice); err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}
		switch choice {
		case 0:
			fmt.Println("Bye!")
			return
		case 1:
			runNewVMWorkflow()
		case 2:
			dummyTest()
		default:
			fmt.Println("Invalid selection!")
		}
	}
}