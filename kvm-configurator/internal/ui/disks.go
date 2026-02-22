// ui/disks.go
// last modified: Feb 22 2026
package ui

import (
	"bufio"
	"configurator/internal/model"
	"configurator/internal/utils"
	"fmt"
	"os"
	"strconv"
	"configurator/internal/style"
)

func PromptAddDisk(r *bufio.Reader, cfg *model.DomainConfig, defaultDiskPath string) error {
	fmt.Println(style.BoxCenter(55, []string{"=== ADD DISK ==="}))

	// disk name
	name, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Disk name (e.g. swap, data, backup): "))
	if name == "" {
		return nil
	}

	// path (directory or complete file name)
	var suggestedPath string
	if primary := cfg.PrimaryDisk(); primary != nil && primary.Path != "" {
		suggestedPath = primary.Path // Accept the path of the first disk
	} else {
		suggestedPath = defaultDiskPath // global default if no disk exists yet
	}
	prompt := fmt.Sprintf("Disk path (default: %s): ", suggestedPath)
	path, _ := utils.Prompt(r, os.Stdout, style.Colourise(prompt, style.ColYellow))
	if path == "" {
		// User has not entered anything > using default
		path = suggestedPath
	}

	// size
	sizeStr, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Size in GiB (0 = default): "))
	size, _ := strconv.Atoi(sizeStr)

	// Bus‑Typ (optional)
	bus, _ := utils.Prompt(r, os.Stdout, style.PromptMsg("Bus (virtio|scsi|sata|usb, default virtio): "))
	if bus == "" {
		bus = "virtio"
	}

	// Attach DiskSpec to DomainConfig
	cfg.Disks = append(cfg.Disks, model.DiskSpec{
		Name:    name,
		Path:    path,
		SizeGiB: size,
		Bus:     bus,
	})
	style.Success("Disk", name, "added")
	return nil
}
