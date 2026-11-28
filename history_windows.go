package main

import (
	"math"
	"unsafe"
)

const (
	idHeaderUndo = 519 + iota
	idHeaderRedo
)

func (a *app) canInlineHistory(command documentCommand) bool {
	e := &a.textTool.inline
	if frame := a.inlineCurrent(); frame == nil || !frame.editable || !e.visible {
		return false
	}
	message := uintptr(0x00c6) // EM_CANUNDO
	if command == docRedo {
		message = 0x0455 // EM_CANREDO
	}
	yes, _, _ := sendMessage.Call(e.handle, message, 0, 0)
	return yes != 0
}

func (a *app) doInlineHistory(command documentCommand) {
	p, e := &a.textTool, &a.textTool.inline
	a.focusInlineEditor()
	message := uintptr(0x0454) // EM_REDO
	if command == docUndo {
		message = 0x00c7 // EM_UNDO
	}
	sendMessage.Call(e.handle, message, 0, 0)
	format := inlineParaFormat{size: uint32(unsafe.Sizeof(inlineParaFormat{}))}
	sendMessage.Call(e.handle, 0x043d, 0, uintptr(unsafe.Pointer(&format))) // EM_GETPARAFORMAT
	e.loading = true
	if format.mask&0x100 != 0 && format.rule == 5 && format.spacing != int32(math.Round(e.previewStyle.leading*20)) {
		leading := float64(format.spacing) / 20
		if format.spacing == int32(math.Round(e.initialStyle.leading*20)) {
			leading = e.initialStyle.leading
		}
		setText(p.leading, formatTextNumber(leading))
	}
	if format.mask&8 != 0 && format.alignment >= 1 && format.alignment <= 4 {
		p.alignment = [...]int{0, 0, 2, 1, 3}[format.alignment]
	}
	e.previewStyle = a.inlineStyle()
	e.loading = false
	a.syncInlineControls(0)
	a.layoutInlineEditor()
	a.updateTextActions()
	a.updateHistoryButtons()
	a.view.invalidateCanvas()
}

func (a *app) updateHistoryButtons() {
	a.refreshFooterInfo()
	if a.activeTab != nil { invalidate(a.activeTab.button) }
	if a.historyButtons[0] != 0 {
		setEnabled(a.historyButtons[0], a.canDocumentCommand(a.view, docUndo, false))
		setEnabled(a.historyButtons[1], a.canDocumentCommand(a.view, docRedo, false))
	}
}

func (a *app) redoDocument() {
	tab := a.activeTab
	if tab == nil || len(tab.redo) == 0 {
		return
	}
	index := len(tab.redo) - 1
	version := tab.redo[index]
	tab.redo[index] = documentContent{}
	tab.redo = tab.redo[:index]
	var applied bool
	if len(version.history) > 0 {
		edit := version.history[len(version.history)-1]
		if len(edit.changedPages) > 0 && edit.draftBase != nil && a.view.isVersion(*edit.draftBase) {
			a.swapLocalVersion(version, edit.changedPages, edit.damage, edit.draftAction == "ocr")
			applied = true
		}
	}
	if !applied {
		a.restoreDocumentVersion(version)
	}
	a.reportStatus("Změna byla znovu použita.")
}

func (a *app) dropRedo(tab *documentTab) {
	if tab == nil {
		return
	}
	current := a.contentOf(tab).draft
	for _, state := range tab.redo {
		if state.draft != "" && state.draft != current {
			a.retiredDrafts = append(a.retiredDrafts, state.draft)
		}
	}
	tab.redo = nil
}
