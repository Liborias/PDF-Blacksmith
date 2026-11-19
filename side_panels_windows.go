package main

import "unsafe"

const (
	toolPanelWidth        = 200
	sidebarRailWidth      = 28
	thumbnailPanelWidth   = 220
	bookmarkPanelWidth    = 236
	minimumWorkspaceWidth = 1000
	minimumDocumentWidth  = 176
	idToggleThumbnails    = 230
)

var sidebarRailColor = paletteColor{28, 27, 32}

func showControl(handle uintptr, visible bool) {
	show := uintptr(0)
	if visible {
		show = 5
	}
	showWindow.Call(handle, show)
}

func (a *app) addDetailButton(id int, label string) {
	handle := a.button(label, id, 0, 0, 145, 0)
	a.detailButtons[id] = handle
	a.panelButtons[handle] = true
}

func (a *app) detailsTab() int {
	if a.tool == idFileTab {
		switch a.mode {
		case idExportTab, idSignTab, idRecentFiles:
		default:
			return idFileDetails
		}
	}
	return a.mode
}

func (a *app) sidebarRails() (rect, rect) {
	v := a.viewport
	width := int32(a.scale(sidebarRailWidth))
	return rect{v.left - width, v.top, v.left, v.bottom},
		rect{v.right, v.top, v.right + width, v.bottom}
}

func (a *app) layoutDocumentPanels() {
	d := a.scale
	left, right := 0, a.width
	if !a.toolsCollapsed {
		left += d(toolPanelWidth)
	}
	if !a.detailsCollapsed {
		right -= d(panelWidth)
	}
	top, bottom := int32(d(headerHeight)), int32(max(d(headerHeight+1), a.height-d(footerHeight)))
	rail := d(sidebarRailWidth)
	if a.view.ready() {
		room := max(0, right-left-2*rail-d(minimumDocumentWidth))
		if !a.thumbnailsCollapsed {
			width := min(d(thumbnailPanelWidth), room)
			if !a.bookmarksCollapsed {
				// Keep room for the bookmark editor on narrow windows.
				width = min(width, room/2, max(d(70), room-d(145)))
			}
			a.thumbnailArea = rect{int32(left), top, int32(left + width), bottom}
			left += width
			room -= width
		}
		if !a.bookmarksCollapsed {
			width := min(d(bookmarkPanelWidth), room)
			a.bookmarkArea = rect{int32(right - width), top, int32(right), bottom}
			right -= width
		}
	}
	a.viewport = rect{int32(left + rail), top, int32(right - rail), bottom}
}

func (a *app) layoutSidebars() {
	d := a.scale
	toolsVisible := !a.full && !a.toolsCollapsed
	detailsVisible := !a.full && !a.detailsCollapsed
	y := int(a.viewport.top) + d(10)
	for _, tab := range workspaceTabs {
		button := a.toolButtons[tab.id]
		show := toolsVisible && (!tab.needsPDF || a.documentTools)
		if show {
			place(button, d(10), y, d(toolPanelWidth-20), d(34))
			y += d(39)
		}
		showControl(button, show)
		invalidate(button)
	}
	leftRail, rightRail := a.sidebarRails()
	if !a.full {
		y := int(a.viewport.top) + d(10)
		place(a.toolsToggle, int(leftRail.left), y, d(sidebarRailWidth), d(88))
		place(a.detailsToggle, int(rightRail.left), y, d(sidebarRailWidth), d(88))
		place(a.thumbnailsToggle, int(leftRail.left), y+d(94), d(sidebarRailWidth), d(88))
		place(a.bookmarksToggle, int(rightRail.left), y+d(94), d(sidebarRailWidth), d(88))
	}
	showControl(a.toolsToggle, !a.full)
	showControl(a.detailsToggle, !a.full)
	showControl(a.thumbnailsToggle, !a.full && a.view.ready())
	showControl(a.bookmarksToggle, !a.full && a.view.ready())
	for _, tab := range [...]uintptr{a.toolsToggle, a.detailsToggle, a.thumbnailsToggle, a.bookmarksToggle} {
		invalidate(tab)
	}

	x, top := a.width-d(panelWidth), int(a.viewport.top)
	fileOptionsVisible := detailsVisible && a.tool == idFileTab
	viewOptionsVisible := detailsVisible && a.tool == idBrowseTab
	for index, option := range fileOptions {
		button := a.detailButtons[option.id]
		if fileOptionsVisible {
			place(button, x+d(16+(index%2)*153), top+d(39+(index/2)*35), d(145), d(30))
		}
		showControl(button, fileOptionsVisible)
		invalidate(button)
	}
	for index, id := range [...]int{idBrowseTab, idPagesTab} {
		button := a.detailButtons[id]
		if viewOptionsVisible {
			place(button, x+d(16+index*153), top+d(39), d(145), d(30))
		}
		showControl(button, viewOptionsVisible)
		invalidate(button)
	}
	if detailsVisible {
		for _, tab := range workspaceTabs {
			if tab.id == a.tool {
				setText(a.sidebarTitle, tab.name)
				break
			}
		}
		place(a.sidebarTitle, x+d(16), top+d(9), d(298), d(22))
		detailY := top + d(38)
		switch {
		case fileOptionsVisible:
			detailY = top + d(181)
		case viewOptionsVisible:
			detailY = top + d(76)
		}
		a.detailsArea = rect{int32(x), int32(detailY), int32(x + d(panelWidth)), a.viewport.bottom}
		if fileOptionsVisible {
			a.updateFileOptions()
		}
	}
	showControl(a.sidebarTitle, detailsVisible)
	showControl(a.sidebar, detailsVisible)
	thumb := a.thumbnailArea
	show := thumb.right > thumb.left && !a.full
	if show {
		width, height := int(thumb.right-thumb.left), int(thumb.bottom-thumb.top)
		place(a.thumbnailSidebar, int(thumb.left), int(thumb.top), width, height)
		a.view.layoutThumbnails(width, height)
	}
	showControl(a.thumbnailSidebar, show)
	a.view.layoutBookmarkPanel()
	a.layoutPanel()
}

func (a *app) paintSidebars(dc uintptr) {
	leftRail, rightRail := a.sidebarRails()
	fillColor(dc, rect{0, leftRail.top, leftRail.left, leftRail.bottom}, panelColor)
	fillColor(dc, rect{rightRail.right, rightRail.top, int32(a.width), rightRail.bottom}, panelColor)
	for _, rail := range [...]rect{leftRail, rightRail} {
		fillColor(dc, rail, sidebarRailColor)
	}
	fillColor(dc, rect{leftRail.right - 1, leftRail.top, leftRail.right, leftRail.bottom}, divider)
	fillColor(dc, rect{rightRail.left, rightRail.top, rightRail.left + 1, rightRail.bottom}, divider)
}

func (a *app) drawPanelToggle(item *buttonDraw) bool {
	var collapsed, onLeft bool
	switch item.handle {
	case a.toolsToggle:
		collapsed, onLeft = a.toolsCollapsed, true
	case a.detailsToggle:
		collapsed = a.detailsCollapsed
	case a.thumbnailsToggle:
		collapsed, onLeft = a.thumbnailsCollapsed, true
	case a.bookmarksToggle:
		collapsed = a.bookmarksCollapsed
	default:
		return false
	}
	arrow := "›"
	if onLeft != collapsed {
		arrow = "‹"
	}
	fillColor(item.dc, item.bounds, sidebarRailColor)
	bounds := item.bounds
	d := a.scale
	bounds.left += int32(d(2))
	bounds.right -= int32(d(2))
	buttonColor, foreground := sidebarRailColor, selectionColor
	if collapsed {
		buttonColor, foreground = paletteColor{47, 43, 50}, textColor
	}
	if item.state&1 != 0 {
		buttonColor = divider
	}
	rounded(item.dc, bounds, d(5), buttonColor, divider)
	arrowBounds := rect{bounds.left, bounds.top + int32(d(1)), bounds.right, bounds.top + int32(d(18))}
	drawText(item.dc, arrow, arrowBounds, a.normalFont, foreground, 1)
	caption := wideString(getText(item.handle))
	length, _, _ := getTextLen.Call(item.handle)
	font, _, _ := gdiSelect.Call(item.dc, a.railFont)
	align := gdi32.NewProc("SetTextAlign")
	previous, _, _ := align.Call(item.dc, 30) // TA_CENTER | TA_BASELINE
	gdiTextColor.Call(item.dc, foreground.win())
	gdiBkMode.Call(item.dc, 1)
	x := (bounds.left+bounds.right)/2 + int32(d(5))
	y := (arrowBounds.bottom + bounds.bottom) / 2
	gdi32.NewProc("TextOutW").Call(item.dc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(caption)), length)
	align.Call(item.dc, previous)
	gdiSelect.Call(item.dc, font)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
	}
	return true
}
