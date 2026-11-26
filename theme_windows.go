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

func (color paletteColor) win() uintptr {
	return uintptr(color.red) | uintptr(color.green)<<8 | uintptr(color.blue)<<16
}

var (
	uiColors         = uiPalettes.colors(uiPalettes.Default)
	titleColor       = uiColors["top"]
	canvasColor      = uiColors["base"]
	panelColor       = uiColors["panel"]
	editorColor      = uiColors["editor"]
	inputColor       = uiColors["input"]
	textColor        = uiColors["text"]
	mutedColor       = uiColors["muted"]
	disabledColor    = mutedColor
	selectionColor   = uiColors["accent"]
	divider          = uiColors["border"]
	sidebarRailColor = uiColors["surface"]
	paperInkColor    = paletteColor{26, 26, 29} // Text on fixed light paper placeholders.

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

func drawCheck(dc uintptr, bounds rect, color paletteColor) {
	width, height := bounds.right-bounds.left, bounds.bottom-bounds.top
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, width/9)), color.win())
	previous, _, _ := gdiSelect.Call(dc, pen)
	gdi32.NewProc("MoveToEx").Call(dc, uintptr(bounds.left+width/4), uintptr(bounds.top+height/2), 0)
	gdi32.NewProc("LineTo").Call(dc, uintptr(bounds.left+width*7/16), uintptr(bounds.top+height*11/16))
	gdi32.NewProc("LineTo").Call(dc, uintptr(bounds.left+width*3/4), uintptr(bounds.top+height/3))
	gdiSelect.Call(dc, previous)
	gdiDelete.Call(pen)
}

func (a *app) drawFocus(dc uintptr, bounds rect) {
	inset := int32(a.scale(5))
	line := bounds.bottom - int32(a.scale(3))
	fillColor(dc, rect{bounds.left + inset, line, bounds.right - inset, line + 1}, uiColors["focus"])
}

func drawText(dc uintptr, text string, bounds rect, font uintptr, color paletteColor, align uintptr) {
	old, _, _ := gdiSelect.Call(dc, font)
	gdiTextColor.Call(dc, color.win())
	gdiBkMode.Call(dc, 1)
	winDrawText.Call(dc, uintptr(unsafe.Pointer(wideString(text))), ^uintptr(0),
		uintptr(unsafe.Pointer(&bounds)), 0x0824|align)
	gdiSelect.Call(dc, old)
}

func (a *app) captionWidth(text string, font uintptr) int {
	dc, _, _ := user32.NewProc("GetDC").Call(a.window)
	if dc == 0 {
		return 0
	}
	old, _, _ := gdiSelect.Call(dc, font)
	var bounds rect
	winDrawText.Call(dc, uintptr(unsafe.Pointer(wideString(text))), ^uintptr(0), uintptr(unsafe.Pointer(&bounds)),
		0x0c20) // DT_CALCRECT | DT_NOPREFIX | DT_SINGLELINE
	gdiSelect.Call(dc, old)
	user32.NewProc("ReleaseDC").Call(a.window, dc)
	return int(bounds.right)
}

func (a *app) area(left, top, right, bottom int) rect {
	return rect{int32(a.scale(left)), int32(a.scale(top)), int32(a.scale(right)), int32(a.scale(bottom))}
}

func (a *app) applyTheme() {
	a.applyPalette()
	a.logo, _, _ = user32.NewProc("LoadImageW").Call(a.instance, appIconID, 1, 128, 128, 0)
	a.loadBrand()
}

func (a *app) applyPalette() {
	uiColors = uiPalettes.colors(a.uiTheme)
	titleColor, canvasColor, panelColor = uiColors["top"], uiColors["base"], uiColors["panel"]
	editorColor, inputColor = uiColors["editor"], uiColors["input"]
	textColor, mutedColor = uiColors["text"], uiColors["muted"]
	disabledColor, selectionColor, divider = mutedColor, uiColors["accent"], uiColors["border"]
	sidebarRailColor = uiColors["surface"]
	for _, entry := range []struct {
		brush *uintptr
		color paletteColor
	}{
		{&a.canvasBrush, canvasColor}, {&a.panelBrush, panelColor}, {&a.inputBrush, inputColor},
	} {
		if brush := makeBrush(entry.color); brush != 0 {
			old := *entry.brush
			*entry.brush = brush
			if old != 0 {
				gdiDelete.Call(old)
			}
		}
	}
	clear(uiIconBitmaps)
	a.applyWindowTheme(a.window)
	a.recolorTooltips()
}

func (a *app) applyWindowTheme(handle uintptr) {
	dwm := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if dwm.Find() == nil {
		dark, round, caption := uint32(0), uint32(2), uint32(titleColor.win())
		if a.uiTheme.Mode == "dark" {
			dark = 1
		}
		dwm.Call(handle, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwm.Call(handle, 33, uintptr(unsafe.Pointer(&round)), 4)
		dwm.Call(handle, 35, uintptr(unsafe.Pointer(&caption)), 4)
	}
}

func (a *app) setUITheme(theme uiTheme) {
	if theme == a.uiTheme {
		return
	}
	a.uiTheme = theme
	a.applyPalette()
	a.updateThemeToggleCaption()
	a.view.invalidateCanvas()
	if a.comparison.view != nil {
		a.comparison.view.invalidateCanvas()
	}
	redraw := user32.NewProc("RedrawWindow")
	redraw.Call(a.window, 0, 0, 0x0085) // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN
	if dialog := a.interfaceDialog; dialog != nil && dialog.window != 0 {
		a.applyWindowTheme(dialog.window)
		redraw.Call(dialog.window, 0, 0, 0x0085)
	}
}

func (a *app) paint(hwnd uintptr) {
	var info paintInfo
	dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
	if !a.full {
		fillColor(dc, rect{0, 0, int32(a.width), int32(a.scale(headerHeight))}, titleColor)
		fillColor(dc, rect{0, int32(a.scale(headerHeight)) - 1, int32(a.width), int32(a.scale(headerHeight))}, divider)
		barTop, barBottom := int32(a.scale(headerHeight)), int32(a.scale(workspaceTop))
		fillColor(dc, rect{0, barTop, int32(a.width), barBottom}, titleColor)
		fillColor(dc, rect{0, barBottom - 1, int32(a.width), barBottom}, divider)
		fillColor(dc, rect{0, a.viewport.bottom, int32(a.width), int32(a.height)}, canvasColor)
		fillColor(dc, rect{0, a.viewport.bottom, int32(a.width), a.viewport.bottom + 1}, divider)
		a.paintBrand(dc, a.area(8, 4, a.headerBrand-22, headerHeight-4))
		a.paintSidebars(dc)
	}
	if a.comparisonVisible() {
		fillColor(dc, a.viewport, canvasColor)
		mid := (a.viewport.left + a.viewport.right) / 2
		fillColor(dc, rect{mid - 1, a.viewport.top, mid + 1, a.viewport.bottom}, divider)
		a.comparison.view.paint(dc)
	}
	a.view.paint(dc)
	user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
}

func (a *app) colorControl(msg uint32, dc, handle uintptr) uintptr {
	gdiTextColor.Call(dc, textColor.win())
	if handle == a.focusHelp {
		gdiTextColor.Call(dc, uiColors["helpText"].win())
	}
	if a.scrollPanes[handle] != nil {
		gdiBkColor.Call(dc, panelColor.win())
		gdiBkMode.Call(dc, 2)
		return a.panelBrush
	}
	if msg != wmCtlColorText {
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
	if a.drawCommandBarButton(item) || a.drawDocumentTab(item) || a.drawWindowButton(item) || a.drawPanelToggle(item) || a.drawPropertySectionButton(item) || a.drawChoice(item) || a.drawTextIcon(item) {
		return true
	}
	bounds := item.bounds
	isOption := a.options[item.handle]
	id := int(item.controlID)
	if a.toolButtons[id] == item.handle {
		a.drawToolNavigation(item)
		return true
	}
	background := canvasColor
	if a.panelButtons[item.handle] {
		background = panelColor
	}
	fillColor(item.dc, bounds, background)
	foreground := textColor
	if item.state&4 != 0 {
		foreground = disabledColor
	}
	if isOption {
		offset, edge := int32(a.scale(2)), int32(a.scale(17))
		top := bounds.top + (bounds.bottom-bounds.top-edge)/2
		box := rect{bounds.left + offset, top, bounds.left + offset + edge, top + edge}
		rounded(item.dc, box, a.scale(5), inputColor, divider)
		if a.checked(item.handle) {
			if item.state&4 == 0 {
				foreground = selectionColor
			}
			drawCheck(item.dc, box, foreground)
		}
		bounds.left += int32(a.scale(28))
		drawText(item.dc, getText(item.handle), bounds, a.normalFont, foreground, 0)
	} else {
		selectedDetail := a.mode
		if a.tool == idFileTab {
			selectedDetail = a.activeFileOption()
		}
		active := (a.detailButtons[id] == item.handle && id == selectedDetail) ||
			(id >= idObjectCropRectangle && id <= idObjectCropCircle && a.checked(item.handle)) ||
			a.primaryButtons[item.handle] || (id == idReaderDeletePages && a.view.pageDrag.overDelete)
		body, edge := uiColors["control"], divider
		if active && item.state&4 == 0 {
			foreground = selectionColor
		}
		if item.state&1 != 0 {
			body = uiColors["hover"]
			foreground = textColor
		}
		if item.state&4 != 0 {
			body, edge = background, divider
			foreground = disabledColor
		}
		rounded(item.dc, bounds, a.scale(5), body, edge)
		if !a.drawToolIconLabel(item, foreground, body) {
			align := uintptr(1)
			if id == idCompareTitleLeft || id == idCompareTitleRight {
				align |= 0x8000 // DT_END_ELLIPSIS
			}
			drawText(item.dc, getText(item.handle), bounds, a.labelFont, foreground, align)
		}
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (a *app) disposeTheme() {
	for _, object := range []uintptr{a.canvasBrush, a.panelBrush, a.inputBrush, a.normalFont, a.titleFont, a.labelFont, a.railFont} {
		if object != 0 {
			gdiDelete.Call(object)
		}
	}
	if a.logo != 0 {
		user32.NewProc("DestroyIcon").Call(a.logo)
	}
}
