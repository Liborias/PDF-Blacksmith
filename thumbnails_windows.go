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
	target                      *documentView
	preview                     dragPreview
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
	numbers := v.order[first:last]
	if number := v.draggedPage(); number != 0 && number != numbers[0] {
		numbers = append([]int{number}, numbers...)
	}
	for _, number := range numbers {
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
	if boundary := v.thumbnailDropBoundary(); boundary >= 0 {
		line := bounds.top
		switch {
		case boundary == index:
		case boundary == index+1 && index+1 == len(v.order):
			line = bounds.bottom - int32(d(2))
		default:
			return
		}
		fillColor(item.dc, rect{bounds.left, line, bounds.right, line + int32(d(2))}, selectionColor)
	}
}

func (v *documentView) thumbnailDropBoundary() int {
	for _, view := range [...]*documentView{v.app.view, v.app.comparison.view} {
		if view != nil && view.pageDrag.target == v {
			return view.pageDrag.boundary
		}
	}
	return -1
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
	a := v.app
	xy := point{int32(int16(position)), int32(int16(position >> 16))}
	drag := &v.pageDrag
	switch msg {
	case 0x0201: // WM_LBUTTONDOWN
		if !v.ready() || a.busy || (!a.comparisonVisible() && a.mode != idPagesTab) ||
			keys&(4|8) != 0 || !contains(v.pageListBounds(), xy.x, xy.y) { // Use native Shift and Ctrl selection.
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
			if !drag.moving && !a.busy && distance >= int32(a.scale(5)) {
				if !v.inPageSelection(drag.page) {
					v.selectOnlyPage(drag.page)
				}
				drag.moving = true
				v.showPageDragPreview()
				v.schedule()
				user32.NewProc("SetTimer").Call(v.pageList, thumbnailDragTimer, 160, 0)
			}
			if drag.moving {
				drag.preview.move(v.thumbnailScreenPoint())
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
			} else if !a.busy && last.overDelete {
				v.deleteSelectedPages()
			} else if !a.busy && last.target == v {
				v.moveSelectedPages(dropPosition(v.selectedPagePositions(), last.boundary))
			} else if !a.busy && last.target != nil {
				a.insertDraggedPages(v, last.target, last.boundary)
			}
			return true
		}
	case 0x0113: // WM_TIMER
		if keys == thumbnailDragTimer && drag.moving {
			if target := drag.target; target != nil {
				xy := v.thumbnailPoint(target)
				bounds := target.pageListBounds()
				bar := a.scrollPanes[target.pageList].bar
				margin := int32(a.scale(24))
				if xy.y < bounds.top+margin {
					bar.setPosition(bar.position - 1)
				} else if xy.y >= bounds.bottom-margin {
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

func (v *documentView) draggedPage() int {
	if drag := &v.pageDrag; drag.moving && drag.page < len(v.order) {
		return v.order[drag.page]
	}
	return 0
}

func (v *documentView) showPageDragPreview() {
	number := v.draggedPage()
	if number == 0 {
		return
	}
	d, drag := v.app.scale, &v.pageDrag
	image := v.thumbnails.frames[number].image
	if image == nil && v.box() == "crop" {
		image = v.reading.frames[number].image
	}
	size := v.pageDimensions(number, "crop")
	if image != nil {
		size = pageSize{float64(image.width), float64(image.height)}
	}
	pictureW, pictureH := d(88), d(112)
	if size.width > 0 && size.height > 0 {
		fit := min(float64(pictureW)/size.width, float64(pictureH)/size.height)
		pictureW, pictureH = max(1, int(size.width*fit)), max(1, int(size.height*fit))
	}
	padding := d(5)
	width, height := max(d(38), pictureW)+2*padding, max(d(38), pictureH)+2*padding
	count := len(v.selectedPagePositions())
	drag.preview.show(v.app.window, v.thumbnailScreenPoint(), point{int32(width), int32(height)}, int32(d(13)), int32(d(7)), func(dc uintptr) {
		bounds := rect{right: int32(width), bottom: int32(height)}
		fillColor(dc, bounds, panelColor)
		rounded(dc, bounds, d(7), panelColor, divider)
		x, y := (width-pictureW)/2, (height-pictureH)/2
		if image != nil {
			gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
			drawBitmap(dc, image, x, y, pictureW, pictureH)
		} else {
			paper := rect{int32(x), int32(y), int32(x + pictureW), int32(y + pictureH)}
			fillColor(dc, paper, paletteColor{235, 234, 232})
			drawText(dc, strconv.Itoa(drag.page+1), paper, v.app.labelFont, paperInkColor, 1)
		}
		if count > 1 {
			label := strconv.Itoa(count)
			w := v.app.captionWidth(label, v.app.labelFont) + d(14)
			right, bottom := width-d(2), height-d(2)
			badge := rect{int32(max(2, right-w)), int32(bottom - d(22)), int32(right), int32(bottom)}
			burgundy := paletteColor{133, 42, 75}
			rounded(dc, badge, d(7), burgundy, burgundy)
			drawText(dc, label, badge, v.app.labelFont, textColor, 1)
		}
	})
}

func (v *documentView) thumbnailScreenPoint() point {
	xy := v.pageDrag.current
	user32.NewProc("ClientToScreen").Call(v.pageList, uintptr(unsafe.Pointer(&xy)))
	return xy
}

func (v *documentView) thumbnailPoint(target *documentView) point {
	xy := v.pageDrag.current
	if target != v {
		xy = v.thumbnailScreenPoint()
		user32.NewProc("ScreenToClient").Call(target.pageList, uintptr(unsafe.Pointer(&xy)))
	}
	return xy
}

func (v *documentView) updatePageDrop() {
	a := v.app
	drag := &v.pageDrag
	oldTarget, oldBoundary, oldDelete := drag.target, drag.boundary, drag.overDelete
	drag.target, drag.boundary, drag.overDelete = nil, -1, false
	canReorder := !v.secondary && a.mode == idPagesTab
	for _, target := range [...]*documentView{a.view, a.comparison.view} {
		if (target == v && !canReorder) || (target != v && !a.comparisonVisible()) || !target.ready() {
			continue
		}
		visible, _, _ := user32.NewProc("IsWindowVisible").Call(target.pageList)
		xy := v.thumbnailPoint(target)
		if visible == 0 || !contains(target.pageListBounds(), xy.x, xy.y) {
			continue
		}
		row := target.pageAt(xy.y)
		line := a.scrollPanes[target.pageList].lineHeight
		if int(xy.y)%line >= line/2 {
			row++
		}
		drag.target, drag.boundary = target, min(row, len(target.order))
		break
	}
	if drag.target == nil && canReorder && a.settingsWidth > 0 && !a.full {
		screen := v.thumbnailScreenPoint()
		hover, _, _ := user32.NewProc("WindowFromPoint").Call(uintptr(uint32(screen.x)) | uintptr(uint32(screen.y))<<32)
		drag.overDelete = hover == v.deleteButton && len(v.selectedPagePositions()) < len(v.order)
	}
	if oldTarget != drag.target || oldBoundary != drag.boundary {
		if oldTarget != nil {
			invalidate(oldTarget.pageList)
		}
		if drag.target != nil && drag.target != oldTarget {
			invalidate(drag.target.pageList)
		}
	}
	if oldDelete != drag.overDelete {
		invalidate(v.deleteButton)
	}
}

func (v *documentView) endPageDrag() {
	if !v.pageDrag.pressed {
		return
	}
	target, overDelete := v.pageDrag.target, v.pageDrag.overDelete
	v.pageDrag.preview.close()
	v.pageDrag = thumbnailDrag{}
	user32.NewProc("KillTimer").Call(v.pageList, thumbnailDragTimer)
	if capture, _, _ := user32.NewProc("GetCapture").Call(); capture == v.pageList {
		user32.NewProc("ReleaseCapture").Call()
	}
	if target != nil {
		invalidate(target.pageList)
	}
	if overDelete {
		invalidate(v.deleteButton)
	}
}

func (a *app) cancelPageDrag() bool {
	cancelled := false
	for _, v := range [...]*documentView{a.view, a.comparison.view} {
		if v != nil && v.pageDrag.pressed {
			v.endPageDrag()
			cancelled = true
		}
	}
	return cancelled
}
