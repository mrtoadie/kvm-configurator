// kvmtools/kvmmodel.go
// last modified: Feb 26 2026
package kvmtools

type Action string

const (
	ActStart    Action = "start"
	ActReboot   Action = "reboot"
	ActShutdown Action = "shutdown"
	ActDestroy  Action = "destroy"
	ActDelete		Action = "undefine"
	ActDiskOps 	Action = "diskops"
	ActRename		Action = "domrename"
)

/* --------------------
VMInfo holds the minimal information we need for the menus
-------------------- */
type VMInfo struct {
	Id   string // empty (“-”) when the VM is stopped
	Name string
	Stat string // canonical: "running" or "shut off"
}