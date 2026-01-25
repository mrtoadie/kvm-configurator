package kvmtools

type Action string

const (
	ActStart    Action = "start"
	ActReboot   Action = "reboot"
	ActShutdown Action = "shutdown"
	ActDestroy  Action = "destroy"
)
