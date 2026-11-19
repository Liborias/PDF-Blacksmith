package main

import "unsafe"

const (
	idOCRTab = 270 + iota
	idOCRScope
	idOCRLanguage
	idOCRRun
)

type ocrPanel struct {
	scope, pageRange, language, start, report uintptr
}

func (a *app) buildOCRPanel() {
	p := &a.ocr
	a.label("Strany pro OCR", 16, 16, 298, true)
	p.scope = a.combo(idOCRScope, 16, 42, 145, pageScopeNames)
	p.pageRange = a.edit(169, 42, 145, false)
	sendMessage.Call(p.pageRange, 0x1501, 1, uintptr(unsafe.Pointer(wideString("Např. 1, 3–5, 8"))))
	a.label("Jazyk rozpoznávání", 16, 86, 298, true)
	p.language = a.combo(idOCRLanguage, 16, 112, 298, []string{"Čeština + angličtina", "Čeština", "Angličtina"})
	p.start = a.button("Rozpoznat a uložit kopii…", idOCRRun, 16, 154, 298, 0)
	a.primaryButtons[p.start] = true
	p.report = a.readout(16, 205, 298, 104)
}

func (a *app) startOCR() {
	if a.busy || !a.view.ready() {
		return
	}
	req := a.view.request("ocr")
	if !a.selectPages(&req, a.ocr.scope, a.ocr.pageRange) {
		return
	}
	languages := [...]string{"ces+eng", "ces", "eng"}
	req.OCRLanguage = languages[max(0, comboIndex(a.ocr.language))]
	req.OpenResult = true
	if !a.chooseOutput(&req, "-OCR") {
		return
	}
	if a.view.changed() {
		a.view.includeEdits(&req)
	}
	setText(a.ocr.report, "Rozpoznávám text na vybraných stranách…")
	a.run(req)
}
