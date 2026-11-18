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

var brandBackground = paletteColor{102, 102, 102}

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
			picture.pixels[offset] = byte((b + remaining*uint32(brandBackground.blue)/255) >> 8)
			picture.pixels[offset+1] = byte((g + remaining*uint32(brandBackground.green)/255) >> 8)
			picture.pixels[offset+2] = byte((r + remaining*uint32(brandBackground.red)/255) >> 8)
		}
	}
	a.brand = picture
}

func (a *app) paintBrand(dc uintptr, block rect) {
	fillColor(dc, block, brandBackground)
	padding := (block.bottom - block.top) / 5
	size := int(block.bottom - block.top - 2*padding)
	left := int(block.left + padding)
	top := int(block.top + padding)
	if a.logo != 0 {
		user32.NewProc("DrawIconEx").Call(dc, uintptr(left), uintptr(top), a.logo, uintptr(size), uintptr(size), 0, 0, 3)
		left += size + int(padding*7/8)
	}
	if a.brand == nil {
		drawText(dc, "Blacksmith", rect{int32(left), block.top, block.right - padding, block.bottom}, a.titleFont, textColor, 0)
		return
	}
	width := int(block.right-padding) - left
	height := width * a.brand.height / a.brand.width
	gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
	drawBitmap(dc, a.brand, left, (int(block.top+block.bottom)-height)/2, width, height)
}

func (a *app) layoutHeader() {
	const windowButtonsWidth = 108
	const compactActions = 2 * 36
	const widePanel, wideFullscreen = 138, 144
	d := a.scale
	tabWidth := 0
	for _, tab := range workspaceTabs {
		if !tab.needsPDF || a.documentTools {
			tabWidth += tab.width
		}
	}
	a.headerBrand = 200
	if a.width < d(a.headerBrand+tabWidth+compactActions+windowButtonsWidth+6) {
		a.headerBrand = 168
	}
	compactPages := a.width < d(a.headerBrand+tabWidth+compactActions+windowButtonsWidth+6)
	x := a.headerBrand
	for _, tab := range workspaceTabs {
		if !tab.needsPDF || a.documentTools {
			width := tab.width
			if tab.id == idPagesTab && compactPages {
				width = 42
			}
			place(a.modes[tab.id], d(x), d(8), d(width-2), d(34))
			x += width
		}
	}
	a.headerTabs = d(x)

	panel, full := 36, 36
	if a.width-a.headerTabs >= d(windowButtonsWidth+widePanel+wideFullscreen+20) {
		panel, full = widePanel, wideFullscreen
	}
	a.headerActions = a.width - d(windowButtonsWidth+panel+full+4)
	place(a.panelButton, a.headerActions, d(8), d(panel-2), d(34))
	place(a.fullButton, a.width-d(windowButtonsWidth+full+4), d(8), d(full-2), d(34))
	for i, handle := range a.windowButtons {
		place(handle, a.width-d(windowButtonsWidth-i*36), 0, d(36), d(headerHeight-1))
		invalidate(handle)
	}
}

var headerIcons = map[int][][]point{
	idBrowseTab: {
		{{1, 9}, {5, 5}, {9, 4}, {13, 5}, {17, 9}},
		{{1, 9}, {5, 13}, {9, 14}, {13, 13}, {17, 9}},
		{{7, 9}, {9, 7}, {11, 9}, {9, 11}, {7, 9}},
	},
	idPagesTab: {
		{{4, 3}, {12, 3}, {16, 7}, {16, 16}, {4, 16}, {4, 3}},
		{{12, 3}, {12, 7}, {16, 7}},
		{{1, 13}, {1, 1}, {10, 1}},
	},
	idProofTab: {
		{{1, 5}, {9, 1}, {17, 5}, {9, 9}, {1, 5}},
		{{1, 9}, {9, 13}, {17, 9}},
		{{1, 13}, {9, 17}, {17, 13}},
	},
	idSignTab: {
		{{3, 12}, {13, 2}, {16, 5}, {6, 15}, {2, 16}, {3, 12}},
		{{11, 4}, {14, 7}},
		{{9, 16}, {17, 16}},
	},
	idTogglePanel: {
		{{1, 2}, {17, 2}, {17, 16}, {1, 16}, {1, 2}},
		{{11, 2}, {11, 16}},
		{{13, 6}, {15, 6}},
		{{13, 10}, {15, 10}},
	},
	idFullscreen: {
		{{1, 6}, {1, 1}, {6, 1}},
		{{12, 1}, {17, 1}, {17, 6}},
		{{17, 12}, {17, 17}, {12, 17}},
		{{6, 17}, {1, 17}, {1, 12}},
	},
}

func (a *app) drawHeaderIconLabel(item *buttonDraw, color paletteColor) bool {
	if int(item.controlID) == idFileTab {
		bounds := item.bounds
		a.arrow(item.dc, int(bounds.right)-a.scale(11), int(bounds.top+bounds.bottom)/2, color)
		bounds.right -= int32(a.scale(18))
		drawText(item.dc, getText(item.handle), bounds, a.labelFont, color, 1)
		return true
	}
	paths := headerIcons[int(item.controlID)]
	if len(paths) == 0 {
		return false
	}
	bounds := item.bounds
	size := int32(a.scale(18))
	iconOnly := bounds.right-bounds.left < int32(a.scale(74))
	left := bounds.left + int32(a.scale(9))
	if iconOnly {
		left = (bounds.left + bounds.right - size) / 2
	}
	top := (bounds.top + bounds.bottom - size) / 2
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(1))), color.win())
	old, _, _ := gdiSelect.Call(item.dc, pen)
	polyline := gdi32.NewProc("Polyline")
	for _, path := range paths {
		var shape [6]point
		for index, vertex := range path {
			shape[index] = point{left + int32(a.scale(int(vertex.x))), top + int32(a.scale(int(vertex.y)))}
		}
		polyline.Call(item.dc, uintptr(unsafe.Pointer(&shape[0])), uintptr(len(path)))
	}
	gdiSelect.Call(item.dc, old)
	gdiDelete.Call(pen)
	if !iconOnly {
		bounds.left = left + size + int32(a.scale(7))
		drawText(item.dc, getText(item.handle), bounds, a.labelFont, color, 0)
	}
	return true
}

func drawBitmap(dc uintptr, bitmap *printBitmap, x, y, width, height int) int32 {
	header := bitmapHeader{size: 40, width: int32(bitmap.width), height: -int32(bitmap.height), planes: 1, bits: 32}
	lines, _, _ := gdi32.NewProc("StretchDIBits").Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0,
		uintptr(bitmap.width), uintptr(bitmap.height), uintptr(unsafe.Pointer(&bitmap.pixels[0])),
		uintptr(unsafe.Pointer(&header)), 0, 0x00cc0020)
	return int32(lines)
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
	if y < a.scale(headerHeight) && (x < a.scale(a.headerBrand) || (x >= a.headerTabs && x < a.headerActions)) {
		return 2
	}
	return 1
}
