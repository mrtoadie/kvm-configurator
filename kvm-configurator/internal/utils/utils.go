// utils/utils.go
// last modified: Feb 23 2026
package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	// internal
	"configurator/internal/style"
)

const (
	CancelChoice  = 0  // user typed 0 > abort
	InvalidChoice = -1 // parsing error or out‑of‑range
)

// FILE UTILS

// ListFiles returns the absolute paths of regular files inside dir.
// If the directory cannot be read, the error is propagated to the caller.
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// propagate – let the UI decide what to do
		return nil, fmt.Errorf("cannot read directory %q: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if e.Type().IsRegular() {
			abs, err := filepath.Abs(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("resolve absolute path for %q: %w", e.Name(), err)
			}
			files = append(files, abs)
		}
	}
	return files, nil
}

// PromptSelection shows a numbered menu of the supplied files and returns the
// user’s choice (1‑based). 0 means “cancel”. Errors are returned explicitly.
func PromptSelection(r *bufio.Reader, w io.Writer, files []string) (int, error) {
	if len(files) == 0 {
		return InvalidChoice, fmt.Errorf("no files to select from")
	}

	// Header – use the shared colour constant
	fmt.Fprintln(w, style.Colourise("\n=== Select ISO ===", style.ColBlue))

	for i, f := range files {
		fmt.Fprintf(w, "[%d] %s\n", i+1, filepath.Base(f))
	}
	prompt := style.Colourise("Please enter number (or 0 to cancel): ", style.ColYellow)

	// Re‑use the universal Prompt helper
	line, err := Prompt(r, w, prompt)
	if err != nil {
		return InvalidChoice, err
	}
	choice, err := strconv.Atoi(line)
	if err != nil {
		return InvalidChoice, fmt.Errorf("invalid number %q", line)
	}
	if choice < CancelChoice || choice > len(files) {
		return InvalidChoice, fmt.Errorf("choice %d out of range (0-%d)", choice, len(files))
	}
	return choice, nil
}

// HELPER

/*
ExpandEnvInStruct recursively expands environment variables in all string fields
of structs, slices, maps, and their nested contents using os.ExpandEnv.
*/
func ExpandEnvInStruct(v any) {
	if v == nil {
		return
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	expandValue(val.Elem())
}

func expandValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			if f.CanSet() {
				expandValue(f)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			expandValue(val.Index(i))
		}
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			if v.Kind() == reflect.String {
				newStr := os.ExpandEnv(v.String())
				val.SetMapIndex(k, reflect.ValueOf(newStr))
			} else {
				expandValue(v)
			}
			// Keys are usually static strings, we leave them untouched.
			_ = k
		}
	case reflect.String:
		val.SetString(os.ExpandEnv(val.String()))
	}
}

// INPUT

// Prompt reads a line from r after printing prompt to w.
// Returns the trimmed line and any I/O error (including io.EOF).
func Prompt(r *bufio.Reader, w io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err // propagate EOF / other errors
	}
	return strings.TrimSpace(line), nil
}

// ASK

// -------------------------------------------------------------------
// 1️⃣ ask – ein einheitlicher Prompt‑Wrapper
// -------------------------------------------------------------------
// label   – das eigentliche Prompt‑Text (z. B. "RAM (MiB)")
// defVal  – optionaler Default‑Wert, der im Prompt angezeigt wird
// r,w     – Reader/Writer, damit wir später leicht mocken können
//
// Rückgabe: (Antwort, error).  Leere Antwort → "" (der Aufrufer entscheidet,
// ob er den Default übernehmen will oder nicht).
// -------------------------------------------------------------------
func Ask(r *bufio.Reader, w io.Writer, label string, defVal string) (string, error) {
	// Baue den Prompt‑String zusammen:
	//   ">> RAM (MiB) (default: 2048): "
	prompt := fmt.Sprintf(">> %s", label)
	if defVal != "" {
		prompt = fmt.Sprintf("%s (default: %s)", prompt, defVal)
	}
	prompt = style.PromptMsg(prompt + ": ")

	// Nutze das bereits vorhandene Prompt‑Utility, das nur das Schreiben übernimmt.
	// (Wir delegieren, damit wir nicht zweimal die gleiche Logik pflegen.)
	return Prompt(r, w, prompt)
}

// -------------------------------------------------------------------
// 2️⃣ MustInt – konvertiert eine Eingabe in ein int, prüft >0.
// -------------------------------------------------------------------
func MustInt(input string) (int, error) {
	if strings.TrimSpace(input) == "" {
		return 0, fmt.Errorf("empty input")
	}
	i, err := strconv.Atoi(input)
	if err != nil || i <= 0 {
		return 0, fmt.Errorf("please enter a positive integer")
	}
	return i, nil
}