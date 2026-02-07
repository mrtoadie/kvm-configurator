// Version 1.0.7
// Autor: 	MrToadie
// GitHub: 	https://github.com/mrtoadie/
// Repo: 		https://github.com/mrtoadie/kvm-configurator
// License: MIT
// last modification: Feb 07 2026
package main

import (
	"bufio"
	//"errors"
	"fmt"
	"os"
	// internal
	"configurator/internal/config"
	"configurator/internal/engine"
	"configurator/internal/utils"
	"configurator/internal/ui"
	"configurator/kvmtools"
)

// MAIN
/*
func main() {
	// [Modul: config] validates if (virt‑install, virsh) is installed
	if err := config.EnsureAll(config.CmdVirtInstall, config.CmdVirsh); err != nil {
			utils.RedError("virt-install not found", "verify $PATH", err)
			os.Exit(1)
	}
	// for debug only
	//ui.Success("✅ Prereqs OK", "virt-install & virsh FOUND!", "")

	// [Modul: config] check if config file exists or invalid
	ok, err := config.Exists()
  if err != nil {
    utils.RedError("Configuration file invalid or corrupt", "", err)
		//os.Exit(1)
  }
  if ok {
		// program starts		
  } else {
		utils.RedError("File does not exist", "verify $PATH", err)
  }
	
	// [Modul: config] loads File‑Config (isopath)
	fp, err := config.LoadFilePaths(config.FileConfig)
	if errors.Is(err, os.ErrNotExist) {
    utils.RedError("Configuration file not found ", ">", err)				
    //os.Exit(1)
	}

	workDir, err := config.ResolveWorkDir(fp)
	if errors.Is(err, os.ErrNotExist) {
		utils.RedError("Cannot resolve work directory", "verify $PATH", err)
    //os.Exit(1)
	}
	
	// [Modul: config] loading global Defaults
	osList, defaults, err := config.LoadOSList(config.FileConfig)
	if errors.Is(err, os.ErrNotExist) {
		utils.RedError("Configuration file not found", ">", err)
    //os.Exit(1)
	}

	variantByName := make(map[string]string, len(osList))
	for _, d := range osList {
		variantByName[d.Name] = d.ID
	}

	r := bufio.NewReader(os.Stdin)
	for {
		//fmt.Println(utils.Colourise("\n=== MAIN MENU ===", utils.ColorBlue))
		fmt.Println(utils.BoxCenter(20, []string{"KVM-CONFIGURATOR"}))
		    fmt.Println(utils.Box(20, []string{
        "[1] New VM",
        "[2] KVM-Tools",
        "[h] Help",
				"[0] Exit",
    }))
		//fmt.Println("[1] New VM")
		//fmt.Println("[2] KVM-Tools")
		//fmt.Println("[h] Help")
		//fmt.Println("[0] Exit")
		fmt.Print(utils.Colourise(" Selection: ", utils.ColorYellow))

		var sel string
		if _, err := fmt.Scanln(&sel); err != nil {
			continue
		}
		switch sel {
		case "0":
			fmt.Println("Bye!")
			return
		case "1":
			if err := engine.RunNewVMWorkflow(
				r,
				osList,
				defaults,
				variantByName,
				workDir,
				fp,
			); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n", utils.ColorRed, err, utils.ColorReset)
			}
		case "2":
			kvmtools.Start(r)
		case "h":
			ui.PrintHelp()
		default:
			fmt.Println(utils.Colourise("\nInvalid selection!", utils.ColorRed))
		}
	}
}
*/

func main() {
	// [Modul: config] validates if (virt‑install, virsh) is installed
	if err := config.EnsureAll(config.CmdVirtInstall, config.CmdVirsh); err != nil {
		// exit if virt-install or virsh are not found
		utils.RedError("virt-install not found", "verify $PATH", err)
		os.Exit(1)
	}
	// for debug only
	//ui.Success("✅ Prereqs OK", "virt-install & virsh FOUND!", "")

	// Get everything from the YAML file in one go
	cfg, err := config.LoadAll(config.FileConfig) // One call, one result
	if err != nil {
		utils.RedError("Failed to load configuration", "", err)
		os.Exit(1)
	}

	// Determine working directory (ISO folder)
	workDir := cfg.IsoPath
	if workDir == "" {
		// fallback dir
		if cwd, e := os.Getwd(); e == nil {
			workDir = cwd
		}
	}

	// // [Modul: config] 
	osList   := cfg.OSList		// the list of supported distributions
	defaults := cfg.Defaults	// global specifications (disk path, size, ...)
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
			"[2] KVM‑Tools",
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
				&config.FilePaths{
					Filepaths: struct {
						IsoPath string `yaml:"isopath"`
						XmlDir  string `yaml:"xmlpath"`
					}{
						IsoPath: cfg.IsoPath,
						XmlDir:  cfg.XmlDir,
					},
				},
			); err != nil {
				// error
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n",
					utils.ColorRed, err, utils.ColorReset)
			}

		case "2":
			kvmtools.Start(r)
		case "h":
			ui.PrintHelp()
		default:
			fmt.Println(utils.Colourise("\nInvalid selection!", utils.ColorRed))
		}
	}
}
// EOF
