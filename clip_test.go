package gotermstyle

import (
	"os"
	"testing"
)

func TestClipLeavesTextThatFits(t *testing.T) {
	if got := clipTo("short", 10, 80); got != "short" {
		t.Errorf("Clip = %q, want it untouched", got)
	}
}

func TestClipShortensToTheRoomThatIsLeft(t *testing.T) {
	got := clipTo(string(make([]rune, 100)), 10, 40)
	if n := len([]rune(got)); n != 30 {
		t.Errorf("clipped to %d columns, want 30", n)
	}
	if []rune(got)[29] != '…' {
		t.Errorf("clipped text does not end in an ellipsis: %q", got)
	}
}

func TestClipGivesUpRatherThanEmitABareEllipsis(t *testing.T) {
	// With no room to say anything, truncating loses more than it saves.
	if got := clipTo("hello", 9, 10); got != "hello" {
		t.Errorf("Clip = %q, want the text whole", got)
	}
}

func TestClipCountsRunesRatherThanBytes(t *testing.T) {
	// Three em-dashes are nine bytes and three columns. A byte count would
	// clip text that fits.
	if got := clipTo("———", 10, 20); got != "———" {
		t.Errorf("Clip = %q, want it untouched", got)
	}
}

func TestColumnsPrefersTheDeclaredWidth(t *testing.T) {
	t.Setenv("COLUMNS", "132")
	if got := Columns(); got != 132 {
		t.Errorf("Columns = %d, want the declared 132", got)
	}
}

func TestAnUnusableColumnsFallsThroughRatherThanFailing(t *testing.T) {
	// A COLUMNS that is not a positive number says nothing about the terminal,
	// so it is ignored rather than treated as zero width.
	for _, declared := range []string{"", "wide", "0", "-5"} {
		t.Setenv("COLUMNS", declared)
		if got := Columns(); got <= 0 {
			t.Errorf("COLUMNS=%q gave %d", declared, got)
		}
	}
}

func TestColumnsFallsBackToTheDefaultWithNothingToAsk(t *testing.T) {
	t.Setenv("COLUMNS", "")
	_ = os.Unsetenv("COLUMNS")
	// go test runs with stdout on a pipe, so the ioctl has no answer either.
	if got := Columns(); got != DefaultColumns && terminalColumns(os.Stdout) == 0 {
		t.Errorf("Columns = %d, want %d", got, DefaultColumns)
	}
}
