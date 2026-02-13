// Version 1.0.9
// Autor: 	MrToadie
// GitHub: 	https://github.com/mrtoadie/
// Repo: 		https://github.com/mrtoadie/kvm-configurator
// License: MIT
// last modification: Feb 13 2026
package main

import (
	"bufio"
	//"errors"
	"fmt"
	"os"

	// internal
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/ui"
	"configurator/internal/utils"
	"configurator/kvmtools"
)

// MAIN
func main() {
	// [Modul: config] validates if (virt‑install, virsh) is installed
	if err := config.EnsureAll(config.CmdVirtInstall, config.CmdVirsh); err != nil {
		// exit if virt-install or virsh are not found
		utils.RedError("virt-install not found", "verify $PATH", err)
		os.Exit(1)
	}
	// for debug only
	//ui.Success("✅ Prereqs OK", "virt-install & virsh FOUND!", "")



	// -----------------------------------------------------------------
	// 1️⃣ Verify that the configuration file exists in $HOME/.config/kvm-configurator
	// -----------------------------------------------------------------
	if ok, err := config.Exists(); err != nil || !ok {
		// Give the user a friendly hint where the file should be.
		utils.RedError(
			"Configuration not found",
			"expected at "+config.ConfigFilePath(),
			err,
		)
		os.Exit(1)
	}
/*
	// -----------------------------------------------------------------
	// 2️⃣ Load the YAML from the correct location
	// -----------------------------------------------------------------
	cfg, err := config.LoadAll(config.ConfigFilePath())
	if err != nil {
		utils.RedError("Failed to load configuration", "", err)
		os.Exit(1)
	}

*/



	// Get everything from the YAML file in one go
	cfg, err := config.LoadAll(config.ConfigFilePath())
	//cfg, err := config.LoadAll(config.FileConfig) // One call, one result
	if err != nil {
		utils.RedError("Failed to load configuration", "", err)
		os.Exit(1)
	}

	// determine xml save path
	xmlDir := cfg.XmlDir
	// determine working directory (ISO folder)
	workDir := cfg.IsoPath
	if workDir == "" {
		// fallback dir
		if cwd, e := os.Getwd(); e == nil {
			workDir = cwd
		}
	}

	// // [Modul: config]
	osList := cfg.OSList     // the list of supported distributions
	defaults := cfg.Defaults // global specifications (disk path, size, ...)
	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	// main menu loop
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(utils.Colourise("\n=== MAIN MENU ===", utils.ColorBlue))
		fmt.Println(utils.Box(20, []string{"KVM-CONFIGURATOR"}))

		fmt.Println(utils.Box(20, []string{
			"[1] New VM",
			"[2] KVM-Tools",
			"[h] Help",
			"[0] Exit",
		}))
		fmt.Print(utils.Colourise(" Selection: ", utils.ColorYellow))

		var sel string
		if _, err := fmt.Scanln(&sel); err != nil {
			// Invalid entry → we simply ask again
			continue
		}

		switch sel {
		case "0":
			fmt.Println("Bye!")
			return

		case "1":
			// The actual “launch sequence” for a new VM
			if err := engine.RunNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				cfg.IsoPath,
				cfg.XmlDir,
			); err != nil {
				// error
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n",
					utils.ColorRed, err, utils.ColorReset)
			}
		case "2":
			//kvmtools.Start(r)
			kvmtools.Start(bufio.NewReader(os.Stdin), xmlDir)
		case "h":
			ui.PrintHelp()
		default:
			fmt.Println(utils.Colourise("\nInvalid selection!", utils.ColorRed))
		}
	}
}
// EOF
