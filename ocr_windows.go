package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
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
	areaMode                                        int
}

func (a *app) buildOCRPanel() {
	p := &a.ocr
	a.beginPropertySection(uiOCRPages, 16, uiOCRPagesHelp)
	p.scope = a.uiCombo(idOCRScope, 16, 42, 145, pageScopeNames)
	p.pageRange = a.edit(169, 42, 145, false)
	a.bindUICue(p.pageRange, uiRangeExample)
	a.beginPropertySection(uiRecognitionLanguage, 86, uiRecognitionLanguageHelp)
	p.language = a.uiCombo(idOCRLanguage, 16, 112, 298, []uiMessage{uiCzechEnglish, uiCzechLanguage, uiEnglishLanguage})
	a.addingSection = 0
	p.start = a.uiButton(uiRecognizeText, idOCRRun, 16, 154, 145, 0)
	p.undo = a.uiButton(uiUndoOCR, idOCRUndo, 169, 154, 145, 0)
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
	a.runDamagedDraft(req, status, report, nil)
}

func (a *app) runDamagedDraft(req request, status string, report uintptr, damage *pageDamage) {
	var err error
	req.Output, err = a.documentDraftPath()
	if err != nil {
		a.report("Pracovní PDF se nepodařilo vytvořit.\n" + err.Error())
		return
	}
	if a.view.hasPageEdits() {
		a.view.includeEdits(&req)
	} else {
		req.PageLocal = req.Action == "text_edit" || req.Action == "object_edit" || req.Action == "ocr"
	}
	if report != 0 {
		setText(report, status)
	}
	if a.run(req) {
		a.draftDamage = damage
		a.reportStatus(status)
	} else {
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
	count := len(a.view.order)
	if req.InsertPages != nil {
		count += len(req.InsertPages.Source.View.Pages)
	}
	if err != nil || result.View == nil {
		return false
	}
	v := a.view
	base := v.documentContent
	data := result.View
	if req.PageLocal && data.Pages == count && len(data.ChangedPages) != 0 {
		for _, page := range data.ChangedPages {
			if page < 1 || page > len(v.sizes) {
				return false
			}
		}
		a.dropRedo(a.activeTab)
		next := base
		next.draft, next.draftInfo = req.Output, info
		next.history = []pageEdit{{draftBase: &base, draftAction: req.Action, changedPages: data.ChangedPages, damage: a.draftDamage}}
		a.swapLocalVersion(next, data.ChangedPages, a.draftDamage, req.Action == "ocr")
		return true
	}
	if len(data.Sizes) != count {
		return false
	}
	a.dropRedo(a.activeTab)
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
	if insert := req.InsertPages; insert != nil {
		v.page = insert.Before
		v.pageSelection = make([]int, len(insert.Source.View.Pages))
		for i := range v.pageSelection {
			v.pageSelection[i] = insert.Before + i
		}
		found = false
	}
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

func (a *app) undoDocumentDraft(edit pageEdit) {
	action := edit.draftAction
	if len(edit.changedPages) > 0 {
		a.swapLocalVersion(*edit.draftBase, edit.changedPages, edit.damage, action == "ocr")
	} else {
		a.restoreDocumentVersion(*edit.draftBase)
	}
	messages := map[string]string{"ocr": "Poslední rozpoznávání OCR bylo vráceno.",
		"insert_pages": "Vložení stránek bylo vráceno.", "object_edit": "Úprava objektu byla vrácena.",
		"text_edit": "Poslední úprava textu byla vrácena."}
	a.report(messages[action])
}

func (a *app) restoreDocumentVersion(base documentContent) {
	v := a.view
	scope, _ := a.scopeControls()
	automaticScope := a.thumbnailScopes[scope]
	v.resetDocument()
	v.documentContent = base
	v.resumeReader()
	v.syncReadingControls()
	v.refreshBookmarks(-1)
	v.updatePageList()
	v.rebuild(false)
	v.updateDocumentInfo()
	v.changedViewport()
	a.updateActions()
	if automaticScope {
		a.applyThumbnailScope()
	}
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
	if a.busy || a.view.hasRenderJobs() || a.comparison.view.hasRenderJobs() || a.view.search.task != nil {
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
