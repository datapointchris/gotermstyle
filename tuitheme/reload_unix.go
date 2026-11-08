//go:build unix

package tuitheme

import (
	"os"
	"os/signal"
	"syscall"
)

// Notify relays each request to reread the theme file to c. The request is
// SIGUSR2, which a theme generator sends every running reader after it moves
// current.json. A Go program that never calls Notify catches SIGUSR2 and
// ignores it, so the signal is safe to send any Go program by name.
func Notify(c chan<- os.Signal) {
	signal.Notify(c, syscall.SIGUSR2)
}

// Stop ends Notify's relaying to c.
func Stop(c chan<- os.Signal) {
	signal.Stop(c)
}
