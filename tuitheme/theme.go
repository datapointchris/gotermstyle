// Package tuitheme reads the TUI theme file: the colors and styles a terminal
// program draws its interface, markdown and code in, keyed by role. A theme
// generator writes one per theme, and every program on this package reads the
// same file, so one theme change recolors them all.
//
// The format is JSON with three groups, ui, markdown and syntax. A value is
// "#rrggbb", a terminal palette slot from "0" to "15", or a style object of fg,
// bg, bold, italic, underline, strikethrough and dim. The schema published
// beside this package documents every key:
//
//	https://raw.githubusercontent.com/datapointchris/gotermstyle/main/tuitheme/schema/v1.json
//
// A role is a group and a key, such as "ui.selection" or "syntax.keyword.return".
// A key the file leaves out falls back along its dots, then to the built-in
// default. The default is written in palette slots, so a program with no theme
// file draws in the terminal's own 16 colors.
//
// Load a theme once, paint each role through a gotermstyle.Palette, and load it
// again when Notify relays a reload:
//
//	path := tuitheme.DefaultPath()
//	theme, err := tuitheme.Load(path)
//	fmt.Println(pal.Paint(theme.Code("ui.selection"), row))
//
//	reloads := make(chan os.Signal, 1)
//	tuitheme.Notify(reloads)
//	<-reloads // then Load(path) again
//
// A Theme never changes once loaded, so it is safe to share between goroutines.
package tuitheme

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed default.json
var defaultJSON []byte

// builtIn is the default every lookup ends at.
var builtIn = mustParse(defaultJSON)

// groups are the format's groups. Any other is ignored, so a newer generator
// never breaks an older reader.
var groups = []string{"ui", "markdown", "syntax"}

// Theme is a theme file's roles over the built-in default.
type Theme struct {
	roles map[string]Style
}

// Default is the built-in theme, in the terminal's 16 colors.
func Default() *Theme {
	return &Theme{}
}

// DefaultPath is where a theme generator installs the current theme:
// $XDG_CONFIG_HOME/tui-theme/current.json, or ~/.config/tui-theme/current.json.
func DefaultPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(dir) {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tui-theme", "current.json")
}

// Load reads the theme file at path. A file that does not exist is the
// built-in default, with no error, because a terminal with no theme file is the
// ordinary case. A file that cannot be read or holds a value of neither form is
// an error, and the caller keeps what it was drawing.
func Load(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("tuitheme: %w", err)
	}
	roles, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("tuitheme: %s: %w", path, err)
	}
	return &Theme{roles: roles}, nil
}

// Style is what role draws in: the file's value for it or the nearest key along
// its dots, else the built-in default's, else a style that sets nothing.
func (t *Theme) Style(role string) Style {
	if s, ok := lookup(t.roles, role); ok {
		return s
	}
	s, _ := lookup(builtIn, role)
	return s
}

// Code is role's style as an SGR sequence, for gotermstyle.Palette.Paint.
func (t *Theme) Code(role string) string {
	return t.Style(role).Code()
}

// lookup tries role, then role with its last dotted part dropped, stopping at
// the group, so markdown.heading.2 tries markdown.heading but never markdown.
func lookup(roles map[string]Style, role string) (Style, bool) {
	group, key, _ := strings.Cut(role, ".")
	for key != "" {
		if s, ok := roles[group+"."+key]; ok {
			return s, true
		}
		cut := strings.LastIndex(key, ".")
		if cut < 0 {
			break
		}
		key = key[:cut]
	}
	return Style{}, false
}

// styleObject is a value's object form. Fields it does not name are ignored.
type styleObject struct {
	Fg            string `json:"fg"`
	Bg            string `json:"bg"`
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Underline     bool   `json:"underline"`
	Strikethrough bool   `json:"strikethrough"`
	Dim           bool   `json:"dim"`
}

func parse(data []byte) (map[string]Style, error) {
	var file map[string]json.RawMessage
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	roles := map[string]Style{}
	for _, group := range groups {
		raw, ok := file[group]
		if !ok {
			continue
		}
		var keys map[string]json.RawMessage
		if json.Unmarshal(raw, &keys) != nil {
			return nil, fmt.Errorf("%s: %s is not an object of roles", group, raw)
		}
		for key, value := range keys {
			s, err := parseValue(value)
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", group, key, err)
			}
			roles[group+"."+key] = s
		}
	}
	return roles, nil
}

func parseValue(raw json.RawMessage) (Style, error) {
	var color string
	if json.Unmarshal(raw, &color) == nil {
		fg, err := parseColor(color)
		return Style{Fg: fg}, err
	}
	var obj styleObject
	if err := json.Unmarshal(raw, &obj); err != nil {
		return Style{}, fmt.Errorf("%s is neither a color nor a style object", raw)
	}
	fg, err := parseColor(obj.Fg)
	if err != nil {
		return Style{}, fmt.Errorf("fg: %w", err)
	}
	bg, err := parseColor(obj.Bg)
	if err != nil {
		return Style{}, fmt.Errorf("bg: %w", err)
	}
	return Style{
		Fg: fg, Bg: bg,
		Bold: obj.Bold, Italic: obj.Italic, Underline: obj.Underline,
		Strikethrough: obj.Strikethrough, Dim: obj.Dim,
	}, nil
}

func mustParse(data []byte) map[string]Style {
	roles, err := parse(data)
	if err != nil {
		panic("tuitheme: default.json: " + err.Error())
	}
	return roles
}
