// utils/status.go
// last modification: January 25 2026
package utils

import "strings"

/* --------------------
	NormalizeStatus converts German/English VM‑states to the internal canonical forms
-------------------- */
func NormalizeStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "running", "laufend":
		return "running"
	case "shut", "off", "shutoff", "ausgeschaltet":
		return "shut off"
	default:
		// unknown – keep as‑is so we never treat it as “running”
		return s
	}
}
