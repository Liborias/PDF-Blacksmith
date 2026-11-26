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
	previewRequested bool
	preview          *printBitmap
	ghost, tint      viewportSurface
	shape            imageCropShape
	drag             imageCropDrag
	selection        *imageCropDrag
	resize           *cropResize
	motion           cropMotion
}

func (crop *imageCrop) endDrag() {
	crop.dragging, crop.resize, crop.motion.gesture = false, nil, nil
}

type cropResize struct {
	handle  cropHandle
	box     searchRect
	start   textPoint
	minimum textPoint
}

func (a *app) cropPreviewBox(image, box searchRect, angle int) (searchRect, bool) {
	crop := &a.objects.crop
	var drag *imageCropDrag
	switch {
	case crop.dragging && crop.motion.gesture != nil:
		motion := crop.motion.gesture
		if motion.mode == moveCropImage {
			return motion.box, true
		}
		return moveCropBounds(motion.box, a.cropMotionDelta(box, angle)), true
	case crop.dragging && crop.resize != nil:
		resize := crop.resize
		delta := textPoint{crop.drag.end[0] - resize.start[0], crop.drag.end[1] - resize.start[1]}
		width, height := box[2]-box[0], box[3]-box[1]
		if angle%180 != 0 {
			width, height = height, width
		}
		bounds := resizeImageCrop(resize.box, image, resize.handle, delta, resize.minimum, a.cropRatioValue()*height/width)
		return bounds, true
	case crop.dragging:
		drag = &crop.drag
	case crop.selection != nil:
		drag = crop.selection
	default:
		if outline := a.cropMotionOutline(box, angle, crop.motion.shifts.mask); len(outline) >= 3 {
			return polygonBounds(outline), true
		}
		return searchRect{}, false
	}
	return rectangleBetween(drag.start, a.cropEnd(*drag, image, box, angle)), true
}

func (a *app) cropRatioValue() float64 {
	preset := imageCropRatios[max(0, comboIndex(a.objects.cropRatio))]
	return float64(preset.width) / float64(preset.height)
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
	e.cropMoveMode = a.uiCombo(idObjectCropMoveMode, 16, 536, 298, []uiMessage{uiCropMovePair, uiCropMoveMask, uiCropMoveImage})
	a.bindUIHelp(e.cropMoveMode, uiCropMoveHelp)
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
	a.invalidateCropMotion()
}

func (a *app) cropEnd(drag imageCropDrag, image, box searchRect, angle int) textPoint {
	return constrainImageCrop(drag.start, drag.end, image, box, angle, a.cropRatioValue())
}

func (a *app) imageCropLimits() (image, box searchRect, angle int, visible bool) {
	e, v := &a.objects, a.view
	box, angle = e.page.Boxes[v.box()], e.rotation(v)
	source := e.selected.cropSourceBounds()
	shift := cropDifference(textPointToPDF(e.crop.motion.shifts.image, box, angle), textPointToPDF(textPoint{}, box, angle))
	image, visible = searchDisplayRect(moveCropBounds(source, shift), box, angle)
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
	e.cropRevision++
	chooseCombo(e.cropMoveMode, int(moveCropPair))
	if mask := e.selected.ImageMask; mask != nil && mask.Shape == "ellipse" {
		e.crop.shape = cropOval
	}
	chooseCombo(e.cropRatio, 0)
	a.updateObjectActions()
	a.layoutPanel()
	a.view.invalidateCanvas()
	a.view.schedule()
}

func (a *app) cancelImageCrop() bool {
	e := &a.objects
	if !e.crop.active {
		return false
	}
	dragging := e.crop.dragging
	e.crop.ghost.close()
	e.crop.tint.close()
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
	if edit, ok := a.cropEditRequest(); ok {
		a.runObjectEdit(a.objects.source, edit)
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
			crop.endDrag()
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
		crop.resize, crop.motion.gesture = nil, nil
		if current, ok := a.cropPreviewBox(image, box, angle); ok {
			radius, minimum := int32(a.scale(7)), float64(a.scale(4))
			for _, handle := range cropHandles {
				p := screenTextPoint(area, handle.point(current))
				if absTextDrag(p.x-x) <= radius && absTextDrag(p.y-y) <= radius {
					crop.resize = &cropResize{handle, current, at,
						textPoint{minimum / float64(area.right-area.left), minimum / float64(area.bottom-area.top)}}
					break
				}
			}
			if crop.resize == nil && a.cropMoveHit(at, current, box, angle) {
				if message == 0x0203 {
					chooseCombo(e.cropMoveMode, int(moveCropImage))
					a.updateObjectActions()
					v.invalidateCanvas()
					return true
				}
				crop.motion.gesture = &cropMotionGesture{mode: a.cropMoveMode(), box: current, start: at}
			}
		}
		if crop.resize != nil || crop.motion.gesture != nil || boundsContain(image, rectangleBetween(at, at)) {
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
			if crop.motion.gesture != nil {
				a.finishCropMotion(box, angle)
			} else if crop.resize != nil {
				if drag.start != drag.end {
					bounds, _ := a.cropPreviewBox(image, box, angle)
					crop.selection = &imageCropDrag{textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}}
				}
			} else {
				start, end := screenTextPoint(area, drag.start), screenTextPoint(area, a.cropEnd(drag, image, box, angle))
				minimum := int32(a.scale(4))
				if absTextDrag(end.x-start.x) >= minimum && absTextDrag(end.y-start.y) >= minimum {
					crop.selection = &drag
				}
			}
			crop.endDrag()
			user32.NewProc("ReleaseCapture").Call()
			a.updateObjectActions()
			if crop.motion.shifts.changed() {
				a.invalidateCropMotion()
			} else {
				crop.motion.preview = nil
			}
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
	a.paintCropMotionPreview(dc, area)
	if gesture := crop.motion.gesture; gesture != nil && gesture.mode != moveCropMask {
		image = moveCropBounds(image, a.cropMotionDelta(box, angle))
	}
	start := screenTextPoint(area, textPoint{image[0], image[1]})
	end := screenTextPoint(area, textPoint{image[2], image[3]})
	a.drawTextFrameOutline(dc, rect{start.x, start.y, end.x, end.y})
	current, drawn := a.cropPreviewBox(image, box, angle)
	if crop.motion.gesture == nil && (!crop.motion.shifts.changed() ||
		!crop.dragging && crop.motion.rendered == a.objects.cropRevision) {
		a.paintCropOverlay(dc, area, box, angle, current, drawn)
	}
	if !drawn {
		return
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	defer gdiDelete.Call(pen)
	defer gdiSelect.Call(dc, old)
	first := textPoint{current[0], current[1]}
	last := textPoint{current[2], current[3]}
	if crop.selection == nil && a.objects.selected.ImageMask != nil && (!crop.dragging || crop.motion.gesture != nil) {
		shift := crop.motion.shifts.mask
		if gesture := crop.motion.gesture; gesture != nil && gesture.mode != moveCropImage {
			shift = sumCropPoints(shift, a.cropMotionDelta(box, angle))
		}
		drawTextOutline(dc, area, a.cropMotionOutline(box, angle, shift), true)
	} else if crop.shape.round() {
		s, t := screenTextPoint(area, first), screenTextPoint(area, last)
		brush, _, _ := gdi32.NewProc("GetStockObject").Call(5) // NULL_BRUSH
		oldBrush, _, _ := gdiSelect.Call(dc, brush)
		gdi32.NewProc("Ellipse").Call(dc, uintptr(s.x), uintptr(s.y), uintptr(t.x), uintptr(t.y))
		gdiSelect.Call(dc, oldBrush)
	} else {
		drawTextOutline(dc, area, rectangleSelection(first, last), true)
	}
	radius := int32(max(2, a.scale(3)))
	for _, handle := range cropHandles {
		p := screenTextPoint(area, handle.point(current))
		fillColor(dc, rect{p.x - radius, p.y - radius, p.x + radius + 1, p.y + radius + 1}, selectionColor)
	}
}
