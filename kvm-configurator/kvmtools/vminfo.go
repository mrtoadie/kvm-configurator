package kvmtools

/* --------------------
VMInfo holds the minimal information we need for the menus
-------------------- */
type VMInfo struct {
	Id   string // empty (“-”) when the VM is stopped
	Name string
	Stat string // canonical: "running" or "shut off"
}
