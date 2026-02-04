// internal/constants.go
// last modification: February 04 2026
package config

const (
    CmdVirtInstall  = "virt-install"
    CmdVirsh        = "virsh"
    FileConfig      = "oslist.yaml"
)
var PrereqCommands = []string{CmdVirtInstall, CmdVirsh}