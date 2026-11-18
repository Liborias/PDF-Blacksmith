package main

import "unsafe"

const (
	thumbnailPixels    = 400
	thumbnailCacheSize = 8 << 20
	thumbnailDragTimer = 7
	wmThumbnailUpdate  = 0x8003
)

type thumbnailDrag struct {
	start, current              point
	page, boundary              int
	pressed, moving, overDelete bool
}

func (v *documentView) thumbnailRange() (int, int) {
	pane := v.app.scrollPanes[v.pageList]
	visible, _, _ := user32.NewProc("IsWindowVisible").Call(v.pageList)
	if visible == 0 || !pane.shown || !v.ready() {
		return 0, 0
	}
	_, rows, top := pane.lines()
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
	const estimate = (thumbnailPixels + 2) * (thumbnailPixels + 2) * 4
	budget := v.thumbnailBudget()
	if budget < estimate {
		return false
	}
	for _, number := range v.order[first:last] {
		if v.thumbnails.frames[number].image != nil || v.thumbnails.failed[number] {
			continue
		}
		v.trimThumbnails(budget - estimate)
		if v.thumbnails.bytes+estimate > budget {
			break
		}
		req := v.request("thumbnail")
		req.View.Page, req.View.Rotation, req.View.Pixels, req.View.Box = number, v.rotations[number-1], thumbnailPixels, "crop"
		v.launch(req, number, thumbnailPixels, true, nil)
		return true
	}
	return false
}

func (v *documentView) drawPageThumbnail(item *buttonDraw, index int, color paletteColor) {
	d, bounds := v.app.scale, item.bounds
	number := v.order[index]
	image := v.thumbnails.frames[number].image
	width, height := d(78), d(93)
	size := v.pageDimensions(number, "crop")
	if image != nil {
		size = pageSize{float64(image.width), float64(image.height)}
	}
	if size.width > 0 && size.height > 0 {
		fit := min(float64(width)/size.width, float64(height)/size.height)
		width, height = max(1, int(size.width*fit)), max(1, int(size.height*fit))
	}
	x := int(bounds.left) + d(7) + (d(78)-width)/2
	y := (int(bounds.top+bounds.bottom) - height) / 2
	if image != nil {
		gdi32.NewProc("SetStretchBltMode").Call(item.dc, 4)
		drawBitmap(item.dc, image, x, y, width, height)
	} else {
		fillColor(item.dc, rect{int32(x), int32(y), int32(x + width), int32(y + height)}, divider)
	}
	label := bounds
	label.left += int32(d(99))
	drawText(item.dc, pageListLabel(index, number), label, v.app.normalFont, color, 0x8000)
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
	user32.NewProc("GetWindowRect").Call(v.app.sidebar, uintptr(unsafe.Pointer(&sidebar)))
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
	xy := point{int32(int16(position)), int32(int16(position >> 16))}
	drag := &v.pageDrag
	switch msg {
	case 0x0201: // WM_LBUTTONDOWN
		if !v.ready() || keys&(4|8) != 0 { // Use native Shift and Ctrl selection.
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
			if !drag.moving && !v.app.busy && distance >= int32(v.app.scale(5)) {
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
