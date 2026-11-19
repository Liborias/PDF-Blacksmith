package main

import (
	"os"
	"unsafe"
)

const idFileMenuFirst = 410

type fileMenu struct {
	handle  uintptr
	buttons []uintptr
	open    bool
}

var fileOptions = []struct {
	id    int
	label string
}{
	{idOpen, "Otevřít soubor"},
	{idRecentFiles, "Nedávné"},
	{idShowOutput, "Otevřít výstup"},
	{idSaveDocument, "Uložit"},
	{idSaveDocumentAs, "Uložit jako"},
	{idExportTab, "Export"},
	{idFileSign, "Podepsaná kopie"},
	{idPrintDocument, "Tisk"},
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
		m.buttons = append(m.buttons, a.button(option.label, idFileMenuFirst+index, 0, 0, 1, 0))
	}
	a.controlParent = 0
	return true
}

func (a *app) showFileMenu() {
	if a.saveInlineBefore(a.showFileMenu) {
		return
	}
	m := &a.fileMenu
	if m.open {
		a.hideFileMenu(true)
		return
	}
	a.activate(idFileTab)
	if a.full || a.toolsCollapsed {
		return
	}
	d := a.scale
	width, height := a.toolsWidth, d(len(m.buttons)*31+8)
	for index, option := range fileOptions {
		width = max(width, a.captionWidth(option.label, a.normalFont)+d(42))
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
	m.open = true
	top := int(a.viewport.top) + d(10) + d(34)
	user32.NewProc("SetWindowPos").Call(m.handle, 0, 0, uintptr(top), uintptr(width), uintptr(height), 0x0040) // SWP_SHOWWINDOW
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
	if restoreFocus {
		setFocus.Call(a.toolButtons[idFileTab])
	}
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
		if msg.window != m.handle && parent != m.handle && msg.window != a.toolButtons[idFileTab] {
			a.hideFileMenu(false)
		}
	case 0x0100: // WM_KEYDOWN
		switch msg.wParam {
		case 0x1b:
			a.hideFileMenu(true)
			return true
		case 9:
			a.hideFileMenu(true)
			msg.window = a.toolButtons[idFileTab]
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
	a.paintPopupEntry(item.dc, item.bounds, popupItem{
		label: option.label, checked: option.id == a.activeFileOption(), disabled: !a.fileOptionEnabled(option.id),
	}, item.state&(1|16|64) != 0)
	return true
}

func (a *app) updateFileOptions() {
	for _, option := range fileOptions {
		setEnabled(a.detailButtons[option.id], a.fileOptionEnabled(option.id))
	}
	if a.detailsTab() == idRecentFiles {
		a.updateRecentActions()
	}
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
	case idRecentFiles, idExportTab, idFileSign:
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
