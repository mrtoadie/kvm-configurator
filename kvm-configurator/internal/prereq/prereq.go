// prereq/prereq.go
package prereq

import (
	"fmt"
	"os"
	"os/exec"
)

/* --------------------
	CommandMissingError
-------------------- */
type CommandMissingError struct {
	Cmd string
}

func (e *CommandMissingError) Error() string {
	return fmt.Sprintf("command %q not found in PATH", e.Cmd)
}

/* --------------------
	RequireCommand checks whether an executable program is located in $PATH
-------------------- */
func RequireCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return &CommandMissingError{Cmd: name}
	}
	return nil
}

/* --------------------
	EnsureAll performs several checks in sequence
-------------------- */
func EnsureAll(commands ...string) error {
	for _, cmd := range commands {
		if err := RequireCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

/* --------------------
	Helper to terminate the program in case of a fatal error
-------------------- */
func FatalIfMissing(err error) {
	if err == nil {
		return
	}
	if missing, ok := err.(*CommandMissingError); ok {
		fmt.Fprintf(os.Stderr,
			"\x1b[31mError: %s is not installed or not in the PATH.\x1b[0m\n"+
				"Please install: (e.g. with `sudo pacman -S %s` or the appropriate package manager).\n",
			missing.Cmd, missing.Cmd)
	} else {
		fmt.Fprintf(os.Stderr, "\x1b[31mUnexpected error: %v\x1b[0m\n", err)
	}
	os.Exit(1)
}