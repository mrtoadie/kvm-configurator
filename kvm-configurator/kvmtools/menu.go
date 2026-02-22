// kvmtools/menu.go
// last modified: Feb 22 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	// internal
	"configurator/internal/style"
)

// only one entry - the rest is in vmmenu.go
type commandInfo struct {
	Description string
}

var menuMap = map[string]commandInfo{
	"[1]": {"Show VMs"},
	"[q]": {"Back to Mainmenu"},
}

// lightweight dispatcher
func Start(r *bufio.Reader, xmlDir string) {
	for {
		printMenu()
		choice := readChoice(r)

		if choice == "q" {
			return
		}

		switch choice {
		case "1":
			VMMenu(r, xmlDir)
		default:
			fmt.Fprintln(os.Stderr,
				style.Err("Invalid selection"))
		}
	}
}

// print kvm-tools menu
func printMenu() {
	// title
	titleBox := style.Box(20, []string{"KVM-TOOLS"})
	fmt.Println(titleBox)

	// sort menu entrys
	keys := make([]string, 0, len(menuMap))
	for k := range menuMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := style.MustTableToLines(func(w *tabwriter.Writer) {
		// optional: header - example
		//fmt.Fprintln(w, "No.\tDescription")
		//fmt.Fprintln(w, "------\t------------")

		// print all entrys vom menuMap
		for _, k := range keys {
			fmt.Fprintf(w, "%s %s\n", k, menuMap[k].Description)
		}
	})

	// draw the box
	menuBox := style.Box(20, lines)
	fmt.Println(menuBox)
}

// Read input and remove whitespace.
func readChoice(r *bufio.Reader) string {
	fmt.Print(style.PromptMsg("\nSelect: "))
	raw, _ := r.ReadString('\n')
	return strings.TrimSpace(raw)
}
