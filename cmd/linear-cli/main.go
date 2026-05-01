package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

var version = "dev"

func main() {
	root := cmd.NewRootCommand(version)
	err := root.Execute()
	if err == nil {
		return
	}
	formatExecError(err, hasJSONFlag(os.Args[1:]), os.Stderr)
	os.Exit(1)
}

// hasJSONFlag scans raw args for --json / --json=<value> before "--".
// Uses last-wins semantics and strconv.ParseBool to mirror Cobra's behavior.
// Used to detect JSON mode even when Cobra has not yet parsed flags
// (e.g. on "unknown command" errors that occur during command lookup).
func hasJSONFlag(args []string) bool {
	result := false
	for _, a := range args {
		if a == "--" {
			break
		}
		if a == "--json" {
			result = true
			continue
		}
		if val, ok := strings.CutPrefix(a, "--json="); ok {
			if b, err := strconv.ParseBool(val); err == nil {
				result = b
			}
		}
	}
	return result
}

func formatExecError(err error, jsonMode bool, w io.Writer) {
	if err == nil {
		return
	}
	if jsonMode {
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		_, _ = fmt.Fprintln(w, string(b))
		return
	}
	_, _ = fmt.Fprintln(w, err)
}
