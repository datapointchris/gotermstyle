//go:build unix

package tuitheme

import (
	"os"
	"syscall"
	"testing"
	"time"
)

// A theme generator signals every reader with SIGUSR2 after it installs a theme.
func TestSIGUSR2AsksForTheFileAgain(t *testing.T) {
	reloads := make(chan os.Signal, 1)
	Notify(reloads)
	defer Stop(reloads)

	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR2); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reloads:
	case <-time.After(5 * time.Second):
		t.Fatal("no reload relayed within 5s of SIGUSR2")
	}
}
