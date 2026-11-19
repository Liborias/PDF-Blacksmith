package main

import (
	"os"
	"path/filepath"
)

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

func (a *app) showFileMenu() {
	a.activate(idFileTab)
	items := make([]popupItem, 0, len(fileOptions))
	for _, option := range fileOptions {
		items = append(items, popupItem{
			label: option.label, command: option.id,
			checked: a.activeFileOption() == option.id, disabled: !a.fileOptionEnabled(option.id),
		})
	}
	if index := a.trackPopup(a.toolButtons[idFileTab], toolPanelWidth, items); index >= 0 {
		a.runFileOption(items[index].command)
	}
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
	v := a.view
	if a.busy || !v.ready() {
		return
	}
	req := v.request("save_document")
	changed := v.changed()
	if changed {
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
		req.Output, req.Overwrite, req.ReplaceSource = v.path, true, true
		req.SourceSize, req.SourceMtime = v.original.Size(), v.original.ModTime().UnixNano()
	}
	a.run(req)
}

func (a *app) reopenSaved(req request) {
	info, err := os.Stat(req.Output)
	if err != nil {
		notice(a.window, "PDF bylo uloženo. Soubor se nepodařilo znovu otevřít.\n"+filepath.Base(req.Output), 0x30)
		return
	}
	v := a.view
	page, continuous, fit, zoom := v.page, v.continuous, v.fitMode, v.zoom
	a.loadDocument(req.Output, req.PDFPassword, info)
	v.page, v.reference = page, page
	v.continuous, v.fitMode, v.zoom = continuous, fit, zoom
	if continuous {
		chooseCombo(v.flow, 0)
	}
	if fit == 0 {
		chooseCombo(v.fit, 2)
	} else {
		chooseCombo(v.fit, fit-1)
	}
}
