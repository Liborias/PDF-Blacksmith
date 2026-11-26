package main

type canvasContentMove struct {
	pick       textSelection
	serial     uint64
	start, end point
	text       []glyphSelection
	tokens     []int
	bounds     []searchRect // Page-normalized coordinates at the start of the gesture.
	box        searchRect
	rotation   int
	requested  bool
	dropped    bool
	settling   bool
	backdrop   *printBitmap
	foreground *printBitmap
	layer      viewportSurface
	coverage   searchRect
}

func (v *documentView) beginCanvasContentMove(pick textSelection, start point) bool {
	a := v.app
	if v.secondary || a.full || a.busy || pick.source == 0 || pick.change != replaceSelection || keyDown(0x20) {
		return false
	}
	// Area tools still start a new selection normally. Alt drags existing content.
	if _, area := v.canvasRegionMode(); (area || a.canvasTextCursor) && !keyDown(0x12) {
		return false
	}
	move := canvasContentMove{pick: pick, start: start, end: start}
	var model *textPage
	var glyphs []layoutGlyph
	var lines []searchRect
	var dim textPoint
	var items []pdfObject
	if m := &v.mixed; pick.sameCanvasPage(m.pick) {
		model, glyphs, lines, dim = m.text, m.glyphs, m.textBounds, m.dim
		items = m.items
		move.box, move.rotation = m.objects.Boxes[pick.box], m.objects.Rotation+v.rotations[pick.source-1]
	} else if p := &a.textTool; a.mode == idTextTab &&
		pick.sameCanvasPage(p.pick) && p.flow.text != "" && !a.inlineDirty() {
		model, glyphs, lines, dim = p.model, p.flow.glyphs(), p.flow.highlightBounds(), p.dim
	} else if e := &a.objects; a.mode == idObjectsTab && pick.sameCanvasPage(e.pick) {
		items, move.box, move.rotation = e.items, e.page.Boxes[pick.box], e.rotation(v)
	}
	for _, item := range items {
		if bounds, visible := searchDisplayRect(item.Bounds, move.box, move.rotation%360); visible {
			move.bounds = append(move.bounds, bounds)
		}
		move.tokens = append(move.tokens, item.Token)
	}
	if model != nil && len(glyphs) > 0 && dim[0] > 0 && dim[1] > 0 {
		for _, line := range lines {
			move.bounds = append(move.bounds, searchRect{line[0] / dim[0], line[1] / dim[1], line[2] / dim[0], line[3] / dim[1]})
		}
		move.text = glyphSelectionFor(model, glyphs)
		move.box, move.rotation = model.Boxes[pick.box], model.Rotation+v.rotations[pick.source-1]
	}
	move.rotation %= 360
	r := searchRect{pick.point[0], pick.point[1], pick.point[0], pick.point[1]}
	for _, bounds := range move.bounds {
		if boundsContain(bounds, r) {
			v.moveSerial++
			move.serial = v.moveSerial
			v.dragging = false
			v.contentMove = move
			setFocus.Call(a.window)
			// A probe may hand off an already captured press to this gesture.
			// Calling SetCapture again can send WM_CAPTURECHANGED even to us.
			if capture, _, _ := user32.NewProc("GetCapture").Call(); capture != a.window {
				user32.NewProc("SetCapture").Call(a.window)
			}
			return true
		}
	}
	return false
}

func (m *canvasContentMove) activePixels(scale int) bool {
	return m.pick.source != 0 && max(absTextDrag(m.end.x-m.start.x), absTextDrag(m.end.y-m.start.y)) > int32(scale)
}

func (m *canvasContentMove) paintBounds() searchRect {
	var result searchRect
	for _, bounds := range m.bounds {
		result = unionPageRects(result, bounds)
	}
	w, h := m.box[2]-m.box[0], m.box[3]-m.box[1]
	if m.rotation%180 != 0 {
		w, h = h, w
	}
	// Include stroke edges, underlines and the selection adornments in the clean patch.
	x, y := 6/max(1, w), 6/max(1, h)
	return searchRect{max(0, result[0]-x), max(0, result[1]-y), min(1, result[2]+x), min(1, result[3]+y)}
}

func (v *documentView) canvasContentMovePointer(message uint32, x, y int32) bool {
	m := &v.contentMove
	if m.pick.source == 0 || m.dropped {
		return false
	}
	switch message {
	case 0x001f, 0x0215: // Cancel mode or capture changed.
		v.cancelCanvasContentMove()
		return false
	case 0x0200, 0x0202: // Move or release.
		previous := m.end
		m.end = point{x, y}
		a := v.app
		dragged := m.activePixels(a.scale(4))
		if message == 0x0200 {
			if dragged {
				v.schedule()
			}
			v.redrawContentDrag(previous)
			if a.textTool.inline.visible && dragged {
				a.layoutInlineEditor()
			} else if !dragged {
				a.layoutInlineEditor()
			}
			return true
		}
		target := v.textAt(x, y)
		if !target.sameCanvasPage(m.pick) || a.busy {
			v.cancelCanvasContentMove()
			return true
		}
		if !dragged {
			v.cancelCanvasContentMove()
			if !a.editFrameAt(target) && v.mixed.pick.source == 0 {
				v.selectionClick(x, y, false)
			}
			return true
		}
		m.dropped = true
		user32.NewProc("ReleaseCapture").Call()
		return v.commitCanvasContentMove(target)
	}
	return message == 0x0201 || message == 0x0203 || message == 0x0207
}

func (v *documentView) commitCanvasContentMove(target textSelection) bool {
	a, m := v.app, &v.contentMove
	box, source := m.box, m.pick.source
	start := textPointToPDF(m.pick.point, box, m.rotation)
	end := textPointToPDF(target.point, box, m.rotation)
	delta := textPoint{end[0] - start[0], end[1] - start[1]}
	operation, grouping := "move_selection", ""
	if len(m.text) == 0 {
		var accepted bool
		grouping, accepted = a.chooseObjectGrouping()
		if !accepted || m.pick.source != source {
			v.cancelCanvasContentMove()
			return true
		}
		operation = "move"
	}
	base := m.paintBounds()
	shift := textPoint{target.point[0] - m.pick.point[0], target.point[1] - m.pick.point[1]}
	damage := &pageDamage{source, m.pick.box, unionPageRects(base, moveCropBounds(base, shift))}
	req := v.request("object_edit")
	req.View.Page, req.View.Rotation = source, v.rotations[source-1]
	req.ObjectEdit = &objectEditRequest{Operation: operation, Grouping: grouping, Tokens: m.tokens, Delta: delta}
	req.TextSelection = m.text
	a.runDamagedDraft(req, a.uiText(uiMovingSelection), 0, damage)
	if !a.busy {
		v.cancelCanvasContentMove()
	}
	return true
}

func (v *documentView) cancelCanvasContentMove() bool {
	if v == nil || v.contentMove.pick.source == 0 {
		return false
	}
	m := &v.contentMove
	v.redrawContentDrag(m.end)
	m.layer.close()
	v.contentMove = canvasContentMove{}
	v.dragging = false
	if capture, _, _ := user32.NewProc("GetCapture").Call(); capture == v.app.window {
		user32.NewProc("ReleaseCapture").Call()
	}
	v.app.layoutInlineEditor()
	return true
}

func (v *documentView) scheduleContentDragPreview() bool {
	m := &v.contentMove
	if m.dropped || m.requested || !m.activePixels(v.app.scale(4)) || v.app.busy {
		return false
	}
	m.requested = true
	req := v.request("object_drag_preview")
	req.View.Page, req.View.Rotation = m.pick.source, v.rotations[m.pick.source-1]
	bounds := m.paintBounds()
	req.View.DragBounds, req.View.DragZoom = &bounds, min(v.zoom, 64)
	req.ObjectEdit = &objectEditRequest{Tokens: m.tokens}
	req.TextSelection = m.text
	v.launch(req, m.pick.source, 0, false, nil)
	return true
}

func (v *documentView) finishContentDragPreview(job readerJob) {
	m, result := &v.contentMove, job.result.View
	if job.generation != v.generation || job.dragSerial != m.serial || result == nil ||
		job.frame.image == nil || !result.DragLayer {
		return
	}
	m.backdrop, m.coverage = job.frame.image, result.DragBounds
	m.foreground = &printBitmap{width: m.backdrop.width, height: m.backdrop.height, pixels: result.dragPixels}
	v.redrawContentDrag(m.end)
}

// Keep the completed gesture on screen until the driver's first correct
// visible pixels replace it. A failed or cancelled edit clears it immediately.
func (v *documentView) settleContentPreview(page int) {
	m := &v.contentMove
	if !m.settling || page != m.pick.source {
		return
	}
	highZoom := false
	for _, tile := range v.visibleDetails {
		if tile.page != page {
			continue
		}
		highZoom = true
		if v.detailFrames[tile].image == nil && !v.failedDetails[tile] {
			return
		}
		if _, pending := v.rasterDamage[tile]; pending {
			return
		}
	}
	if !highZoom && v.reading.frames[page].image == nil && !v.reading.failed[page] {
		return
	}
	if !highZoom {
		if _, pending := v.rasterDamage[screenTile{page: page}]; pending {
			return
		}
	}
	v.cancelCanvasContentMove()
}

func dragScreenBounds(page rect, coverage searchRect, dx, dy int32) rect {
	first := screenTextPoint(page, textPoint{coverage[0], coverage[1]})
	last := screenTextPoint(page, textPoint{coverage[2], coverage[3]})
	return rect{first.x + dx, first.y + dy, last.x + dx, last.y + dy}
}

func (v *documentView) redrawContentDrag(previous point) {
	m := &v.contentMove
	if page, source := v.textSheetAt(0, 0, m.pick.source); source != 0 {
		bounds := m.paintBounds()
		if m.backdrop != nil {
			bounds = m.coverage
		}
		for _, cursor := range []point{m.start, previous, m.end} {
			area := dragScreenBounds(page, bounds, cursor.x-m.start.x, cursor.y-m.start.y)
			area = intersectInlineRect(rect{area.left - 3, area.top - 3, area.right + 3, area.bottom + 3}, v.viewport)
			if area.right > area.left && area.bottom > area.top {
				v.invalidateOverlay(area)
			}
		}
	}
}

func (v *documentView) paintCanvasContentMove(dc uintptr) {
	m := &v.contentMove
	if !m.activePixels(v.app.scale(4)) {
		return
	}
	page, source := v.textSheetAt(0, 0, m.pick.source)
	if source == 0 {
		return
	}
	clip := intersectInlineRect(v.viewport, page)
	saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
	defer gdi32.NewProc("RestoreDC").Call(dc, saved)
	gdi32.NewProc("IntersectClipRect").Call(dc, uintptr(clip.left), uintptr(clip.top), uintptr(clip.right), uintptr(clip.bottom))
	dx, dy := m.end.x-m.start.x, m.end.y-m.start.y
	if m.foreground != nil && cropPreviewSurface(dc, &m.layer, m.foreground, false) {
		base := dragScreenBounds(page, m.coverage, 0, 0)
		pixels := m.backdrop
		drawBitmap(dc, pixels, int(base.left), int(base.top), int(base.right-base.left), int(base.bottom-base.top))
		alphaBlend.Call(dc, uintptr(base.left+dx), uintptr(base.top+dy),
			uintptr(base.right-base.left), uintptr(base.bottom-base.top),
			m.layer.dc, 0, 0, uintptr(m.layer.width), uintptr(m.layer.height), 255<<16|1<<24)
		return
	}
	// While the layers are being prepared, indicate the position without
	// moving or capturing pixels from the underlying page.
	page.left, page.right = page.left+dx, page.right+dx
	page.top, page.bottom = page.top+dy, page.bottom+dy
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, v.app.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	for _, bounds := range m.bounds {
		drawTextOutline(dc, page, rectangleSelection(textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}), true)
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}
