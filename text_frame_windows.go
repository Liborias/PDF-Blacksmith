package main

func textFrameDragRect(start, end point) rect {
	return rect{min(start.x, end.x), min(start.y, end.y), max(start.x, end.x), max(start.y, end.y)}
}

func (v *documentView) finishTextFrameDrag(pick textSelection, page rect, start, end point) {
	bounds := intersectInlineRect(page, textFrameDragRect(start, end))
	if bounds.right-bounds.left < int32(v.app.scale(3)) || bounds.bottom-bounds.top < int32(v.app.scale(3)) {
		return
	}
	top := normalizedTextPoint(page, bounds.left, bounds.top)
	bottom := normalizedTextPoint(page, bounds.right, bounds.bottom)
	pick.point, pick.polygon = top, rectangleSelection(top, bottom)
	v.dispatchCanvasSelection(pick)
}

func (a *app) openNewText(v *documentView, model *textPage, pick textSelection) {
	if a.busy || a.textTool.mode != pick.mode || pick.insertPoint && !a.canvasTextCursor {
		return
	}
	_, _, rotation := v.nativeTextPageGeometry(model, pick)
	pick, ok := a.activateSelection(v, pick, idTextTab, model.Boxes, rotation)
	if !ok {
		return
	}
	_, dim, rotation := a.view.nativeTextPageGeometry(model, pick)
	var bounds searchRect
	if pick.insertPoint {
		x, y := pick.point[0]*dim[0], pick.point[1]*dim[1]
		if x < 0 || y < 0 || x > dim[0]-1 || y > dim[1]-1 {
			return
		}
		bounds = searchRect{x, y, x, y}
	} else {
		region := polygonBounds(pick.polygon)
		bounds = searchRect{max(0, region[0]) * dim[0], max(0, region[1]) * dim[1],
			min(1, region[2]) * dim[0], min(1, region[3]) * dim[1]}
		bounds[2] = bounds[0] + float64(int((bounds[2]-bounds[0])*100))/100
		if bounds[2]-bounds[0] < 1 || bounds[3]-bounds[1] < 1 {
			return
		}
	}
	a.clearTextSelection()
	p := &a.textTool
	p.model, p.pick, p.dim, p.rotation = model, pick, dim, rotation
	p.inline.frames = []inlineTextFrame{{newBounds: &bounds, editable: true}}
	a.focusInlineFrame(0, pick.point, false)
	if pick.insertPoint {
		a.reportUI(uiNewPointTextHint)
	} else {
		a.reportUI(uiNewTextFrameHint)
	}
}

func (a *app) drawTextFrameOutline(dc uintptr, bounds rect) {
	if bounds.right <= bounds.left || bounds.bottom <= bounds.top {
		return
	}
	pen, _, _ := gdiPen.Call(1, 1, selectionColor.win())   // PS_DASH
	brush, _, _ := gdi32.NewProc("GetStockObject").Call(5) // NULL_BRUSH
	oldPen, _, _ := gdiSelect.Call(dc, pen)
	oldBrush, _, _ := gdiSelect.Call(dc, brush)
	gdi32.NewProc("Rectangle").Call(dc, uintptr(bounds.left), uintptr(bounds.top), uintptr(bounds.right), uintptr(bounds.bottom))
	gdiSelect.Call(dc, oldBrush)
	gdiSelect.Call(dc, oldPen)
	gdiDelete.Call(pen)
}

func (a *app) configureNewTextFrame(frame *inlineTextFrame) {
	p, e := &a.textTool, &a.textTool.inline
	e.sourceBounds = *frame.newBounds
	e.leftLimit, e.rightLimit = 0, p.dim[0]
	e.firstBaseline = e.sourceBounds[1] + 10.8
	e.fixedWidth = !p.pick.insertPoint
	e.foreground = 0
	e.background = 0xffffff
	sendMessage.Call(e.handle, 0x0443, 0, e.background) // EM_SETBKGNDCOLOR
	a.updateInlineFontChoices()
	setText(p.original, "")
	setText(p.size, "12,00")
	setText(p.leading, "1,20")
	setText(p.tracking, "0,00")
	setText(p.width, formatTextNumber(e.sourceBounds[2]-e.sourceBounds[0]))
	p.alignment = 0
	for _, control := range []uintptr{p.bold, p.italic, p.underline} {
		a.setChecked(control, false)
	}
}
