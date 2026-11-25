package main

import (
	"os"
	"unsafe"
)

const idFileMenuFirst = 410

type fileMenu struct {
	handle  uintptr
	anchor  uintptr
	buttons []uintptr
	open    bool
}

type fileOption struct {
	id    int
	label uiMessage
}

var fileOptions = []fileOption{
	{idOpen, uiOpenFile},
	{idRecentFiles, uiRecent},
	{idShowOutput, uiOpenOutput},
	{idSaveDocument, uiSave},
	{idSaveDocumentAs, uiSaveAs},
	{idExportTab, uiExport},
	{idFileSign, uiSignedCopy},
	{idPrintDocument, uiPrint},
}

func (a *app) fileOptionEnabled(option int) bool {
	if a.busy {
		return false
	}
	switch option {
	case idOpen, idRecentFiles:
		return true
	case idShowOutput:
		return a.lastOutput != ""
	case idSaveDocument:
		return a.view.ready() && a.view.changed()
	default:
		return a.view.ready()
	}
}

func (a *app) activeFileOption() int {
	if a.mode == idSignTab {
		return idFileSign
	}
	return a.mode
}

func (a *app) buildFileMenu(className *uint16) bool {
	m := &a.fileMenu
	m.handle, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	if m.handle == 0 {
		return false
	}
	a.controlParent = m.handle
	for index, option := range fileOptions {
		handle := a.button("", idFileMenuFirst+index, 0, 0, 1, 0)
		a.bindUILabel(handle, option.label)
		watchNavigationHover(handle)
		m.buttons = append(m.buttons, handle)
	}
	a.controlParent = 0
	return true
}

func (a *app) showFileMenu() {
	a.showFileMenuAt(a.toolButtons[idFileTab])
}

func (a *app) showFileMenuAt(anchor uintptr) {
	if anchor == 0 || a.saveInlineBefore(func() { a.showFileMenuAt(anchor) }) {
		return
	}
	m := &a.fileMenu
	if m.open {
		same := anchor == m.anchor
		a.hideFileMenu(same)
		if same {
			return
		}
	}
	a.activate(idFileTab)
	if a.full || a.toolsWidth == 0 && anchor == a.toolButtons[idFileTab] {
		return
	}
	d := a.scale
	width, height := max(a.toolsWidth, d(168)), d(len(m.buttons)*31+8)
	for index, option := range fileOptions {
		width = max(width, a.captionWidth(a.uiText(option.label), a.normalFont)+d(42))
		setEnabled(m.buttons[index], a.fileOptionEnabled(option.id))
	}
	width = min(width, a.width)
	for index, button := range m.buttons {
		place(button, 1, d(index*31)+1, width-2, d((index+1)*31)-d(index*31))
	}
	radius := max(2, d(9))
	region, _, _ := gdi32.NewProc("CreateRoundRectRgn").Call(0, uintptr(-radius), uintptr(width+1), uintptr(height+1), uintptr(radius), uintptr(radius))
	if region != 0 {
		if ok, _, _ := user32.NewProc("SetWindowRgn").Call(m.handle, region, 0); ok == 0 {
			gdiDelete.Call(region)
		}
	}
	var anchorBounds rect
	user32.NewProc("GetWindowRect").Call(anchor, uintptr(unsafe.Pointer(&anchorBounds)))
	pos := point{anchorBounds.left, anchorBounds.bottom}
	user32.NewProc("ScreenToClient").Call(a.window, uintptr(unsafe.Pointer(&pos)))
	left := 0
	if anchor != a.toolButtons[idFileTab] {
		left = max(0, min(int(pos.x), a.width-width))
	}
	m.open, m.anchor = true, anchor
	user32.NewProc("SetWindowPos").Call(m.handle, 0, uintptr(left), uintptr(pos.y), uintptr(width), uintptr(height), 0x0040) // SWP_SHOWWINDOW
	invalidate(anchor)
	a.focusFileMenu(-1, 1)
	user32.NewProc("RedrawWindow").Call(m.handle, 0, 0, 0x0085) // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN
}

func (a *app) hideFileMenu(restoreFocus bool) {
	m := &a.fileMenu
	if !m.open {
		return
	}
	m.open = false
	showControl(m.handle, false)
	invalidate(m.anchor)
	if restoreFocus {
		setFocus.Call(m.anchor)
	}
	m.anchor = 0
}

func (a *app) focusFileMenu(from, step int) {
	buttons := a.fileMenu.buttons
	for i := 1; i <= len(buttons); i++ {
		index := (from + i*step + len(buttons)) % len(buttons)
		if a.fileOptionEnabled(fileOptions[index].id) {
			setFocus.Call(buttons[index])
			return
		}
	}
}

func (a *app) fileMenuInput(msg *winMessage) bool {
	m := &a.fileMenu
	if !m.open {
		return false
	}
	switch msg.message {
	case 0x0201, 0x0204, 0x0207, 0x00a1: // Mouse down, including the title bar.
		parent, _, _ := user32.NewProc("GetParent").Call(msg.window)
		if msg.window != m.handle && parent != m.handle && msg.window != m.anchor {
			a.hideFileMenu(false)
		}
	case 0x0100: // WM_KEYDOWN
		switch msg.wParam {
		case 0x1b:
			a.hideFileMenu(true)
			return true
		case 9:
			anchor := m.anchor
			a.hideFileMenu(true)
			msg.window = anchor
		case 0x0d, 0x26, 0x28: // Enter, up and down.
			for index, button := range m.buttons {
				if msg.window == button {
					if msg.wParam == 0x0d {
						a.command(idFileMenuFirst+index, 0)
						return true
					}
					step := 1
					if msg.wParam == 0x26 {
						step = -1
					}
					a.focusFileMenu(index, step)
					return true
				}
			}
		}
	}
	return false
}

func (a *app) drawFileMenuButton(item *buttonDraw) bool {
	index := int(item.controlID) - idFileMenuFirst
	if item.controlType != 4 || index < 0 || index >= len(a.fileMenu.buttons) || item.handle != a.fileMenu.buttons[index] {
		return false
	}
	option := fileOptions[index]
	selected := option.id == a.activeFileOption() || (option.id == idRecentFiles && a.mode == idFileDetails)
	a.paintPopupEntry(item.dc, item.bounds, popupItem{
		label: a.uiText(option.label), checked: selected, disabled: !a.fileOptionEnabled(option.id),
	}, a.navigationHot(item))
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (a *app) runFileOption(option int) {
	if a.tool != idFileTab || !a.fileOptionEnabled(option) {
		return
	}
	switch option {
	case idOpen:
		a.openDialog()
	case idShowOutput:
		a.openDocument(a.lastOutput, a.lastPassword)
	case idSaveDocument, idSaveDocumentAs:
		a.saveDocument(option == idSaveDocumentAs)
	case idPrintDocument:
		a.printDocument()
	case idRecentFiles:
		a.mode = idFileDetails
		a.rightDock.revealPrimary()
		a.layout()
		if a.recent.visibleCount() > 0 {
			setFocus.Call(a.recent.list)
			a.revealFocusedControl()
		}
	case idExportTab, idFileSign:
		a.activate(option)
	}
}

func (a *app) saveDocument(saveAs bool) {
	if a.saveInlineBefore(func() { a.saveDocument(saveAs) }) {
		return
	}
	v := a.view
	if a.busy || !v.ready() {
		return
	}
	req := v.request("save_document")
	changed := v.changed()
	if v.hasPageEdits() {
		v.includeEdits(&req)
	}
	if saveAs {
		suffix := "-kopie"
		if changed {
			suffix = "-upraveno"
		}
		if !a.chooseOutput(&req, suffix) {
			return
		}
	} else {
		if !changed || notice(a.window, "Uložit změny a přepsat původní PDF?\n\n"+v.path, 0x134) != 6 {
			return
		}
		a.saveOriginal()
		return
	}
	a.run(req)
}

func (a *app) saveOriginal() bool {
	v := a.view
	if !unchanged(v.path, v.original) {
		notice(a.window, "Původní PDF se mezitím změnilo. Změny můžete uložit do nového souboru, pokud je pracovní PDF dostupné.", 0x30)
		return false
	}
	req := v.request("save_document")
	if v.hasPageEdits() {
		v.includeEdits(&req)
	}
	req.Output, req.Overwrite, req.ReplaceSource = v.path, true, true
	req.ReplaceDocument = v.path
	req.SourceSize, req.SourceMtime = v.original.Size(), v.original.ModTime().UnixNano()
	return a.run(req)
}

func (a *app) reopenSaved(req request) bool {
	info, err := os.Stat(req.Output)
	if err != nil {
		return false
	}
	v := a.view
	page, continuous, fit, zoom := v.page, v.continuous, v.fitMode, v.zoom
	a.loadDocument(req.Output, req.PDFPassword, info)
	v.page, v.reference = page, page
	v.continuous, v.fitMode, v.zoom = continuous, fit, zoom
	v.syncReadingControls()
	return true
}
