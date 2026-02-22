// kvmtools/vmmenu.go
// last modified: Feb 22 2026
package kvmtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	// internal
	"configurator/internal/style"
)

// fetchAllVMs -calls `virsh list --all` and returns []*VMInfo
func fetchAllVMs() ([]*VMInfo, error) {
	if _, lookErr := exec.LookPath("virsh"); lookErr != nil {
		return nil, fmt.Errorf("virsh not found – Please check PATH")
	}
	out, err := exec.Command("virsh", "list", "--all").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("virsh call failed: %v\n%s", err, out)
	}
	return parseVMs(out)
}

// parseVMs – converts raw Virsh output to []*VMInfo
func parseVMs(raw []byte) ([]*VMInfo, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	var vms []*VMInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header/divider line/empty lines
		if strings.HasPrefix(line, "Id") ||
			strings.HasPrefix(line, "---") ||
			line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue // malformed – ignore
		}

		id := fields[0]                                    // may be "-"
		rawStat := fields[len(fields)-1]                   // last column
		stat := style.NormalizeStatus(rawStat)             // canonical
		name := strings.Join(fields[1:len(fields)-1], " ") // everything in between

		vms = append(vms, &VMInfo{Id: id, Name: name, Stat: stat})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error parsing virsh output: %v", err)
	}
	return vms, nil
}

// sortVMsAlphabetically – sorted by name (case‑insensitive)
func sortVMsAlphabetically(vms []*VMInfo) []*VMInfo {
	sorted := make([]*VMInfo, len(vms))
	copy(sorted, vms)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})
	return sorted
}

// printVMTable – prints the VM list formatted
func printVMTable(vms []*VMInfo) {
	fmt.Println(style.BoxCenter(51, []string{"AVALABLE VIRTUAL MACHINES"}))
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "No.\tName\tState")
		for i, vm := range vms {
			fmt.Fprintf(w, "%d\t%s\t%s\n", i+1, vm.Name, vm.Stat)
		}
		w.Flush()
	})
	fmt.Print(style.Box(51, lines))
}

// pickAction – only shows permitted actions for the respective status
func pickAction(r *bufio.Reader, vm *VMInfo) Action {
	actions := []struct {
		Key   string
		Desc  string
		Cmd   Action
		Check func(*VMInfo) bool // true > allowed
	}{
		{"1", "Start", ActStart, func(v *VMInfo) bool { return v.Stat != "running" }},
		{"2", "Restart", ActReboot, func(v *VMInfo) bool { return v.Stat == "running" }},
		{"3", "Shutdown", ActShutdown, func(v *VMInfo) bool { return v.Stat == "running" }},
		{"4", "Force-Shutdown", ActDestroy, func(v *VMInfo) bool { return v.Stat == "running" }},
		{"5", "Disk-Operations", ActDiskOps, func(v *VMInfo) bool { return true }},
		{"6", "Rename VM", ActRename, func(v *VMInfo) bool { return true }},
		{"0", "Undefine", ActDelete, func(v *VMInfo) bool { return v.Stat == "shut off" }},
		{"q", "Back to VM overview", "", nil},
	}

	// print actions menu
	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Action\tDescription")
		for _, a := range actions {
			if a.Check != nil && !a.Check(vm) {
				continue
			}
			fmt.Fprintf(w, "%s\t%s\n", a.Key, a.Desc)
		}
		w.Flush()
	})
	fmt.Print(style.Box(51, lines))

	fmt.Print(style.PromptMsg("\nSelect action (or q to exit): "))
	choiceRaw, _ := r.ReadString('\n')
	choice := strings.TrimSpace(choiceRaw)

	for _, a := range actions {
		if choice == a.Key && (a.Check == nil || a.Check(vm)) {
			return a.Cmd
		}
	}
	fmt.Fprintln(os.Stderr, style.Err("Invalid selection"))
	return ""
}

// runVMAction – executes the selected virsh command
func runVMAction(action Action, vmName string) error {
	if action == "" {
		return nil // abort / q
	}
	cmd := exec.Command("virsh", string(action), vmName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// VMMenu – public entry point
// xmlDir: Path in which the libvirt XML files are located (e.g. "./xml")
func VMMenu(r *bufio.Reader, xmlDir string) {
	for {
		// fetch all VMs
		vms, err := fetchAllVMs()
		if err != nil {
			fmt.Fprintln(os.Stderr,
				style.Colourise("Error reading the VM list: "+err.Error(),
					style.ColRed))
			return
		}
		if len(vms) == 0 {
			fmt.Println(style.Err("No VMs found"))
			return
		}

		// sort and print
		sorted := sortVMsAlphabetically(vms)
		printVMTable(sorted)

		// make selection
		fmt.Print(style.PromptMsg("\nSelect VM number (or q to exit): "))
		choiceRaw, _ := r.ReadString('\n')
		choice := strings.TrimSpace(choiceRaw)
		if choice == "q" || choice == "quit" {
			return
		}
		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(sorted) {
			fmt.Fprintln(os.Stderr,
				style.Err("Invalid selection"))
			continue
		}
		selected := sorted[idx-1]

		// Determine action
		action := pickAction(r, selected)
		if action == "" {
			continue // user canceled or invalid input
		}

		if action == ActDiskOps {
			// Start Disk Ops submenu (only pass VM name)
			if err := DiskOpsMenu(r, selected.Name); err != nil {
				fmt.Fprintln(os.Stderr, style.Colourise(err.Error(), style.ColRed))
			}
			continue
		}

		if action == ActRename {
			if err := RenameVM(r, selected.Name, xmlDir); err != nil {
				fmt.Fprintln(os.Stderr, style.Colourise(err.Error(), style.ColRed))
			}
			//back to the VM overview
			continue
		}

		// run – special case “Undefine + Disk Cleanup”
		if action == ActDelete {
			if err := deleteVMWithDisks(r, selected.Name, xmlDir); err != nil {
				fmt.Fprintln(os.Stderr, style.Colourise(err.Error(), style.ColRed))
			}
		} else {
			if err := runVMAction(action, selected.Name); err != nil {
				fmt.Fprintln(os.Stderr, style.Colourise(err.Error(), style.ColRed))
			} else {
				fmt.Println(style.Ok("Action successfully completed"))
			}
		}
	}
}

// deleteVMWithDisks – undefine + optionales Disk‑Cleanup
func deleteVMWithDisks(r *bufio.Reader, vmName, xmlDir string) error {
	// undefine
	if err := runVMAction(ActDelete, vmName); err != nil {
		return err
	}
	fmt.Printf("\nVM %s became undefined.\n", vmName)

	// determine disk paths (XML > fallback virsh)
	var diskPaths []string
	xmlPath := filepath.Join(xmlDir, vmName+".xml")
	if paths, err := GetDiskPathsFromXML(xmlPath); err == nil && len(paths) > 0 {
		diskPaths = paths
	} else {
		if paths, err2 := GetDiskPathsViaVirsh(vmName); err2 == nil {
			diskPaths = paths
		}
	}

	if len(diskPaths) == 0 {
		fmt.Println("No hard drives found to delete.")
		return nil
	}

	ok, err := AskYesNo(r,
		fmt.Sprintf("Should %d disk files really be deleted?", len(diskPaths)))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Disk deletion aborted.")
		return nil
	}

	// remove files
	var failures []string
	for _, p := range diskPaths {
		if err := os.Remove(p); err != nil {
			failures = append(failures, fmt.Sprintf("%s (%v)", p, err))
		} else {
			fmt.Printf("%s deleted.\n", p)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("Some files could not be deleted: %s",
			strings.Join(failures, "; "))
	}
	fmt.Println("All associated hard drives of " + vmName + " have been successfully removed.")
	return nil
}

// EOF
