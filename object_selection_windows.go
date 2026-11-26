package main

import "slices"

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
	if pick.combined {
		v.finishCanvasRegion(page, pick)
		return
	}
	rotation := (page.Rotation + v.rotations[pick.source-1] + 360) % 360
	hits := objectsAt(page.Objects, page.Boxes[pick.box], rotation, pick)
	if (a.mode != idObjectsTab || v.secondary) && len(hits) == 0 {
		text := &v.nativeText
		if pick.insertPoint && text.source == pick.source && text.generation == pick.generation && text.model != nil {
			a.openNewText(v, text.model, pick)
			return
		}
		if !v.secondary && a.mode == idTextTab && !pick.region() && pick.change == replaceSelection {
			a.clearTextSelection()
		}
		return
	}
	pick.insertPoint = false
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
	v.applyObjectSelection(page, pick, combineSelection(previous, hits, pick.change, objectKey))
}

func objectKey(item pdfObject) int { return item.Token }

func (v *documentView) applyObjectSelection(page *objectPage, pick textSelection, hits []pdfObject) {
	a, e := v.app, &v.app.objects
	showedCrop := a.hasImageCropTarget()
	showedReset := showedCrop && e.selected.ImageMask != nil
	same := e.page == page && e.source == pick.source && slices.EqualFunc(e.items, hits, func(first, second pdfObject) bool { return first.Token == second.Token })
	e.page, e.source, e.pick = page, pick.source, pick
	if !same {
		a.cancelImageCrop()
		e.resetSelection()
		e.items = hits
		if len(hits) > 0 {
			e.selected = hits[0]
		}
	}
	switch len(e.items) {
	case 0:
		a.setUIMessage(e.info, uiNoEditableObject)
	case 1:
		a.setUIMessage(e.info, map[string]uiMessage{"path": uiVectorShape, "image": uiSelectedImage, "form": uiPDFGroup}[e.selected.Kind])
	default:
		a.setUIMessage(e.info, uiSelectedObjects, len(e.items))
	}
	a.updateObjectActions()
	showCrop := a.hasImageCropTarget()
	if showedCrop != showCrop || showedReset != (showCrop && e.selected.ImageMask != nil) {
		a.layoutPanel()
	}
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
	area := polygonBounds(pick.polygon)
	var pdfArea searchRect
	if pick.mode == textRectangle {
		first, last := pick.polygon[0], pick.polygon[2]
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
		} else if !boundsContain(b, area) && polygonTouchesRect(pick.polygon, b) {
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
