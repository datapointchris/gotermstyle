package tuitheme

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTheme(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRoleFallsBackAlongItsDotsThenToTheTerminalsSixteenColors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "current.json")
	writeTheme(t, path, `{
		"ui": {"selection": {"fg": "#c5c8c6", "bg": "#373b41"}},
		"markdown": {"heading": {"fg": "#4eb6fe", "bold": true}},
		"panes": {"left": "#ffffff"}
	}`)
	themed, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	unthemed, err := Load(filepath.Join(dir, "absent.json"))
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		theme      *Theme
		role, want string
	}{
		{themed, "ui.selection", "\033[0;38;2;197;200;198;48;2;55;59;65m"},
		{themed, "markdown.heading.2", "\033[0;1;38;2;78;182;254m"},
		{themed, "ui.muted", "\033[0;90m"},
		{unthemed, "ui.selection", "\033[0;30;47m"},
		{unthemed, "markdown.heading.4", "\033[0;1m"},
		{unthemed, "ui.text", ""},
	} {
		if got := test.theme.Code(test.role); got != test.want {
			t.Errorf("%s = %q, want %q", test.role, got, test.want)
		}
	}
}
