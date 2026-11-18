package main

import (
	"syscall"
	"unsafe"
)

const (
	appName        = "PDF Blacksmith"
	storageName    = "PDF Blaksmith"
	appIconID      = 2 // rsrc allocates the manifest first, then the icon group.
	wmPaint        = 0x000f
	wmEraseBkgnd   = 0x0014
	wmDrawItem     = 0x002b
	wmCtlColorEdit = 0x0133
	wmCtlColorList = 0x0134
	wmCtlColorText = 0x0138
	bsOwnerDraw    = 0x000b
)

type paletteColor struct{ red, green, blue uint8 }

func (color paletteColor) win() uintptr {
	return uintptr(color.red) | uintptr(color.green)<<8 | uintptr(color.blue)<<16
}

var (
	titleColor  = paletteColor{22, 22, 24}
	canvasColor = paletteColor{26, 26, 29}
	panelColor  = paletteColor{33, 33, 39}
	inputColor  = paletteColor{19, 19, 23}
	textColor   = paletteColor{245, 242, 244}
	mutedColor  = paletteColor{186, 178, 188}
	accentColor = paletteColor{173, 77, 109}
	divider     = paletteColor{65, 61, 70}

	gdiDelete     = gdi32.NewProc("DeleteObject")
	gdiSelect     = gdi32.NewProc("SelectObject")
	gdiBrush      = gdi32.NewProc("CreateSolidBrush")
	gdiPen        = gdi32.NewProc("CreatePen")
	gdiRoundRect  = gdi32.NewProc("RoundRect")
	gdiTextColor  = gdi32.NewProc("SetTextColor")
	gdiBkColor    = gdi32.NewProc("SetBkColor")
	gdiBkMode     = gdi32.NewProc("SetBkMode")
	winFill       = user32.NewProc("FillRect")
	winDrawText   = user32.NewProc("DrawTextW")
	winInvalidate = user32.NewProc("InvalidateRect")
	copyDrawItem  = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlMoveMemory")
)

type paintInfo struct {
	dc        uintptr
	erase     int32
	bounds    rect
	restore   int32
	incUpdate int32
	reserved  [32]byte
}

type buttonDraw struct {
	controlType, controlID, itemID, action, state uint32
	handle, dc                                    uintptr
	bounds                                        rect
	itemData                                      uintptr
}

func makeBrush(color paletteColor) uintptr {
	brush, _, _ := gdiBrush.Call(color.win())
	return brush
}

func invalidate(hwnd uintptr) {
	winInvalidate.Call(hwnd, 0, 0)
}

func fillColor(dc uintptr, bounds rect, color paletteColor) {
	brush := makeBrush(color)
	winFill.Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
	gdiDelete.Call(brush)
}

func rounded(dc uintptr, bounds rect, radius int, fill, border paletteColor) {
	brush := makeBrush(fill)
	pen, _, _ := gdiPen.Call(0, 1, border.win())
	previousBrush, _, _ := gdiSelect.Call(dc, brush)
	previousPen, _, _ := gdiSelect.Call(dc, pen)
	gdiRoundRect.Call(dc, uintptr(bounds.left), uintptr(bounds.top), uintptr(bounds.right),
		uintptr(bounds.bottom), uintptr(radius), uintptr(radius))
	gdiSelect.Call(dc, previousBrush)
	gdiSelect.Call(dc, previousPen)
	gdiDelete.Call(brush)
	gdiDelete.Call(pen)
}

func drawCheck(dc uintptr, bounds rect) {
	width, height := bounds.right-bounds.left, bounds.bottom-bounds.top
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, width/9)), textColor.win())
	previous, _, _ := gdiSelect.Call(dc, pen)
	gdi32.NewProc("MoveToEx").Call(dc, uintptr(bounds.left+width/4), uintptr(bounds.top+height/2), 0)
	gdi32.NewProc("LineTo").Call(dc, uintptr(bounds.left+width*7/16), uintptr(bounds.top+height*11/16))
	gdi32.NewProc("LineTo").Call(dc, uintptr(bounds.left+width*3/4), uintptr(bounds.top+height/3))
	gdiSelect.Call(dc, previous)
	gdiDelete.Call(pen)
}

func drawText(dc uintptr, text string, bounds rect, font uintptr, color paletteColor, align uintptr) {
	old, _, _ := gdiSelect.Call(dc, font)
	gdiTextColor.Call(dc, color.win())
	gdiBkMode.Call(dc, 1)
	winDrawText.Call(dc, uintptr(unsafe.Pointer(wideString(text))), ^uintptr(0),
		uintptr(unsafe.Pointer(&bounds)), 0x0824|align)
	gdiSelect.Call(dc, old)
}

func (a *app) area(left, top, right, bottom int) rect {
	return rect{int32(a.scale(left)), int32(a.scale(top)), int32(a.scale(right)), int32(a.scale(bottom))}
}

func (a *app) applyTheme() {
	a.canvasBrush, a.panelBrush, a.inputBrush = makeBrush(canvasColor), makeBrush(panelColor), makeBrush(inputColor)
	a.logo, _, _ = user32.NewProc("LoadImageW").Call(a.instance, appIconID, 1, 128, 128, 0)
	dwm := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if dwm.Find() == nil {
		dark, round, caption := uint32(1), uint32(2), uint32(canvasColor.win())
		dwm.Call(a.window, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwm.Call(a.window, 33, uintptr(unsafe.Pointer(&round)), 4)
		dwm.Call(a.window, 35, uintptr(unsafe.Pointer(&caption)), 4)
	}
	a.loadBrand()
}

func (a *app) paint(hwnd uintptr) {
	var info paintInfo
	dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
	fillColor(dc, rect{0, 0, int32(a.width), int32(a.height)}, canvasColor)
	if !a.full {
		fillColor(dc, rect{0, 0, int32(a.width), int32(a.scale(49))}, titleColor)
		fillColor(dc, rect{0, int32(a.scale(49)), int32(a.width), int32(a.scale(50))}, divider)
		fillColor(dc, rect{0, int32(a.scale(90)), int32(a.width), int32(a.scale(91))}, divider)
		fillColor(dc, rect{0, a.viewport.bottom, int32(a.width), a.viewport.bottom + 1}, divider)
		a.paintBrand(dc)
	}
	a.view.paint(dc)
	user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
}

func (a *app) colorControl(msg uint32, dc, handle uintptr) uintptr {
	gdiTextColor.Call(dc, textColor.win())
	if msg != wmCtlColorText || a.readOnlyFields[handle] {
		gdiBkColor.Call(dc, inputColor.win())
		gdiBkMode.Call(dc, 2)
		return a.inputBrush
	}
	gdiBkMode.Call(dc, 1)
	if handle == a.documentName {
		return a.canvasBrush
	}
	if a.panelLabels[handle] {
		return a.panelBrush
	}
	return a.canvasBrush
}

func (a *app) drawButton(item *buttonDraw) bool {
	if item.controlType != 4 { // ODT_BUTTON
		return false
	}
	if a.drawWindowButton(item) || a.drawChoice(item) {
		return true
	}
	bounds := item.bounds
	isOption := a.options[item.handle]
	background := canvasColor
	if a.panelButtons[item.handle] {
		background = panelColor
	}
	fillColor(item.dc, bounds, background)
	foreground := textColor
	if item.state&4 != 0 {
		foreground = mutedColor
	}
	if isOption {
		offset, edge := int32(a.scale(2)), int32(a.scale(17))
		top := bounds.top + (bounds.bottom-bounds.top-edge)/2
		box := rect{bounds.left + offset, top, bounds.left + offset + edge, top + edge}
		if a.checked(item.handle) {
			rounded(item.dc, box, a.scale(5), paletteColor{113, 42, 70}, accentColor)
			drawCheck(item.dc, box)
		} else {
			rounded(item.dc, box, a.scale(5), inputColor, paletteColor{116, 109, 122})
		}
		bounds.left += int32(a.scale(28))
		drawText(item.dc, getText(item.handle), bounds, a.normalFont, foreground, 0)
	} else {
		_, tab := a.modes[int(item.controlID)]
		active := int(item.controlID) == a.mode || a.primaryButtons[item.handle]
		body, edge := paletteColor{42, 42, 47}, divider
		if tab {
			body, edge = canvasColor, canvasColor
			if active {
				body, edge = panelColor, panelColor
			}
		} else if active {
			body, edge = paletteColor{91, 42, 63}, paletteColor{128, 58, 84}
		}
		if item.state&1 != 0 {
			body = paletteColor{69, 52, 62}
		}
		if item.state&4 != 0 {
			body, edge = background, divider
			foreground = paletteColor{119, 114, 122}
		}
		rounded(item.dc, bounds, a.scale(5), body, edge)
		if tab && active {
			fillColor(item.dc, rect{bounds.left + 1, bounds.bottom - int32(a.scale(3)), bounds.right - 1, bounds.bottom - 1}, accentColor)
		}
		drawText(item.dc, getText(item.handle), bounds, a.labelFont, foreground, 1)
	}
	if item.state&16 != 0 {
		focus := rect{bounds.left + 3, bounds.top + 3, bounds.right - 3, bounds.bottom - 3}
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&focus)))
	}
	return true
}

func (a *app) disposeTheme() {
	for _, object := range []uintptr{a.canvasBrush, a.panelBrush, a.inputBrush, a.normalFont, a.titleFont, a.labelFont} {
		if object != 0 {
			gdiDelete.Call(object)
		}
	}
	if a.logo != 0 {
		user32.NewProc("DestroyIcon").Call(a.logo)
	}
}
