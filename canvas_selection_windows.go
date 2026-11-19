package main

func keyDown(key uintptr) bool {
	state, _, _ := user32.NewProc("GetKeyState").Call(key)
	return state&0x8000 != 0
}

func canvasSelectionChange() selectionChange {
	shift, ctrl := keyDown(0x10), keyDown(0x11)
	switch {
	case shift && ctrl:
		return subtractSelection
	case ctrl:
		return toggleSelection
	case shift:
		return addSelection
	default:
		return replaceSelection
	}
}

func (pick textSelection) region() bool { return len(pick.polygon) >= 3 }
func (pick textSelection) crossing() bool {
	return pick.mode == textRectangle && pick.region() && pick.polygon[2][0] < pick.polygon[0][0]
}

func (v *documentView) selectionClick(x, y int32, paragraph bool) {
	a := v.app
	if a.busy || a.full || !v.ready() {
		return
	}
	pick := v.textAt(x, y)
	if pick.source == 0 {
		return
	}
	pick.mode = textObject
	pick.paragraphRequested = paragraph
	if paragraph || a.mode != idTextTab || a.textTool.mode == textBlock {
		pick.mode = textBlock
	}
	pick.automatic = true
	v.dispatchCanvasSelection(pick, false)
}

func (v *documentView) dispatchCanvasSelection(pick textSelection, objects bool) {
	a := v.app
	position := a.selectionPagePosition(v, pick.source)
	target := a.activeTab
	if v.secondary {
		target = a.comparison.tab
	}
	apply := func() {
		view := a.view
		if a.activeTab != target {
			if !a.comparisonVisible() || a.comparison.tab != target {
				return
			}
			view = a.comparison.view
		}
		if position < 0 || position >= len(view.order) {
			return
		}
		pick.source, pick.generation, pick.tool = view.order[position], view.generation, a.mode
		if objects {
			view.requestObjects(pick)
		} else {
			view.requestNativeText(pick)
		}
	}
	if !a.saveInlineBefore(apply) {
		apply()
	}
}

func (a *app) selectionPagePosition(v *documentView, source int) int {
	for index, page := range v.order {
		if page == source {
			return index
		}
	}
	return -1
}

func (a *app) activateSelection(v *documentView, pick textSelection, tool int, boxes map[string]searchRect, rotation int) (textSelection, bool) {
	if v.secondary {
		tab := a.comparison.tab
		a.selectDocumentTab(tab)
		if a.activeTab != tab {
			return pick, false
		}
	}
	if a.mode != tool {
		a.activate(tool)
	}
	primary := a.view
	box := primary.box()
	if pick.box != box {
		old, current := boxes[pick.box], boxes[box]
		convert := func(point textPoint) textPoint {
			return pdfPointToText(textPointToPDF(point, old, rotation), current, rotation)
		}
		pick.point = convert(pick.point)
		for index, point := range pick.polygon {
			pick.polygon[index] = convert(point)
		}
	}
	pick.generation, pick.tool, pick.box = primary.generation, tool, box
	return pick, a.mode == tool
}

func (v *documentView) middleSelection(message uint32, x, y int32) bool {
	a := v.app
	switch message {
	case 0x0207: // WM_MBUTTONDOWN
		if a.busy || a.full || !v.ready() {
			return false
		}
		pick := v.textAt(x, y)
		if pick.source == 0 {
			return false
		}
		pick.mode = textRectangle
		v.marquee = pick
		setFocus.Call(a.window)
		user32.NewProc("SetCapture").Call(a.window)
		return true
	case 0x0200, 0x0208: // WM_MOUSEMOVE, WM_MBUTTONUP
		pick := v.marquee
		if pick.source == 0 {
			return false
		}
		area, source := v.textSheetAt(x, y, pick.source)
		if source != 0 {
			pick.polygon = rectangleSelection(pick.point, normalizedTextPoint(area, x, y))
			v.marquee = pick
		}
		if message == 0x0208 {
			v.marquee = textSelection{}
			user32.NewProc("ReleaseCapture").Call()
			if source != 0 && pick.generation == v.generation &&
				absTextDrag(x-screenTextPoint(area, pick.point).x) >= int32(a.scale(3)) &&
				absTextDrag(y-screenTextPoint(area, pick.point).y) >= int32(a.scale(3)) {
				pick.automatic = a.mode != idObjectsTab && a.mode != idTextTab || v.secondary
				v.dispatchCanvasSelection(pick, a.mode != idTextTab)
			}
		}
		v.invalidateCanvas()
		return true
	case 0x0215, 0x001f: // WM_CAPTURECHANGED, WM_CANCELMODE
		v.cancelMiddleSelection()
	}
	return false
}

func (v *documentView) cancelMiddleSelection() bool {
	if v == nil || v.marquee.source == 0 {
		return false
	}
	v.marquee = textSelection{}
	v.invalidateCanvas()
	user32.NewProc("ReleaseCapture").Call()
	return true
}

func (a *app) drawSelectionRegion(dc uintptr, area rect, pick textSelection) {
	style, width := uintptr(0), uintptr(max(1, a.scale(2)))
	if pick.crossing() {
		style, width = 1, 1 // PS_DASH is available for cosmetic pens.
	}
	pen, _, _ := gdiPen.Call(style, width, selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	background, _, _ := gdi32.NewProc("SetBkMode").Call(dc, 1) // TRANSPARENT
	drawTextOutline(dc, area, pick.polygon, true)
	gdi32.NewProc("SetBkMode").Call(dc, background)
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}
