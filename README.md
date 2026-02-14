# kvm-configurator
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) ![GitHub License](https://img.shields.io/github/license/mrtoadie/kvm-configurator) ![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/mrtoadie/kvm-configurator/total)


**kvm-configurator** creates a **virtual machine** and registers it with qemu. In addition, the definition is saved as an **XML** file.

The program can help you if you don't want to or can't use tools like virt-manager. Or if you don't feel like using commands like this:
```bash
virt-install \
  --name guest1-rhel7 \
  --memory 2048 \
  --vcpus 2 \
  --disk size=8 \
  --cdrom /path/to/rhel7.iso \
  --os-variant rhel7
```
**kvm-configurator** is my first ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) project that goes beyond just playing around.

![kvm-configurator-demo](assets/kvm-configurator_demo.gif)


## Features
- **Easy**: Assisted creation of virtual machines
- **Automatoin**: Created VMs are automatically registered (not started) and are immediately ready for use
- **Customizable**: Default values can be customized individually via a YAML file
- **Reuse & backup**: Create VM configurations are also saved as XML files

## Project Structure

```
kvm-configurator/
│
├─ oslist.yaml                # Central YAML configuration (OS list + filepaths)
│
├─ internal/
│   ├─ config/                # Loading and validating YAML data
│   │   └─ config.go
│   ├─ model/                 # Data models & helper logic
│   │   └─ model.go
│   ├─ fileutils/             # File utilities (ListFiles, PromptSelection)
│   │   └─ fileutils.go
│   ├─ engine/                # Core logic: Calling virt-install & XML handling
│   │   └─ engine.go
│   ├─ ui/                    # User interaction (menus, inputs, summary, colours)
│   │   ├─ colours.go         
│   │   ├─ progress.go
│   │   └─ ui.go
│   ├─ utils/                    
│   │   ├─ status.go         
│   │   └─ tabwriter.go
│   └─ prereq/                # Checks whether necessary programs are installed
│       └─ prereq.go
├─ kvmtools/                  # kvm-tools
│   ├─ action.go
│   ├─ menu.go                
│   ├─ vminfo.go
│   └─ vmmenu.go                
│
└─ main.go                    # Entry point, orchestrates the entire workflow
```
## Install
### Arch Linux
```bash
yay -S kvm-configurator
```
### Compiled version
Download the two files oslist.yaml and kvm-config_x.x.
Set kvm-config_x.x as an executable file:
```bash
chmod +x configurator
```
and start the program with 
```bash
./configurator
```

### Tested under
:white_check_mark: [Arch Linux](https://archlinux.org/)

:white_check_mark: [NixOS](https://nixos.org/)

:white_check_mark: [GuideOS](https://guideos.de/) (Debian-based)

:white_check_mark: [Solus](https://getsol.us/)

:white_check_mark: [Ubuntu 25.04 & 25.10](https://ubuntu.com/)

## Release Notes
[Release notes](https://github.com/mrtoadie/kvm-configurator/wiki/Release-Notes)
