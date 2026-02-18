// kvmtools/action.go
// last modification: Feb 16 2026
package kvmtools

type Action string

const (
	ActStart    Action = "start"
	ActReboot   Action = "reboot"
	ActShutdown Action = "shutdown"
	ActDestroy  Action = "destroy"
	ActDelete		Action = "undefine"
	ActDiskOps 	Action = "diskops"
)
// EOF