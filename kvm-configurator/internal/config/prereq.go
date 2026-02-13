// config/prereq.go
// last modification: Feb 07 2026
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

/*
	not good, need improvment
*/
// function to check if config file exists

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


// only checks that oslist.yaml exists in current dir
/*
func Exists() (bool, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	path := filepath.Join(cwd, FileConfig)
	info, err := os.Stat(path)
	
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}*/
// EOF