package main

import "fmt"

const (
	commandBarHeight = 38
	workspaceTop     = headerHeight + commandBarHeight
)

const (
	idHeaderFile = 570 + iota
	idHeaderView
	idHeaderTools
	idHeaderEdit
	idHeaderForms
	idHeaderHelp
	idHeaderHelpIcon
	idHeaderFlow
	idHeaderFitPage
	idHeaderFitWidth
	idHeaderSave
	idHeaderSettings
	idHeaderThemeMode
	idHeaderActualSize
	idHeaderOverview
	idHeaderPointer
	idHeaderRectangle
	idHeaderLasso
	idHeaderZoomPercent
	idHeaderPreviousPage
	idHeaderPageCount
	idHeaderNextPage
	idHeaderCollapseDocks
)

type commandBarButton struct {
	id               int
	label            uiMessage
	symbol, shortcut string
	menu             bool
}

type commandBarControl struct {
	commandBarButton
	handle uintptr
}

type commandBar struct {
	controls   []commandBarControl
	openMenu   int
	fitMode    readingZoomMode
	overview   bool
	areaMode   int
	searchEdit uintptr
	searchSync bool
}

var headerButtons = []commandBarButton{
	{id: idHeaderFile, label: uiFile, menu: true},
	{id: idHeaderView, label: uiView, menu: true},
	{id: idHeaderTools, label: uiTools, menu: true},
	{id: idHeaderEdit, label: uiEdit, menu: true},
	{id: idHeaderForms, label: uiForms, menu: true},
	{id: idHeaderHelp, label: uiHelpMenu, menu: true},
	{id: idHeaderPointer, label: uiSwitchToCanvasText, symbol: "i-cursor"},
	{id: idHeaderRectangle, label: uiCanvasRectangle, symbol: "i-rect_select"},
	{id: idHeaderLasso, label: uiCanvasLasso, symbol: "i-lasso"},
	{id: idToggleLayoutGrid, label: uiShowLayoutGrid, symbol: "i-toggle_grid"},
	{id: idHeaderUndo, label: uiUndo, symbol: "i-undo", shortcut: "Ctrl+Z"},
	{id: idHeaderRedo, label: uiRedo, symbol: "i-redo", shortcut: "Ctrl+Y"},
	{id: idHeaderFlow, label: uiSwitchToContinuous, symbol: "i-view_pages"},
	{id: idHeaderOverview, label: uiPageOverview, symbol: "i-nav_thumbnails"},
	{id: idReaderZoomOut, label: uiZoomOut, symbol: "i-zoom_out"},
	{id: idHeaderZoomPercent, label: uiMessage("ui.priblizeni.e4a6a87be9")},
	{id: idReaderZoomIn, label: uiZoomIn, symbol: "i-zoom_in"},
	{id: idHeaderActualSize, label: uiActualSize, symbol: "i-zoom_1_1"},
	{id: idHeaderFitPage, label: uiFitPage, symbol: "i-fit_page"},
	{id: idHeaderFitWidth, label: uiFitWidth, symbol: "i-fit_width"},
	{id: idSearchTab, label: uiFind, symbol: "i-search", shortcut: "Ctrl+F"},
	{id: idHeaderPreviousPage, label: uiPreviousPage, symbol: "‹"},
	{id: idHeaderPageCount, label: uiGoToPage},
	{id: idHeaderNextPage, label: uiNextPage, symbol: "›"},
	{id: idCompareOpen, label: uiSideBySide, symbol: "i-side_by_side"},
	{id: idHeaderCollapseDocks, label: uiCollapseBothDocks, symbol: "i-side_by_side"},
	{id: idHeaderSave, label: uiSave, symbol: "i-save", shortcut: "Ctrl+S"},
	{id: idHeaderThemeMode, label: uiSwitchToLight, symbol: themeToggleIcon},
	{id: idHeaderSettings, label: uiSettings, symbol: "i-nav_settings"},
	{id: idHeaderHelpIcon, label: uiHelp, symbol: "?"},
}

func (a *app) commandBarCaption(button commandBarButton) string {
	key := button.label
	if button.id == idHeaderPointer {
		if _, area := a.view.canvasRegionMode(); area || a.canvasTextCursor {
			key = uiSwitchToCanvasArrow
		}
	}
	if button.id == idHeaderLasso && a.commandBar.areaMode == textCrossing {
		key = uiCrossingSelection
	}
	if button.id == idHeaderFlow && a.view.continuous {
		key = uiSwitchToPages
	}
	if button.id == idHeaderThemeMode && a.uiTheme.Mode == "light" {
		key = uiSwitchToDark
	}
	if button.id == idHeaderCollapseDocks && a.docksTemporarilyCollapsed {
		key = uiRestoreBothDocks
	}
	caption := a.uiText(key)
	if button.shortcut != "" {
		caption += " (" + button.shortcut + ")"
	}
	return caption
}

func (a *app) updateThemeToggleCaption() {
	for _, control := range a.commandBar.controls {
		if control.id == idHeaderThemeMode {
			caption := a.commandBarCaption(control.commandBarButton)
			setText(control.handle, caption)
			a.setTooltip(control.handle, caption)
			return
		}
	}
}

func (a *app) localizeCommandBar() {
	a.localizeHeaderSearch()
	for _, control := range a.commandBar.controls {
		if control.id == idHeaderPageCount {
			continue
		}
		caption := a.commandBarCaption(control.commandBarButton)
		setText(control.handle, caption)
		if !control.menu {
			a.setTooltip(control.handle, caption)
		}
	}
	a.syncHeaderZoomLabel()
	a.syncHeaderPageNavigation()
}

func (a *app) buildCommandBar() {
	for _, button := range headerButtons {
		caption := a.commandBarCaption(button)
		var handle uintptr
		if button.id == idHeaderPageCount {
			// Read-only count, not a keyboard tab stop or a new navigation action.
			handle = a.add("BUTTON", "— / —", button.id, 0, 0, 1, 30, bsOwnerDraw, 0, a.normalFont)
			setEnabled(handle, false)
		} else {
			handle = a.button(caption, button.id, 0, 0, 1, 0x04000000)
			watchNavigationHover(handle)
		}
		if !button.menu && button.id != idHeaderPageCount {
			a.setTooltip(handle, caption)
		}
		a.commandBar.controls = append(a.commandBar.controls, commandBarControl{button, handle})
		switch button.id {
		case idHeaderUndo:
			a.historyButtons[0] = handle
		case idHeaderRedo:
			a.historyButtons[1] = handle
		case idHeaderForms:
			setEnabled(handle, false)
		}
	}
	a.commandBar.searchEdit = a.edit(0, 0, 1, false)
	sendMessage.Call(a.commandBar.searchEdit, 0x00c5, 256, 0) // EM_LIMITTEXT
	a.localizeHeaderSearch()
	setEnabled(a.commandBar.searchEdit, false)
}

func (a *app) layoutCommandBar() {
	d := a.scale
	x := d(8)
	menuWidth := 0
	var icons []int
	for _, control := range a.commandBar.controls {
		if control.id == idHeaderUndo || control.id == idHeaderRedo || control.id == idHeaderZoomPercent || isHeaderPageNavigation(control.id) {
			continue
		}
		if control.menu {
			menuWidth += a.captionWidth(a.uiText(control.label), a.normalFont) + d(20) + d(1)
		} else {
			icons = append(icons, control.id)
		}
	}
	// The HTML's optional order, omitting commands not implemented in the
	// native bar yet. The emergency order is used only if icons alone overflow.
	optional := []int{idHeaderHelpIcon, idHeaderSettings, idHeaderCollapseDocks, idHeaderFitWidth, idHeaderFitPage, idToggleLayoutGrid}
	emergency := []int{idCompareOpen, idHeaderOverview, idHeaderActualSize, idHeaderFlow,
		idHeaderThemeMode, idHeaderLasso, idHeaderRectangle, idReaderZoomOut, idReaderZoomIn,
		idSearchTab, idHeaderSave, idHeaderPointer}
	available := a.width-menuWidth-d(28)
	makePlan := func(space int) headerToolbarPlan {
		return planHeaderToolbar(space, icons, optional, emergency,
			d(30), d(4), d(27), d(2), d(78), d(112), d(145), d(76))
	}
	plan := makePlan(available)
	// Keep the navigation together, without evicting any existing icon.
	pageArrow, pageCount := d(27), d(90)
	pageCost := 2*pageArrow + pageCount + 3*d(4)
	showPages := false
	if available >= pageCost {
		candidate := makePlan(available-pageCost)
		if headerPageNavigationFits(plan, candidate, d(78)) {
			plan, showPages = candidate, true
			plan.used += 2*pageArrow + pageCount + 3*plan.gap
		}
	}
	iconStart := a.width - d(8) - plan.used
	showSearchEdit := false
	for _, control := range a.commandBar.controls {
		if control.id == idHeaderUndo || control.id == idHeaderRedo {
			bounds, show := a.headerHistoryBounds(control.id - idHeaderUndo)
			if show {
				place(control.handle, int(bounds.left), int(bounds.top), int(bounds.right-bounds.left), int(bounds.bottom-bounds.top))
				invalidate(control.handle)
			}
			showControl(control.handle, show)
			continue
		}
		if plan.hidden[control.id] || control.id == idHeaderZoomPercent && plan.zoom == 0 || isHeaderPageNavigation(control.id) && !showPages {
			showControl(control.handle, false)
			continue
		}
		width, gap := plan.iconWidth, plan.gap
		if control.id == idHeaderZoomPercent {
			width = plan.zoom
		} else if control.id == idHeaderPageCount {
			width = pageCount
		} else if isHeaderPageNavigation(control.id) {
			width = pageArrow
		}
		if control.menu {
			width, gap = a.captionWidth(a.uiText(control.label), a.normalFont)+d(20), d(1)
		} else {
			x = max(x, iconStart)
		}
		show := !a.full && x+width+d(8) <= a.width
		if show {
			place(control.handle, x, d(headerHeight+4), width, d(30))
			invalidate(control.handle)
		}
		showControl(control.handle, show)
		x += width + gap
		if control.id == idSearchTab && plan.search > 0 {
			showSearchEdit = show && x+plan.search+d(8) <= a.width
			if showSearchEdit {
				place(a.commandBar.searchEdit, x, d(headerHeight+5), plan.search, d(28))
				invalidate(a.commandBar.searchEdit)
			}
			x += plan.search + plan.gap
		}
	}
	showControl(a.commandBar.searchEdit, showSearchEdit)
}

func (a *app) updateCommandBar() {
	v := a.view
	ready := v.ready() && !a.busy
	fitChanged := a.commandBar.fitMode != v.fitMode || a.commandBar.overview != v.overview
	a.commandBar.fitMode = v.fitMode
	a.commandBar.overview = v.overview
	areaMode, selectingArea := v.canvasRegionMode()
	if !selectingArea {
		areaMode = 0
	}
	selectionChanged := a.commandBar.areaMode != areaMode
	a.commandBar.areaMode = areaMode
	canSelect := a.canUseHeaderSelection()
	a.syncHeaderPageNavigation()
	for _, control := range a.commandBar.controls {
		switch control.id {
		case idHeaderCollapseDocks:
			setEnabled(control.handle, !a.busy)
		case idHeaderSave:
			setEnabled(control.handle, a.canDocumentCommand(v, docSave, false))
		case idCompareOpen:
			setEnabled(control.handle, !a.busy && (a.comparisonVisible() || ready && len(a.tabs) > 1))
		case idHeaderFlow, idHeaderOverview, idReaderZoomOut, idReaderZoomIn, idHeaderFitPage, idHeaderFitWidth, idHeaderActualSize, idSearchTab, idToggleLayoutGrid, idHeaderZoomPercent:
			setEnabled(control.handle, ready)
		case idHeaderEdit:
			setEnabled(control.handle, ready)
		case idHeaderPointer, idHeaderRectangle, idHeaderLasso:
			setEnabled(control.handle, canSelect)
			if selectionChanged {
				invalidate(control.handle)
			}
		}
		if fitChanged && (control.id == idHeaderFitPage || control.id == idHeaderFitWidth || control.id == idHeaderActualSize || control.id == idHeaderOverview || control.id == idHeaderFlow) {
			invalidate(control.handle)
		}
		if control.id == idHeaderFlow || control.id == idHeaderPointer || control.id == idHeaderLasso {
			label := a.commandBarCaption(control.commandBarButton)
			if getText(control.handle) != label {
				setText(control.handle, label)
				a.setTooltip(control.handle, label)
				invalidate(control.handle)
			}
		}
	}
}

func (a *app) drawCommandBarButton(item *buttonDraw) bool {
	if item.controlType != 4 {
		return false
	}
	for _, control := range a.commandBar.controls {
		if control.handle != item.handle {
			continue
		}
		body, ink := titleColor, textColor
		hovered, disabled := a.navigationHot(item), item.state&4 != 0
		if control.id == idHeaderPageCount {
			fillColor(item.dc, item.bounds, titleColor)
			if !a.view.ready() {
				ink = mutedColor
			}
			drawText(item.dc, getText(control.handle), item.bounds, a.normalFont, ink, 1|0x8000)
			return true
		}
		if control.id == idHeaderZoomPercent {
			fillColor(item.dc, item.bounds, titleColor)
			body = inputColor
			if hovered && !disabled {
				body = uiColors["hover"]
			}
			if disabled {
				ink = disabledColor
			}
			rounded(item.dc, item.bounds, a.scale(5), body, divider)
			bounds := item.bounds
			a.arrow(item.dc, int(bounds.right)-a.scale(12), int(bounds.top+bounds.bottom)/2, ink)
			bounds.left += int32(a.scale(6))
			bounds.right -= int32(a.scale(22))
			drawText(item.dc, getText(control.handle), bounds, a.normalFont, ink, 0x8000)
			if item.state&16 != 0 {
				a.drawFocus(item.dc, item.bounds)
			}
			return true
		}
		active := control.id == a.commandBar.openMenu ||
			control.id == idHeaderFile && a.fileMenu.open && a.fileMenu.anchor == control.handle
		symbol := control.symbol
		switch control.id {
		case idHeaderPointer:
			if a.commandBar.areaMode == 0 && a.canvasTextCursor {
				symbol = "i-text_caret"
			}
			active = a.commandBar.areaMode == 0
		case idHeaderRectangle:
			active = a.commandBar.areaMode == textRectangle
		case idHeaderLasso:
			active = polygonSelectionMode(a.commandBar.areaMode)
			if a.commandBar.areaMode == textCrossing {
				symbol = "i-crossing_selection"
			}
		case idHeaderFlow:
			if a.view.continuous {
				symbol = "i-view_continuous"
			}
			active = !a.view.overview
		case idHeaderOverview:
			active = a.view.overview
		case idCompareOpen:
			active = a.comparisonVisible()
		case idHeaderCollapseDocks:
			active = a.docksTemporarilyCollapsed
		case idHeaderFitPage:
			active = !a.view.overview && a.view.fitMode == readingZoomFitPage
		case idHeaderFitWidth:
			active = !a.view.overview && a.view.fitMode == readingZoomFitWidth
		case idHeaderActualSize:
			active = !a.view.overview && a.view.fitMode == readingZoomActual
		case idToggleLayoutGrid:
			active = a.checked(a.layoutAids.grid)
		case idSearchTab:
			active = a.mode == idSearchTab && a.settingsWidth > 0
		}
		if disabled {
			ink, hovered, active = disabledColor, false, false
		}
		activeIcon := active && !control.menu
		switch {
		case activeIcon && hovered:
			body = uiColors["toolbarActiveHover"]
		case activeIcon:
			body = uiColors["toolbarActive"]
		case active || hovered:
			body = uiColors["hover"]
		}
		if activeIcon {
			fillColor(item.dc, item.bounds, titleColor)
			bounds := item.bounds
			rounded(item.dc, bounds, a.scale(5), body, uiColors["accentStrong"])
			inset, thickness := int32(a.scale(3)), int32(a.scale(2))
			bottom := bounds.bottom - int32(a.scale(1))
			fillColor(item.dc, rect{bounds.left + inset, bottom - thickness, bounds.right - inset, bottom}, selectionColor)
		} else {
			fillColor(item.dc, item.bounds, body)
		}
		if control.menu || symbol == "?" || isHeaderPageNavigation(control.id) {
			caption, font := a.uiText(control.label), a.normalFont
			if symbol == "?" || isHeaderPageNavigation(control.id) {
				caption, font = symbol, a.titleFont
			}
			drawText(item.dc, caption, item.bounds, font, ink, 1)
		} else {
			size := a.scale(18)
			x, y := int(item.bounds.left+item.bounds.right)/2-size/2, int(item.bounds.top+item.bounds.bottom)/2-size/2
			drawUIIcon(item.dc, symbol, x, y, size, ink, body)
		}
		if item.state&16 != 0 {
			focus := item.bounds
			if activeIcon {
				focus.bottom -= int32(a.scale(2))
			}
			a.drawFocus(item.dc, focus)
		}
		return true
	}
	return false
}

type headerMenuItem struct {
	popupItem
	run func()
}

func (a *app) headerMenuCommand(id int, label string, enabled, checked bool) headerMenuItem {
	return headerMenuItem{popupItem{label: label, disabled: !enabled, checked: checked}, func() { a.command(id, 0) }}
}

func (a *app) showHeaderMenu(id int) {
	v := a.view
	ready := v.ready() && !a.busy
	var entries []headerMenuItem
	run := a.headerMenuCommand
	switch id {
	case idHeaderView:
		entries = []headerMenuItem{
			{popupItem{label: a.uiText(uiPageByPage), checked: !v.overview && !v.continuous, disabled: !ready}, func() { a.setHeaderFlow(false) }},
			{popupItem{label: a.uiText(uiContinuous), checked: !v.overview && v.continuous, disabled: !ready}, func() { a.setHeaderFlow(true) }},
			run(idHeaderFitPage, a.uiText(uiFitPage), ready, ready && !v.overview && v.fitMode == readingZoomFitPage),
			run(idHeaderFitWidth, a.uiText(uiFitWidth), ready, ready && !v.overview && v.fitMode == readingZoomFitWidth),
			run(idHeaderActualSize, a.uiText(uiActualSize), ready, ready && !v.overview && v.fitMode == readingZoomActual),
			run(idHeaderOverview, a.uiText(uiPageOverview), ready, v.overview),
			run(idFullscreen, a.uiText(uiFullscreen), !a.busy, false),
			run(idHeaderCollapseDocks, a.dockCollapseCaption(), !a.busy, a.docksTemporarilyCollapsed),
			run(idCompareOpen, getText(a.comparison.button), !a.busy && (a.comparisonVisible() || ready && len(a.tabs) > 1), a.comparisonVisible()),
			run(idLayoutAids, a.uiText(uiLayoutAids), ready, a.mode == idLayoutAids),
			run(idToggleLayoutGrid, a.uiText(uiShowLayoutGrid), ready, a.checked(a.layoutAids.grid)),
		}
		// Use direct preset actions when the dropdown is hidden, so the popup
		// is never anchored to a hidden native control.
		if !a.headerVisible(idHeaderZoomPercent) {
			for _, percent := range headerZoomPresets {
				entries = append(entries, headerMenuItem{
					popupItem{label: fmt.Sprintf("%d %%", percent), disabled: !ready,
						checked: ready && !v.overview && fmt.Sprintf("%.0f", v.zoomPercent()) == fmt.Sprint(percent)},
					func() { v.setZoomPercent(percent) },
				})
			}
		}
		for _, id := range []int{idReaderZoomOut, idReaderZoomIn} {
			if !a.headerVisible(id) {
				key := uiZoomOut
				if id == idReaderZoomIn {
					key = uiZoomIn
				}
				entries = append(entries, run(id, a.uiText(key), ready, false))
			}
		}
	case idHeaderTools:
		for _, tab := range workspaceTabs {
			if tab.id != idFileTab && tab.id != idBrowseTab {
				entries = append(entries, run(tab.id, a.uiText(tab.name), !a.busy && (!tab.needsPDF || ready), tab.id == a.tool))
			}
		}
		entries = append(entries, run(idToggleLayers, a.uiText(uiLayersAndObjects), ready, a.rightAux.panel == auxiliaryLayers))
		// The optional right-hand toolbar commands must remain available when
		// the window is too narrow to show their buttons.
		for _, control := range a.commandBar.controls {
			switch control.id {
			case idHeaderThemeMode, idHeaderSettings:
				if !a.headerVisible(control.id) {
					entries = append(entries, run(control.id, a.commandBarCaption(control.commandBarButton), true, false))
				}
			case idHeaderPointer, idHeaderRectangle, idHeaderLasso:
				if !a.headerVisible(control.id) {
					entries = append(entries, run(control.id, a.commandBarCaption(control.commandBarButton), a.canUseHeaderSelection(), false))
				}
			}
		}
	case idHeaderEdit:
		editing := a.inlineCurrent() != nil
		for _, action := range []documentCommand{docUndo, docRedo, docCut, docCopy, docPaste, docDelete, docSelectAll} {
			item := documentActions[action]
			entries = append(entries, headerMenuItem{
				popupItem{label: a.documentActionLabel(action), shortcut: item.shortcut, disabled: !a.canDocumentCommand(v, action, editing)},
				func() { a.doDocumentCommand(v, action, editing) },
			})
		}
	}
	items := make([]popupItem, len(entries))
	for index, entry := range entries {
		items[index] = entry.popupItem
	}
	anchor := a.headerControl(id)
	a.commandBar.openMenu = id
	invalidate(anchor)
	choice := a.trackPopup(anchor, 286, items)
	a.commandBar.openMenu = 0
	invalidate(anchor)
	if choice >= 0 {
		entries[choice].run()
	}
}

func (a *app) headerVisible(id int) bool {
	handle := a.headerControl(id)
	visible, _, _ := user32.NewProc("IsWindowVisible").Call(handle)
	return visible != 0
}

func (a *app) headerControl(id int) uintptr {
	for _, control := range a.commandBar.controls {
		if control.id == id {
			return control.handle
		}
	}
	return 0
}

func (a *app) setHeaderFlow(continuous bool) {
	if !a.view.ready() || a.busy {
		return
	}
	a.view.setPageOverview(false)
	if a.view.continuous == continuous {
		return
	}
	index := 1
	if continuous {
		index = 0
	}
	chooseCombo(a.view.flow, index)
	a.command(idReaderFlow, 1)
}

func (a *app) commandBarCommand(id, notification int) bool {
	if notification != 0 {
		return false
	}
	switch id {
	case idHeaderFile:
		a.showFileMenuAt(a.headerControl(id))
	case idHeaderSettings:
		a.showInterfaceSettings()
	case idHeaderThemeMode:
		a.toggleThemeMode()
	case idHeaderView, idHeaderTools, idHeaderEdit:
		a.showHeaderMenu(id)
	case idHeaderHelp, idHeaderHelpIcon:
		a.togglePropertyHelp(a.headerControl(id), nil)
	case idHeaderPointer, idHeaderRectangle, idHeaderLasso:
		a.setHeaderSelection(id)
	case idHeaderFlow:
		a.setHeaderFlow(!a.view.continuous)
	case idHeaderOverview:
		a.togglePageOverview()
	case idHeaderFitPage:
		a.view.setZoomMode(readingZoomFitPage)
	case idHeaderFitWidth:
		a.view.setZoomMode(readingZoomFitWidth)
	case idHeaderActualSize:
		a.view.setZoomMode(readingZoomActual)
	case idHeaderZoomPercent:
		a.showHeaderZoom()
	case idHeaderCollapseDocks:
		a.toggleTemporaryDocks()
	case idHeaderPreviousPage, idHeaderNextPage:
		if a.busy || !a.view.ready() {
			return true
		}
		if id == idHeaderPreviousPage {
			a.command(idReaderPrevious, 0)
		} else {
			a.command(idReaderNext, 0)
		}
	case idHeaderPageCount:
		return true
	case idHeaderSave:
		a.doDocumentCommand(a.view, docSave, false)
	default:
		return false
	}
	return true
}

func (a *app) canUseHeaderSelection() bool {
	if a.full || a.busy || !a.documentTools || !a.view.ready() {
		return false
	}
	v := a.focusedDocumentView()
	return v.ready() && !v.overview
}

func (a *app) setHeaderSelection(id int) {
	if !a.canUseHeaderSelection() || a.saveInlineBefore(func() { a.setHeaderSelection(id) }) {
		return
	}
	setFocus.Call(a.window)
	if id == idHeaderPointer {
		_, area := a.view.canvasRegionMode()
		if !area && !a.canvasTextCursor {
			a.activate(idTextTab)
			a.textCommand(idTextSelectObject, 0)
		} else if p := &a.textTool; a.mode == idTextTab && (area || p.flow.text != "" || a.view.mixed.pick.source != 0) {
			// Keep the selection in the Text panel when switching to the arrow.
			a.cancelCanvasSelections()
			a.cancelCanvasSelectionProbes()
			a.canvasTextCursor = false
			p.mode, p.explicitParagraph = textObject, false
			a.syncTextSelectionModes()
		} else {
			a.activate(idObjectsTab)
			a.objectCommand(idObjectPoint, 0)
		}
		return
	}
	if a.mode == idTextTab {
		a.textCommand(idTextSelectRectangle+id-idHeaderRectangle, 0)
		return
	}
	if a.mode == idOCRTab {
		a.cancelCanvasSelections()
		a.cancelCanvasSelectionProbes()
		mode := textRectangle + id - idHeaderRectangle
		if mode == textLasso && a.ocr.areaMode == textLasso {
			mode = textCrossing
		}
		a.ocr.areaMode = mode
		a.updateCommandBar()
		return
	}
	a.activate(idObjectsTab)
	a.objectCommand(idObjectRectangle+id-idHeaderRectangle, 0)
}

// Use the reader's displayed value, including fit modes and overview scale.
func (a *app) syncHeaderZoomLabel() {
	handle := a.headerControl(idHeaderZoomPercent)
	if handle == 0 || a.view == nil {
		return
	}
	label := "—"
	if a.view.ready() {
		label = fmt.Sprintf("%.0f %%", a.view.displayedZoomPercent())
	}
	if getText(handle) != label {
		setText(handle, label)
		invalidate(handle)
	}
}

func (a *app) showHeaderZoom() {
	if a.busy || !a.view.ready() {
		return
	}
	presets := headerZoomPresets
	items := make([]popupItem, len(presets))
	for i, percent := range presets {
		items[i] = popupItem{label: fmt.Sprintf("%d %%", percent),
			checked: !a.view.overview && fmt.Sprintf("%.0f", a.view.zoomPercent()) == fmt.Sprint(percent)}
	}
	tab := a.activeTab
	index := a.trackPopup(a.headerControl(idHeaderZoomPercent), 100, items)
	if index >= 0 && index < len(presets) && a.activeTab == tab {
		a.view.setZoomPercent(presets[index])
	}
}

func isHeaderPageNavigation(id int) bool {
	return id == idHeaderPreviousPage || id == idHeaderPageCount || id == idHeaderNextPage
}

func (a *app) syncHeaderPageNavigation() {
	if a.view == nil {
		return
	}
	v := a.view
	count, previous, next := headerPageNavigationState(v.ready(), a.busy, v.page, len(v.order))
	if handle := a.headerControl(idHeaderPageCount); handle != 0 && getText(handle) != count {
		setText(handle, count)
		invalidate(handle)
	}
	if handle := a.headerControl(idHeaderPreviousPage); handle != 0 {
		setEnabled(handle, previous)
	}
	if handle := a.headerControl(idHeaderNextPage); handle != 0 {
		setEnabled(handle, next)
	}
}
