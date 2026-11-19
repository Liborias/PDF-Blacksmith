package main

import "slices"

func (a *app) scopeControls() (uintptr, uintptr) {
	switch a.mode {
	case idOCRTab:
		return a.ocr.scope, a.ocr.pageRange
	case idExportTab:
		return a.exportScope, a.exportRange
	case idSplitTab:
		return a.splitScope, a.splitRange
	default:
		return 0, 0
	}
}

func (a *app) applyThumbnailScope() {
	scope, field := a.scopeControls()
	selection := a.view.pageSelection
	if a.mode == idMergeTab {
		a.syncMergeThumbnailScope(selection)
	}
	if scope == 0 || (len(selection) == 0 && !a.thumbnailScopes[scope]) {
		return
	}
	positions := pageNumbers(selection)
	switch {
	case len(positions) == 0:
		chooseCombo(scope, exportAllPages)
		setText(field, "")
		delete(a.thumbnailScopes, scope)
	case contiguousPages(positions):
		chooseCombo(scope, exportPageRanges)
		setText(field, pageRanges(positions))
		a.thumbnailScopes[scope] = true
	default:
		chooseCombo(scope, exportThumbnails)
		setText(field, "")
		a.thumbnailScopes[scope] = true
	}
	a.layoutPanel()
}

func (a *app) resetThumbnailScopes() {
	for scope := range a.thumbnailScopes {
		chooseCombo(scope, exportAllPages)
	}
	clear(a.thumbnailScopes)
}

func (a *app) syncMergeThumbnailScope(selection []int) {
	entry := a.merge.selected()
	v := a.view
	if entry == nil || len(selection) == 0 {
		return
	}
	path, info := v.readSource()
	if info == nil || path != entry.input.Path || info.Size() != entry.input.SourceSize ||
		info.ModTime().UnixNano() != entry.input.SourceMtime {
		return
	}
	order := v.savedOrder
	if entry.input.View != nil {
		order = entry.input.View.Order
	}
	if len(v.order) != entry.info.Pages || !slices.Equal(v.order, order) {
		return
	}
	entry.pageRange = pageRanges(pageNumbers(selection))
	setText(a.merge.pageRange, entry.pageRange)
}
