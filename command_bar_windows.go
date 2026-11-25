package main

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
)

type commandBarButton struct {
	id            int
	label, symbol string
	menu          bool
}

type commandBarControl struct {
	commandBarButton
	handle uintptr
}

type commandBar struct {
	controls          []commandBarControl
	openMenu, fitMode int
}

var headerButtons = []commandBarButton{
	{id: idHeaderFile, label: "Soubor", menu: true},
	{id: idHeaderView, label: "Zobrazení", menu: true},
	{id: idHeaderTools, label: "Nástroje", menu: true},
	{id: idHeaderEdit, label: "Úpravy", menu: true},
	{id: idHeaderForms, label: "Formuláře", menu: true},
	{id: idHeaderHelp, label: "Pomoc", menu: true},
	{id: idHeaderUndo, label: "Zpět (Ctrl+Z)", symbol: "i-undo"},
	{id: idHeaderRedo, label: "Znovu (Ctrl+Y)", symbol: "i-redo"},
	{id: idHeaderFlow, label: "Přepnout na souvislé zobrazení", symbol: "i-view_continuous"},
	{id: idReaderZoomOut, label: "Oddálit", symbol: "i-zoom_out"},
	{id: idReaderZoomIn, label: "Přiblížit", symbol: "i-zoom_in"},
	{id: idHeaderFitPage, label: "Přizpůsobit stránce", symbol: "i-fit_page"},
	{id: idHeaderFitWidth, label: "Přizpůsobit šířce", symbol: "i-fit_width"},
	{id: idSearchTab, label: "Najít v PDF (Ctrl+F)", symbol: "i-search"},
	{id: idCompareOpen, label: "Zobrazení vedle sebe", symbol: "i-side_by_side"},
	{id: idHeaderSave, label: "Uložit (Ctrl+S)", symbol: "i-save"},
	{id: idHeaderHelpIcon, label: "Nápověda", symbol: "?"},
}

func (a *app) buildCommandBar() {
	for _, button := range headerButtons {
		handle := a.button(button.label, button.id, 0, 0, 1, 0x04000000)
		watchNavigationHover(handle)
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
}

func (a *app) layoutCommandBar() {
	d := a.scale
	x := d(8)
	for _, control := range a.commandBar.controls {
		width, gap := d(30), d(4)
		if control.menu {
			width, gap = a.captionWidth(control.label, a.normalFont)+d(20), d(1)
		} else if control.id == idHeaderUndo {
			x += d(12)
		}
		show := !a.full && x+width+d(8) <= a.width
		if show {
			place(control.handle, x, d(headerHeight+4), width, d(30))
			invalidate(control.handle)
		}
		showControl(control.handle, show)
		x += width + gap
	}
}

func (a *app) updateCommandBar() {
	v := a.view
	ready := v.ready() && !a.busy
	fitChanged := a.commandBar.fitMode != v.fitMode
	a.commandBar.fitMode = v.fitMode
	for _, control := range a.commandBar.controls {
		switch control.id {
		case idHeaderSave:
			setEnabled(control.handle, a.canDocumentCommand(v, docSave, false))
		case idCompareOpen:
			setEnabled(control.handle, !a.busy && (a.comparisonVisible() || ready && len(a.tabs) > 1))
		case idHeaderFlow, idReaderZoomOut, idReaderZoomIn, idHeaderFitPage, idHeaderFitWidth, idSearchTab:
			setEnabled(control.handle, ready)
		case idHeaderEdit:
			setEnabled(control.handle, ready)
		}
		if fitChanged && (control.id == idHeaderFitPage || control.id == idHeaderFitWidth) {
			invalidate(control.handle)
		}
		if control.id == idHeaderFlow {
			label := control.label
			if v.continuous {
				label = "Přepnout na zobrazení po stránkách"
			}
			if getText(control.handle) != label {
				setText(control.handle, label)
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
		active := control.id == a.commandBar.openMenu ||
			control.id == idHeaderFile && a.fileMenu.open && a.fileMenu.anchor == control.handle
		symbol := control.symbol
		switch control.id {
		case idHeaderFlow:
			if a.view.continuous {
				symbol = "i-view_pages"
				active = true
			}
		case idCompareOpen:
			active = a.comparisonVisible()
		case idHeaderFitPage:
			active = a.view.fitMode == 1
		case idHeaderFitWidth:
			active = a.view.fitMode == 2
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
		if control.menu || symbol == "?" {
			caption, font := control.label, a.normalFont
			if symbol == "?" {
				caption, font = "?", a.titleFont
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
			{popupItem{label: "Po stránkách", checked: !v.continuous, disabled: !ready}, func() { a.setHeaderFlow(false) }},
			{popupItem{label: "Souvisle", checked: v.continuous, disabled: !ready}, func() { a.setHeaderFlow(true) }},
			run(idHeaderFitPage, "Přizpůsobit stránce", ready, v.ready() && v.fitMode == 1),
			run(idHeaderFitWidth, "Přizpůsobit šířce", ready, v.ready() && v.fitMode == 2),
			run(idFullscreen, "Celá obrazovka", !a.busy, false),
			run(idCompareOpen, getText(a.comparison.button), !a.busy && (a.comparisonVisible() || ready && len(a.tabs) > 1), a.comparisonVisible()),
		}
	case idHeaderTools:
		for _, tab := range workspaceTabs {
			if tab.id != idFileTab && tab.id != idBrowseTab {
				entries = append(entries, run(tab.id, tab.name, !a.busy && (!tab.needsPDF || ready), tab.id == a.tool))
			}
		}
	case idHeaderEdit:
		editing := a.inlineCurrent() != nil
		for _, action := range []documentCommand{docUndo, docRedo, docCut, docCopy, docPaste, docDelete, docSelectAll} {
			item := documentActions[action]
			entries = append(entries, headerMenuItem{
				popupItem{label: item.label, shortcut: item.shortcut, disabled: !a.canDocumentCommand(v, action, editing)},
				func() { a.doDocumentCommand(v, action, editing) },
			})
		}
	case idHeaderHelp, idHeaderHelpIcon:
		entries = []headerMenuItem{{popupItem{label: "Klávesové zkratky"}, func() {
			notice(a.window, "Ctrl+O   Otevřít PDF\nCtrl+F   Hledat v PDF\nCtrl+S   Uložit\nCtrl+Z / Ctrl+Y   Zpět / Znovu\nF11   Celá obrazovka\nCtrl+kolečko   Přiblížení dokumentu nebo miniatur", 0x40)
		}}}
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

func (a *app) headerControl(id int) uintptr {
	for _, control := range a.commandBar.controls {
		if control.id == id {
			return control.handle
		}
	}
	return 0
}

func (a *app) setHeaderFlow(continuous bool) {
	if !a.view.ready() || a.busy || a.view.continuous == continuous {
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
	case idHeaderView, idHeaderTools, idHeaderEdit, idHeaderHelp, idHeaderHelpIcon:
		a.showHeaderMenu(id)
	case idHeaderFlow:
		a.setHeaderFlow(!a.view.continuous)
	case idHeaderFitPage, idHeaderFitWidth:
		if a.view.ready() && !a.busy {
			mode := 0
			if id == idHeaderFitWidth {
				mode = 1
			}
			chooseCombo(a.view.fit, mode)
			a.command(idReaderFit, 1)
		}
	case idHeaderSave:
		a.doDocumentCommand(a.view, docSave, false)
	default:
		return false
	}
	return true
}
