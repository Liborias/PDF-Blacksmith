package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unsafe"
)

const (
	idOCRTab = 270 + iota
	idOCRScope
	idOCRLanguage
	idOCRRun
	idOCRUndo
)

type ocrPanel struct {
	scope, pageRange, language, start, undo, report uintptr
}

func (a *app) buildOCRPanel() {
	p := &a.ocr
	a.label("Strany pro OCR", 16, 16, 298, true)
	p.scope = a.combo(idOCRScope, 16, 42, 145, pageScopeNames)
	p.pageRange = a.edit(169, 42, 145, false)
	sendMessage.Call(p.pageRange, 0x1501, 1, uintptr(unsafe.Pointer(wideString("Např. 1, 3–5, 8"))))
	a.label("Jazyk rozpoznávání", 16, 86, 298, true)
	p.language = a.combo(idOCRLanguage, 16, 112, 298, []string{"Čeština + angličtina", "Čeština", "Angličtina"})
	p.start = a.button("Rozpoznat text", idOCRRun, 16, 154, 145, 0)
	p.undo = a.button("Vrátit OCR", idOCRUndo, 169, 154, 145, 0)
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
	a.runDocumentDraft(req, "Rozpoznávám text na vybraných stranách…", a.ocr.report)
}

func (a *app) runDocumentDraft(req request, status string, report uintptr) {
	var err error
	req.Output, err = a.documentDraftPath()
	if err != nil {
		a.report("Pracovní PDF se nepodařilo vytvořit.\n" + err.Error())
		return
	}
	if a.view.hasPageEdits() {
		a.view.includeEdits(&req)
	}
	setText(report, status)
	if !a.run(req) {
		a.retiredDrafts = append(a.retiredDrafts, req.Output)
		a.pruneDrafts()
	}
}

func (a *app) documentDraftPath() (string, error) {
	if a.workDir == "" {
		root, err := os.MkdirTemp("", "pdf-blacksmith-")
		if err != nil {
			return "", err
		}
		a.workDir = root
	}
	folder, err := os.MkdirTemp(a.workDir, "edit-")
	if err != nil {
		return "", err
	}
	name := filepath.Base(a.view.path)
	if !strings.EqualFold(filepath.Ext(name), ".pdf") {
		name += ".pdf"
	}
	return filepath.Join(folder, name), nil
}

func (a *app) applyDocumentDraft(req request, result response) bool {
	info, err := os.Stat(req.Output)
	if err != nil || result.View == nil || len(result.View.Sizes) != len(a.view.order) {
		return false
	}
	v := a.view
	base := v.documentContent
	scope, _ := a.scopeControls()
	automaticScope := a.thumbnailScopes[scope]
	w, h := v.viewportSize()
	anchor, found := v.spread.anchor(v.scrollX+w/2, v.scrollY+h/2)
	query := getText(v.search.input)
	v.resetDocument()
	v.path, v.password, v.original = base.path, base.password, base.original
	v.draft, v.draftInfo = req.Output, info
	v.page, v.zoom, v.continuous, v.fitMode = base.page, base.zoom, base.continuous, base.fitMode
	v.thumbnailScale, v.pageSelection = base.thumbnailScale, slices.Clone(base.pageSelection)
	v.history = []pageEdit{{draftBase: &base, draftAction: req.Action}}
	v.resumeReader()
	v.syncReadingControls()
	v.applyDescription(result.View)
	if automaticScope {
		a.applyThumbnailScope()
	}
	if found {
		v.restore(anchor, w/2, h/2)
	}
	setText(v.search.input, query)
	v.updateDocumentInfo()
	v.changedViewport()
	return true
}

func (a *app) undoDocumentDraft(base documentContent, action string) {
	v := a.view
	scope, _ := a.scopeControls()
	automaticScope := a.thumbnailScopes[scope]
	a.retiredDrafts = append(a.retiredDrafts, v.draft)
	v.resetDocument()
	v.documentContent = base
	v.resumeReader()
	v.syncReadingControls()
	v.refreshBookmarks(-1)
	v.updatePageList()
	v.rebuild(false)
	v.updateDocumentInfo()
	v.changedViewport()
	if automaticScope {
		a.applyThumbnailScope()
	}
	if action == "ocr" {
		a.report("Poslední rozpoznávání OCR bylo vráceno.")
	} else {
		a.report("Poslední úprava textu byla vrácena.")
	}
	a.pruneDrafts()
}

func (a *app) retireDocumentDrafts(content documentContent) {
	if content.draft != "" {
		a.retiredDrafts = append(a.retiredDrafts, content.draft)
	}
	for _, edit := range content.history {
		if edit.draftBase != nil {
			a.retireDocumentDrafts(*edit.draftBase)
		}
	}
}

func (a *app) pruneDrafts() {
	if a.busy || a.view.running || a.view.prefetch.page != 0 || a.view.search.task != nil {
		return
	}
	a.retiredDrafts = slices.DeleteFunc(a.retiredDrafts, func(path string) bool {
		if path == "" {
			return true
		}
		if slices.ContainsFunc(a.merge.entries, func(entry mergeEntry) bool { return entry.input.Path == path }) {
			return false
		}
		return os.RemoveAll(filepath.Dir(path)) == nil
	})
}

func (a *app) cleanupDrafts() {
	if a.workDir == "" {
		return
	}
	for range 10 {
		if os.RemoveAll(a.workDir) == nil {
			return
		}
		time.Sleep(80 * time.Millisecond)
	}
}
