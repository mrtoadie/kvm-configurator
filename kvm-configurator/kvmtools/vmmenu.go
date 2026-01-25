// kvmtools/vmmenu.go
// last modification: January 25 2026
package kvmtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sort"
	// internal
	"configurator/internal/ui"
	"configurator/internal/utils"
)

/* --------------------
	fetchAllVMs runs `virsh list --all` and returns a slice of *VMInfo
	The order is the raw order from virsh; sorting is done later on demand
-------------------- */
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

/* --------------------
	parseVMs turns the raw virsh output into a slice of *VMInfo
-------------------- */
func parseVMs(raw []byte) ([]*VMInfo, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	var vms []*VMInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header / separator / empty lines
		if strings.HasPrefix(line, "Id") ||
			strings.HasPrefix(line, "---") ||
			line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue // malformed line – ignore
		}

		id := fields[0]                                 // may be "-"
		rawStat := fields[len(fields)-1]                // last column
		stat := utils.NormalizeStatus(rawStat)          // canonical form
		name := strings.Join(fields[1:len(fields)-1], " ") // everything in‑between

		vms = append(vms, &VMInfo{Id: id, Name: name, Stat: stat})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error parsing virsh output: %v", err)
	}
	return vms, nil
}

/* --------------------
	sortVMsAlphabetically returns a new slice sorted by VM name (case‑insensitive)
-------------------- */
func sortVMsAlphabetically(vms []*VMInfo) []*VMInfo {
	sorted := make([]*VMInfo, len(vms))
	copy(sorted, vms)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})
	return sorted
}

/* --------------------
	printVMTable prints a slice of VMs. The slice is already ordered the way we want
-------------------- */
func printVMTable(vms []*VMInfo) {
	w := utils.NewTabWriter()
	fmt.Fprintln(w, ui.Colourise("\n=== Available VMs ===", ui.Blue))
	fmt.Fprintln(w, "No.\tName\tState")
	for i, vm := range vms {
		fmt.Fprintf(w, "%d\t%s\t%s\n", i+1, vm.Name, vm.Stat)
	}
	w.Flush()
}

/* --------------------
	pickAction shows only the actions that make sense for the given VM status
-------------------- */
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
		{"q", "Back to VM overview", "", nil},
	}

	// Show the allowed actions
	fmt.Println()
	w := utils.NewTabWriter()
	fmt.Fprintln(w, ui.Colourise("Action\tDescription", ui.Blue))
	for _, a := range actions {
		if a.Check != nil && !a.Check(vm) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", a.Key, a.Desc)
	}
	w.Flush()

	fmt.Print(ui.Colourise("\nSelect action (or q to exit): ", ui.Yellow))
	choiceRaw, _ := r.ReadString('\n')
	choice := strings.TrimSpace(choiceRaw)

	for _, a := range actions {
		if choice == a.Key && (a.Check == nil || a.Check(vm)) {
			return a.Cmd
		}
	}
	fmt.Fprintln(os.Stderr, ui.Colourise("Invalid selection", ui.Red))
	return ""
}

/* --------------------
	runVMAction executes the chosen virsh command
-------------------- */
func runVMAction(action Action, vmName string) error {
	if action == "" {
		return nil // user aborted or chose “q”
	}
	cmd := exec.Command("virsh", string(action), vmName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

/* --------------------
	VMMenu – public entry point, called from menu.Start()
-------------------- */
func VMMenu(r *bufio.Reader) {
	for {
		// fetch all VMs
		vms, err := fetchAllVMs()
		if err != nil {
			fmt.Fprintln(os.Stderr,
				ui.Colourise("Error reading the VM list: "+err.Error(),
					ui.Red))
			return
		}
		if len(vms) == 0 {
			fmt.Println(ui.Colourise("No VMs found", ui.Yellow))
			return
		}
		// sort alphabetically
		sorted := sortVMsAlphabetically(vms)

		// display table
		printVMTable(sorted)

		// 4️⃣ choose VM
		fmt.Print(ui.Colourise("\nSelect VM number (or q to exit): ", ui.Yellow))
		choiceRaw, _ := r.ReadString('\n')
		choice := strings.TrimSpace(choiceRaw)
		if choice == "q" || choice == "quit" {
			return
		}
		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(sorted) {
			fmt.Fprintln(os.Stderr,
				ui.Colourise("Invalid selection", ui.Red))
			continue
		}
		selected := sorted[idx-1]

		// pick & run action
		action := pickAction(r, selected)
		if action == "" {
			continue // user cancelled or invalid choice
		}
		if err := runVMAction(action, selected.Name); err != nil {
			fmt.Fprintln(os.Stderr, ui.Colourise(err.Error(), ui.Red))
		} else {
			fmt.Println(ui.Colourise("Action successfully completed", ui.Green))
		}
	}
}
// EOF