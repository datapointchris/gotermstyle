//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package gotermstyle

import "os"

// terminalColumns has no query to make here, so Columns answers from $COLUMNS or
// DefaultColumns. This covers windows, plan9, js and wasip1, and the unixes whose
// syscall package has no SYS_IOCTL: solaris, illumos and aix. A width from the
// environment or the default is enough, because $COLUMNS is the override a
// caller sets anyway; this file exists so the package builds everywhere Go does.
func terminalColumns(*os.File) int {
	return 0
}
