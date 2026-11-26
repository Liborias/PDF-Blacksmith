package main

import "unsafe"

func cropPreviewSurface(dc uintptr, surface *viewportSurface, bitmap *printBitmap, tint bool) bool {
	if surface.dc != 0 {
		return true
	}
	width, height := int32(bitmap.width), int32(bitmap.height)
	if width <= 0 || height <= 0 || len(bitmap.pixels) != int(width)*int(height)*4 {
		return false
	}
	source, _, _ := createSurfaceDC.Call(dc)
	if source == 0 {
		return false
	}
	var pixels *byte
	header := bitmapHeader{size: 40, width: width, height: -height, planes: 1, bits: 32}
	dib, _, _ := gdi32.NewProc("CreateDIBSection").Call(source, uintptr(unsafe.Pointer(&header)), 0,
		uintptr(unsafe.Pointer(&pixels)), 0, 0)
	if dib == 0 {
		deleteSurfaceDC.Call(source)
		return false
	}
	old, _, _ := gdiSelect.Call(source, dib)
	if old == 0 || old == ^uintptr(0) {
		gdiDelete.Call(dib)
		deleteSurfaceDC.Call(source)
		return false
	}
	dst := unsafe.Slice(pixels, len(bitmap.pixels))
	copy(dst, bitmap.pixels)
	if tint {
		for i := 0; i < len(dst); i += 4 {
			dst[i], dst[i+1], dst[i+2] = dst[i+3], dst[i+3], dst[i+3]
		}
	}
	*surface = viewportSurface{dc: source, bitmap: dib, previous: old, width: width, height: height, complete: true}
	return true
}

func cropScreenRegion(dc uintptr, area rect, outline []textPoint) uintptr {
	if len(outline) < 3 {
		return 0
	}
	points := make([]point, len(outline))
	for i, pos := range outline {
		points[i] = screenTextPoint(area, pos)
	}
	gdi32.NewProc("LPtoDP").Call(dc, uintptr(unsafe.Pointer(&points[0])), uintptr(len(points)))
	region, _, _ := gdi32.NewProc("CreatePolygonRgn").Call(uintptr(unsafe.Pointer(&points[0])), uintptr(len(points)), 1)
	return region
}

func (a *app) paintCropOverlay(dc uintptr, area rect, box searchRect, angle int, draft searchRect, drawn bool) {
	e, crop := &a.objects, &a.objects.crop
	if crop.preview == nil || !cropPreviewSurface(dc, &crop.ghost, crop.preview, false) {
		return
	}
	source := e.selected.cropSourceBounds()
	bounds := rectangleBetween(pdfPointToText(textPoint{source[0], source[1]}, box, angle),
		pdfPointToText(textPoint{source[2], source[3]}, box, angle))
	moved := crop.motion.shifts.changed()
	if moved {
		bounds = crop.motion.shifts.imageBounds(bounds)
	}
	first := screenTextPoint(area, textPoint{bounds[0], bounds[1]})
	last := screenTextPoint(area, textPoint{bounds[2], bounds[3]})
	if first.x >= last.x || first.y >= last.y {
		return
	}
	var previous, candidate uintptr
	if mask := e.selected.ImageMask; mask != nil {
		outline := displayContour(mask.Outline, box, angle)
		if moved {
			outline = a.cropMotionOutline(box, angle, crop.motion.shifts.mask)
		}
		previous = cropScreenRegion(dc, area, outline)
		if previous == 0 {
			return
		}
		defer gdiDelete.Call(previous)
	}
	hasDraft := drawn && (crop.dragging || crop.selection != nil)
	if hasDraft {
		s, t := screenTextPoint(area, textPoint{draft[0], draft[1]}), screenTextPoint(area, textPoint{draft[2], draft[3]})
		points := [2]point{s, t}
		gdi32.NewProc("LPtoDP").Call(dc, uintptr(unsafe.Pointer(&points[0])), 2)
		s, t = points[0], points[1]
		create := "CreateRectRgn"
		if crop.shape.round() {
			create = "CreateEllipticRgn"
		}
		candidate, _, _ = gdi32.NewProc(create).Call(uintptr(s.x), uintptr(s.y), uintptr(t.x), uintptr(t.y))
		if candidate == 0 {
			return
		}
		defer gdiDelete.Call(candidate)
	}
	blend := func(source viewportSurface, opacity uintptr, keep, remove1, remove2 uintptr) {
		saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
		if saved == 0 {
			return
		}
		defer gdi32.NewProc("RestoreDC").Call(dc, saved)
		selectClip := gdi32.NewProc("ExtSelectClipRgn")
		if keep != 0 {
			selectClip.Call(dc, keep, 1) // RGN_AND
		}
		for _, outside := range []uintptr{remove1, remove2} {
			if outside != 0 {
				selectClip.Call(dc, outside, 4) // RGN_DIFF
			}
		}
		alphaBlend.Call(dc, uintptr(first.x), uintptr(first.y), uintptr(last.x-first.x), uintptr(last.y-first.y),
			source.dc, 0, 0, uintptr(source.width), uintptr(source.height), opacity<<16|1<<24)
	}
	if moved {
		cropRegion := previous
		if candidate != 0 {
			cropRegion = candidate
		}
		blend(crop.ghost, 77, 0, cropRegion, 0)
		return
	}
	if previous != 0 {
		blend(crop.ghost, 77, 0, previous, candidate)
		if hasDraft {
			blend(crop.ghost, 255, candidate, previous, 0)
		}
	}
	if hasDraft && cropPreviewSurface(dc, &crop.tint, crop.preview, true) {
		blend(crop.tint, 178, previous, candidate, 0)
	}
}
