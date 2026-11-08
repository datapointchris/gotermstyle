//go:build unix

package tuitheme

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSIGUSR2AsksForTheFileAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.json")
	reloads := make(chan os.Signal, 1)
	Notify(reloads)
	defer Stop(reloads)

	writeTheme(t, path, `{"ui": {"selection": {"bg": "#2d4f67"}}}`)
	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR2); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reloads:
	case <-time.After(5 * time.Second):
		t.Fatal("no reload relayed within 5s of SIGUSR2")
	}
	theme, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := theme.Code("ui.selection"), "\033[0;48;2;45;79;103m"; got != want {
		t.Errorf("ui.selection after reload = %q, want %q", got, want)
	}
}
