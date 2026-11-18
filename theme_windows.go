package main

import (
	"syscall"
	"unsafe"
)

const (
	appName        = "PDF Blaksmith"
	appIconID      = 2 // rsrc allocates the manifest first, then the icon group.
	windowWidth    = 780
	windowHeight   = 708
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
	canvasColor = paletteColor{23, 23, 27}
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
	msimg32       = syscall.NewLazyDLL("msimg32.dll")
	gdiGradient   = msimg32.NewProc("GradientFill")
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

type gradientVertex struct {
	x, y       int32
	r, g, b, a uint16
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

func gradient(dc uintptr, bounds rect, left, right paletteColor, vertical bool) {
	start := gradientVertex{bounds.left, bounds.top, uint16(left.red) << 8, uint16(left.green) << 8, uint16(left.blue) << 8, 0}
	end := gradientVertex{bounds.right, bounds.bottom, uint16(right.red) << 8, uint16(right.green) << 8, uint16(right.blue) << 8, 0}
	data := [2]gradientVertex{start, end}
	index := [2]uint32{0, 1}
	var mode uintptr
	if vertical {
		mode = 1
	}
	gdiGradient.Call(dc, uintptr(unsafe.Pointer(&data[0])), 2, uintptr(unsafe.Pointer(&index[0])), 1, mode)
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

func (a *app) applyTheme(width, height int32) {
	a.canvasBrush, a.panelBrush, a.inputBrush = makeBrush(canvasColor), makeBrush(panelColor), makeBrush(inputColor)
	a.logo, _, _ = user32.NewProc("LoadImageW").Call(a.instance, appIconID, 1, 128, 128, 0)
	dwm := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if dwm.Find() == nil {
		dark, round, caption := uint32(1), uint32(2), uint32(canvasColor.win())
		dwm.Call(a.window, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwm.Call(a.window, 33, uintptr(unsafe.Pointer(&round)), 4)
		dwm.Call(a.window, 35, uintptr(unsafe.Pointer(&caption)), 4)
	}
	user32.NewProc("SetLayeredWindowAttributes").Call(a.window, 0, 249, 2)
	radius := a.scale(18)
	region, _, _ := gdi32.NewProc("CreateRoundRectRgn").Call(0, 0, uintptr(width+1), uintptr(height+1), uintptr(radius), uintptr(radius))
	result, _, _ := user32.NewProc("SetWindowRgn").Call(a.window, region, 1)
	if result == 0 {
		gdiDelete.Call(region)
	}
}

func (a *app) paint(hwnd uintptr) {
	var info paintInfo
	dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
	fillColor(dc, a.area(0, 0, windowWidth, windowHeight), canvasColor)
	gradient(dc, a.area(0, 0, windowWidth, 108), paletteColor{62, 27, 42}, paletteColor{27, 27, 33}, false)
	gradient(dc, a.area(0, 108, windowWidth, 110), accentColor, paletteColor{49, 39, 50}, false)
	if a.logo != 0 {
		size := uintptr(a.scale(96))
		user32.NewProc("DrawIconEx").Call(dc, uintptr(a.scale(31)), uintptr(a.scale(5)), a.logo, size, size, 0, 0, 3)
	}
	drawText(dc, appName, a.area(145, 19, 735, 66), a.titleFont, textColor, 0)
	drawText(dc, "Tiskové separace, barevné profily a digitální podpisy", a.area(148, 66, 740, 91), a.normalFont, mutedColor, 0)
	rounded(dc, a.area(24, 248, 756, 481), a.scale(20), panelColor, divider)
	user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
}

func (a *app) colorControl(msg uint32, dc, handle uintptr) uintptr {
	gdiTextColor.Call(dc, textColor.win())
	if msg != wmCtlColorText || handle == a.report {
		gdiBkColor.Call(dc, inputColor.win())
		gdiBkMode.Call(dc, 2)
		return a.inputBrush
	}
	gdiBkMode.Call(dc, 1)
	if a.panelLabels[handle] {
		return a.panelBrush
	}
	return a.canvasBrush
}

func (a *app) drawButton(item *buttonDraw) bool {
	if item.controlType != 4 { // ODT_BUTTON
		return false
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
			drawText(item.dc, "✓", box, a.labelFont, textColor, 1)
		} else {
			rounded(item.dc, box, a.scale(5), inputColor, paletteColor{116, 109, 122})
		}
		bounds.left += int32(a.scale(28))
		drawText(item.dc, getText(item.handle), bounds, a.normalFont, foreground, 0)
	} else {
		active := int(item.controlID) == a.mode || int(item.controlID) == idRun
		border, body := divider, paletteColor{45, 43, 50}
		if active {
			border, body = accentColor, paletteColor{112, 42, 69}
		}
		if item.state&1 != 0 {
			body = paletteColor{70, 36, 50}
		}
		if item.state&4 != 0 {
			body, border = background, divider
		}
		rounded(item.dc, bounds, a.scale(12), body, border)
		if active && item.state&5 == 0 {
			inset := int32(a.scale(6))
			inner := rect{bounds.left + inset, bounds.top + 2, bounds.right - inset, bounds.bottom - 2}
			gradient(item.dc, inner, paletteColor{140, 57, 86}, body, true)
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
