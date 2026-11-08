//go:build !unix

package tuitheme

import "os"

// Notify relays nothing here, because this platform has no SIGUSR2. A program
// still builds, and rereads its theme when it next loads it.
func Notify(chan<- os.Signal) {}

// Stop does nothing here, because Notify relays nothing.
func Stop(chan<- os.Signal) {}
