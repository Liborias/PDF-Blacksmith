package main

import (
	"strconv"
	"unsafe"
)

const (
	thumbnailCacheSize = 8 << 20
	thumbnailDragTimer = 7
	wmThumbnailUpdate  = 0x8003
)

type thumbnailDrag struct {
	start, current              point
	page, boundary              int
	pressed, moving, overDelete bool
}

func (v *documentView) buildThumbnails() {
	a := v.app
	v.thumbnailHeading = a.label("Náhledy stránek", 16, 4, 160, true)
	a.panelLabels[v.thumbnailHeading] = true
	v.thumbnailHint = a.label("Ctrl + kolečko", 196, 4, 123, false)
	a.panelLabels[v.thumbnailHint] = true
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100 | 0x0800 // LBS_EXTENDEDSEL
	v.pageList = a.add("LISTBOX", "Náhledy stránek", idReaderPageList, 16, 31, 298, 1, wsTabstop|listStyle, 0, a.normalFont)
	rowHeight := a.scale(112)
	sendMessage.Call(v.pageList, 0x01a0, 0, uintptr(rowHeight))
	a.addScrollPane(v.pageList, rowHeight, true)
}

func (v *documentView) thumbnailWidth() int {
	return v.thumbnailPanelPixels * v.thumbnailScale / 100
}

func (v *documentView) layoutThumbnails(width, height int) {
	d := v.app.scale
	v.thumbnailPanelPixels = width
	margin := min(d(12), width/10)
	top := d(31)
	heading := "Náhledy stránek"
	if width < d(137) {
		heading = "Náhledy"
	}
	setText(v.thumbnailHeading, heading)
	place(v.thumbnailHeading, margin, d(5), width-2*margin, d(22))
	showHint := width >= d(130)
	if showHint {
		place(v.thumbnailHint, margin, d(27), width-2*margin, d(19))
		top = d(50)
	}
	showControl(v.thumbnailHint, showHint)
	frame := rect{int32(margin), int32(top), int32(width - margin), int32(height)}
	pane := v.app.scrollPanes[v.pageList]
	rowHeight := max(1, min(int(frame.bottom-frame.top), v.thumbnailWidth()*3/2+d(32)))
	changed := rowHeight != pane.lineHeight
	if changed {
		pane.lineHeight = rowHeight
		sendMessage.Call(v.pageList, 0x01a0, 0, uintptr(rowHeight)) // LB_SETITEMHEIGHT
	}
	pane.arrange(frame, true)
	if changed {
		pane.reflow()
		invalidate(v.pageList)
	}
}

func thumbnailWheelSteps(delta int) int {
	if delta > 0 {
		return max(1, delta/120)
	}
	if delta < 0 {
		return min(-1, delta/120)
	}
	return 0
}

func (v *documentView) zoomThumbnails(screen point, delta int) {
	if delta == 0 || !v.ready() || v.pageDrag.pressed {
		return
	}
	scale := max(10, min(v.thumbnailScale+5*thumbnailWheelSteps(delta), 60))
	if scale == v.thumbnailScale {
		return
	}
	user32.NewProc("ScreenToClient").Call(v.pageList, uintptr(unsafe.Pointer(&screen)))
	x, y := screen.x, max(0, int(screen.y))
	pane := v.app.scrollPanes[v.pageList]
	_, _, top := pane.lines()
	row := min(len(v.order)-1, top+y/pane.lineHeight)
	if !contains(v.pageListBounds(), x, screen.y) {
		row, y = top, 0
	}
	v.thumbnailScale = scale
	var bounds rect
	getClientRect.Call(v.thumbnailSidebar, uintptr(unsafe.Pointer(&bounds)))
	v.layoutThumbnails(int(bounds.right), int(bounds.bottom))
	pane.bar.setPosition(row - y/pane.lineHeight)
	v.schedule()
}

func (v *documentView) thumbnailRange() (int, int) {
	pane := v.app.scrollPanes[v.pageList]
	visible, _, _ := user32.NewProc("IsWindowVisible").Call(v.pageList)
	if visible == 0 || !pane.shown || !v.ready() {
		return 0, 0
	}
	_, rows, top := pane.lines()
	top = max(0, min(top, len(v.order)))
	return top, min(len(v.order), top+rows+1)
}

func (v *documentView) thumbnailBudget() int64 {
	return min(thumbnailCacheSize, v.cacheLimit/16)
}

func (v *documentView) trimThumbnails(budget int64) {
	first, last := v.thumbnailRange()
	for number := range v.thumbnails.frames {
		if v.thumbnails.bytes <= budget {
			return
		}
		keep := false
		for _, visible := range v.order[first:last] {
			if visible == number {
				keep = true
				break
			}
		}
		if !keep || budget == 0 {
			v.thumbnails.remove(number)
		}
	}
}

func (v *documentView) scheduleThumbnails() bool {
	first, last := v.thumbnailRange()
	if first == last {
		return false
	}
	v.updateCacheLimit()
	pixels := 640
	rowHeight := v.app.scrollPanes[v.pageList].lineHeight
	for _, size := range []int{128, 256, 400} {
		if rowHeight <= size {
			pixels = size
			break
		}
	}
	estimate := int64((pixels + 2) * (pixels + 2) * 4)
	budget := v.thumbnailBudget()
	if budget < estimate {
		return false
	}
	for _, number := range v.order[first:last] {
		if v.thumbnails.frames[number].pixels >= pixels || v.thumbnails.failed[number] {
			continue
		}
		var used int64
		if old := v.thumbnails.frames[number].image; old != nil {
			used = int64(len(old.pixels))
		}
		v.trimThumbnails(budget - estimate + used)
		if v.thumbnails.bytes+estimate-used > budget {
			break
		}
		req := v.request("thumbnail")
		req.View.Page, req.View.Rotation, req.View.Pixels, req.View.Box = number, v.rotations[number-1], pixels, "crop"
		v.launch(req, number, pixels, true, nil)
		return true
	}
	return false
}

func (v *documentView) drawPageThumbnail(item *buttonDraw, index int, color paletteColor) {
	d, bounds := v.app.scale, item.bounds
	number := v.order[index]
	image := v.thumbnails.frames[number].image
	labelHeight, gap := d(18), d(2)
	width, height := v.thumbnailWidth(), max(1, int(bounds.bottom-bounds.top)-labelHeight-gap-d(12))
	size := v.pageDimensions(number, "crop")
	if image != nil {
		size = pageSize{float64(image.width), float64(image.height)}
	}
	if size.width > 0 && size.height > 0 {
		fit := min(float64(width)/size.width, float64(height)/size.height)
		width, height = max(1, int(size.width*fit)), max(1, int(size.height*fit))
	}
	x := (int(bounds.left+bounds.right) - width) / 2
	y := (int(bounds.top+bounds.bottom) - height - gap - labelHeight) / 2
	if image != nil {
		gdi32.NewProc("SetStretchBltMode").Call(item.dc, 4)
		drawBitmap(item.dc, image, x, y, width, height)
	} else {
		fillColor(item.dc, rect{int32(x), int32(y), int32(x + width), int32(y + height)}, divider)
	}
	labelTop := int32(y + height + gap)
	labelWidth := max(width, d(45))
	labelLeft := (int(bounds.left+bounds.right) - labelWidth) / 2
	label := rect{int32(labelLeft), labelTop, int32(labelLeft + labelWidth), labelTop + int32(labelHeight)}
	drawText(item.dc, strconv.Itoa(index+1), label, v.app.normalFont, color, 1)
	if item.state&16 != 0 {
		v.app.drawFocus(item.dc, label)
	}
	if v.pageDrag.moving {
		line := bounds.top
		switch {
		case v.pageDrag.boundary == index:
		case v.pageDrag.boundary == index+1 && index+1 == len(v.order):
			line = bounds.bottom - int32(d(2))
		default:
			return
		}
		fillColor(item.dc, rect{bounds.left, line, bounds.right, line + int32(d(2))}, selectionColor)
	}
}

func (v *documentView) pageListBounds() rect {
	var bounds rect
	getClientRect.Call(v.pageList, uintptr(unsafe.Pointer(&bounds)))
	var sidebar rect
	var origin point
	user32.NewProc("GetWindowRect").Call(v.thumbnailSidebar, uintptr(unsafe.Pointer(&sidebar)))
	user32.NewProc("ClientToScreen").Call(v.pageList, uintptr(unsafe.Pointer(&origin)))
	bounds.top = max(bounds.top, sidebar.top-origin.y)
	bounds.bottom = min(bounds.bottom, sidebar.bottom-origin.y)
	return bounds
}

func (v *documentView) pageAt(y int32) int {
	pane := v.app.scrollPanes[v.pageList]
	_, _, first := pane.lines()
	return first + int(y)/pane.lineHeight
}

func (v *documentView) pageListMouse(msg uint32, keys, position uintptr) bool {
	if v.secondary {
		return false
	}
	xy := point{int32(int16(position)), int32(int16(position >> 16))}
	drag := &v.pageDrag
	switch msg {
	case 0x0201: // WM_LBUTTONDOWN
		if !v.ready() || v.app.mode != idPagesTab || keys&(4|8) != 0 { // Use native Shift and Ctrl selection.
			return false
		}
		if row := v.pageAt(xy.y); row >= 0 && row < len(v.order) {
			setFocus.Call(v.pageList)
			*drag = thumbnailDrag{start: xy, current: xy, page: row, boundary: -1, pressed: true}
			user32.NewProc("SetCapture").Call(v.pageList)
			return true
		}
	case 0x0200: // WM_MOUSEMOVE
		if drag.pressed {
			drag.current = xy
			deltaX, deltaY := xy.x-drag.start.x, xy.y-drag.start.y
			distance := max(max(deltaX, -deltaX), max(deltaY, -deltaY))
			if !drag.moving && v.app.mode == idPagesTab && !v.app.busy && distance >= int32(v.app.scale(5)) {
				if !v.inPageSelection(drag.page) {
					v.selectOnlyPage(drag.page)
				}
				drag.moving = true
				user32.NewProc("SetTimer").Call(v.pageList, thumbnailDragTimer, 160, 0)
			}
			if drag.moving {
				v.updatePageDrop()
			}
			return true
		}
	case 0x0202: // WM_LBUTTONUP
		if drag.pressed {
			drag.current = xy
			if drag.moving {
				v.updatePageDrop()
			}
			last := *drag
			v.endPageDrag()
			if !last.moving {
				v.selectOnlyPage(last.page)
			} else if !v.app.busy && last.overDelete {
				v.deleteSelectedPages()
			} else if !v.app.busy && last.boundary >= 0 {
				v.moveSelectedPages(dropPosition(v.selectedPagePositions(), last.boundary))
			}
			return true
		}
	case 0x0113: // WM_TIMER
		if keys == thumbnailDragTimer && drag.moving {
			bounds := v.pageListBounds()
			if contains(bounds, drag.current.x, drag.current.y) {
				bar := v.app.scrollPanes[v.pageList].bar
				margin := int32(v.app.scale(24))
				if drag.current.y < bounds.top+margin {
					bar.setPosition(bar.position - 1)
				} else if drag.current.y >= bounds.bottom-margin {
					bar.setPosition(bar.position + 1)
				}
				v.updatePageDrop()
			}
			return true
		}
	case 0x0215, 0x001f: // WM_CAPTURECHANGED, WM_CANCELMODE
		v.endPageDrag()
	}
	return false
}

func (v *documentView) updatePageDrop() {
	drag := &v.pageDrag
	oldBoundary, oldDelete := drag.boundary, drag.overDelete
	drag.boundary, drag.overDelete = -1, false
	if contains(v.pageListBounds(), drag.current.x, drag.current.y) {
		row := v.pageAt(drag.current.y)
		line := v.app.scrollPanes[v.pageList].lineHeight
		if int(drag.current.y)%line >= line/2 {
			row++
		}
		drag.boundary = min(row, len(v.order))
	} else {
		screen := drag.current
		user32.NewProc("ClientToScreen").Call(v.pageList, uintptr(unsafe.Pointer(&screen)))
		hover, _, _ := user32.NewProc("WindowFromPoint").Call(uintptr(uint32(screen.x)) | uintptr(uint32(screen.y))<<32)
		drag.overDelete = hover == v.deleteButton && len(v.selectedPagePositions()) < len(v.order)
	}
	if oldBoundary != drag.boundary {
		invalidate(v.pageList)
	}
	if oldDelete != drag.overDelete {
		invalidate(v.deleteButton)
	}
}

func (v *documentView) endPageDrag() {
	if !v.pageDrag.pressed {
		return
	}
	v.pageDrag = thumbnailDrag{}
	user32.NewProc("KillTimer").Call(v.pageList, thumbnailDragTimer)
	if capture, _, _ := user32.NewProc("GetCapture").Call(); capture == v.pageList {
		user32.NewProc("ReleaseCapture").Call()
	}
	invalidate(v.pageList)
	invalidate(v.deleteButton)
}
