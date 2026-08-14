//go:build !unix

package gotermstyle

import "os"

// terminalColumns has no answer off unix, so Columns falls through to $COLUMNS
// or DefaultColumns. The fleet is macOS, Arch and WSL; this exists so the
// package still builds for anyone who imports it elsewhere.
func terminalColumns(*os.File) int {
	return 0
}
