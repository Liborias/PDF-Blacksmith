package main

import (
	"slices"
	"unsafe"
)

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
	if pick.source == 0 && contains(v.viewport, x, y) && canvasSelectionChange() == replaceSelection {
		v.clearCanvasContent()
		return
	}
	v.selectCanvasPoint(pick, paragraph)
}

func (v *documentView) selectCanvasPoint(pick textSelection, paragraph bool) {
	if pick.source == 0 {
		return
	}
	a := v.app
	pick.mode = textObject
	pick.paragraphRequested = paragraph
	if paragraph || a.mode != idTextTab || a.textTool.mode == textBlock {
		pick.mode = textBlock
	}
	pick.automatic = true
	pick.combined = pick.change != replaceSelection && v.hasCanvasSelection(pick)
	v.dispatchCanvasSelection(pick)
}

func (v *documentView) dispatchCanvasSelection(pick textSelection) {
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
		a.cancelCanvasRegionProbes()
		if !pick.combined {
			view.clearMixedSelection()
		}
		view.nativeText.next, view.nativeText.flight = textSelection{}, textSelection{}
		view.objectProbe.next, view.objectProbe.flight = textSelection{}, textSelection{}
		view.requestNativeText(pick)
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

type canvasSelectionGesture struct {
	pick     textSelection
	vertices []point
	release  uint32
	hover    point
	hasHover bool
}

func (g *canvasSelectionGesture) active() bool { return len(g.vertices) != 0 }

func (v *documentView) canvasRegionMode() (int, bool) {
	switch v.app.mode {
	case idTextTab:
		mode := v.app.textTool.mode
		return mode, mode == textRectangle || mode == textLasso
	case idOCRTab:
		mode := v.app.ocr.areaMode
		return mode, mode == textRectangle || mode == textLasso
	case idObjectsTab:
		switch v.app.objects.mode {
		case 1:
			return textRectangle, true
		case 2:
			return textLasso, true
		}
	}
	return 0, false
}

func (v *documentView) canvasPointer(message uint32, x, y int32) bool {
	a := v.app
	if !v.secondary && v.canvasContentMovePointer(message, x, y) {
		return true
	}
	if v.selection.active() && v.continueCanvasSelection(message, x, y) {
		return true
	}
	if !v.secondary && a.objects.pressed {
		return v.objectPointer(message, x, y)
	}
	if v.dragging || a.busy || a.full || !v.ready() {
		return false
	}
	switch message {
	case 0x0201, 0x0203, 0x0207: // Left press/double click or middle press.
	default:
		return false
	}
	if !contains(v.viewport, x, y) {
		return false
	}
	pick := v.textAt(x, y)
	a.cancelCanvasRegionProbes()
	mode, region := v.canvasRegionMode()
	if message == 0x0203 {
		if pick.source != 0 && (!region || mode != textLasso) {
			v.selectCanvasPoint(pick, true)
		}
		return true
	}
	pos := point{x, y}
	if message == 0x0207 {
		v.beginCanvasSelection(pick, pos, textRectangle, 0x0208)
		return true
	}
	if keyDown(0x20) {
		return false
	}
	if a.mode == idTextTab && a.textTool.mode == textFrame && pick.change == replaceSelection {
		if pick.source != 0 {
			v.beginCanvasSelection(pick, pos, textFrame, 0x0202)
		} else {
			setFocus.Call(a.window)
		}
		return true
	}
	if v.beginCanvasContentMove(pick, pos) {
		return true
	}
	if region {
		v.beginCanvasSelection(pick, pos, mode, 0x0202)
		return true
	}
	if pick.source == 0 {
		if a.mode == idObjectsTab || a.mode == idTextTab {
			v.beginCanvasSelection(pick, pos, textRectangle, 0x0202)
			return true
		}
		return false
	}
	if !v.secondary && a.mode == idObjectsTab && v.objectPointer(message, x, y) {
		return true
	}
	if a.mode == idTextTab {
		if !v.secondary && pick.change == replaceSelection && a.editFrameAt(pick) {
			return true
		}
		setFocus.Call(a.window)
		v.selectCanvasPoint(pick, a.textTool.explicitParagraph)
		return true
	}
	return false
}

func (v *documentView) beginCanvasSelection(pick textSelection, pos point, mode int, release uint32) {
	if pick.source == 0 {
		pick = v.canvasPick()
	}
	pick.mode = mode
	vertices := []point{pos, pos}
	if mode == textLasso {
		vertices, release = vertices[:1], 0
	}
	v.selection = canvasSelectionGesture{pick: pick, vertices: vertices, release: release}
	v.app.comparison.canvasFocused = v.secondary
	setFocus.Call(v.app.window)
	if release != 0 {
		user32.NewProc("SetCapture").Call(v.app.window)
	}
	v.invalidateCanvas()
}

func (v *documentView) continueCanvasSelection(message uint32, x, y int32) bool {
	g := &v.selection
	if g.pick.generation != v.generation || g.pick.tool != v.app.mode ||
		message == 0x001f || message == 0x0215 && g.release != 0 {
		v.cancelCanvasSelection()
		return false
	}
	if g.pick.mode == textLasso {
		return v.continueCanvasLasso(message, x, y)
	}
	if message != 0x0200 && message != g.release {
		return message == 0x0201 || message == 0x0203 || message == 0x0207
	}
	last := point{min(max(x, v.viewport.left), v.viewport.right), min(max(y, v.viewport.top), v.viewport.bottom)}
	g.vertices[1] = last
	if message == g.release {
		first, threshold := g.vertices[0], int32(v.app.scale(3))
		if absTextDrag(last.x-first.x) < threshold || absTextDrag(last.y-first.y) < threshold {
			mode, change := g.pick.mode, g.pick.change
			click := g.release == 0x0202 && mode != textFrame
			v.cancelCanvasSelection()
			if mode == textFrame {
				if frame := v.app.inlineCurrent(); frame != nil && frame.newBounds != nil && !v.app.inlineDirty() {
					v.app.clearTextSelection()
				}
			}
			if click {
				pick := v.textAt(x, y)
				if pick.source == 0 && contains(v.viewport, x, y) && change == replaceSelection {
					v.clearCanvasContent()
				} else {
					pick.change = change
					v.selectCanvasPoint(pick, false)
				}
			}
		} else {
			v.finishCanvasSelection()
		}
	} else {
		v.invalidateCanvas()
	}
	return true
}

func (v *documentView) continueCanvasLasso(message uint32, x, y int32) bool {
	g := &v.selection
	pos := point{x, y}
	switch message {
	case 0x0201, 0x0203:
		if !contains(v.viewport, x, y) {
			if message == 0x0201 {
				v.cancelCanvasSelection()
			}
			return false
		}
		first := g.vertices[0]
		close := len(g.vertices) >= 3 && max(absTextDrag(pos.x-first.x), absTextDrag(pos.y-first.y)) <= int32(v.app.scale(9))
		if message == 0x0203 || close {
			v.finishCanvasSelection()
		} else if len(g.vertices) < 256 && pos != g.vertices[len(g.vertices)-1] {
			g.vertices = append(g.vertices, pos)
			g.hasHover = false
			v.invalidateCanvas()
		}
		return true
	case 0x0200, 0x02a3: // Move or leave.
		show := message == 0x0200 && contains(v.viewport, x, y)
		if show && !g.hasHover {
			trackMouseLeave(v.app.window)
		}
		if show != g.hasHover || show && pos != g.hover {
			g.hasHover, g.hover = show, pos
			v.invalidateCanvas()
		}
		return show
	case 0x0207:
		v.cancelCanvasSelection()
	}
	return false
}

func selectionScreenPolygon(g canvasSelectionGesture) []textPoint {
	toPoint := func(p point) textPoint { return textPoint{float64(p.x), float64(p.y)} }
	if g.pick.mode != textLasso {
		return rectangleSelection(toPoint(g.vertices[0]), toPoint(g.vertices[1]))
	}
	polygon := make([]textPoint, len(g.vertices))
	for i, vertex := range g.vertices {
		polygon[i] = toPoint(vertex)
	}
	return polygon
}

func (v *documentView) canvasRegionPage(startSource int, polygon []textPoint, mode int) (rect, int) {
	var best rect
	var source int
	var bestOverlap float64
	bounds := polygonBounds(polygon)
	_, h := v.viewportSize()
	for _, page := range v.spread.visible(v.scrollY, h) {
		number := v.order[page.index]
		if startSource != 0 && number != startSource {
			continue
		}
		area := v.pageBounds(page)
		visible := searchRect{float64(max(area.left, v.viewport.left)), float64(max(area.top, v.viewport.top)),
			float64(min(area.right, v.viewport.right)), float64(min(area.bottom, v.viewport.bottom))}
		overlap := rectangleOverlapArea(bounds, visible)
		if overlap > bestOverlap && (mode != textLasso || polygonTouchesRect(polygon, visible)) {
			best, source, bestOverlap = area, number, overlap
		}
	}
	return best, source
}

func (v *documentView) finishCanvasSelection() bool {
	g := v.selection
	if len(g.vertices) < 3 && g.pick.mode == textLasso {
		return false
	}
	polygon := selectionScreenPolygon(g)
	pick := g.pick
	v.cancelCanvasSelection()
	area, source := v.canvasRegionPage(pick.source, polygon, pick.mode)
	if source == 0 {
		return true
	}
	if pick.mode == textFrame {
		v.finishTextFrameDrag(pick, area, g.vertices[0], g.vertices[1])
		return true
	}
	pick.source = source
	for i, vertex := range polygon {
		polygon[i] = normalizedTextPoint(area, int32(vertex[0]), int32(vertex[1]))
	}
	pick.polygon, pick.point = polygon, polygon[0]
	pick.combined = pick.tool != idOCRTab
	pick.automatic = !pick.combined
	v.dispatchCanvasSelection(pick)
	return true
}

func (v *documentView) cancelCanvasSelection() bool {
	if v == nil || !v.selection.active() {
		return false
	}
	captured := v.selection.release != 0
	v.selection = canvasSelectionGesture{}
	v.invalidateCanvas()
	if captured {
		if capture, _, _ := user32.NewProc("GetCapture").Call(); capture == v.app.window {
			user32.NewProc("ReleaseCapture").Call()
		}
	}
	if v.app.mode == idTextTab && v.app.inlineDirty() {
		postMessage.Call(v.app.window, wmInlineCommit, uintptr(v.app.textTool.inline.serial), 0)
	}
	return true
}

func (a *app) cancelCanvasSelections() bool {
	primary := a.view.cancelCanvasSelection()
	move := a.view.cancelCanvasContentMove()
	return a.comparison.view.cancelCanvasSelection() || primary || move
}

func (v *documentView) paintCanvasSelection(dc uintptr) {
	g := &v.selection
	if !g.active() || g.pick.generation != v.generation {
		return
	}
	if g.pick.mode == textFrame {
		if area, source := v.textSheetAt(0, 0, g.pick.source); source != 0 {
			v.app.drawTextFrameOutline(dc, intersectInlineRect(area, textFrameDragRect(g.vertices[0], g.vertices[1])))
		}
		return
	}
	if g.pick.mode == textRectangle {
		first, last := g.vertices[0], g.vertices[1]
		preview := textSelection{mode: textRectangle, polygon: rectangleSelection(
			normalizedTextPoint(v.viewport, first.x, first.y), normalizedTextPoint(v.viewport, last.x, last.y))}
		v.app.drawSelectionRegion(dc, v.viewport, preview)
		return
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, v.app.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	vertices := g.vertices
	if g.hasHover {
		vertices = append(slices.Clone(vertices), g.hover)
	}
	drawScreenOutline(dc, vertices)
	n := int32(max(2, v.app.scale(2)))
	for _, pos := range g.vertices {
		fillColor(dc, rect{pos.x - n, pos.y - n, pos.x + n + 1, pos.y + n + 1}, selectionColor)
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}

func (a *app) setCanvasCursor() bool {
	if a.mode != idTextTab || a.full || a.busy || keyDown(0x20) {
		return false
	}
	cursorID := uintptr(32513) // IDC_IBEAM
	if a.textTool.mode == textFrame {
		cursorID = 32515 // IDC_CROSS
	} else if !a.canvasTextCursor || a.textTool.mode >= textRectangle {
		return false
	}
	var p point
	user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&p)))
	user32.NewProc("ScreenToClient").Call(a.window, uintptr(unsafe.Pointer(&p)))
	v := a.view
	if a.comparisonVisible() && contains(a.comparison.view.viewport, p.x, p.y) {
		v = a.comparison.view
	}
	if v.overview || v.dragging || v.selection.active() || !v.ready() || !contains(v.viewport, p.x, p.y) {
		return false
	}
	if _, source := v.textSheetAt(p.x, p.y, 0); source == 0 {
		return false
	}
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, cursorID)
	user32.NewProc("SetCursor").Call(cursor)
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
