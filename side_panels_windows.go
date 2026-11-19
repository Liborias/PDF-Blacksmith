package main

import (
	"path/filepath"
	"unsafe"
)

const (
	sidebarRailWidth      = 28
	minimumWorkspaceWidth = 1000
	minimumDocumentWidth  = 176
	idToggleThumbnails    = 230
	idToggleLeftPanels    = 231
	idToggleRightPanels   = 232
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
		left += a.toolsWidth
	}
	rail := d(sidebarRailWidth)
	thumbOpen, bookOpen := a.view.ready() && !a.thumbnailsCollapsed && !a.comparisonVisible(), a.view.ready() && !a.bookmarksCollapsed
	reserved := 0
	if thumbOpen {
		reserved += d(70)
	}
	if bookOpen {
		reserved += d(145)
	}
	a.settingsWidth = 0
	if !a.detailsCollapsed {
		limit := max(d(panelWidth), right-left-2*rail-reserved-d(minimumDocumentWidth))
		a.settingsWidth = min(a.preferredSettingsWidth(), limit)
		right -= a.settingsWidth
	}
	top, bottom := int32(d(headerHeight)), int32(max(d(headerHeight+1), a.height-d(footerHeight)))
	room := right - left - 2*rail
	minimumPreview := max(d(minimumDocumentWidth), min(d(265), room-reserved))
	room = max(0, room-minimumPreview)
	if thumbOpen {
		width := d(182)
		for _, sizes := range a.view.sizes[:min(16, len(a.view.sizes))] {
			if size := sizes["crop"]; size[0] > size[1] {
				width = d(220)
				break
			}
		}
		if bookOpen {
			width = min(width, max(d(70), room-a.view.preferredBookmarkWidth()))
		}
		width = min(width, room)
		a.thumbnailArea = rect{int32(left), top, int32(left + width), bottom}
		left += width
		room -= width
	}
	if bookOpen {
		width := min(a.view.preferredBookmarkWidth(), room)
		a.bookmarkArea = rect{int32(right - width), top, int32(right), bottom}
		right -= width
	}
	a.viewport = rect{int32(left + rail), top, int32(right - rail), bottom}
}

func (a *app) preferredSettingsWidth() int {
	d, tab := a.scale, a.detailsTab()
	width := d(panelWidth)
	if tab == idSearchTab {
		width = max(width, a.view.search.widest+d(80))
	}
	if tab == idRecentFiles {
		for _, path := range a.recent.files {
			width = max(width, a.captionWidth(filepath.Base(path), a.normalFont)+d(80))
		}
	}
	for index, item := range a.panelItems[tab] {
		if choice := a.choices[item.handle]; choice != nil && item.width >= 240 {
			if _, shown := a.panelItemLayout(tab, index, item); shown {
				width = max(width, a.captionWidth(choice.names[choice.index], a.normalFont)+d(item.x+64))
			}
		}
	}
	return min(width, d(454))
}

func (a *app) settingsItem(tab, index int, item panelItem) (panelItem, bool) {
	item, shown := a.panelItemLayout(tab, index, item)
	extra := max(0, a.settingsWidth-a.scale(panelWidth))
	item.x = a.scale(item.x) + (item.x-16)*extra/298
	item.width = a.scale(item.width) + item.width*extra/298
	return item, shown
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
			place(button, d(10), y, a.toolsWidth-d(20), d(34))
			y += d(39)
		}
		showControl(button, show)
		invalidate(button)
	}
	leftRail, rightRail := a.sidebarRails()
	if !a.full {
		a.layoutRail(leftRail, a.leftPanelsToggle, a.toolsToggle, a.thumbnailsToggle)
		a.layoutRail(rightRail, a.rightPanelsToggle, a.detailsToggle, a.bookmarksToggle)
	}
	showControl(a.leftPanelsToggle, !a.full)
	showControl(a.rightPanelsToggle, !a.full)
	showControl(a.toolsToggle, !a.full)
	showControl(a.detailsToggle, !a.full)
	showControl(a.thumbnailsToggle, !a.full && a.view.ready() && !a.comparisonVisible())
	showControl(a.bookmarksToggle, !a.full && a.view.ready())
	for _, tab := range [...]uintptr{a.leftPanelsToggle, a.toolsToggle, a.thumbnailsToggle, a.rightPanelsToggle, a.detailsToggle, a.bookmarksToggle} {
		invalidate(tab)
	}

	x, top := a.width-a.settingsWidth, int(a.viewport.top)
	gridWidth := d(145) + max(0, a.settingsWidth-d(panelWidth))/2
	fileOptionsVisible := detailsVisible && a.tool == idFileTab
	viewOptionsVisible := detailsVisible && a.tool == idBrowseTab
	for index, option := range fileOptions {
		button := a.detailButtons[option.id]
		if fileOptionsVisible {
			place(button, x+d(16)+(index%2)*(gridWidth+d(8)), top+d(39+(index/2)*35), gridWidth, d(30))
		}
		showControl(button, fileOptionsVisible)
		invalidate(button)
	}
	for index, id := range [...]int{idBrowseTab, idPagesTab} {
		button := a.detailButtons[id]
		if viewOptionsVisible {
			place(button, x+d(16)+index*(gridWidth+d(8)), top+d(39), gridWidth, d(30))
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
		place(a.sidebarTitle, x+d(16), top+d(9), a.settingsWidth-d(52), d(22))
		detailY := top + d(38)
		switch {
		case fileOptionsVisible:
			detailY = top + d(181)
		case viewOptionsVisible:
			detailY = top + d(76)
		}
		a.detailsArea = rect{int32(x), int32(detailY), int32(a.width), a.viewport.bottom}
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
	if show {
		user32.NewProc("RedrawWindow").Call(a.thumbnailSidebar, 0, 0, 0x0085) // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN
	}
	a.view.layoutBookmarkPanel()
	a.layoutPanel()
}

func (a *app) layoutRail(rail rect, group, primary, auxiliary uintptr) {
	d := a.scale
	x, y, width := int(rail.left), int(rail.top)+d(10), d(sidebarRailWidth)
	place(group, x, y, width, d(104))
	place(primary, x, y+d(110), width, d(88))
	place(auxiliary, x, y+d(204), width, d(88))
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
	var collapsed, onLeft, group bool
	switch item.handle {
	case a.leftPanelsToggle:
		collapsed, onLeft, group = a.sideCollapsed(true), true, true
	case a.rightPanelsToggle:
		collapsed, group = a.sideCollapsed(false), true
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
	if group {
		arrow += arrow
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

func (a *app) sideCollapsed(left bool) bool {
	if left {
		return a.toolsCollapsed && (a.thumbnailsCollapsed || !a.view.ready() || a.comparisonVisible())
	}
	return a.detailsCollapsed && (a.bookmarksCollapsed || !a.view.ready())
}

func (a *app) togglePanelSide(left bool) {
	if left && a.comparisonVisible() {
		a.toolsCollapsed = !a.toolsCollapsed
		a.layout()
		return
	}
	primary, auxiliary, previous := &a.detailsCollapsed, &a.bookmarksCollapsed, &a.rightPanelsOpen
	if left {
		primary, auxiliary, previous = &a.toolsCollapsed, &a.thumbnailsCollapsed, &a.leftPanelsOpen
		a.view.endPageDrag()
	}
	if !a.sideCollapsed(left) {
		*previous = [2]bool{!*primary, !*auxiliary}
		*primary, *auxiliary = true, true
	} else if previous[0] || (previous[1] && a.view.ready()) {
		*primary, *auxiliary = !previous[0], !previous[1]
	} else {
		*primary = false
	}
	a.layout()
	a.view.schedule()
}
