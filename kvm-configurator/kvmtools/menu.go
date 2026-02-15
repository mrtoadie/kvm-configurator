// kvmtools/menu.go
// last modification: Feb 15 2026
package kvmtools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	//"text/tabwriter"
	// internal
	"configurator/internal/utils"
)

// only one entry - the rest is in vmmenu.go
type commandInfo struct {
	Description string
}

var menuMap = map[string]commandInfo{
	"1": {"Show VMs"},
	"q": {"Back to Mainmenu"},
}

// lightweight dispatcher
func Start(r *bufio.Reader, xmlDir string) {
    for {
        printMenu()
        choice := readChoice(r)

        if choice == "q" {
            fmt.Println(utils.Colourise("\nBack to Mainmenu", utils.ColorYellow))
            return
        }

        switch choice {
        case "1":
            VMMenu(r, xmlDir)
        default:
            fmt.Fprintln(os.Stderr,
                utils.Colourise("Invalid selection", utils.ColorRed))
        }
    }
}

// print kvm-tools menu
func printMenu() {
	fmt.Println(utils.Box(20, []string{"KVM-TOOLS"}))
	w := utils.NewTabWriter()
	for key, info := range menuMap {
		fmt.Fprintf(w, "%s\t%s\n", key, info.Description)
	}
	w.Flush()
}

// Read input and remove whitespace.
func readChoice(r *bufio.Reader) string {
	fmt.Print(utils.Colourise("\nSelect: ", utils.ColorYellow))
	raw, _ := r.ReadString('\n')
	return strings.TrimSpace(raw)
}
// EOF