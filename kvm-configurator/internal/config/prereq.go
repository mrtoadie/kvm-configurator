// config/prereq.go
// last modified: Feb 22 2026
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CommandMissingError
type CommandMissingError struct {
	Cmd string
}

func (e *CommandMissingError) Error() string {
	return fmt.Sprintf("command %q not found in PATH", e.Cmd)
}

// RequireCommand checks whether an executable program is located in $PATH
func RequireCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return &CommandMissingError{Cmd: name}
	}
	return nil
}

// EnsureAll performs several checks in sequence
func EnsureAll(commands ...string) error {
	for _, cmd := range commands {
		if err := RequireCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

// check if config file exists
func Exists() (bool, error) {
	//var fileName = ".config/kvm-configurator/"
  home, err := os.UserHomeDir()
  if err != nil {
    return false, err
  }
	//path := filepath.Join(home, fileName)
	path := filepath.Join(home, ConfigDir, FileConfig)
  info, err := os.Stat(path)
  if err == nil {
    return !info.IsDir(), nil
  }
  return false, err
}