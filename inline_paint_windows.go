package main

import (
	"math"
	"syscall"
	"unsafe"
)

var (
	transparentBlt      = syscall.NewLazyDLL("msimg32.dll").NewProc("TransparentBlt")
	inlinePaintSubclass = syscall.NewCallback(inlineEditProc)
)

func inlineEditProc(hwnd uintptr, message uint32, wParam, lParam, id, data uintptr) uintptr {
	a := window
	if a != nil {
		e := &a.textTool.inline
		if message == 0x0008 { // WM_KILLFOCUS
			e.caretHidden = false
			if e.visible && !a.textSettingsFocus(wParam) && wParam != 0 {
				postMessage.Call(a.window, wmInlineCommit, uintptr(e.serial), 0)
			}
		}
		if e.visible {
			view := a.view
			x, y := int32(int16(lParam))+e.rect.left, int32(int16(lParam>>16))+e.rect.top
			_, selectingArea := view.canvasRegionMode()
			if message == 0x0207 || selectingArea && (message == 0x0201 || message == 0x0203 ||
				message == 0x0200 && view.selection.active()) {
				if view.canvasPointer(message, x, y) {
					return 0
				}
			} else if message == 0x0203 || message == 0x0201 && canvasSelectionChange() != replaceSelection {
				view.selectionClick(x, y, message == 0x0203)
				return 0
			}
		}
		if e.visible && a.originalInlineMouse(message, int32(int16(lParam))+e.rect.left,
			int32(int16(lParam>>16))+e.rect.top) {
			return 0
		}
		surface := &a.view.canvas
		if message == wmPaint && e.visible && surface.complete {
			clip := e.clip
			width, height := clip.right-clip.left, clip.bottom-clip.top
			if e.capture.ensure(surface.dc, width, height) {
				var info paintInfo
				dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
				if dc != 0 {
					x, y := clip.left-e.rect.left, clip.top-e.rect.top
					copySurface.Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height), surface.dc,
						uintptr(clip.left-a.view.viewport.left), uintptr(clip.top-a.view.viewport.top), 0x00cc0020)
					if e.pristine {
						a.paintOriginalInline(dc)
						user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
						return 0
					}
					if e.caretHidden {
						user32.NewProc("ShowCaret").Call(hwnd)
						e.caretHidden = false
					}
					key := e.background
					fillColor(e.capture.dc, rect{right: width, bottom: height},
						paletteColor{uint8(key), uint8(key >> 8), uint8(key >> 16)})
					saved, _, _ := gdi32.NewProc("SaveDC").Call(e.capture.dc)
					surfaceOrigin.Call(e.capture.dc, uintptr(-x), uintptr(-y), 0)
					sendMessage.Call(hwnd, 0x0318, e.capture.dc, 4|8) // WM_PRINTCLIENT, PRF_CLIENT | PRF_ERASEBKGND
					gdi32.NewProc("RestoreDC").Call(e.capture.dc, saved)
					transparentBlt.Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height),
						e.capture.dc, 0, 0, uintptr(width), uintptr(height), key)
					user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
					return 0
				}
			}
		}
	}
	result, _, _ := defSubclass.Call(hwnd, uintptr(message), wParam, lParam)
	return result
}

func (a *app) scheduleInlineBackground() bool {
	p, e, v := &a.textTool, &a.textTool.inline, a.view
	frame := a.inlineCurrent()
	if frame == nil || !e.visible || e.backdropRequested || p.pick.generation != v.generation || a.busy {
		return false
	}
	bounds := frame.flow.bounds
	w, h := bounds[2]-bounds[0]+4, bounds[3]-bounds[1]+4
	zoom := min(4., 1800/max(w, h))
	x := max(0, int(math.Floor((bounds[0]-2)*zoom)))
	y := max(0, int(math.Floor((bounds[1]-2)*zoom)))
	right := min(int(math.Round(p.dim[0]*zoom)), int(math.Ceil((bounds[2]+2)*zoom)))
	bottom := min(int(math.Round(p.dim[1]*zoom)), int(math.Ceil((bounds[3]+2)*zoom)))
	if right <= x || bottom <= y {
		return false
	}
	region := screenRegion{Zoom: zoom, X: x, Y: y, Width: right - x, Height: bottom - y}
	e.backdropRegion = region
	e.backdropRequested = true
	req := v.request("text_background")
	req.View.Page, req.View.Rotation = p.pick.source, v.rotations[p.pick.source-1]
	req.View.Region = &region
	req.TextSelection = a.inlineGlyphSelection(frame)
	v.launch(req, p.pick.source, 0, false, nil)
	return true
}

func (a *app) finishInlineBackground(job readerJob) {
	e := &a.textTool.inline
	if job.generation == a.view.generation && job.textFrame == e.serial && job.frame.image != nil {
		e.backdrop = job.frame.image
		a.view.invalidateCanvas()
	}
}

func (a *app) paintInlineBackdrop(dc uintptr, area rect, lines []searchRect, toPixel func(searchRect) rect) {
	e := &a.textTool.inline
	// Sample at the height of the letters so alternating table rows keep their color.
	sample := func(box rect) uintptr {
		y := (box.top + box.bottom) / 2
		x := max(area.left, box.left-2)
		paper, _, _ := gdi32.NewProc("GetPixel").Call(dc, uintptr(x), uintptr(y))
		if paper > 0xffffff {
			return e.background
		}
		return paper
	}
	if len(lines) > 0 {
		paper := sample(toPixel(lines[0]))
		if paper == e.foreground {
			paper ^= 0x00ffffff
		}
		if paper != e.background {
			e.background = paper
			sendMessage.Call(e.handle, 0x0443, 0, paper) // EM_SETBKGNDCOLOR, also the compositing key.
		}
	}
	if bitmap := e.backdrop; bitmap != nil {
		r := e.backdropRegion
		region := toPixel(searchRect{float64(r.X) / r.Zoom, float64(r.Y) / r.Zoom,
			float64(r.X+r.Width) / r.Zoom, float64(r.Y+r.Height) / r.Zoom})
		drawBitmap(dc, bitmap, int(region.left), int(region.top), int(region.right-region.left), int(region.bottom-region.top))
		return
	}
	// Keep typing responsive while PDFium prepares the original background.
	for _, line := range lines {
		box := toPixel(line)
		paper := sample(box)
		fillColor(dc, box, paletteColor{uint8(paper), uint8(paper >> 8), uint8(paper >> 16)})
	}
}
