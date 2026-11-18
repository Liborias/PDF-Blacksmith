package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"unsafe"
)

const (
	idMinimize = 180 + iota
	idMaximize
	idCloseWindow
)

//go:embed assets/blacksmith.png
var wordmark []byte

func (a *app) loadBrand() {
	img, err := png.Decode(bytes.NewReader(wordmark))
	if err != nil {
		return
	}
	area := img.Bounds()
	picture := &printBitmap{width: area.Dx(), height: area.Dy(), pixels: make([]byte, area.Dx()*area.Dy()*4)}
	for y := 0; y < picture.height; y++ {
		for x := 0; x < picture.width; x++ {
			r, g, b, alpha := img.At(area.Min.X+x, area.Min.Y+y).RGBA()
			remaining := 65535 - alpha
			offset := (y*picture.width + x) * 4
			picture.pixels[offset] = byte((b + remaining*uint32(titleColor.blue)/255) >> 8)
			picture.pixels[offset+1] = byte((g + remaining*uint32(titleColor.green)/255) >> 8)
			picture.pixels[offset+2] = byte((r + remaining*uint32(titleColor.red)/255) >> 8)
		}
	}
	a.brand = picture
}

func (a *app) paintBrand(dc uintptr) {
	if a.logo != 0 {
		size := uintptr(a.scale(36))
		user32.NewProc("DrawIconEx").Call(dc, uintptr(a.scale(8)), uintptr(a.scale(6)), a.logo, size, size, 0, 0, 3)
	}
	if a.brand == nil {
		drawText(dc, "Blacksmith", a.area(47, 4, 172, 46), a.titleFont, textColor, 0)
		return
	}
	width := a.scale(123)
	height := width * a.brand.height / a.brand.width
	gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
	drawBitmap(dc, a.brand, a.scale(45), (a.scale(49)-height)/2, width, height)
}

func drawBitmap(dc uintptr, bitmap *printBitmap, x, y, width, height int) {
	header := bitmapHeader{size: 40, width: int32(bitmap.width), height: -int32(bitmap.height), planes: 1, bits: 32}
	gdi32.NewProc("StretchDIBits").Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0,
		uintptr(bitmap.width), uintptr(bitmap.height), uintptr(unsafe.Pointer(&bitmap.pixels[0])),
		uintptr(unsafe.Pointer(&header)), 0, 0x00cc0020)
}

func (a *app) maximized() bool {
	result, _, _ := user32.NewProc("IsZoomed").Call(a.window)
	return result != 0
}

func (a *app) toggleMaximized() {
	if a.maximized() {
		showWindow.Call(a.window, 9)
	} else {
		showWindow.Call(a.window, 3)
	}
	a.layout()
}

func (a *app) drawWindowButton(item *buttonDraw) bool {
	id := int(item.controlID)
	if id < idMinimize || id > idCloseWindow {
		return false
	}
	body := titleColor
	if item.state&1 != 0 {
		body = divider
		if id == idCloseWindow {
			body = paletteColor{133, 35, 53}
		}
	}
	fillColor(item.dc, item.bounds, body)
	bounds := item.bounds
	x, y := int((bounds.left+bounds.right)/2), int((bounds.top+bounds.bottom)/2)
	d := a.scale(5)
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(1))), textColor.win())
	old, _, _ := gdiSelect.Call(item.dc, pen)
	move := gdi32.NewProc("MoveToEx")
	line := gdi32.NewProc("LineTo")
	segment := func(x1, y1, x2, y2 int) {
		move.Call(item.dc, uintptr(x1), uintptr(y1), 0)
		line.Call(item.dc, uintptr(x2), uintptr(y2))
	}
	square := func(x1, y1, size int) {
		move.Call(item.dc, uintptr(x1), uintptr(y1), 0)
		for _, p := range [][2]int{{x1 + size, y1}, {x1 + size, y1 + size}, {x1, y1 + size}, {x1, y1}} {
			line.Call(item.dc, uintptr(p[0]), uintptr(p[1]))
		}
	}
	switch id {
	case idMinimize:
		segment(x-d, y, x+d+1, y)
	case idMaximize:
		if a.maximized() {
			k := a.scale(3)
			square(x-d, y-d+k, 2*d-k)
			segment(x-d+k, y-d, x+d, y-d)
			segment(x+d, y-d, x+d, y+d-k)
		} else {
			square(x-d, y-d, 2*d)
		}
	case idCloseWindow:
		segment(x-d, y-d, x+d+1, y+d+1)
		segment(x+d, y-d, x-d-1, y+d+1)
	}
	gdiSelect.Call(item.dc, old)
	gdiDelete.Call(pen)
	if item.state&16 != 0 {
		bounds.left += 5
		bounds.top += 5
		bounds.right -= 5
		bounds.bottom -= 5
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&bounds)))
	}
	return true
}

func (a *app) nonClientSize(h, param uintptr) {
	if a.full || param == 0 {
		return
	}
	maximized, _, _ := user32.NewProc("IsZoomed").Call(h)
	if maximized == 0 {
		return
	}
	var proposed rect
	copyDrawItem.Call(uintptr(unsafe.Pointer(&proposed)), param, unsafe.Sizeof(proposed))
	monitor, _, _ := user32.NewProc("MonitorFromRect").Call(uintptr(unsafe.Pointer(&proposed)), 2)
	info := monitorInfo{size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := user32.NewProc("GetMonitorInfoW").Call(monitor, uintptr(unsafe.Pointer(&info))); ok != 0 {
		proposed.left = max(proposed.left, info.work.left)
		proposed.top = max(proposed.top, info.work.top)
		proposed.right = min(proposed.right, info.work.right)
		proposed.bottom = min(proposed.bottom, info.work.bottom)
		copyDrawItem.Call(param, uintptr(unsafe.Pointer(&proposed)), unsafe.Sizeof(proposed))
	}
}

func (a *app) hitTest(hwnd, param uintptr) uintptr {
	if a.full {
		return 1
	}
	pos := point{int32(int16(param)), int32(int16(param >> 16))}
	user32.NewProc("ScreenToClient").Call(hwnd, uintptr(unsafe.Pointer(&pos)))
	x, y := int(pos.x), int(pos.y)
	border := a.scale(6)
	if !a.maximized() && a.width > 0 && a.height > 0 {
		left, right := x < border, x >= a.width-border
		top, bottom := y < border, y >= a.height-border
		switch {
		case left && top:
			return 13
		case right && top:
			return 14
		case left && bottom:
			return 16
		case right && bottom:
			return 17
		case left:
			return 10
		case right:
			return 11
		case top:
			return 12
		case bottom:
			return 15
		}
	}
	if y < a.scale(49) && (x < a.scale(175) || (x > a.scale(311) && x < a.width-a.scale(494))) {
		return 2
	}
	return 1
}
