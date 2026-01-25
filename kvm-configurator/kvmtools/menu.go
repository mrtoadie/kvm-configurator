// kvmtools/menu.go
// last modification: January 25 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	// internal
	"configurator/internal/ui"
)

/* --------------------
	only one entry - the rest is in vmmenu.go
-------------------- */
type commandInfo struct {
	Description string
}

var menuMap = map[string]commandInfo{
	"1": {"Show VMs"},
	"q": {"Back to Mainmenu"},
}

/* --------------------
	lightweight dispatcher
-------------------- */
func Start(r *bufio.Reader) {
	for {
		printMenu()
		choice := readChoice(r)

		if choice == "q" {
			fmt.Println(ui.Colourise("\nBack to Mainmenu", ui.Yellow))
			return
		}

		switch choice {
		case "1":
			VMMenu(bufio.NewReader(os.Stdin))
		default:
			fmt.Fprintln(os.Stderr,
				ui.Colourise("Invalid selection", ui.Red))
		}
	}
}

/* --------------------
	print kvm-tools menu
-------------------- */
func printMenu() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	fmt.Fprintln(w, ui.Colourise("\n=== KVM-TOOLS ===", ui.Blue))
	for key, info := range menuMap {
		fmt.Fprintf(w, "%s\t%s\n", key, info.Description)
	}
	w.Flush()
}

/* --------------------
	Read input and remove whitespace.
-------------------- */
func readChoice(r *bufio.Reader) string {
	fmt.Print(ui.Colourise("\nSelect: ", ui.Yellow))
	raw, _ := r.ReadString('\n')
	return strings.TrimSpace(raw)
}
// EOF