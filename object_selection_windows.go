package main

import (
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
		space, _, _ := user32.NewProc("GetKeyState").Call(0x20)
		if pick.source == 0 || space&0x8000 != 0 {
			return false
		}
		if !e.defining {
			if e.selectionContains(pick, v) {
				e.start, e.cursor, e.pressed = point{x, y}, point{x, y}, true
				e.pick.point = pick.point
				setFocus.Call(a.window)
				user32.NewProc("SetCapture").Call(a.window)
				return true
			}
			e.resetSelection()
			e.pick, e.defining, e.start = pick, true, point{x, y}
			e.polygon = []textPoint{pick.point}
			setFocus.Call(a.window)
			if e.mode == 1 {
				user32.NewProc("SetCapture").Call(a.window)
			}
			a.updateObjectActions()
		} else if e.mode == 2 && e.pick.source == pick.source {
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
		if area, visible := v.textSheetAt(x, y, e.pick.source); visible != 0 {
			e.cursor = point{x, y}
			if e.mode == 1 {
				e.polygon = rectangleSelection(e.pick.point, normalizedTextPoint(area, x, y))
				if message == 0x0202 {
					e.defining = false
					user32.NewProc("ReleaseCapture").Call()
					a.finishObjectRegion()
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
		e.pick.polygon = e.polygon
		if e.page != nil && e.pick.source == e.source {
			a.pickObject()
		} else {
			setText(e.info, "Zjišťuji objekty na stránce…")
			v.schedule()
		}
	}
	e.polygon = nil
	v.invalidateCanvas()
}

func (a *app) paintObjectRegion(dc uintptr, area rect) {
	e := &a.objects
	polygon := e.polygon
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
