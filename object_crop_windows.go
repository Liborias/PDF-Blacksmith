package main

import "fmt"

var imageCropRatios = [...]struct{ width, height int }{
	{0, 1}, {1, 1}, {4, 3}, {3, 4}, {16, 9}, {9, 16}, {3, 2}, {2, 3}, {5, 4}, {4, 5}, {21, 9},
}

type imageCropShape int

const (
	cropRectangle imageCropShape = iota
	cropSquare
	cropOval
	cropCircle
)

func (shape imageCropShape) round() bool {
	return shape == cropOval || shape == cropCircle
}

func (shape imageCropShape) locksRatio() bool {
	return shape == cropSquare || shape == cropCircle
}

type imageCropDrag struct {
	start, end textPoint
}

type imageCrop struct {
	active, dragging bool
	shape            imageCropShape
	drag             imageCropDrag
	selection        *imageCropDrag
}

func (a *app) buildCropOptions() {
	e := &a.objects
	for i, label := range [...]uiMessage{uiCropRectangle, uiCropSquare, uiCropOval, uiCropCircle} {
		e.cropShapes[i] = a.uiButton(label, idObjectCropRectangle+i, 16+i%2*153, 340+i/2*37, 145, 0)
	}
	names := make([]string, len(imageCropRatios))
	messages := make([]uiMessage, len(names))
	messages[0], messages[1] = uiCropFreeRatio, uiCropSquareRatio
	for i, ratio := range imageCropRatios[2:] {
		names[i+2] = fmt.Sprintf("%d : %d", ratio.width, ratio.height)
	}
	e.cropRatioLabel = a.uiLabel(uiCropAspectRatio, 16, 418, 145, false)
	e.cropRatio = a.comboWithMessages(idObjectCropRatio, 169, 414, 145, names, messages)
	a.bindUIHelp(e.cropRatio, uiCropAspectRatio)
}

func (a *app) chooseImageCropShape(shape imageCropShape) {
	a.beginImageCrop()
	e := &a.objects
	if !e.crop.active || e.crop.dragging {
		return
	}
	e.crop.shape = shape
	if shape.locksRatio() {
		chooseCombo(e.cropRatio, 1)
	}
	a.updateObjectActions()
	a.view.invalidateCanvas()
}

func (a *app) cropEnd(drag imageCropDrag, image, box searchRect, angle int) textPoint {
	preset := imageCropRatios[max(0, comboIndex(a.objects.cropRatio))]
	ratio := float64(preset.width) / float64(preset.height)
	return constrainImageCrop(drag.start, drag.end, image, box, angle, ratio)
}

func (a *app) imageCropLimits() (image, box searchRect, angle int, visible bool) {
	e, v := &a.objects, a.view
	box, angle = e.page.Boxes[v.box()], e.rotation(v)
	image, visible = searchDisplayRect(e.selected.cropSourceBounds(), box, angle)
	return
}

func (object pdfObject) cropSourceBounds() searchRect {
	if object.ImageMask != nil {
		return object.ImageMask.SourceBounds
	}
	return object.Bounds
}

func (a *app) imageCropBounds() (searchRect, bool) {
	e := &a.objects
	if !a.canCropImage() || !e.crop.active || e.crop.dragging || e.crop.selection == nil {
		return searchRect{}, false
	}
	image, box, angle, visible := a.imageCropLimits()
	if !visible {
		return searchRect{}, false
	}
	drag := *e.crop.selection
	end := a.cropEnd(drag, image, box, angle)
	bounds := rectangleBetween(textPointToPDF(drag.start, box, angle), textPointToPDF(end, box, angle))
	return bounds, (e.crop.shape.round() || bounds != e.selected.cropSourceBounds()) &&
		bounds[2] > bounds[0] && bounds[3] > bounds[1]
}

func (a *app) hasImageCropTarget() bool {
	e, v := &a.objects, a.view
	return a.mode == idObjectsTab && v.ready() && !v.overview &&
		e.page != nil && e.pick.generation == v.generation && len(e.items) == 1 && e.selected.Kind == "image"
}

func (a *app) canCropImage() bool {
	return !a.full && !a.busy && a.hasImageCropTarget()
}

func (a *app) beginImageCrop() {
	if !a.canCropImage() || a.objects.crop.active {
		return
	}
	a.cancelCanvasSelections()
	a.cancelCanvasSelectionProbes()
	a.cancelObjectGesture()
	e := &a.objects
	e.crop.active = true
	chooseCombo(e.cropRatio, 0)
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
	chooseCombo(e.cropRatio, 0)
	a.updateObjectActions()
	a.layoutPanel()
	a.view.invalidateCanvas()
	return true
}

func (a *app) applyImageCrop() {
	e := &a.objects
	if bounds, ok := a.imageCropBounds(); ok {
		edit := objectEditRequest{Operation: "replace_crop", Token: e.selected.Token, Crop: &bounds}
		if e.crop.shape.round() {
			edit.CropShape = "ellipse"
		}
		a.runObjectEdit(e.source, edit)
	}
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
	image, box, angle, visible := a.imageCropLimits()
	if !visible {
		a.cancelImageCrop()
		return true
	}
	at := normalizedTextPoint(area, x, y)
	switch message {
	case 0x0201, 0x0203:
		if at[0] >= image[0] && at[0] <= image[2] && at[1] >= image[1] && at[1] <= image[3] {
			crop.drag = imageCropDrag{at, at}
			crop.dragging = true
			a.comparison.canvasFocused = false
			setFocus.Call(a.window)
			user32.NewProc("SetCapture").Call(a.window)
			a.updateObjectActions()
			v.invalidateCanvas()
		}
	case 0x0200, 0x0202:
		crop.drag.end = at
		if message == 0x0202 {
			drag := crop.drag
			start, end := screenTextPoint(area, drag.start), screenTextPoint(area, a.cropEnd(drag, image, box, angle))
			minimum := int32(a.scale(4))
			if absTextDrag(end.x-start.x) >= minimum && absTextDrag(end.y-start.y) >= minimum {
				crop.selection = &drag
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
	crop := &a.objects.crop
	image, box, angle, visible := a.imageCropLimits()
	if !visible {
		return
	}
	start := screenTextPoint(area, textPoint{image[0], image[1]})
	end := screenTextPoint(area, textPoint{image[2], image[3]})
	a.drawTextFrameOutline(dc, rect{start.x, start.y, end.x, end.y})
	drag := crop.selection
	if crop.dragging {
		drag = &crop.drag
	}
	mask := a.objects.selected.ImageMask
	if drag == nil && mask == nil {
		return
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	defer gdiDelete.Call(pen)
	defer gdiSelect.Call(dc, old)
	if drag == nil {
		outline := make([]textPoint, len(mask.Outline))
		for i, point := range mask.Outline {
			outline[i] = pdfPointToText(point, box, angle)
		}
		drawTextOutline(dc, area, outline, true)
		return
	}
	finish := a.cropEnd(*drag, image, box, angle)
	if crop.shape.round() {
		first, last := screenTextPoint(area, drag.start), screenTextPoint(area, finish)
		brush, _, _ := gdi32.NewProc("GetStockObject").Call(5) // NULL_BRUSH
		oldBrush, _, _ := gdiSelect.Call(dc, brush)
		gdi32.NewProc("Ellipse").Call(dc, uintptr(min(first.x, last.x)), uintptr(min(first.y, last.y)),
			uintptr(max(first.x, last.x)), uintptr(max(first.y, last.y)))
		gdiSelect.Call(dc, oldBrush)
	} else {
		drawTextOutline(dc, area, rectangleSelection(drag.start, finish), true)
	}
}
