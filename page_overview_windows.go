package main

import (
	"math"
	"slices"
	"strconv"
)

const idFitOverviewPage = 235

func (v *documentView) overviewBounds() rect {
	bounds := v.viewport
	bounds.top = min(bounds.bottom, bounds.top+int32(v.app.scale(32)))
	bounds.right = max(bounds.left, bounds.right-int32(v.app.scale(14)))
	return bounds
}

func (v *documentView) relayoutOverview(anchor, relativeY int, rowFraction float64) {
	bounds, d := v.overviewBounds(), v.app.scale
	v.overviewGrid = layoutOverview(len(v.order), int(bounds.right-bounds.left), int(float64(d(172))*v.overviewScale), d(12), d(28))
	if anchor >= 0 && anchor < v.overviewGrid.count {
		_, top := v.overviewGrid.card(anchor)
		v.overviewScroll = top + int(math.Round(rowFraction*float64(v.overviewGrid.cardHeight))) - relativeY
	}
	v.updateOverviewBar()
}

func (v *documentView) updateOverviewBar() {
	bar, bounds := v.overviewBar, v.overviewBounds()
	height := int(bounds.bottom - bounds.top)
	v.overviewScroll = max(0, min(v.overviewScroll, v.overviewGrid.height-height))
	visible := v.overview && v.ready() && (!v.secondary || v.app.comparisonVisible())
	if visible {
		place(bar.handle, int(bounds.right), int(bounds.top), v.app.scale(14), height)
	}
	bar.update(v.overviewGrid.height, height, v.overviewScroll, visible)
}

func (v *documentView) rebuildOverview() {
	if !v.overview || !v.ready() {
		v.overviewGrid = overviewLayout{}
		v.updateOverviewBar()
		return
	}
	old, bounds := v.overviewGrid, v.overviewBounds()
	index, relativeY := -1, 0
	if v.page >= 0 && v.page < old.count {
		_, top := old.card(v.page)
		if top < v.overviewScroll+int(bounds.bottom-bounds.top) && top+old.cardHeight > v.overviewScroll {
			index, relativeY = v.page, top-v.overviewScroll
		}
	}
	v.relayoutOverview(index, relativeY, 0)
}

func (v *documentView) overviewVisiblePages() (int, int) {
	bounds := v.overviewBounds()
	return v.overviewGrid.visible(v.overviewScroll, int(bounds.bottom-bounds.top))
}

func (v *documentView) scrollOverview(offset int) {
	v.overviewScroll = offset
	v.updateOverviewBar()
	v.invalidateCanvas()
	v.schedule()
}

func (v *documentView) revealOverviewPage(index int) {
	grid := v.overviewGrid
	if index < 0 || index >= grid.count {
		return
	}
	_, top := grid.card(index)
	bounds := v.overviewBounds()
	height := int(bounds.bottom - bounds.top)
	if top < v.overviewScroll || top >= v.overviewScroll+height {
		v.overviewScroll = top
	} else if bottom := top + grid.cardHeight + grid.gap; grid.cardHeight <= height && bottom > v.overviewScroll+height {
		v.overviewScroll = bottom - height
	}
	v.updateOverviewBar()
}

func (v *documentView) setPageOverview(show bool) {
	if !v.ready() || v.overview == show {
		return
	}
	if show {
		v.pageSelection = slices.Clone(v.selectedPagePositions())
		v.overviewAnchor = v.page
	}
	v.overview = show
	v.reference = v.page
	v.pageWheel = pageWheelGesture{}
	v.finishZoomTimer()
	v.rebuild(false)
	if show {
		v.revealOverviewPage(v.page)
	}
	v.resetPageListSelection()
	v.changedViewport()
}

func (a *app) togglePageOverview() {
	if a.busy || !a.view.ready() || a.saveInlineBefore(a.togglePageOverview) {
		return
	}
	a.cancelPageDrag()
	a.cancelObjectGesture()
	a.view.cancelMiddleSelection()
	a.clearTextSelection()
	a.clearObjectSelection()
	v := a.view
	v.setPageOverview(!v.overview)
	if v.overview {
		a.applyThumbnailScope()
	}
	a.comparison.canvasFocused = false
	setFocus.Call(a.window)
}

func (v *documentView) fitOverviewSelection() {
	if !v.overview || len(v.pageSelection) == 0 || v.app.busy {
		return
	}
	if !v.inPageSelection(v.page) {
		v.page = v.pageSelection[0]
	}
	v.overview, v.continuous = false, false
	v.syncReadingControls()
	v.setZoomMode(readingZoomFitPage)
	setFocus.Call(v.app.window)
}

func (v *documentView) selectOverviewPage(index int, toggle, extend bool) {
	if !v.ready() || index < 0 || index >= len(v.order) {
		return
	}
	v.pageSelection = overviewSelection(v.pageSelection, min(max(v.overviewAnchor, 0), len(v.order)-1), index, toggle, extend)
	if !extend {
		v.overviewAnchor = index
	}
	v.gotoPage(index, false)
	v.resetPageListSelection()
	if !v.secondary {
		v.app.applyThumbnailScope()
	}
}

func (v *documentView) selectAllOverviewPages() {
	v.pageSelection = make([]int, len(v.order))
	for i := range v.pageSelection {
		v.pageSelection[i] = i
	}
	v.resetPageListSelection()
	if !v.secondary {
		v.app.applyThumbnailScope()
	}
	v.changedViewport()
}

func (v *documentView) overviewPointer(msg uint32, x, y int32) bool {
	if msg < 0x0200 || msg > 0x0209 { // Client mouse messages only.
		return false
	}
	if !contains(v.viewport, x, y) {
		return false
	}
	if (msg == 0x0201 || msg == 0x0203) && v.ready() {
		setFocus.Call(v.app.window)
		bounds := v.overviewBounds()
		if contains(bounds, x, y) {
			index := v.overviewGrid.at(int(x-bounds.left), int(y-bounds.top)+v.overviewScroll)
			v.selectOverviewPage(index, keyDown(0x11), keyDown(0x10))
		}
	}
	return true
}

func (v *documentView) overviewKey(key uintptr) bool {
	if !v.ready() {
		return false
	}
	index := v.page
	switch key {
	case 0x25:
		index--
	case 0x27:
		index++
	case 0x26:
		index -= v.overviewGrid.columns
	case 0x28:
		index += v.overviewGrid.columns
	case 0x24:
		index = 0
	case 0x23:
		index = len(v.order) - 1
	case 0x21, 0x22, 0x20:
		bounds := v.overviewBounds()
		distance := int(bounds.bottom-bounds.top) * 9 / 10
		if key == 0x21 {
			distance = -distance
		}
		v.scrollOverview(v.overviewScroll + distance)
		return true
	default:
		return false
	}
	v.selectOverviewPage(min(max(index, 0), len(v.order)-1), false, keyDown(0x10))
	return true
}

func (v *documentView) zoomOverview(delta, x, y float64) {
	next := min(2.6, max(0.48, v.overviewScale*math.Pow(1.2, delta)))
	if next == v.overviewScale {
		return
	}
	bounds := v.overviewBounds()
	relativeY := int(y) - int(bounds.top-v.viewport.top)
	index := v.overviewGrid.at(int(x), relativeY+v.overviewScroll)
	var fraction float64
	if index >= 0 {
		_, top := v.overviewGrid.card(index)
		fraction = float64(relativeY+v.overviewScroll-top) / float64(v.overviewGrid.cardHeight)
	}
	v.overviewScale = next
	v.relayoutOverview(index, relativeY, fraction)
	v.changedViewport()
}

func (v *documentView) paintOverview(dc uintptr) {
	a, grid, bounds := v.app, v.overviewGrid, v.overviewBounds()
	caption := v.viewport
	caption.left += int32(a.scale(12))
	caption.bottom = bounds.top
	drawText(dc, a.uiLocale.formatMessage(localizedMessage{uiOverviewSelection, []any{len(v.pageSelection)}}),
		caption, a.labelFont, textColor, 0x8000)
	if grid.columns == 0 {
		return
	}
	saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
	gdi32.NewProc("IntersectClipRect").Call(dc, uintptr(bounds.left), uintptr(bounds.top), uintptr(bounds.right), uintptr(bounds.bottom))
	defer gdi32.NewProc("RestoreDC").Call(dc, saved)
	gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
	focus, _, _ := user32.NewProc("GetFocus").Call()
	first, last := v.overviewVisiblePages()
	for index := first; index < last; index++ {
		x, y := grid.card(index)
		x, y = x+int(bounds.left), y+int(bounds.top)-v.overviewScroll
		card := rect{int32(x), int32(y), int32(x + grid.cardWidth), int32(y + grid.cardHeight)}
		background, outline := uiColors["surface"], divider
		if v.inPageSelection(index) {
			background, outline = uiColors["selected"], selectionColor
		}
		rounded(dc, card, a.scale(6), background, outline)
		if focus == a.window && a.focusedDocumentView() == v && index == v.page {
			a.drawFocus(dc, card)
		}
		number := v.order[index]
		size := v.pageDimensions(number, "crop")
		captionHeight, margin := a.scale(23), a.scale(7)
		fit := min(float64(max(1, grid.cardWidth-2*margin))/size.width,
			float64(max(1, grid.cardHeight-captionHeight-2*margin))/size.height)
		width, height := max(1, int(size.width*fit)), max(1, int(size.height*fit))
		left, top := x+(grid.cardWidth-width)/2, y+(grid.cardHeight-captionHeight-height)/2
		if image := v.thumbnails.frames[number].image; image != nil {
			drawBitmap(dc, image, left, top, width, height)
		} else {
			fillColor(dc, rect{int32(left), int32(top), int32(left + width), int32(top + height)}, uiColors["subtle"])
		}
		label := rect{card.left, card.bottom - int32(captionHeight), card.right, card.bottom}
		drawText(dc, strconv.Itoa(index+1), label, a.normalFont, textColor, 1)
	}
}
