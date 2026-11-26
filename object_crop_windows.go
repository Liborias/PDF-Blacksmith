package main

type imageCrop struct {
	active, dragging, hasRect bool
	start, end                textPoint
	bounds                    searchRect
}

func (a *app) canCropImage() bool {
	e, v := &a.objects, a.view
	return a.mode == idObjectsTab && !a.full && !a.busy && v.ready() && !v.overview &&
		e.page != nil && e.pick.generation == v.generation && len(e.items) == 1 && e.selected.Kind == "image"
}

func (a *app) beginImageCrop() {
	if !a.canCropImage() || a.objects.crop.active {
		return
	}
	a.cancelCanvasSelections()
	a.cancelCanvasRegionProbes()
	a.cancelObjectGesture()
	e := &a.objects
	e.crop.active = true
	a.setUIMessage(e.cropHint, uiCropImageHint)
	a.updateObjectActions()
	a.layoutPanel()
	a.view.invalidateCanvas()
}

func (a *app) cancelImageCrop() bool {
	e := &a.objects
	if !e.crop.active {
		return false
	}
	dragging := e.crop.dragging
	e.crop = imageCrop{}
	if dragging {
		user32.NewProc("ReleaseCapture").Call()
	}
	setText(e.cropHint, "")
	a.updateObjectActions()
	a.layoutPanel()
	a.view.invalidateCanvas()
	return true
}

func (a *app) applyImageCrop() {
	e := &a.objects
	if !a.canCropImage() || !e.crop.active || !e.crop.hasRect || e.crop.dragging {
		return
	}
	bounds := e.crop.bounds
	a.runObjectEdit(e.source, objectEditRequest{Operation: "crop", Token: e.selected.Token, Crop: &bounds})
}

func (v *documentView) imageCropPointer(message uint32, x, y int32) bool {
	a, e := v.app, &v.app.objects
	crop := &e.crop
	if !crop.active || !a.canCropImage() {
		return false
	}
	switch message {
	case 0x0215, 0x001f: // Capture changed or cancelled.
		if crop.dragging {
			crop.dragging = false
			a.updateObjectActions()
			v.invalidateCanvas()
		}
		return false
	case 0x0201, 0x0203:
		if keyDown(0x20) || !contains(v.viewport, x, y) {
			return false
		}
	case 0x0200, 0x0202:
		if !crop.dragging {
			return false
		}
	default:
		return false
	}

	area, source := v.textSheetAt(0, 0, e.source)
	if source == 0 {
		a.cancelImageCrop()
		return true
	}
	box, angle := e.page.Boxes[v.box()], e.rotation(v)
	image, visible := searchDisplayRect(e.selected.Bounds, box, angle)
	if !visible {
		a.cancelImageCrop()
		return true
	}
	at := normalizedTextPoint(area, x, y)
	switch message {
	case 0x0201, 0x0203:
		if at[0] >= image[0] && at[0] <= image[2] && at[1] >= image[1] && at[1] <= image[3] {
			crop.start, crop.end, crop.dragging = at, at, true
			a.comparison.canvasFocused = false
			setFocus.Call(a.window)
			user32.NewProc("SetCapture").Call(a.window)
			a.updateObjectActions()
			v.invalidateCanvas()
		}
	case 0x0200, 0x0202:
		crop.end = textPoint{min(image[2], max(image[0], at[0])), min(image[3], max(image[1], at[1]))}
		if message == 0x0202 {
			start, end := screenTextPoint(area, crop.start), screenTextPoint(area, crop.end)
			minimum := int32(a.scale(4))
			if absTextDrag(end.x-start.x) >= minimum && absTextDrag(end.y-start.y) >= minimum {
				bounds := rectangleBetween(textPointToPDF(crop.start, box, angle), textPointToPDF(crop.end, box, angle))
				if bounds != e.selected.Bounds {
					crop.bounds, crop.hasRect = bounds, true
				}
			}
			crop.dragging = false
			user32.NewProc("ReleaseCapture").Call()
			a.updateObjectActions()
		}
		v.invalidateCanvas()
	}
	return true
}

func (a *app) paintImageCrop(dc uintptr, area rect) {
	e := &a.objects
	box, angle := e.page.Boxes[a.view.box()], e.rotation(a.view)
	if bounds, visible := searchDisplayRect(e.selected.Bounds, box, angle); visible {
		start := screenTextPoint(area, textPoint{bounds[0], bounds[1]})
		end := screenTextPoint(area, textPoint{bounds[2], bounds[3]})
		a.drawTextFrameOutline(dc, rect{start.x, start.y, end.x, end.y})
	}
	bounds, visible := searchDisplayRect(e.crop.bounds, box, angle)
	if e.crop.dragging {
		bounds, visible = rectangleBetween(e.crop.start, e.crop.end), true
	}
	if !visible {
		return
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	drawTextOutline(dc, area, rectangleSelection(textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}), true)
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}
