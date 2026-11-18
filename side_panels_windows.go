package main

const (
	toolPanelWidth   = 200
	sidebarRailWidth = 24
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
	if a.tool == idFileTab && a.mode != idExportTab && a.mode != idSignTab {
		return idFileDetails
	}
	return a.mode
}

func (a *app) sidebarRails() (rect, rect) {
	v := a.viewport
	width := int32(a.scale(sidebarRailWidth))
	return rect{v.left - width, v.top, v.left, v.bottom},
		rect{v.right, v.top, v.right + width, v.bottom}
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
		y := (int(a.viewport.top+a.viewport.bottom) - d(44)) / 2
		place(a.toolsToggle, int(leftRail.left), y, d(sidebarRailWidth), d(44))
		place(a.detailsToggle, int(rightRail.left), y, d(sidebarRailWidth), d(44))
	}
	if a.toolsCollapsed {
		setText(a.toolsToggle, "Rozbalit panel nástrojů")
	} else {
		setText(a.toolsToggle, "Sbalit panel nástrojů")
	}
	if a.detailsCollapsed {
		setText(a.detailsToggle, "Rozbalit panel nastavení")
	} else {
		setText(a.detailsToggle, "Sbalit panel nastavení")
	}
	showControl(a.toolsToggle, !a.full)
	showControl(a.detailsToggle, !a.full)
	invalidate(a.toolsToggle)
	invalidate(a.detailsToggle)

	x, top := int(rightRail.right), int(a.viewport.top)
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
	if !detailsVisible {
		showControl(a.thumbnailSidebar, false)
	}
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
	var arrow string
	switch item.handle {
	case a.toolsToggle:
		arrow = "‹"
		if a.toolsCollapsed {
			arrow = "›"
		}
	case a.detailsToggle:
		arrow = "›"
		if a.detailsCollapsed {
			arrow = "‹"
		}
	default:
		return false
	}
	fillColor(item.dc, item.bounds, sidebarRailColor)
	bounds := item.bounds
	bounds.left += 2
	bounds.right -= 2
	buttonColor := paletteColor{47, 43, 50}
	if item.state&1 != 0 {
		buttonColor = divider
	}
	rounded(item.dc, bounds, a.scale(5), buttonColor, divider)
	drawText(item.dc, arrow, bounds, a.titleFont, textColor, 1)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
	}
	return true
}
