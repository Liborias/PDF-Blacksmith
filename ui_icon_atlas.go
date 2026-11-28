package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
)

const themeToggleIcon = "i-theme_quick_inline"

// These masks are generated from the v3.26 SVG symbols by tools/generate_ui_icons.py.
//
//go:embed assets/ui_icons.png
var uiIconImage []byte

//go:embed assets/ui_icons.json
var uiIconIndex []byte

// Original v3.26 1:1 symbol, generated separately to preserve the existing atlas.
//go:embed assets/ui_zoom_1_1.png
var uiActualSizeIcon []byte

type uiIconAtlas struct {
	pixels image.Image
	tile   int
	slots  map[string]int
}

func loadUIIconAtlas() uiIconAtlas {
	var index struct {
		Tile    int      `json:"tile_size"`
		Symbols []string `json:"symbols"`
	}
	if err := json.Unmarshal(uiIconIndex, &index); err != nil {
		panic(err)
	}
	pixels, err := png.Decode(bytes.NewReader(uiIconImage))
	if err != nil {
		panic(err)
	}
	if index.Tile <= 0 || pixels.Bounds().Dy() != index.Tile || pixels.Bounds().Dx() != index.Tile*len(index.Symbols) {
		panic("invalid UI icon atlas dimensions")
	}
	atlas := uiIconAtlas{pixels: pixels, tile: index.Tile, slots: make(map[string]int, len(index.Symbols))}
	for i, name := range index.Symbols {
		if _, ok := atlas.slots[name]; ok {
			panic("duplicate UI icon: " + name)
		}
		atlas.slots[name] = i
	}
	// Extend the atlas in memory; the original sprite sheet and its slot order stay intact.
	actual, err := png.Decode(bytes.NewReader(uiActualSizeIcon))
	if err != nil {
		panic(err)
	}
	if actual.Bounds().Dx() != atlas.tile || actual.Bounds().Dy() != atlas.tile {
		panic("invalid actual-size icon dimensions")
	}
	const actualName = "i-zoom_1_1"
	if _, ok := atlas.slots[actualName]; ok {
		panic("duplicate UI icon: " + actualName)
	}
	width := pixels.Bounds().Dx()
	combined := image.NewRGBA(image.Rect(0, 0, width+atlas.tile, atlas.tile))
	draw.Draw(combined, image.Rect(0, 0, width, atlas.tile), pixels, pixels.Bounds().Min, draw.Src)
	draw.Draw(combined, image.Rect(width, 0, width+atlas.tile, atlas.tile), actual, actual.Bounds().Min, draw.Src)
	atlas.slots[actualName] = len(index.Symbols)
	atlas.pixels = combined
	return atlas
}

func (atlas uiIconAtlas) mask(name string, size int) []byte {
	slot, ok := atlas.slots[name]
	if !ok || size <= 0 {
		return nil
	}
	mask := make([]byte, size*size)
	tile, left := atlas.tile, slot*atlas.tile
	// Average coverage of the larger source sprite at the actual device size.
	for y := range size {
		top, bottom := y*tile/size, (y+1)*tile/size
		bottom = max(top+1, bottom)
		for x := range size {
			first, last := x*tile/size, (x+1)*tile/size
			last = max(first+1, last)
			sum := 0
			for row := top; row < bottom; row++ {
				for col := first; col < last; col++ {
					_, _, _, alpha := atlas.pixels.At(left+col, row).RGBA()
					sum += int(alpha >> 8)
				}
			}
			mask[y*size+x] = byte(sum / ((last - first) * (bottom - top)))
		}
	}
	return mask
}

var uiIcons = loadUIIconAtlas()
