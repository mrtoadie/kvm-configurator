// ui/disks.go
// last modification: Feb 08 2026
package ui

import (
	"fmt"
	"strconv"
	"bufio"
	"configurator/internal/utils"
	"configurator/internal/model"
)

func PromptAddDisk(r *bufio.Reader, cfg *model.DomainConfig, defaultDiskPath string) error {
	fmt.Println(utils.BoxCenter(55, []string{"=== ADD DISK ==="}))

	// disk name
	name, _ := ReadLine(r, utils.Colourise("Disk name (e.g. swap, data, backup): ", utils.ColorYellow))
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
	path, _ := ReadLine(r, utils.Colourise(prompt, utils.ColorYellow))
	if path == "" {
		// User has not entered anything > using default
		path = suggestedPath
	}

	// size
	sizeStr, _ := ReadLine(r, utils.Colourise("Size in GiB (0 = default): ", utils.ColorYellow))
	size, _ := strconv.Atoi(sizeStr)

	// Bus‑Typ (optional)
	bus, _ := ReadLine(r, utils.Colourise("Bus (virtio|scsi|sata|usb, default virtio): ", utils.ColorYellow))
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
	utils.Success("Disk", name, "added")
	return nil
}