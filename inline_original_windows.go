package main

import "math"

const wmInlineCommit = 0x801a

type inlineGlyphPosition struct {
	start, end, line int
	bounds           searchRect
}

func (a *app) originalInlinePosition(x, y int32) int {
	p, e := &a.textTool, &a.textTool.inline
	w, h := float64(e.pageArea.right-e.pageArea.left), float64(e.pageArea.bottom-e.pageArea.top)
	if w <= 0 || h <= 0 {
		return 0
	}
	docX := float64(x-e.pageArea.left) * p.dim[0] / w
	docY := float64(y-e.pageArea.top) * p.dim[1] / h
	result, best := 0, math.Inf(1)
	for _, glyph := range e.glyphPositions {
		box := glyph.bounds
		dy := max(0, box[1]-docY, docY-box[3])
		dx := max(0, box[0]-docX, docX-box[2])
		distance := dy*4 + dx
		if distance < best {
			best, result = distance, glyph.start
			if docX > (box[0]+box[2])/2 {
				result = glyph.end
			}
		}
	}
	return result
}

func (a *app) paintOriginalInline(dc uintptr) {
	p, e := &a.textTool, &a.textTool.inline
	focus, _, _ := user32.NewProc("GetFocus").Call()
	if focus != e.handle {
		return
	}
	if !e.caretHidden {
		hide := user32.NewProc("HideCaret")
		hidden, _, _ := hide.Call(e.handle)
		e.caretHidden = hidden != 0
	}
	start, end := inlineSelection(e.handle)
	toLocal := func(x, y float64) point {
		pos := screenTextPoint(e.pageArea, textPoint{x / p.dim[0], y / p.dim[1]})
		return point{pos.x - e.rect.left, pos.y - e.rect.top}
	}
	var caret rect
	var row inlineGlyphPosition
	var mark rect
	lastLine := -1
	color := paletteColor{87, 117, 149}
	drawRow := func() {
		if mark.right > mark.left {
			y := toLocal(0, row.bounds[3]).y + int32(max(1, a.scale(1)))
			fillColor(dc, rect{mark.left, y, mark.right, y + 1}, color)
		}
	}
	for _, glyph := range e.glyphPositions {
		if glyph.line != lastLine {
			drawRow()
			lastLine, mark, row = glyph.line, rect{}, glyph
		}
		box := glyph.bounds
		row.bounds[3] = max(row.bounds[3], box[3])
		top, bottom := toLocal(box[0], box[1]), toLocal(box[2], box[3])
		if start == end && start >= glyph.start {
			x := bottom.x
			if start == glyph.start {
				x = top.x
			}
			caret = rect{x, top.y, x + int32(max(1, a.scale(1))), bottom.y}
		}
		if start != end && start < glyph.end && end > glyph.start {
			if mark.right <= mark.left {
				mark.left = top.x
			}
			mark.right = max(mark.right, bottom.x+1)
		}
	}
	drawRow()
	if start == end && caret.bottom > caret.top {
		fillColor(dc, caret, color)
	}
}

func (a *app) originalInlineMouse(message uint32, x, y int32) bool {
	e := &a.textTool.inline
	if message == 0x0215 {
		e.mouseSelecting = false
	}
	if !e.pristine {
		return false
	}
	switch message {
	case 0x0201:
		e.mouseSelecting = true
		e.mouseAnchor = a.originalInlinePosition(x, y)
		setFocus.Call(e.handle)
		selectInline(e.handle, e.mouseAnchor, e.mouseAnchor)
		user32.NewProc("SetCapture").Call(e.handle)
		invalidate(e.handle)
		return true
	case 0x0200, 0x0202:
		if e.mouseSelecting {
			offset := a.originalInlinePosition(x, y)
			selectInline(e.handle, min(e.mouseAnchor, offset), max(e.mouseAnchor, offset))
			if message == 0x0202 {
				e.mouseSelecting = false
				user32.NewProc("ReleaseCapture").Call()
			}
			invalidate(e.handle)
			return true
		}
	}
	return false
}

func (a *app) textSettingsFocus(handle uintptr) bool {
	if handle == a.textTool.inline.handle {
		return true
	}
	for _, control := range a.panelItems[idTextTab] {
		if control.handle == handle {
			return true
		}
	}
	return false
}

func (a *app) commitInlineOnBlur(serial uintptr) {
	e := &a.textTool.inline
	focus, _, _ := user32.NewProc("GetFocus").Call()
	if e.serial == uint64(serial) && a.mode == idTextTab && !a.busy && !a.textSettingsFocus(focus) &&
		!a.focusedDocumentView().selection.active() && a.inlineDirty() {
		a.startTextEdit()
	}
}
