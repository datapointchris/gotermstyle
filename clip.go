package gotermstyle

import (
	"os"
	"strconv"
)

// DefaultColumns is the width assumed when nothing will say what it really is.
const DefaultColumns = 80

// Columns is the terminal width. $COLUMNS is consulted first so a caller can
// override it and a test can set it, then the terminal itself, then
// DefaultColumns.
//
// This is the order Python's shutil.get_terminal_size uses, which is what the
// counterpart library resolves through.
func Columns() int {
	if declared := os.Getenv("COLUMNS"); declared != "" {
		if n, err := strconv.Atoi(declared); err == nil && n > 0 {
			return n
		}
	}
	if n := terminalColumns(os.Stdout); n > 0 {
		return n
	}
	return DefaultColumns
}

// Clip shortens text to fit the terminal alongside used other columns.
//
// Callers pass the length of the uncolored surrounding text and color the
// result afterwards, so a clip can never land inside an escape sequence and
// leak a raw code onto the screen.
//
// With no room to say anything the text is returned whole: truncating to a bare
// ellipsis loses more than it saves, and a row that overflows is more use than
// a row that says nothing.
//
// Every character this clips around is assumed one column wide. That holds for
// the ↳ · … this family is built for and does not hold for CJK or emoji, which
// is the one thing a wide-rune library would buy and is a deliberate
// acceptance rather than an oversight.
func Clip(text string, used int) string {
	return clipTo(text, used, Columns())
}

func clipTo(text string, used, columns int) string {
	room := columns - used
	runes := []rune(text)
	if len(runes) <= room || room < 2 {
		return text
	}
	return string(runes[:room-1]) + "…"
}
