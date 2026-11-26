package main

type canvasContentMove struct {
	pick       textSelection
	start, end point
	text       []glyphSelection
	tokens     []int
	bounds     []searchRect // Page-normalized coordinates at the start of the gesture.
	box        searchRect
	rotation   int
}

func (v *documentView) beginCanvasContentMove(pick textSelection, start point) bool {
	a := v.app
	if v.secondary || a.full || a.busy || pick.source == 0 || pick.change != replaceSelection || keyDown(0x20) {
		return false
	}
	// The area and text tools must still accept a new selection over existing content.
	if _, area := v.canvasRegionMode(); (area || a.canvasTextCursor) && !keyDown(0x12) {
		return false
	}
	move := canvasContentMove{pick: pick, start: start, end: start}
	var model *textPage
	var glyphs []layoutGlyph
	var lines []searchRect
	var dim textPoint
	if m := &v.mixed; pick.sameCanvasPage(m.pick) {
		model, glyphs, lines, dim = m.text, m.glyphs, m.textBounds, m.dim
		angle := m.objects.Rotation + v.rotations[pick.source-1]
		for _, item := range m.items {
			if bounds, visible := searchDisplayRect(item.Bounds, m.objects.Boxes[pick.box], angle); visible {
				move.bounds = append(move.bounds, bounds)
			}
			move.tokens = append(move.tokens, item.Token)
		}
	} else if p := &a.textTool; a.mode == idTextTab &&
		pick.sameCanvasPage(p.pick) && p.flow.text != "" && !a.inlineDirty() {
		model, glyphs, lines, dim = p.model, p.flow.glyphs(), p.flow.highlightBounds(), p.dim
	}
	if model == nil || len(glyphs) == 0 {
		return false
	}
	for _, line := range lines {
		move.bounds = append(move.bounds, searchRect{line[0] / dim[0], line[1] / dim[1], line[2] / dim[0], line[3] / dim[1]})
	}
	found := false
	r := searchRect{pick.point[0], pick.point[1], pick.point[0], pick.point[1]}
	for _, bounds := range move.bounds {
		if boundsContain(bounds, r) {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	move.text = glyphSelectionFor(model, glyphs)
	move.box, move.rotation = model.Boxes[pick.box], model.Rotation+v.rotations[pick.source-1]
	v.contentMove = move
	setFocus.Call(a.window)
	user32.NewProc("SetCapture").Call(a.window)
	return true
}

func (m canvasContentMove) dragged(scale int) bool {
	return max(absTextDrag(m.end.x-m.start.x), absTextDrag(m.end.y-m.start.y)) > int32(scale)
}

func (v *documentView) canvasContentMovePointer(message uint32, x, y int32) bool {
	m := &v.contentMove
	if m.pick.source == 0 {
		return false
	}
	switch message {
	case 0x001f, 0x0215: // Cancel mode or capture changed.
		v.cancelCanvasContentMove()
		return false
	case 0x0200, 0x0202: // Move or release.
		m.end = point{x, y}
		if message == 0x0200 {
			v.invalidateCanvas()
			return true
		}
		move := *m
		v.cancelCanvasContentMove()
		target := v.textAt(x, y)
		if !target.sameCanvasPage(move.pick) || v.app.busy {
			return true
		}
		a := v.app
		if !move.dragged(a.scale(4)) {
			if !a.editFrameAt(target) && v.mixed.pick.source == 0 {
				v.selectCanvasPoint(target, false)
			}
			return true
		}
		start := textPointToPDF(move.pick.point, move.box, move.rotation)
		end := textPointToPDF(target.point, move.box, move.rotation)
		req := v.request("object_edit")
		req.View.Page, req.View.Rotation = move.pick.source, v.rotations[move.pick.source-1]
		req.ObjectEdit = &objectEditRequest{Operation: "move_selection", Tokens: move.tokens,
			Delta: textPoint{end[0] - start[0], end[1] - start[1]}}
		req.TextSelection = move.text
		a.runDocumentDraft(req, a.uiText(uiMovingSelection), 0)
		return true
	}
	return message == 0x0201 || message == 0x0203 || message == 0x0207
}

func (v *documentView) cancelCanvasContentMove() bool {
	if v == nil || v.contentMove.pick.source == 0 {
		return false
	}
	v.contentMove = canvasContentMove{}
	if capture, _, _ := user32.NewProc("GetCapture").Call(); capture == v.app.window {
		user32.NewProc("ReleaseCapture").Call()
	}
	v.invalidateCanvas()
	return true
}

func (v *documentView) paintCanvasContentMove(dc uintptr, source int, area rect) {
	m := &v.contentMove
	if m.pick.source != source || !m.dragged(v.app.scale(4)) {
		return
	}
	area.left, area.right = area.left+m.end.x-m.start.x, area.right+m.end.x-m.start.x
	area.top, area.bottom = area.top+m.end.y-m.start.y, area.bottom+m.end.y-m.start.y
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, v.app.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	for _, bounds := range m.bounds {
		drawTextOutline(dc, area, rectangleSelection(textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}), true)
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}
