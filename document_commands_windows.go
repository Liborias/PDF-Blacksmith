package main

import (
	"syscall"
	"unsafe"
)

type documentCommand int

const (
	docUndo documentCommand = iota
	docRedo
	docCut
	docCopy
	docPaste
	docDelete
	docSelectAll
	docApplyText
	docBold
	docItalic
	docUnderline
	docAlignLeft
	docAlignCenter
	docAlignRight
	docAlignJustify
	docFind
	docZoomIn
	docZoomOut
	docFitPage
	docPrevious
	docNext
	docFullscreen
	docSave
	docPrint
	docDuplicate
)

type documentAction struct {
	label, shortcut string
}

var documentActions = map[documentCommand]documentAction{
	docUndo: {"Zpět", "Ctrl+Z"}, docRedo: {"Znovu", "Ctrl+Y"},
	docCut: {"Vyjmout", "Ctrl+X"}, docCopy: {"Kopírovat", "Ctrl+C"},
	docPaste: {"Vložit", "Ctrl+V"}, docDelete: {"Odstranit", "Delete"},
	docSelectAll: {"Vybrat vše", "Ctrl+A"}, docApplyText: {"Použít úpravu", "Ctrl+Enter"},
	docBold: {"Tučné", "Ctrl+B"}, docItalic: {"Kurzíva", "Ctrl+I"}, docUnderline: {"Podtržení", "Ctrl+U"},
	docAlignLeft: {"Zarovnat vlevo", "Ctrl+L"}, docAlignCenter: {"Zarovnat na střed", "Ctrl+E"},
	docAlignRight: {"Zarovnat vpravo", "Ctrl+R"}, docAlignJustify: {"Zarovnat do bloku", "Ctrl+J"},
	docFind: {"Najít v PDF", "Ctrl+F"}, docZoomIn: {"Přiblížit", "Ctrl++"}, docZoomOut: {"Oddálit", "Ctrl+−"},
	docFitPage: {"Přizpůsobit stránce", "Ctrl+0"}, docPrevious: {"Předchozí strana", "Page Up"},
	docNext: {"Další strana", "Page Down"}, docFullscreen: {"Celá obrazovka", "F11"},
	docSave: {"Uložit", "Ctrl+S"}, docPrint: {"Tisk", "Ctrl+P"},
	docDuplicate: {"Duplikovat", "Ctrl+D"},
}

var editMenuMessages = map[documentCommand]uiMessage{
	docUndo: uiUndo, docRedo: uiRedo, docCut: uiCut, docCopy: uiCopy,
	docPaste: uiPaste, docDelete: uiDelete, docSelectAll: uiSelectAll,
}

func (a *app) documentActionLabel(command documentCommand) string {
	if key, exists := editMenuMessages[command]; exists {
		return a.uiText(key)
	}
	return documentActions[command].label
}

type documentShortcut struct {
	key   uintptr
	shift bool
}

var inlineShortcuts = map[documentShortcut]documentCommand{
	{'Z', false}: docUndo, {'Z', true}: docRedo, {'Y', false}: docRedo,
	{'X', false}: docCut, {'C', false}: docCopy, {'V', false}: docPaste, {'A', false}: docSelectAll,
	{0x0d, false}: docApplyText, {'B', false}: docBold, {'I', false}: docItalic, {'U', false}: docUnderline,
	{'L', false}: docAlignLeft, {'E', false}: docAlignCenter, {'R', false}: docAlignRight, {'J', false}: docAlignJustify,
}

var viewShortcuts = map[documentShortcut]documentCommand{
	{'Z', false}: docUndo, {'Y', false}: docRedo, {'Z', true}: docRedo, {'C', false}: docCopy, {'A', false}: docSelectAll, {0x0d, false}: docApplyText,
	{0xbb, false}: docZoomIn, {0xbb, true}: docZoomIn, {0x6b, false}: docZoomIn,
	{0xbd, false}: docZoomOut, {0x6d, false}: docZoomOut, {'0', false}: docFitPage, {0x60, false}: docFitPage,
}

// Only page-focused keys travel through this table. Sidebar text fields and
// other native controls keep their standard keyboard behavior.
func (a *app) documentKey(msg *winMessage) bool {
	e := &a.textTool.inline
	inline := e.visible && e.handle != 0 && msg.window == e.handle
	page := msg.window == a.window || a.full
	if !inline && !page {
		return false
	}
	ctrl, _, _ := user32.NewProc("GetKeyState").Call(0x11)
	shift, _, _ := user32.NewProc("GetKeyState").Call(0x10)
	alt, _, _ := user32.NewProc("GetKeyState").Call(0x12)
	if alt&0x8000 != 0 {
		return false
	}
	hasShift := shift&0x8000 != 0
	if msg.wParam == 0x5d || msg.wParam == 0x79 && hasShift {
		if inline {
			a.trackDocumentContext(a.view, point{e.clip.left, e.clip.bottom}, true)
		} else {
			v := a.focusedDocumentView()
			a.trackDocumentContext(v, point{(v.viewport.left + v.viewport.right) / 2, (v.viewport.top + v.viewport.bottom) / 2}, false)
		}
		return true
	}
	if ctrl&0x8000 == 0 {
		if !inline && !a.busy && !a.full && a.view.ready() && !a.focusedDocumentView().overview {
			if a.mode == idObjectsTab {
				if id := map[uintptr]int{'V': idObjectPoint, 'M': idObjectRectangle, 'L': idObjectLasso}[msg.wParam]; id != 0 {
					return a.objectCommand(id, 0)
				}
			}
			if mode := map[uintptr]int{'T': idTextSelectObject, 'V': idTextSelectParagraph, 'M': idTextSelectRectangle, 'L': idTextSelectLasso}[msg.wParam]; mode != 0 && !hasShift {
				a.activate(idTextTab)
				return a.textCommand(mode, 0)
			}
		}
		return false
	}
	stroke := documentShortcut{msg.wParam, hasShift}
	if inline {
		if command, ok := inlineShortcuts[stroke]; ok {
			a.doDocumentCommand(a.view, command, true)
			return true
		}
	}
	if command, ok := viewShortcuts[stroke]; ok {
		a.doDocumentCommand(a.focusedDocumentView(), command, false)
		return true
	}
	return false
}

func (a *app) focusedDocumentView() *documentView {
	if a.comparisonVisible() && a.comparison.canvasFocused {
		return a.comparison.view
	}
	return a.view
}

func (a *app) documentMouseContext(msg *winMessage) bool {
	e := &a.textTool.inline
	if msg.message != 0x0205 || msg.window != a.window && (!e.visible || msg.window != e.handle) {
		return false
	} // WM_RBUTTONUP
	p := point{int32(int16(msg.lParam)), int32(int16(msg.lParam >> 16))}
	if msg.window == e.handle {
		user32.NewProc("ClientToScreen").Call(e.handle, uintptr(unsafe.Pointer(&p)))
		user32.NewProc("ScreenToClient").Call(a.window, uintptr(unsafe.Pointer(&p)))
		a.trackDocumentContext(a.view, p, true)
		return true
	}
	v := a.view
	if a.comparisonVisible() && contains(a.comparison.view.viewport, p.x, p.y) {
		v = a.comparison.view
	}
	if !contains(v.viewport, p.x, p.y) {
		return false
	}
	if v == a.view && a.mode == idTextTab && a.inlineFrameAt(v.textAt(p.x, p.y)) >= 0 && !a.inlineDirty() {
		a.editFrameAt(v.textAt(p.x, p.y))
		a.trackDocumentContext(v, p, a.inlineCurrent() != nil)
	} else {
		a.trackDocumentContext(v, p, false)
	}
	return true
}

func (a *app) trackDocumentContext(v *documentView, pos point, editing bool) {
	if a.busy {
		return
	}
	groups := [][]documentCommand{
		{docUndo, docCopy, docSelectAll},
		{docFind, docSave, docPrint},
		{docZoomIn, docZoomOut, docFitPage, docPrevious, docNext, docFullscreen},
	}
	if editing {
		groups = [][]documentCommand{
			{docUndo, docRedo}, {docCut, docCopy, docPaste, docDelete, docSelectAll},
			{docBold, docItalic, docUnderline}, {docAlignLeft, docAlignCenter, docAlignRight, docAlignJustify},
		}
	} else if v == a.view && a.mode == idObjectsTab {
		groups[0] = []documentCommand{docUndo, docCopy, docPaste, docDuplicate, docDelete}
	}
	var commands []documentCommand
	var items []popupItem
	for i, group := range groups {
		if i > 0 {
			commands = append(commands, 0)
			items = append(items, popupItem{separator: true})
		}
		for _, command := range group {
			action := documentActions[command]
			commands = append(commands, command)
			items = append(items, popupItem{label: a.documentActionLabel(command), shortcut: action.shortcut, disabled: !a.canDocumentCommand(v, command, editing)})
		}
	}
	user32.NewProc("ClientToScreen").Call(a.window, uintptr(unsafe.Pointer(&pos)))
	if index := a.trackPopupAt(pos, 280, items); index >= 0 {
		a.doDocumentCommand(v, commands[index], editing)
	}
}

func (a *app) canDocumentCommand(v *documentView, command documentCommand, editing bool) bool {
	if a.busy {
		return false
	}
	e, frame := &a.textTool.inline, a.inlineCurrent()
	if !editing && v == a.view && a.mode == idObjectsTab {
		switch command {
		case docCopy, docDuplicate, docDelete:
			return a.objects.selected.Token != 0
		case docPaste:
			return a.objects.owner == a.activeTab && len(a.objects.clipboard) > 0
		}
	}
	inFrame := editing && frame != nil
	writable := inFrame && frame.editable
	hasSelection := false
	if inFrame {
		var start, end int32
		sendMessage.Call(e.handle, 0x00b0, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end))) // EM_GETSEL
		hasSelection = start != end
	}
	switch command {
	case docUndo:
		return v == a.view && (a.canInlineHistory(command) || !a.inlineDirty() && (v.canCanvasHistory(command) || len(v.history) > 0))
	case docRedo:
		return v == a.view && (a.canInlineHistory(command) || !a.inlineDirty() && (v.canCanvasHistory(command) || a.activeTab != nil && len(a.activeTab.redo) > 0))
	case docCut, docDelete:
		return writable && hasSelection
	case docCopy:
		if inFrame {
			return hasSelection
		}
		return v == a.view && a.mode == idTextTab && a.textTool.flow.text != ""
	case docPaste:
		ok, _, _ := sendMessage.Call(e.handle, 0x0432, 13, 0)
		return writable && ok != 0
	case docSelectAll:
		return inFrame || !a.full && v.ready()
	case docApplyText, docBold, docItalic, docUnderline, docAlignLeft, docAlignCenter, docAlignRight, docAlignJustify:
		return frame != nil && frame.editable && a.mode == idTextTab
	case docPrevious:
		return v.page > 0 && v.ready()
	case docNext:
		return v.page+1 < len(v.order) && v.ready()
	case docSave:
		return v.ready() && (v.changed() || v == a.view && a.inlineDirty())
	default:
		return v.ready()
	}
}

func (a *app) doDocumentCommand(v *documentView, command documentCommand, editing bool) {
	if !a.canDocumentCommand(v, command, editing) {
		return
	}
	if !editing && v == a.view && a.mode == idObjectsTab {
		if id := map[documentCommand]int{docCopy: idObjectCopy, docPaste: idObjectPaste,
			docDuplicate: idObjectDuplicate, docDelete: idObjectDelete}[command]; id != 0 {
			a.objectCommand(id, 0)
			return
		}
	}
	e := &a.textTool.inline
	if (command == docUndo || command == docRedo) && a.canInlineHistory(command) {
		a.doInlineHistory(command)
		return
	}
	if editing {
		message, ok := map[documentCommand]uintptr{docCut: 0x0300, docCopy: 0x0301, docPaste: 0x0302, docDelete: 0x0303}[command]
		if ok {
			setFocus.Call(e.handle)
			sendMessage.Call(e.handle, message, 0, 0)
			return
		}
		if command == docSelectAll {
			setFocus.Call(e.handle)
			sendMessage.Call(e.handle, 0x00b1, 0, ^uintptr(0))
			return
		}
	}
	format, ok := map[documentCommand]int{docBold: idTextBold, docItalic: idTextItalic, docUnderline: idTextUnderline,
		docAlignLeft: idTextAlignLeft, docAlignCenter: idTextAlignCenter, docAlignRight: idTextAlignRight, docAlignJustify: idTextAlignJustify}[command]
	if ok {
		a.textCommand(format, 0)
		if editing {
			setFocus.Call(e.handle)
		}
		return
	}
	switch command {
	case docUndo:
		if !v.stepCanvasHistory(command) {
			v.undo()
		}
	case docRedo:
		if !v.stepCanvasHistory(command) {
			a.redoDocument()
		}
	case docCopy:
		a.copyTextSelection()
	case docApplyText:
		a.startTextEdit()
	case docFind, docSave, docPrint, docSelectAll:
		if command == docSelectAll && v.overview {
			v.selectAllOverviewPages()
			return
		}
		action := a.showSearch
		if command == docSelectAll {
			action = func() {
				active := a.view
				if !active.ready() {
					return
				}
				a.activate(idTextTab)
				pick := textSelection{source: active.order[active.page], generation: active.generation,
					mode: textRectangle, tool: idTextTab, box: active.box(),
					polygon: rectangleSelection(textPoint{}, textPoint{1, 1})}
				a.selectText(pick)
			}
		}
		if command == docSave {
			action = func() { a.saveDocument(false) }
		}
		if command == docPrint {
			action = a.printDocument
		}
		after := action
		if v != a.view {
			tab := a.comparison.tab
			after = func() {
				a.selectDocumentTab(tab)
				if a.activeTab == tab {
					action()
				}
			}
		}
		if !a.saveInlineBefore(after) {
			after()
		}
	case docPrevious:
		v.gotoPage(v.page-1, false)
	case docNext:
		v.gotoPage(v.page+1, false)
	case docFullscreen:
		a.toggleFullscreen()
	case docFitPage:
		v.setZoomMode(readingZoomFitPage)
	case docZoomIn, docZoomOut:
		width, height := v.viewportSize()
		delta := 1.
		if command == docZoomOut {
			delta = -1
		}
		v.zoomAt(delta, width/2, height/2)
	}
}

func (a *app) copyTextSelection() {
	frame := a.inlineCurrent()
	if a.textTool.flow.text == "" && (frame == nil || frame.newBounds == nil) {
		return
	}
	value := a.textTool.flow.text
	if e := &a.textTool.inline; frame != nil && (len(e.frames) == 1 || a.inlineDirty()) {
		var start, end int32
		sendMessage.Call(e.handle, 0x00b0, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
		if start != end {
			sendMessage.Call(e.handle, 0x0301, 0, 0)
			return
		}
		value = getText(e.handle)
	}
	a.copyClipboardText(value)
}

func (a *app) copyClipboardText(value string) {
	data := syscall.StringToUTF16(multiline(value))
	bytes := uintptr(len(data) * 2)
	memory, _, _ := kernel32.NewProc("GlobalAlloc").Call(2, bytes)
	if memory == 0 {
		return
	}
	defer func() {
		if memory != 0 {
			kernel32.NewProc("GlobalFree").Call(memory)
		}
	}()
	destination, _, _ := kernel32.NewProc("GlobalLock").Call(memory)
	if destination == 0 {
		return
	}
	copyDrawItem.Call(destination, uintptr(unsafe.Pointer(&data[0])), bytes)
	kernel32.NewProc("GlobalUnlock").Call(memory)
	if ok, _, _ := user32.NewProc("OpenClipboard").Call(a.window); ok == 0 {
		return
	}
	defer user32.NewProc("CloseClipboard").Call()
	user32.NewProc("EmptyClipboard").Call()
	if ok, _, _ := user32.NewProc("SetClipboardData").Call(13, memory); ok != 0 {
		memory = 0
	}
}
