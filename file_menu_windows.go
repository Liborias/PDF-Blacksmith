package main

import (
	"os"
	"path/filepath"
)

func (a *app) openFileMenu() {
	ready := !a.busy && a.view.ready()
	entries := []popupItem{
		{command: idOpen, label: "Otevřít soubor…", shortcut: "Ctrl+O", disabled: a.busy},
		{command: idShowOutput, label: "Otevřít výstup", disabled: a.busy || a.lastOutput == ""},
		{separator: true},
		{command: idSaveDocument, label: "Uložit", shortcut: "Ctrl+S", disabled: !ready || !a.view.changed()},
		{command: idSaveDocumentAs, label: "Uložit jako…", shortcut: "Ctrl+Shift+S", disabled: !ready},
		{command: idExportTab, label: "Export…", disabled: !ready},
		{command: idSignTab, label: "Uložit podepsanou kopii…", disabled: !ready},
		{separator: true},
		{command: idPrintDocument, label: "Tisk…", shortcut: "Ctrl+P", disabled: !ready},
		{command: idClosePDF, label: "Zavřít PDF", shortcut: "Ctrl+W", disabled: a.busy || a.view.path == ""},
	}
	if index := a.trackPopup(a.modes[idFileTab], 300, entries); index >= 0 {
		a.command(entries[index].command, 0)
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
