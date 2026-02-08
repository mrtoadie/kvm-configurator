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

func PromptAddDisk(r *bufio.Reader, cfg *model.DomainConfig) error {
	fmt.Println(utils.BoxCenter(55, []string{"=== ADD DISK ==="}))

	// disk name
	name, _ := ReadLine(r, utils.Colourise("Disk name (e.g. system, data, backup): ", utils.ColorYellow))
	if name == "" {
		return nil // Abbruch
	}

	// path (directory or complete file name)
	path, _ := ReadLine(r, utils.Colourise("Disk path (dir or full file): ", utils.ColorYellow))

	// size
	sizeStr, _ := ReadLine(r, utils.Colourise("Size in GiB (0 = default): ", utils.ColorYellow))
	size, _ := strconv.Atoi(sizeStr)

	// Bus‑Typ (optional)
	bus, _ := ReadLine(r, utils.Colourise("Bus (virtio|scsi|ide, default virtio): ", utils.ColorYellow))
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