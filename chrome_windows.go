package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"path/filepath"
	"slices"
	"strings"
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
	d := a.scale
	a.headerBrand = 200
	a.windowButtonStart = a.width - d(windowButtonsWidth)
	for i, handle := range a.windowButtons {
		place(handle, a.width-d(windowButtonsWidth-i*36), 0, d(36), d(headerHeight-1))
		invalidate(handle)
	}
}

func (a *app) layoutDocumentTabs() {
	d := a.scale
	left, right := d(a.headerBrand), a.windowButtonStart-d(4)
	widths := make([]int, len(a.tabs))
	names := make(map[string]int, len(a.tabs))
	for _, tab := range a.tabs {
		names[strings.ToLower(filepath.Base(a.contentOf(tab).path))]++
	}
	duplicates := false
	total := 0
	for index, tab := range a.tabs {
		path := a.contentOf(tab).path
		caption := filepath.Base(path)
		if names[strings.ToLower(caption)] > 1 {
			caption += " · " + filepath.Base(filepath.Dir(path))
			duplicates = true
		}
		if getText(tab.button) != caption {
			setText(tab.button, caption)
		}
		widths[index] = min(d(225), max(d(112), a.captionWidth(caption, a.normalFont)+d(54)))
		total += widths[index] + d(3)
	}
	showList := !a.full && (total > right-left || duplicates) && len(a.tabs) > 1
	if showList {
		right -= d(28)
		place(a.tabList, right, d(8), d(26), d(33))
	}
	showControl(a.tabList, showList)
	a.tabBounds = rect{}
	first := 0
	if current := slices.Index(a.tabs, a.activeTab); current >= 0 {
		throughActive := 0
		for _, width := range widths[:current+1] {
			throughActive += width + d(3)
		}
		for first < current && throughActive > right-left {
			throughActive -= widths[first] + d(3)
			first++
		}
	}
	x := left
	for index, tab := range a.tabs {
		width := min(widths[index], right-x)
		show := !a.full && index >= first && width >= d(84)
		if show {
			place(tab.button, x, d(8), width-d(25), d(33))
			place(tab.close, x+width-d(25), d(8), d(25), d(33))
			x += width + d(3)
		}
		showControl(tab.button, show)
		showControl(tab.close, show)
		invalidate(tab.button)
		invalidate(tab.close)
	}
	if x > left || showList {
		end := x
		if showList {
			end = right + d(26)
		}
		a.tabBounds = rect{int32(left), int32(d(8)), int32(end), int32(d(41))}
	}
}

func (a *app) drawDocumentTab(item *buttonDraw) bool {
	id := int(item.controlID)
	if id == idListDocumentTabs {
		fillColor(item.dc, item.bounds, titleColor)
		a.arrow(item.dc, int(item.bounds.left+item.bounds.right)/2, int(item.bounds.top+item.bounds.bottom)/2, textColor)
	} else if id == idSelectDocumentTab || id == idClosePDF {
		for _, tab := range a.tabs {
			if item.handle != tab.button && item.handle != tab.close {
				continue
			}
			body, foreground := sidebarRailColor, mutedColor
			if tab == a.activeTab {
				body, foreground = panelColor, textColor
			}
			if item.state&1 != 0 {
				body = divider
			}
			r := item.bounds
			fillColor(item.dc, r, body)
			if tab == a.activeTab {
				fillColor(item.dc, rect{r.left, r.top, r.right, r.top + 1}, selectionColor)
			}
			label, align := "×", uintptr(1)
			if item.handle == tab.button {
				label, align = getText(tab.button), 0x8000
				r.left += int32(a.scale(9))
				if a.contentOf(tab).changed() {
					label = "● " + label
				}
			}
			drawText(item.dc, label, r, a.normalFont, foreground, align)
			break
		}
	} else {
		return false
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

var pagesIcon = [][]point{
	{{4, 3}, {12, 3}, {16, 7}, {16, 16}, {4, 16}, {4, 3}},
	{{12, 3}, {12, 7}, {16, 7}},
	{{1, 13}, {1, 1}, {10, 1}},
}

var toolIcons = map[int][][]point{
	idFileTab: {
		{{1, 4}, {6, 4}, {8, 6}, {17, 6}, {17, 16}, {1, 16}, {1, 4}},
	},
	idSearchTab: {
		{{8, 2}, {12, 4}, {14, 8}, {12, 12}, {8, 14}, {4, 12}, {2, 8}, {4, 4}},
		{{4, 4}, {8, 2}}, {{12, 12}, {17, 17}},
	},
	idTextTab: {{{3, 5}, {3, 2}, {15, 2}, {15, 5}}, {{9, 2}, {9, 16}}, {{6, 16}, {12, 16}}},
	idOCRTab: {
		{{1, 5}, {1, 1}, {5, 1}}, {{13, 1}, {17, 1}, {17, 5}},
		{{17, 13}, {17, 17}, {13, 17}}, {{5, 17}, {1, 17}, {1, 13}},
		{{5, 13}, {9, 5}, {13, 13}}, {{7, 10}, {11, 10}},
	},
	idBrowseTab: {
		{{1, 9}, {5, 5}, {9, 4}, {13, 5}, {17, 9}},
		{{1, 9}, {5, 13}, {9, 14}, {13, 13}, {17, 9}},
		{{7, 9}, {9, 7}, {11, 9}, {9, 11}, {7, 9}},
	},
	idPagesTab:      pagesIcon,
	idPageOrderTool: pagesIcon,
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
	idFullscreen: {
		{{1, 6}, {1, 1}, {6, 1}},
		{{12, 1}, {17, 1}, {17, 6}},
		{{17, 12}, {17, 17}, {12, 17}},
		{{6, 17}, {1, 17}, {1, 12}},
	},
	idSplitTab: {
		{{2, 2}, {12, 2}, {16, 6}, {16, 8}},
		{{12, 2}, {12, 6}, {16, 6}},
		{{2, 2}, {2, 8}},
		{{2, 12}, {2, 16}, {16, 16}, {16, 12}},
		{{1, 10}, {4, 10}}, {{7, 10}, {10, 10}}, {{13, 10}, {17, 10}},
	},
	idMergeTab: {
		{{1, 2}, {7, 2}, {7, 7}, {1, 7}, {1, 2}},
		{{1, 11}, {7, 11}, {7, 16}, {1, 16}, {1, 11}},
		{{7, 5}, {12, 9}}, {{7, 13}, {12, 9}}, {{10, 7}, {12, 9}, {10, 11}},
		{{14, 5}, {17, 5}, {17, 13}, {14, 13}},
	},
}

func (a *app) drawToolIconLabel(item *buttonDraw, color paletteColor) bool {
	paths := toolIcons[int(item.controlID)]
	if len(paths) == 0 {
		return false
	}
	bounds := item.bounds
	size := int32(a.scale(18))
	left := bounds.left + int32(a.scale(9))
	top := (bounds.top + bounds.bottom - size) / 2
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(1))), color.win())
	old, _, _ := gdiSelect.Call(item.dc, pen)
	polyline := gdi32.NewProc("Polyline")
	for _, path := range paths {
		var shape [8]point
		for index, vertex := range path {
			shape[index] = point{left + int32(a.scale(int(vertex.x))), top + int32(a.scale(int(vertex.y)))}
		}
		polyline.Call(item.dc, uintptr(unsafe.Pointer(&shape[0])), uintptr(len(path)))
	}
	gdiSelect.Call(item.dc, old)
	gdiDelete.Call(pen)
	bounds.left = left + size + int32(a.scale(7))
	bounds.right -= int32(a.scale(5))
	if int(item.controlID) == idFileTab {
		a.arrow(item.dc, int(bounds.right)-a.scale(8), int(bounds.top+bounds.bottom)/2, color)
		bounds.right -= int32(a.scale(20))
	}
	drawText(item.dc, getText(item.handle), bounds, a.labelFont, color, 0x8000)
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
	if (id < idMinimize || id > idCloseWindow) && id != idClosePDF {
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
	color := textColor
	if id == idClosePDF {
		d = a.scale(4)
		color = mutedColor
		if item.state&4 != 0 {
			color = disabledColor
		} else if item.state&(1|16) != 0 {
			color = selectionColor
		}
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(1))), color.win())
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
	case idCloseWindow, idClosePDF:
		segment(x-d, y-d, x+d+1, y+d+1)
		segment(x+d, y-d, x-d-1, y+d+1)
	}
	gdiSelect.Call(item.dc, old)
	gdiDelete.Call(pen)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
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
	if y < a.scale(headerHeight) && x < a.windowButtonStart && !contains(a.tabBounds, pos.x, pos.y) {
		return 2
	}
	return 1
}
