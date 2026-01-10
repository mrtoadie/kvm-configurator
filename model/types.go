package model

// DomainConfig holds the values that will be passed to virt‑install.
type DomainConfig struct {
	Name        string
	MemMiB      int
	VCPU        int
	Disk        string
	DiskSize    int
	Network     string
	ISO         string
	NestedVirt  string // e.g. "vmx" or "svm"
	BootOrder   string
}

// Distro represents one entry from the YAML “oslist”.
type Distro struct {
	Name       string `yaml:"name"`
	ID         string `yaml:"id"`
	CPU        int    `yaml:"cpu"`
	RAM        int    `yaml:"ram"`
	DiskSize   int    `yaml:"disksize"`
	DiskPath   string `yaml:"disk_path"`
	NestedVirt string `yaml:"nvirt"` // optional per‑distro default
}

// Defaults are the global fall‑backs defined in the YAML file.
type Defaults struct {
	DiskPath   string `yaml:"disk_path"`
	DiskSize   int    `yaml:"disksize"`
	NestedVirt string `yaml:"nvirt"` // optional global default
}