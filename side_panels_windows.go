package main

import (
	"path/filepath"
	"unsafe"
)

const (
	sidebarRailWidth      = 28
	minimumWorkspaceWidth = 980
	minimumDocumentWidth  = 176
	idToggleThumbnails    = 230
	idToggleLeftPanels    = 231
	idToggleRightPanels   = 232
)

func showControl(handle uintptr, visible bool) {
	show := uintptr(0)
	if visible {
		show = 5
	}
	showWindow.Call(handle, show)
}

func (a *app) addDetailButton(id int, label uiMessage) {
	handle := a.shellButton(label, id, 145, 0)
	a.detailButtons[id] = handle
	a.panelButtons[handle] = true
}

func (a *app) detailsTab() int {
	if a.layoutAids.temporary {
		return idLayoutAids
	}
	if a.tool == idFileTab {
		switch a.mode {
		case idExportTab, idSignTab:
		default:
			return idFileDetails
		}
	}
	return a.mode
}

func (a *app) sidebarRails() (rect, rect) {
	width := int32(min(a.scale(sidebarRailWidth), a.width/2))
	left, right := int32(a.toolsWidth), int32(a.width-a.settingsWidth)
	top, bottom := a.viewport.top, a.viewport.bottom
	return rect{left, top, left + width, bottom},
		rect{right - width, top, right, bottom}
}

func (a *app) visibleRightAuxiliary() auxiliaryPanel {
	if a.full || a.effectiveRightDock().collapsed || !a.view.ready() {
		return auxiliaryNone
	}
	return a.rightAux.panel
}

func (a *app) toggleRightAuxiliary(panel auxiliaryPanel) {
	collapsed := a.rightDock.collapsed
	if collapsed {
		a.rightAux.setGroupCollapsed(&a.rightDock, false)
	}
	if !collapsed || a.rightAux.panel != panel {
		a.rightAux.togglePanel(panel)
	}
	a.layout()
	a.view.invalidateCanvas()
	a.view.schedule()
}

func (a *app) layoutDocumentPanels() {
	d := a.scale
	var minimum, preferred dockWidths
	if a.leftDock.showsPrimary() {
		minimum[dockTools] = d(168)
		preferred[dockTools] = a.preferredToolsWidth
	}
	if a.effectiveRightDock().showsPrimary() {
		minimum[dockProperties] = d(panelWidth)
		preferred[dockProperties] = a.preferredSettingsWidth()
	}
	ready, compare := a.view.ready(), a.comparisonVisible()
	if a.leftDock.showsAuxiliary(ready && !compare) {
		minimum[dockThumbnails] = d(70)
		preferred[dockThumbnails] = a.preferredThumbnailsWidth()
	}
	switch a.visibleRightAuxiliary() {
	case auxiliaryBookmarks:
		minimum[dockRightAuxiliary] = d(145)
		preferred[dockRightAuxiliary] = a.view.preferredBookmarkWidth()
	case auxiliaryResults:
		minimum[dockRightAuxiliary] = d(145)
		preferred[dockRightAuxiliary] = a.view.preferredSearchWidth()
	case auxiliaryLayers:
		minimum[dockRightAuxiliary] = d(145)
		preferred[dockRightAuxiliary] = max(d(248), a.captionWidth(a.uiText(uiLayersAndObjects), a.labelFont)+d(26))
	}
	preferredDocument := d(265)
	if compare {
		preferredDocument = d(420)
	}
	rail := min(d(sidebarRailWidth), a.width/2)
	widths := fitDockWidths(a.width-2*rail, d(minimumDocumentWidth), preferredDocument, minimum, preferred)
	a.toolsWidth, a.settingsWidth = widths[dockTools], widths[dockProperties]
	top, bottom := int32(d(workspaceTop)), int32(max(d(workspaceTop+1), a.height-d(footerHeight)))
	left, right := a.toolsWidth+rail, a.width-a.settingsWidth-rail
	docLeft, docRight := left+widths[dockThumbnails], right-widths[dockRightAuxiliary]
	a.thumbnailArea = rect{int32(left), top, int32(docLeft), bottom}
	a.rightAuxArea = rect{int32(docRight), top, int32(right), bottom}
	a.viewport = rect{int32(docLeft), top, int32(docRight), bottom}
}

func (a *app) preferredThumbnailsWidth() int {
	for _, sizes := range a.view.sizes[:min(16, len(a.view.sizes))] {
		if size := sizes["crop"]; size[0] > size[1] {
			return a.scale(220)
		}
	}
	return a.scale(182)
}

func (a *app) preferredSettingsWidth() int {
	d, tab := a.scale, a.detailsTab()
	width := d(panelWidth)
	if tab == idFileDetails {
		for _, path := range a.recent.files[:min(recentFilePreview, len(a.recent.files))] {
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

func (a *app) scaleSettingsItem(item panelItem) panelItem {
	extra := a.settingsWidth - a.scale(panelWidth)
	item.x = a.scale(item.x) + (item.x-16)*extra/298
	item.width = max(1, a.scale(item.width)+item.width*extra/298)
	return item
}

func (a *app) layoutSidebars() {
	d := a.scale
	toolsVisible := !a.full && a.toolsWidth > 0
	detailsVisible := !a.full && a.settingsWidth > 0
	y := int(a.viewport.top) + d(10)
	step := d(39)
	if a.documentTools {
		step = min(step, max(d(27), (int(a.viewport.bottom)-y)/len(workspaceTabs)))
	}
	for _, tab := range workspaceTabs {
		button := a.toolButtons[tab.id]
		show := toolsVisible && (!tab.needsPDF || a.documentTools)
		if show {
			place(button, d(10), y, max(1, a.toolsWidth-d(20)), step-d(5))
			y += step
		}
		showControl(button, show)
		invalidate(button)
	}
	leftRail, rightRail := a.sidebarRails()
	ready := a.view.ready()
	a.layoutRail(leftRail, a.leftPanelsToggle, a.toolsToggle, a.leftDock, ready && !a.comparisonVisible(), uiCollapseLeft, uiExpandLeft, a.thumbnailsToggle)
	a.layoutRail(rightRail, a.rightPanelsToggle, a.detailsToggle, a.effectiveRightDock(), ready, uiCollapseRight, uiExpandRight, a.resultsToggle, a.bookmarksToggle, a.layers.toggle)

	x, top := a.width-a.settingsWidth, int(a.viewport.top)
	gridWidth := max(1, d(145)+(a.settingsWidth-d(panelWidth))/2)
	tab := a.detailsTab()
	viewOptionsVisible := detailsVisible && a.tool == idBrowseTab && tab != idLayoutAids
	for index, id := range [...]int{idBrowseTab, idPagesTab} {
		button := a.detailButtons[id]
		if viewOptionsVisible {
			place(button, x+d(16)+index*(gridWidth+d(8)), top+d(39), gridWidth, d(30))
		}
		showControl(button, viewOptionsVisible)
		invalidate(button)
	}
	if detailsVisible {
		if tab == idLayoutAids {
			setText(a.sidebarTitle, a.uiText(uiLayoutAids))
		} else {
			for _, tool := range workspaceTabs {
				if tool.id == a.tool {
					setText(a.sidebarTitle, a.uiText(tool.name))
					break
				}
			}
		}
		place(a.sidebarTitle, x+d(16), top+d(9), max(1, a.settingsWidth-d(52)), d(22))
		detailY := top + d(38)
		switch {
		case tab == idFileDetails:
			detailY = top
		case viewOptionsVisible:
			detailY = top + d(76)
		}
		a.detailsArea = rect{int32(x), int32(detailY), int32(a.width), a.viewport.bottom}
	}
	showControl(a.sidebarTitle, detailsVisible && tab != idFileDetails)
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
	a.view.layoutSearchResults()
	a.layoutLayers()
	a.syncLayerSelection()
	a.layoutPanel()
}

func (a *app) layoutRail(rail rect, group, primary uintptr, dock dockVisibility, auxiliaryAvailable bool, collapse, expand uiMessage, auxiliary ...uintptr) {
	d := a.scale
	x, y, width := int(rail.left), int(rail.top)+d(10), int(rail.right-rail.left)
	visible := !a.full && width > 0
	panels := visible && !dock.collapsed
	if visible {
		label := a.uiText(collapse)
		if dock.collapsed {
			label = a.uiText(expand)
		}
		if getText(group) != label {
			setText(group, label)
			a.setTooltip(group, label)
		}
		place(group, x, y, width, d(30))
	}
	showControl(group, visible)
	invalidate(group)
	y += d(36)
	for _, handle := range append([]uintptr{primary}, auxiliary...) {
		show := panels && (handle == primary || auxiliaryAvailable)
		if show {
			height := max(d(88), a.captionWidth(getText(handle), a.labelFont)+d(28))
			place(handle, x, y, width, height)
			y += height + d(6)
		}
		showControl(handle, show)
		invalidate(handle)
	}
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
		collapsed, onLeft, group = a.leftDock.collapsed, true, true
	case a.rightPanelsToggle:
		collapsed, group = a.effectiveRightDock().collapsed, true
	case a.toolsToggle:
		collapsed, onLeft = a.leftDock.primaryCollapsed, true
	case a.detailsToggle:
		collapsed = a.effectiveRightDock().primaryCollapsed
	case a.thumbnailsToggle:
		collapsed, onLeft = a.leftDock.auxiliaryCollapsed, true
	case a.bookmarksToggle:
		collapsed = a.rightAux.panel != auxiliaryBookmarks
	case a.resultsToggle:
		collapsed = a.rightAux.panel != auxiliaryResults
	case a.layers.toggle:
		collapsed = a.rightAux.panel != auxiliaryLayers
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
	buttonColor, foreground := uiColors["panelTabActive"], textColor
	if collapsed {
		buttonColor = uiColors["panelTabIdle"]
	}
	if item.state&(1|64) != 0 {
		buttonColor = uiColors["panelTabHover"]
	}
	rounded(item.dc, bounds, d(5), buttonColor, divider)
	if group {
		drawText(item.dc, arrow+arrow, bounds, a.normalFont, foreground, 1)
		if item.state&16 != 0 {
			a.drawFocus(item.dc, bounds)
		}
		return true
	}
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

func (a *app) togglePanelSide(left bool) {
	if left {
		a.leftDock.collapsed = !a.leftDock.collapsed
		if !a.comparisonVisible() {
			a.view.endPageDrag()
		}
	} else {
		hidden := !a.effectiveRightDock().collapsed
		a.layoutAids.temporary = false
		a.rightAux.setGroupCollapsed(&a.rightDock, hidden)
	}
	a.layout()
	a.view.schedule()
}
