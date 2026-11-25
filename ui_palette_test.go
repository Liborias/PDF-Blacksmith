package main

import (
	"encoding/json"
	"math"
	"testing"
)

func TestUIPaletteCatalog(t *testing.T) {
	var reference struct {
		Roles map[string]string `json:"roles"`
	}
	if err := json.Unmarshal(uiPaletteData, &reference); err != nil {
		t.Fatal(err)
	}
	if len(uiPalettes.Palettes) != 6 || len(reference.Roles) != 44 {
		t.Fatalf("got %d palettes with %d roles, want 6 palettes with 44 roles",
			len(uiPalettes.Palettes), len(reference.Roles))
	}
	for _, name := range []string{"wine", "gray", "cream", "blue", "olive", "teal"} {
		for _, mode := range []string{"dark", "light"} {
			colors := uiPalettes.colors(uiTheme{name, mode})
			if len(colors) != len(reference.Roles) {
				t.Errorf("%s/%s: got %d colors, want %d", name, mode, len(colors), len(reference.Roles))
			}
			for role := range reference.Roles {
				if _, exists := colors[role]; !exists {
					t.Errorf("%s/%s: missing %s", name, mode, role)
				}
			}
		}
	}
	if uiPalettes.Default != (uiTheme{"wine", "dark"}) {
		t.Fatalf("unexpected default theme: %+v", uiPalettes.Default)
	}
	colors := uiPalettes.colors(uiPalettes.Default)
	if colors["editor"] != (paletteColor{27, 27, 33}) || colors["input"] == colors["editor"] {
		t.Fatal("expected the specified burgundy dark editor color, separate from input fills")
	}
}

func TestUIPaletteChromeContrast(t *testing.T) {
	pairs := []struct {
		foreground  string
		backgrounds []string
		minimum     float64
	}{
		{"text", []string{"base", "top", "panel", "control", "input", "hover", "tabActive", "tabHover",
			"panelTabIdle", "panelTabHover", "panelTabActive", "popoverBg", "selected", "toolbarActive", "toolbarActiveHover"}, 4.5},
		{"muted", []string{"base", "panel", "editor", "control", "tabIdle", "hover", "popoverBg", "selected", "top"}, 4.5},
		{"accent", []string{"base", "panel", "control", "input"}, 4.5},
		{"accent", []string{"hover", "selected", "toolbarActive", "toolbarActiveHover"}, 3},
		{"focus", []string{"base", "top", "panel", "control", "input", "hover", "tabIdle", "tabActive", "tabHover",
			"panelTabIdle", "panelTabHover", "panelTabActive", "popoverBg", "selected", "toolbarActive", "toolbarActiveHover"}, 3},
	}
	for name := range uiPalettes.Palettes {
		for _, mode := range []string{"dark", "light"} {
			colors := uiPalettes.colors(uiTheme{name, mode})
			for _, pair := range pairs {
				for _, background := range pair.backgrounds {
					fg, bg := paletteLuminance(colors[pair.foreground]), paletteLuminance(colors[background])
					ratio := (max(fg, bg) + 0.05) / (min(fg, bg) + 0.05)
					if ratio < pair.minimum {
						t.Errorf("%s/%s: %s on %s has %.2f:1, want at least %.1f:1",
							name, mode, pair.foreground, background, ratio, pair.minimum)
					}
				}
			}
		}
	}
}

func TestNavigationStyle(t *testing.T) {
	for name := range uiPalettes.Palettes {
		for _, mode := range []string{"dark", "light"} {
			colors := uiPalettes.colors(uiTheme{name, mode})
			base := colors["popoverBg"]
			for _, state := range []struct {
				disabled, selected, hovered bool
				fill, ink                   string
				accent                      bool
			}{
				{false, false, false, "popoverBg", "text", false},
				{false, true, false, "selected", "text", true},
				{false, false, true, "hover", "text", true},
				{false, true, true, "hover", "text", true},
				{true, true, true, "popoverBg", "muted", false},
			} {
				got := colors.navigation(base, state.disabled, state.selected, state.hovered)
				want := navigationStyle{fill: colors[state.fill], ink: colors[state.ink], accent: state.accent}
				if got != want {
					t.Errorf("%s/%s in %s state: got %+v, want %+v", name, mode, state.fill, got, want)
				}
			}
		}
	}
}

func paletteLuminance(color paletteColor) float64 {
	linear := func(channel uint8) float64 {
		value := float64(channel) / 255
		if value <= 0.04045 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(color.red) + 0.7152*linear(color.green) + 0.0722*linear(color.blue)
}
