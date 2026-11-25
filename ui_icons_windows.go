package main

type uiIconKey struct {
	symbol          string
	size            int
	ink, background paletteColor
}

var uiIconBitmaps = make(map[uiIconKey]*printBitmap)

func drawUIIcon(dc uintptr, symbol string, x, y, size int, ink, background paletteColor) {
	key := uiIconKey{symbol, size, ink, background}
	bitmap := uiIconBitmaps[key]
	if bitmap == nil {
		mask := uiIcons.mask(symbol, size)
		if mask == nil {
			return
		}
		pixels := make([]byte, 4*len(mask))
		blend := func(fg, bg, alpha byte) byte {
			return byte((int(fg)*int(alpha) + int(bg)*(255-int(alpha)) + 127) / 255)
		}
		for i, alpha := range mask {
			pixels[i*4] = blend(ink.blue, background.blue, alpha)
			pixels[i*4+1] = blend(ink.green, background.green, alpha)
			pixels[i*4+2] = blend(ink.red, background.red, alpha)
			pixels[i*4+3] = 255
		}
		bitmap = &printBitmap{width: size, height: size, pixels: pixels}
		uiIconBitmaps[key] = bitmap
	}
	drawBitmap(dc, bitmap, x, y, size, size)
}
