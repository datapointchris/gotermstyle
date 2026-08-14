//go:build unix

package gotermstyle

import (
	"os"
	"syscall"
	"unsafe"
)

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

// terminalColumns asks the terminal how wide it is. syscall is standard
// library, so this costs no dependency; golang.org/x/term would have.
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
