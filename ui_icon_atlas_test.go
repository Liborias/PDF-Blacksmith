package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestUIIconAtlasUsesVisibleApprovedSymbols(t *testing.T) {
	data, err := os.ReadFile("docs/pracovni/mapa_ikon_FINAL_v6.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Icons []struct {
			Symbol string `json:"symbol"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	approved := make(map[string]bool)
	for _, item := range registry.Icons {
		approved[item.Symbol] = true
	}
	approved[themeToggleIcon] = true // Exact inline SVG of the reference toolbar switch.
	for name := range uiIcons.slots {
		if !approved[name] || !strings.HasPrefix(name, "i-") {
			t.Errorf("unapproved UI icon %q", name)
		}
		for _, size := range []int{18, 27, 36} {
			mask := uiIcons.mask(name, size)
			var transparent, visible int
			for _, alpha := range mask {
				if alpha == 0 {
					transparent++
				}
				if alpha >= 128 {
					visible++
				}
			}
			if len(mask) != size*size || transparent == 0 || visible == 0 {
				t.Errorf("%s at %d px: invalid icon coverage", name, size)
			}
		}
	}
	if len(uiIcons.slots) == 0 || uiIcons.mask("i-does-not-exist", 18) != nil {
		t.Fatal("unexpected handling of the icon registry")
	}
}
