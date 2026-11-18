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
	case idOpen:
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

func (a *app) buildFilePanel() {
	a.fileHint = a.add("STATIC", "", 0, 16, 16, 298, 44, 0x80, 0, a.normalFont)
	a.filePath = a.readout(16, 65, 298, 54)
	a.fileRun = a.button("", idFileRun, 16, 134, 298, 0)
	a.primaryButtons[a.fileRun] = true
}

func (a *app) updateFilePanel() {
	path := a.view.path
	var hint, command string
	switch a.mode {
	case idOpen:
		hint, command = "PDF můžete také přetáhnout do okna aplikace.", "Vybrat soubor PDF…"
	case idShowOutput:
		hint, command = "Poslední uložený výstup", "Otevřít tento výstup"
		path = a.lastOutput
	case idSaveDocument:
		hint, command = "Úpravy se uloží do otevřeného PDF po potvrzení přepsání.", "Uložit změny"
	case idSaveDocumentAs:
		hint, command = "Vyberte název a umístění kopie otevřeného PDF.", "Vybrat umístění…"
	case idPrintDocument:
		hint, command = "V tiskovém dialogu vyberte tiskárnu, rozsah stran a počet kopií.", "Otevřít tiskový dialog…"
	}
	setText(a.fileHint, hint)
	setText(a.filePath, path)
	setText(a.fileRun, command)
	setEnabled(a.fileRun, a.fileOptionEnabled(a.mode))
}

func (a *app) updateFileOptions() {
	for _, option := range fileOptions {
		setEnabled(a.detailButtons[option.id], a.fileOptionEnabled(option.id))
	}
	if a.detailsTab() == idFileDetails {
		a.updateFilePanel()
	}
}

func (a *app) runFileOption() {
	if a.tool != idFileTab || !a.fileOptionEnabled(a.mode) {
		return
	}
	switch a.mode {
	case idOpen:
		a.openDialog()
	case idShowOutput:
		a.openDocument(a.lastOutput, a.lastPassword)
	case idSaveDocument, idSaveDocumentAs:
		a.saveDocument(a.mode == idSaveDocumentAs)
	case idPrintDocument:
		a.printDocument()
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
