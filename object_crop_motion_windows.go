package main

type cropMotionGesture struct {
	mode  cropMotionMode
	box   searchRect
	start textPoint
}

type cropMotion struct {
	shifts              cropShifts
	gesture             *cropMotionGesture
	preview             *printBitmap
	coverage            searchRect
	requested, rendered uint64
}

func (a *app) cropMoveMode() cropMotionMode {
	return cropMotionMode(max(0, comboIndex(a.objects.cropMoveMode)))
}

func (a *app) cropMotionOutline(box searchRect, angle int, delta textPoint) []textPoint {
	mask := a.objects.selected.ImageMask
	if mask == nil {
		return nil
	}
	outline := displayContour(mask.ClipOutline, box, angle)
	for i := range outline {
		outline[i] = sumCropPoints(outline[i], delta)
	}
	return outline
}

func (a *app) cropMotionDelta(box searchRect, angle int) textPoint {
	e, movement := &a.objects, &a.objects.crop.motion
	gesture, mask := movement.gesture, e.selected.ImageMask
	if gesture == nil || mask == nil {
		return textPoint{}
	}
	return constrainCropMotion(displayContour(mask.SourceOutline, box, angle),
		a.cropMotionOutline(box, angle, textPoint{}), movement.shifts, gesture.mode, gesture.box,
		cropDifference(e.crop.drag.end, gesture.start))
}

func (a *app) cropMoveHit(at textPoint, current searchRect, box searchRect, angle int) bool {
	e := &a.objects
	if e.selected.ImageMask == nil || !boundsContain(current, rectangleBetween(at, at)) {
		return false
	}
	if e.crop.selection == nil {
		return pointInTextRegion(at, a.cropMotionOutline(box, angle, e.crop.motion.shifts.mask))
	}
	if !e.crop.shape.round() {
		return true
	}
	x := (2*at[0] - current[0] - current[2]) / (current[2] - current[0])
	y := (2*at[1] - current[1] - current[3]) / (current[3] - current[1])
	return x*x+y*y <= 1
}

func (a *app) finishCropMotion(box searchRect, angle int) {
	crop := &a.objects.crop
	delta := a.cropMotionDelta(box, angle)
	if crop.motion.gesture.mode != moveCropMask {
		crop.motion.shifts.image = sumCropPoints(crop.motion.shifts.image, delta)
	}
	if crop.motion.gesture.mode != moveCropImage {
		crop.motion.shifts.mask = sumCropPoints(crop.motion.shifts.mask, delta)
		if crop.selection != nil {
			crop.selection.start = sumCropPoints(crop.selection.start, delta)
			crop.selection.end = sumCropPoints(crop.selection.end, delta)
		}
	}
}

func (a *app) cropEditRequest() (objectEditRequest, bool) {
	e := &a.objects
	if !a.canCropImage() || !e.crop.active || e.crop.dragging {
		return objectEditRequest{}, false
	}
	edit := objectEditRequest{Operation: "replace_crop", Token: e.selected.Token}
	if bounds, ok := a.imageCropBounds(); ok {
		edit.Crop = &bounds
		if e.crop.shape.round() {
			edit.CropShape = "ellipse"
		}
	}
	if shifts := e.crop.motion.shifts; shifts.changed() {
		_, box, angle, _ := a.imageCropLimits()
		origin := textPointToPDF(textPoint{}, box, angle)
		mask := cropDifference(textPointToPDF(shifts.mask, box, angle), origin)
		image := cropDifference(textPointToPDF(shifts.image, box, angle), origin)
		edit.Operation, edit.MaskDelta, edit.ImageDelta = "translate_crop", &mask, &image
	}
	return edit, edit.Crop != nil || edit.MaskDelta != nil
}

func (a *app) invalidateCropMotion() {
	a.objects.cropRevision++
	a.view.invalidateCanvas()
	a.view.schedule()
}

func (a *app) paintCropMotionPreview(dc uintptr, area rect) {
	motion := a.objects.crop.motion
	if motion.preview == nil || !motion.shifts.changed() {
		return
	}
	p := screenTextPoint(area, textPoint{motion.coverage[0], motion.coverage[1]})
	q := screenTextPoint(area, textPoint{motion.coverage[2], motion.coverage[3]})
	saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
	if saved == 0 {
		return
	}
	defer gdi32.NewProc("RestoreDC").Call(dc, saved)
	gdi32.NewProc("IntersectClipRect").Call(dc, uintptr(p.x), uintptr(p.y), uintptr(q.x+1), uintptr(q.y+1))
	drawBitmap(dc, motion.preview, int(area.left), int(area.top), int(area.right-area.left), int(area.bottom-area.top))
}

func (a *app) cropMotionCoverage() searchRect {
	image, box, angle, _ := a.imageCropLimits()
	source, visible := searchDisplayRect(a.objects.selected.cropSourceBounds(), box, angle)
	if !visible {
		return image
	}
	return searchRect{min(source[0], image[0]), min(source[1], image[1]),
		max(source[2], image[2]), max(source[3], image[3])}
}
