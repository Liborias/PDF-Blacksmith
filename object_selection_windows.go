package main

import (
	"fmt"
	"slices"
)

func (e *objectEditor) resetSelection() {
	e.serial++
	e.selected, e.items, e.image, e.capsule = pdfObject{}, nil, nil, nil
	e.snapshot, e.snapshotComplete, e.copyWaiting = false, false, false
}

func (e *objectEditor) tokens() []int {
	if len(e.items) < 2 {
		return nil
	}
	tokens := make([]int, len(e.items))
	for i, object := range e.items {
		tokens[i] = object.Token
	}
	return tokens
}

func (e *objectEditor) selectionContains(pick textSelection, v *documentView) bool {
	if e.page == nil || len(e.items) == 0 || e.source != pick.source {
		return false
	}
	at := textPointToPDF(pick.point, e.page.Boxes[pick.box], e.rotation(v))
	return slices.ContainsFunc(e.items, func(item pdfObject) bool {
		b := item.Bounds
		return at[0] >= b[0] && at[0] <= b[2] && at[1] >= b[1] && at[1] <= b[3]
	})
}

func (v *documentView) objectRegionPointer(message uint32, x, y int32) bool {
	a, e := v.app, &v.app.objects
	switch message {
	case 0x0201:
		pick := v.textAt(x, y)
		if pick.source == 0 || keyDown(0x20) {
			return false
		}
		if !e.defining {
			if pick.change == replaceSelection && e.selectionContains(pick, v) {
				e.start, e.cursor, e.pressed, e.dragReady = point{x, y}, point{x, y}, true, true
				e.pressPick, e.pick = pick, pick
				setFocus.Call(a.window)
				user32.NewProc("SetCapture").Call(a.window)
				return true
			}
			pick.mode = textRectangle
			if e.mode == 2 {
				pick.mode = textLasso
			}
			e.regionPick, e.defining = pick, true
			e.polygon = []textPoint{pick.point}
			setFocus.Call(a.window)
			if e.mode == 1 {
				user32.NewProc("SetCapture").Call(a.window)
			}
			a.updateObjectActions()
		} else if e.mode == 2 && e.regionPick.source == pick.source {
			area, _ := v.textSheetAt(x, y, pick.source)
			first := screenTextPoint(area, e.polygon[0])
			if len(e.polygon) >= 3 && max(absTextDrag(x-first.x), absTextDrag(y-first.y)) <= int32(a.scale(8)) {
				a.finishObjectRegion()
			} else if len(e.polygon) < 256 {
				e.polygon = append(e.polygon, pick.point)
			}
		}
		e.cursor = point{x, y}
		v.invalidateCanvas()
		return true
	case 0x0200, 0x0202:
		if !e.defining {
			return false
		}
		if area, visible := v.textSheetAt(x, y, e.regionPick.source); visible != 0 {
			e.cursor = point{x, y}
			if e.mode == 1 {
				e.polygon = rectangleSelection(e.regionPick.point, normalizedTextPoint(area, x, y))
				if message == 0x0202 {
					e.defining = false
					user32.NewProc("ReleaseCapture").Call()
					if absTextDrag(x-screenTextPoint(area, e.regionPick.point).x) < int32(a.scale(3)) ||
						absTextDrag(y-screenTextPoint(area, e.regionPick.point).y) < int32(a.scale(3)) {
						e.polygon = nil
						v.selectionClick(x, y, false)
					} else {
						a.finishObjectRegion()
					}
				}
			}
		}
		v.invalidateCanvas()
		return true
	}
	return false
}

func (a *app) finishObjectRegion() {
	e, v := &a.objects, a.view
	e.defining = false
	if len(e.polygon) >= 3 {
		pick := e.regionPick
		pick.polygon = e.polygon
		v.dispatchCanvasSelection(pick, true)
	}
	e.regionPick, e.polygon = textSelection{}, nil
	v.invalidateCanvas()
}

func (a *app) paintObjectRegion(dc uintptr, area rect) {
	e := &a.objects
	polygon := e.polygon
	if e.mode == 1 {
		pick := e.regionPick
		pick.polygon = polygon
		a.drawSelectionRegion(dc, area, pick)
		return
	}
	if e.mode == 2 {
		polygon = append(slices.Clone(polygon), normalizedTextPoint(area, e.cursor.x, e.cursor.y))
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	drawTextOutline(dc, area, polygon, e.mode == 1)
	if e.mode == 2 {
		for _, vertex := range e.polygon {
			p := screenTextPoint(area, vertex)
			n := int32(max(2, a.scale(2)))
			fillColor(dc, rect{p.x - n, p.y - n, p.x + n + 1, p.y + n + 1}, selectionColor)
		}
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}

type objectProbeState struct {
	model              *objectPage
	source, generation int
	next, flight       textSelection
}

func (v *documentView) requestObjects(pick textSelection) {
	p := &v.objectProbe
	if p.model != nil && p.source == pick.source && p.generation == pick.generation {
		p.next, p.flight = textSelection{}, textSelection{}
		v.matchObjectSelection(p.model, pick)
		return
	}
	p.next = pick
	v.schedule()
}

func (v *documentView) matchObjectSelection(page *objectPage, pick textSelection) {
	a, e := v.app, &v.app.objects
	if pick.generation != v.generation || pick.tool != a.mode || a.busy {
		return
	}
	rotation := (page.Rotation + v.rotations[pick.source-1] + 360) % 360
	hits := objectsAt(page.Objects, page.Boxes[pick.box], rotation, pick)
	if len(hits) == 0 && pick.automatic && pick.region() {
		v.requestNativeText(pick)
		return
	}
	if (a.mode != idObjectsTab || v.secondary) && len(hits) == 0 {
		if !v.secondary && a.mode == idTextTab && !pick.region() && pick.change == replaceSelection {
			a.clearTextSelection()
		}
		return
	}
	if a.mode != idObjectsTab || v.secondary {
		var activated bool
		pick, activated = a.activateSelection(v, pick, idObjectsTab, page.Boxes, rotation)
		if !activated {
			return
		}
		v = a.view
		v.objectProbe.model, v.objectProbe.source, v.objectProbe.generation = page, pick.source, pick.generation
	}
	var previous []pdfObject
	if e.page == page && e.source == pick.source {
		previous = e.items
	}
	hits = combineSelection(previous, hits, pick.change, func(item pdfObject) int { return item.Token })
	same := e.page == page && e.source == pick.source && slices.EqualFunc(e.items, hits, func(first, second pdfObject) bool { return first.Token == second.Token })
	e.page, e.source, e.pick = page, pick.source, pick
	if !same {
		e.resetSelection()
		e.items = hits
		if len(hits) > 0 {
			e.selected = hits[0]
		}
	}
	switch len(e.items) {
	case 0:
		setText(e.info, "Na tomto místě není upravitelný objekt.")
	case 1:
		setText(e.info, map[string]string{"path": "Vektorový tvar", "image": "Obrázek", "form": "Skupina PDF"}[e.selected.Kind])
	default:
		setText(e.info, fmt.Sprintf("Vybrané objekty: %d", len(e.items)))
	}
	a.updateObjectActions()
	v.invalidateCanvas()
	a.moveObjectPreview()
	v.schedule()
}

func objectsAt(items []pdfObject, page searchRect, rotation int, pick textSelection) []pdfObject {
	if !pick.region() {
		for i := len(items) - 1; i >= 0; i-- {
			item := items[i]
			b, ok := searchDisplayRect(item.Bounds, page, rotation)
			if !ok || pick.automatic && isSheetBackground(item, b) {
				continue
			}
			if pick.point[0] >= b[0]-pick.tolerance[0] && pick.point[0] <= b[2]+pick.tolerance[0] &&
				pick.point[1] >= b[1]-pick.tolerance[1] && pick.point[1] <= b[3]+pick.tolerance[1] {
				return []pdfObject{item}
			}
		}
		return nil
	}
	var area, pdfArea searchRect
	if pick.mode == textRectangle {
		first, last := pick.polygon[0], pick.polygon[2]
		area = rectangleBetween(first, last)
		pdfArea = rectangleBetween(textPointToPDF(first, page, rotation), textPointToPDF(last, page, rotation))
	}
	var hits []pdfObject
	for _, item := range items {
		b, ok := searchDisplayRect(item.Bounds, page, rotation)
		if !ok || isSheetBackground(item, b) {
			continue
		}
		if pick.mode == textRectangle {
			crossing := pick.crossing()
			if rectangleSelects(area, b, crossing) && (!crossing || objectCrosses(item, pdfArea)) {
				hits = append(hits, item)
			}
		} else if pointInTextRegion(textPoint{(b[0] + b[2]) / 2, (b[1] + b[3]) / 2}, pick.polygon) {
			hits = append(hits, item)
		}
	}
	return hits
}

func objectCrosses(object pdfObject, area searchRect) bool {
	outline := object.Outline
	if outline == nil {
		// The rectangle must reach an edge of an enclosing image or Form.
		bounds := object.Bounds
		return area[0] <= bounds[0] || area[1] <= bounds[1] || area[2] >= bounds[2] || area[3] >= bounds[3]
	}
	area[0] -= outline.Margin[0]
	area[1] -= outline.Margin[1]
	area[2] += outline.Margin[0]
	area[3] += outline.Margin[1]
	for _, contour := range outline.Contours {
		for i := 1; i < len(contour); i++ {
			if rectangleTouchesSegment(area, contour[i-1], contour[i]) {
				return true
			}
		}
	}
	return false
}

func isSheetBackground(object pdfObject, bounds searchRect) bool {
	if object.Outline != nil && !object.Outline.Rectangle {
		return false
	}
	const edgeTolerance = .002
	return bounds[0] <= edgeTolerance && bounds[1] <= edgeTolerance &&
		bounds[2] >= 1-edgeTolerance && bounds[3] >= 1-edgeTolerance
}
