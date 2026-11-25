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
			"panelTabIdle", "panelTabHover", "panelTabActive"}, 4.5},
		{"muted", []string{"base", "panel", "editor", "control", "tabIdle", "hover"}, 4.5},
		{"accent", []string{"base", "panel", "control", "input"}, 4.5},
		{"focus", []string{"base", "top", "panel", "control", "input", "hover", "tabIdle", "tabActive", "tabHover",
			"panelTabIdle", "panelTabHover", "panelTabActive"}, 3},
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
