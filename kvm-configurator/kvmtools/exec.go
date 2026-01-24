// kvmtools/exec.go
// last modification: January 24 2026
package kvmtools

import (
	"fmt"
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// cmdsMap definiert die Zuordnung von Menü‑Key → Virsh‑Befehl.
// Die Struktur ist exakt dieselbe wie in deinem alten tools.go,
// sodass du keine Änderungen an den eigentlichen Befehlen brauchst.
var cmdsMap = map[string]CmdInfo{
	"1": {Command: "list --all", Description: "List all VMs", Interactive: true},
	"2": {Command: "list", Description: "List running VMs", Interactive: false},
	"3": {Command: "reboot", Description: "Restart selected VM", Interactive: true},
	"4": {Command: "shutdown", Description: "Shutdown selected VM", Interactive: true},
	"5": {Command: "destroy", Description: "Force-stop selected VM", Interactive: true},
	"6": {Command: "nodeinfo", Description: "Host system infos", Interactive: false},
	"7": {Command: "start", Description: "start selected VM", Interactive: true},
}

// CmdInfo hält Metadaten zu jedem Befehl.
type CmdInfo struct {
	Command     string
	Description string
	Interactive bool
}

// Run ist das öffentliche Interface, das vom Menü aufgerufen wird.
func Run(choice string) error {
	info, ok := cmdsMap[choice]
	if !ok {
		return fmt.Errorf("Invalid selection!")
	}

	// Interaktive Befehle benötigen zusätzliche Dialoge (VM auswählen)
	if strings.HasPrefix(info.Command, "list") && info.Interactive {
		return handleVirshListAndStart()
	}
	if info.Interactive {
		return handleVirshAction(info.Command)
	}
	// Direkter Aufruf ohne weitere Benutzereingaben
	return run(info.Command)
}

/* ---------- Hilfsfunktionen (identisch zu deinem alten tools.go) ---------- */

func run(cmdStr string) error {
	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleVirshListAndStart() error {
	out, err := exec.Command("virsh", "list", "--all").CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh call failed: %v\n%s", err, out)
	}
	vmMap, err := parseVirshList(out)
	if err != nil {
		return err
	}
	if len(vmMap) == 0 {
		fmt.Println("No VMs found")
		return nil
	}
	fmt.Println("\nFound VMs:")
	for i, n := range vmMap {
		fmt.Printf("%d)\t%s\n", i, n)
	}
	fmt.Print("\nNumber of the VM to be started (or blank to cancel): ")
	var line string
	fmt.Scanln(&line)
	line = strings.TrimSpace(line)
	if line == "" {
		fmt.Println("Aborted – no VM started.")
		return nil
	}
	num, err := strconv.Atoi(line)
	if err != nil || vmMap[num] == "" {
		return fmt.Errorf("Invalid selection!")
	}
	selectedVM := vmMap[num]
	fmt.Printf("Start VM \"%s\" …\n", selectedVM)
	startCmd := exec.Command("virsh", "start", selectedVM)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("Start failed: %v", err)
	}
	fmt.Println("\x1b[32mVM successfully started\x1b[0m")
	return nil
}

func handleVirshAction(action string) error {
	out, err := exec.Command("virsh", "list", "--all").CombinedOutput()
	if err != nil {
		return fmt.Errorf("virsh call failed: %v\n%s", err, out)
	}
	vmMap, err := parseVirshList(out)
	if err != nil {
		return err
	}
	if len(vmMap) == 0 {
		fmt.Println("No VMs found")
		return nil
	}
	fmt.Println("\nFound VMs:")
	for i, n := range vmMap {
		fmt.Printf("%d)\t%s\n", i, n)
	}
	fmt.Printf("\nNumber of the VM to %s (or blank to cancel): ", action)
	var line string
	fmt.Scanln(&line)
	line = strings.TrimSpace(line)
	if line == "" {
		fmt.Println("Abort – no action performed")
		return nil
	}
	num, err := strconv.Atoi(line)
	if err != nil || vmMap[num] == "" {
		return fmt.Errorf("Invalid selection!")
	}
	selectedVM := vmMap[num]

	fmt.Printf("%s VM \"%s\" …\n", strings.Title(action), selectedVM)
	cmd := exec.Command("virsh", action, selectedVM)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %v", strings.Title(action), err)
	}
	fmt.Printf("✅ VM \"%s\" successfully %sed.\n", selectedVM, action)
	return nil
}

// parseVirshList bleibt unverändert – liefert eine Map[Index]→VM‑Name.
func parseVirshList(raw []byte) (map[int]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	vmMap := make(map[int]string)
	idx := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Id") || strings.HasPrefix(line, "---") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := strings.Join(fields[1:len(fields)-1], " ")
		vmMap[idx] = name
		idx++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error parsing virsh output: %v", err)
	}
	return vmMap, nil
}