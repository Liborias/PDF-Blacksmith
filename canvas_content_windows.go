package main

import "slices"

type mixedCanvasSelection struct {
	pick       textSelection
	text       *textPage
	objects    *objectPage
	glyphs     []layoutGlyph
	items      []pdfObject
	textBounds []searchRect
	dim        textPoint
}

func glyphKey(g layoutGlyph) [2]int { return [2]int{g.run, g.index} }

func glyphsInRegion(glyphs []layoutGlyph, dim textPoint, pick textSelection) []layoutGlyph {
	var area searchRect
	if pick.mode == textRectangle {
		area = rectangleBetween(pick.polygon[0], pick.polygon[2])
	}
	return slices.DeleteFunc(glyphs, func(g layoutGlyph) bool {
		b := searchRect{g.bounds[0] / dim[0], g.bounds[1] / dim[1], g.bounds[2] / dim[0], g.bounds[3] / dim[1]}
		if pick.mode == textRectangle {
			return !rectangleSelects(area, b, pick.crossing())
		}
		return !polygonTouchesRect(pick.polygon, b)
	})
}

func (v *documentView) cancelCanvasSelectionProbe() bool {
	if v == nil {
		return false
	}
	canceled := v.regionText != nil
	v.regionText = nil
	for _, pick := range []*textSelection{&v.nativeText.next, &v.nativeText.flight, &v.objectProbe.next, &v.objectProbe.flight} {
		if pick.source != 0 {
			*pick = textSelection{}
			canceled = true
		}
	}
	return canceled
}

func (a *app) cancelCanvasSelectionProbes() {
	a.view.cancelCanvasSelectionProbe()
	a.comparison.view.cancelCanvasSelectionProbe()
}

func (v *documentView) failCanvasRegion() {
	v.cancelCanvasSelectionProbe()
	v.app.reportStatus(v.app.uiText(uiCanvasRegionUnavailable))
}

func (v *documentView) clearMixedSelection() bool {
	old := v.mixed.pick
	if old.source == 0 {
		return false
	}
	v.mixed = mixedCanvasSelection{}
	a := v.app
	switch old.tool {
	case idTextTab:
		setText(a.textTool.report, "")
	case idObjectsTab:
		a.setUIMessage(a.objects.info, uiSelectPageObject)
	}
	v.invalidateCanvas()
	return true
}

func (v *documentView) clearCanvasContent() bool {
	if v == nil {
		return false
	}
	canceled := v.cancelCanvasSelectionProbe()
	return v.clearMixedSelection() || canceled
}

func (pick textSelection) sameCanvasPage(previous textSelection) bool {
	return pick.source != 0 && pick.source == previous.source && pick.generation == previous.generation && pick.box == previous.box
}

func (v *documentView) hasCanvasSelection(pick textSelection) bool {
	if v.secondary {
		return false
	}
	a := v.app
	return pick.sameCanvasPage(v.mixed.pick) ||
		(a.mode == idTextTab && a.textTool.flow.text != "" && pick.sameCanvasPage(a.textTool.pick)) ||
		(a.mode == idObjectsTab && len(a.objects.items) != 0 && pick.sameCanvasPage(a.objects.pick))
}

func (v *documentView) previousCanvasContent(pick textSelection, text *textPage, page *objectPage) ([]layoutGlyph, []pdfObject) {
	if pick.change == replaceSelection || v.secondary {
		return nil, nil
	}
	var glyphs []layoutGlyph
	var items []pdfObject
	if m := &v.mixed; pick.sameCanvasPage(m.pick) {
		if m.text == text {
			glyphs = m.glyphs
		}
		if m.objects == page {
			items = m.items
		}
		return glyphs, items
	}
	if p := &v.app.textTool; p.model == text && pick.sameCanvasPage(p.pick) {
		glyphs = p.flow.glyphs()
	}
	if e := &v.app.objects; e.page == page && pick.sameCanvasPage(e.pick) {
		items = e.items
	}
	return glyphs, items
}

func (v *documentView) finishCanvasRegion(page *objectPage, pick textSelection) {
	text := v.regionText
	v.regionText = nil
	if text == nil {
		return
	}
	a := v.app
	previousGlyphs, previousObjects := v.previousCanvasContent(pick, text, page)
	glyphs, dim, rotation := v.nativeTextGeometry(text, pick)
	var selectedGlyphs []layoutGlyph
	if pick.region() {
		selectedGlyphs = glyphsInRegion(glyphs, dim, pick)
	} else {
		flow, _ := textAtPoint(text, glyphs, dim, pick)
		selectedGlyphs = flow.glyphs()
	}
	objectRotation := (page.Rotation + v.rotations[pick.source-1] + 360) % 360
	var selectedObjects []pdfObject
	if pick.region() || len(selectedGlyphs) == 0 {
		selectedObjects = objectsAt(page.Objects, page.Boxes[pick.box], objectRotation, pick)
	}
	glyphs = combineSelection(previousGlyphs, selectedGlyphs, pick.change, glyphKey)
	items := combineSelection(previousObjects, selectedObjects, pick.change, objectKey)
	flow := layoutText(glyphs)
	hasText := flow.text != ""
	mixed := hasText && len(items) != 0

	tool := idObjectsTab
	switch {
	case mixed && pick.tool == idTextTab, hasText && !mixed:
		tool = idTextTab
	case pick.region() && len(items) == 0 && pick.tool == idTextTab:
		// Keep an empty region available for OCR once neither type was hit.
		tool = idTextTab
	case len(items) == 0 && pick.tool != idObjectsTab:
		v.clearMixedSelection()
		return
	}

	oldBox, oldTool := pick.box, pick.tool
	boxes, angle := page.Boxes, objectRotation
	if tool == idTextTab {
		boxes, angle = text.Boxes, rotation
	}
	pick, ok := a.activateSelection(v, pick, tool, boxes, angle)
	if !ok {
		return
	}
	v = a.view
	v.clearMixedSelection()
	v.nativeText.model, v.nativeText.source, v.nativeText.generation = text, pick.source, pick.generation
	v.objectProbe.model, v.objectProbe.source, v.objectProbe.generation = page, pick.source, pick.generation

	if oldBox != pick.box && (mixed || tool == idTextTab) {
		selected := make(map[[2]int]struct{}, len(glyphs))
		for _, glyph := range glyphs {
			selected[glyphKey(glyph)] = struct{}{}
		}
		glyphs, dim, rotation = v.nativeTextGeometry(text, pick)
		glyphs = slices.DeleteFunc(glyphs, func(g layoutGlyph) bool {
			_, found := selected[glyphKey(g)]
			return !found
		})
		flow = layoutText(glyphs)
	}
	if tool == idTextTab && oldTool != idTextTab {
		a.textTool.mode, a.textTool.explicitParagraph = pick.mode, false
		a.syncTextSelectionModes()
	}
	if !mixed {
		if tool == idTextTab {
			a.applyTextSelection(v, text, pick, flow, nil, dim, rotation)
		} else {
			v.applyObjectSelection(page, pick, items)
		}
		return
	}
	a.clearTextSelection()
	a.clearObjectSelection()
	v.mixed = mixedCanvasSelection{pick: pick, text: text, objects: page, glyphs: glyphs, items: items, textBounds: flow.highlightBounds(), dim: dim}
	report := a.objects.info
	if tool == idTextTab {
		report = a.textTool.report
	}
	a.setUIMessage(report, uiMixedCanvasSelection, len(glyphs), len(items))
	v.invalidateCanvas()
}

func (v *documentView) paintMixedSelection(dc uintptr, source int, area rect) {
	m := &v.mixed
	if v.app.full || source != m.pick.source || v.generation != m.pick.generation {
		return
	}
	m.paintTextHighlight(dc, area, int32(max(1, v.app.scale(2))))

	rotation := (m.objects.Rotation + v.rotations[source-1] + 360) % 360
	box := m.objects.Boxes[m.pick.box]
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, v.app.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	for _, item := range m.items {
		if b, ok := searchDisplayRect(item.Bounds, box, rotation); ok {
			drawTextOutline(dc, area, rectangleSelection(textPoint{b[0], b[1]}, textPoint{b[2], b[3]}), true)
		}
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}

func (m *mixedCanvasSelection) paintTextHighlight(dc uintptr, area rect, stroke int32) {
	mask, _, _ := createSurfaceDC.Call(dc)
	if mask != 0 {
		defer deleteSurfaceDC.Call(mask)
	}
	bitmap, _, _ := createSurfaceBitmap.Call(dc, 1, 1)
	if bitmap != 0 {
		defer gdiDelete.Call(bitmap)
	}
	var tinted bool
	if mask != 0 && bitmap != 0 {
		if old, _, _ := gdiSelect.Call(mask, bitmap); old != 0 && old != ^uintptr(0) {
			defer gdiSelect.Call(mask, old)
			fillColor(mask, rect{right: 1, bottom: 1}, selectionColor)
			tinted = true
		}
	}
	for _, b := range m.textBounds {
		first := screenTextPoint(area, textPoint{b[0] / m.dim[0], b[1] / m.dim[1]})
		last := screenTextPoint(area, textPoint{b[2] / m.dim[0], b[3] / m.dim[1]})
		if tinted && last.x > first.x && last.y > first.y {
			alphaBlend.Call(dc, uintptr(first.x), uintptr(first.y), uintptr(last.x-first.x), uintptr(last.y-first.y), mask, 0, 0, 1, 1, 87<<16)
		}
		fillColor(dc, rect{first.x, last.y, last.x, last.y + stroke}, selectionColor)
	}
}
