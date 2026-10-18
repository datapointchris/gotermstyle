//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package gotermstyle

import (
	"os"
	"syscall"
	"unsafe"
)

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

// terminalColumns asks the terminal how wide it is, through the TIOCGWINSZ
// ioctl. The build constraint lists the platforms whose syscall package defines
// both SYS_IOCTL and TIOCGWINSZ; android and ios satisfy it as linux and darwin.
// solaris, illumos and aix are unix too, but their syscall package has no
// SYS_IOCTL, so a bare unix constraint fails to compile there.
//
// syscall is standard library, so this costs no dependency; golang.org/x/term
// would have.
func terminalColumns(file *os.File) int {
	var ws winsize
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		file.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return 0
	}
	return int(ws.cols)
}
