package main

import "unsafe"

var (
	createSurfaceDC     = gdi32.NewProc("CreateCompatibleDC")
	createSurfaceBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	deleteSurfaceDC     = gdi32.NewProc("DeleteDC")
	copySurface         = gdi32.NewProc("BitBlt")
	surfaceOrigin       = gdi32.NewProc("SetViewportOrgEx")
)

// The surface holds one complete frame. Rendering a new page or scale
// doesn't alter it until all the visible raster data is ready.
type viewportSurface struct {
	dc, bitmap, previous uintptr
	width, height        int32
	complete             bool
}

func (s *viewportSurface) bytes() int64 {
	return int64(s.width) * int64(s.height) * 4
}

func (s *viewportSurface) close() {
	if s.dc != 0 {
		gdiSelect.Call(s.dc, s.previous)
		gdiDelete.Call(s.bitmap)
		deleteSurfaceDC.Call(s.dc)
	}
	*s = viewportSurface{}
}

func (s *viewportSurface) ensure(reference uintptr, width, height int32) bool {
	if s.dc != 0 && s.width == width && s.height == height {
		return true
	}
	dc, _, _ := createSurfaceDC.Call(reference)
	if dc == 0 {
		return false
	}
	bitmap, _, _ := createSurfaceBitmap.Call(reference, uintptr(width), uintptr(height))
	if bitmap == 0 {
		deleteSurfaceDC.Call(dc)
		return false
	}
	previous, _, _ := gdiSelect.Call(dc, bitmap)
	if previous == 0 || previous == ^uintptr(0) {
		gdiDelete.Call(bitmap)
		deleteSurfaceDC.Call(dc)
		return false
	}
	complete := false
	if s.complete {
		fillColor(dc, rect{0, 0, width, height}, editorColor)
		copied, _, _ := copySurface.Call(dc, 0, 0, uintptr(min(width, s.width)), uintptr(min(height, s.height)),
			s.dc, 0, 0, 0x00cc0020)
		complete = copied != 0
	}
	s.close()
	*s = viewportSurface{dc, bitmap, previous, width, height, complete}
	return true
}

func (v *documentView) invalidateCanvas() {
	v.canvasDirty, v.canvasDirtyArea = true, rect{}
	bounds := v.viewport
	winInvalidate.Call(v.app.window, uintptr(unsafe.Pointer(&bounds)), 0)
}

// Keep the cached canvas and repaint just the destination of the new raster.
func (v *documentView) invalidateRasterArea(bounds rect) {
	bounds = intersectInlineRect(bounds, v.viewport)
	if bounds.right <= bounds.left || bounds.bottom <= bounds.top {
		return
	}
	if !v.canvasDirty {
		v.canvasDirtyArea = bounds
	} else if previous := v.canvasDirtyArea; previous != (rect{}) {
		v.canvasDirtyArea = rect{min(previous.left, bounds.left), min(previous.top, bounds.top),
			max(previous.right, bounds.right), max(previous.bottom, bounds.bottom)}
	}
	v.canvasDirty = true
	v.invalidateOverlay(bounds)
}

func (v *documentView) invalidateOverlay(bounds rect) {
	winInvalidate.Call(v.app.window, uintptr(unsafe.Pointer(&bounds)), 0)
}

func (v *documentView) canvasReady() bool {
	if v.describing || v.invalidSource {
		return false
	}
	if v.overview {
		return true
	}
	cache := v.activeCache()
	_, height := v.viewportSize()
	visible := v.spread.visible(v.scrollY, height)
	for _, page := range visible {
		number := v.order[page.index]
		if cache.frames[number].image != nil || cache.failed[number] ||
			(!v.proofMode() && (v.hasDetailFrame(number, v.pageBounds(page)) || v.detailFailed(number))) {
			return true
		}
	}
	return len(visible) == 0
}

func (v *documentView) paint(dc uintptr) {
	bounds := v.viewport
	width, height := bounds.right-bounds.left, bounds.bottom-bounds.top
	if width <= 0 || height <= 0 {
		return
	}
	ready := v.canvasReady()
	surface := &v.canvas
	if !surface.ensure(dc, width, height) {
		// The frame can still be drawn if GDI is low on memory. Never draw
		// a partially rendered page in this fallback.
		if ready {
			v.paintCanvas(dc)
		} else if surface.complete {
			surface.blit(dc, bounds)
		} else {
			v.paintLoading(dc)
		}
		v.paintCanvasContentMove(dc)
		return
	}
	if (ready && v.canvasDirty) || !surface.complete {
		saved, _, _ := gdi32.NewProc("SaveDC").Call(surface.dc)
		surfaceOrigin.Call(surface.dc, uintptr(-bounds.left), uintptr(-bounds.top), 0)
		if ready {
			if dirty := v.canvasDirtyArea; surface.complete && dirty != (rect{}) {
				gdi32.NewProc("IntersectClipRect").Call(surface.dc, uintptr(dirty.left), uintptr(dirty.top),
					uintptr(dirty.right), uintptr(dirty.bottom))
			}
			v.paintCanvas(surface.dc)
			surface.complete, v.canvasDirty, v.canvasDirtyArea = true, false, rect{}
		} else {
			v.paintLoading(surface.dc)
		}
		gdi32.NewProc("RestoreDC").Call(surface.dc, saved)
	}
	surface.blit(dc, bounds)
	v.paintCanvasContentMove(dc)
}

func (s *viewportSurface) blit(dc uintptr, bounds rect) {
	copySurface.Call(dc, uintptr(bounds.left), uintptr(bounds.top),
		uintptr(min(s.width, bounds.right-bounds.left)), uintptr(min(s.height, bounds.bottom-bounds.top)),
		s.dc, 0, 0, 0x00cc0020)
}

func (v *documentView) paintLoading(dc uintptr) {
	fillColor(dc, v.viewport, editorColor)
	message := "Načítám stránku…"
	if v.describing {
		message = "Načítám dokument…"
	} else if v.invalidSource {
		message = "Vstupní PDF se změnilo. Otevřete je znovu."
	}
	if v.readerError != "" {
		message = v.readerError
	}
	drawText(dc, message, v.viewport, v.app.titleFont, mutedColor, 1)
}
