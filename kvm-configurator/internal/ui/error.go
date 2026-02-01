// internal/ui/error.go
// last modification: January 26 2026
package ui

import (
	"fmt"
	"os"
	"errors"
)

var (
	// General configuration issues
	ErrConfigMissing   = errors.New("Configuration file not found")
	ErrConfigInvalid   = errors.New("Configuration file invalid or corrupt") // not used yet
	ErrWorkDirInvalid  = errors.New("Cannot resolve work directory")

	// Prerequisite issues
	ErrVirtInstallMissing = errors.New("„virt-install“ not in $PATH") // not used yet
	ErrVirshMissing       = errors.New("“virsh” not in $PATH") // not used yet

	// VM creation issues
	ErrISONotFound      = errors.New("ISO file not accessible") // not used yet
	ErrDiskCreationFail = errors.New("Disk argument could not be constructed") // not used yet
	ErrVirtInstallFail  = errors.New("virt-install failed") // not used yet
	ErrVirshDefineFail  = errors.New("virsh define failed") // not used yet
	ErrVMCreationFail		= errors.New("VM creation failed")

	//
	ErrSelection	= errors.New("Invalid Selection!")
)

/* --------------------
	Helper function: Format and output errors with context
	Report outputs a nicely colored error.
	* ctx* – short context text (e.g., “VM creation”)
	* err* – the actual error (can be one of the predefined ones)
-------------------- */
func Report(ctx string, err error) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf("%s: %v", ctx, err)
	fmt.Fprintln(os.Stderr, Colourise(msg, Red))
}

/* --------------------
	Uniform error message + exit
	example: ui.Fatal(err, "Error loading file-config")
-------------------- */
func Fatal(err error, ctx string) {
	if err != nil {
		Report(ctx, err)
		os.Exit(1)
	}
}

func WarnSoft(err error, ctx string) {
	if err != nil {
			msg := fmt.Sprintf("%s %v", ctx, err)
	fmt.Fprintln(os.Stderr, Colourise(msg, Yellow))
	}
}