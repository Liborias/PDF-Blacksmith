package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
)

//go:embed docs/pracovni/mapa_barev_FINAL_v6.0.json
var uiPaletteData []byte

type paletteColor struct{ red, green, blue uint8 }

func (color *paletteColor) UnmarshalText(text []byte) error {
	if len(text) != 7 || text[0] != '#' {
		return fmt.Errorf("invalid UI color %q", text)
	}
	rgb, err := strconv.ParseUint(string(text[1:]), 16, 24)
	if err != nil {
		return fmt.Errorf("invalid UI color %q: %w", text, err)
	}
	*color = paletteColor{uint8(rgb >> 16), uint8(rgb >> 8), uint8(rgb)}
	return nil
}

type uiPalette map[string]paletteColor

type navigationStyle struct {
	fill, ink paletteColor
	accent    bool
}

func (colors uiPalette) navigation(base paletteColor, disabled, selected, hovered bool) navigationStyle {
	style := navigationStyle{fill: base, ink: colors["text"]}
	switch {
	case disabled:
		style.ink = colors["muted"]
	case hovered:
		style.fill, style.accent = colors["hover"], true
	case selected:
		style.fill, style.accent = colors["selected"], true
	}
	return style
}

type uiTheme struct {
	Palette string `json:"palette"`
	Mode    string `json:"mode"`
}

type uiPaletteVariants struct {
	Dark  uiPalette `json:"dark"`
	Light uiPalette `json:"light"`
}

type uiPaletteCatalog struct {
	Default  uiTheme                      `json:"default"`
	Palettes map[string]uiPaletteVariants `json:"palettes"`
}

func (catalog uiPaletteCatalog) colors(theme uiTheme) uiPalette {
	variants := catalog.Palettes[theme.Palette]
	switch theme.Mode {
	case "dark":
		return variants.Dark
	case "light":
		return variants.Light
	default:
		return nil
	}
}

func loadUIPalettes() uiPaletteCatalog {
	var catalog uiPaletteCatalog
	if err := json.Unmarshal(uiPaletteData, &catalog); err != nil {
		panic(fmt.Errorf("invalid UI palette catalog: %w", err))
	}
	if catalog.colors(catalog.Default) == nil {
		panic("default UI palette is missing")
	}
	return catalog
}

var uiPalettes = loadUIPalettes()
