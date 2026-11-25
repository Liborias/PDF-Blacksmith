package main

import (
	"encoding/json"
	"testing"
)

func TestInterfacePreferencesRoundTrip(t *testing.T) {
	for _, palette := range uiPaletteOrder {
		for _, mode := range []string{"dark", "light"} {
			want := interfacePreferences{Locale: uiEnglish, Theme: uiTheme{Palette: palette, Mode: mode}}
			data, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			if got := decodeInterfacePreferences(data, uiCzech); got != want {
				t.Errorf("%s/%s got %+v, want %+v", palette, mode, got, want)
			}
			next := want.Theme.oppositeMode()
			if next.Palette != palette || next.Mode == mode || next.oppositeMode() != want.Theme {
				t.Errorf("cannot switch %s/%s without changing palette: %+v", palette, mode, next)
			}
		}
	}
}

func TestInterfacePreferencesFallback(t *testing.T) {
	for _, invalid := range []string{
		"",
		"{",
		`{"locale":"de-DE","theme":{"palette":"unknown","mode":"light"}}`,
		`{"theme":{"palette":"olive","mode":"other"}}`,
	} {
		want := interfacePreferences{Locale: uiEnglish, Theme: uiPalettes.Default}
		if got := decodeInterfacePreferences([]byte(invalid), uiEnglish); got != want {
			t.Errorf("%q got %+v, want %+v", invalid, got, want)
		}
	}
}
