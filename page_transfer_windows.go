package main

import "slices"

type pageInsertion struct {
	Source mergeInput `json:"source"`
	Before int        `json:"before"`
}

func (a *app) insertDraggedPages(source, target *documentView, before int) {
	from, to := a.activeTab, a.comparison.tab
	if source.secondary {
		from, to = to, from
	}
	if source == target || from == nil || to == nil {
		return
	}
	a.insertPagesBetweenTabs(from, to, slices.Clone(source.selectedPagePositions()), before)
}

func (a *app) insertPagesBetweenTabs(from, to *documentTab, positions []int, before int) {
	if a.busy || !a.comparisonVisible() || !a.comparisonKeeps(from) || !a.comparisonKeeps(to) {
		return
	}
	if a.saveInlineBefore(func() { a.insertPagesBetweenTabs(from, to, positions, before) }) {
		return
	}
	source := a.contentOf(from)
	path, info := source.path, source.original
	if source.draft != "" {
		path, info = source.draft, source.draftInfo
	}
	if !unchanged(path, info) {
		a.report("Zdrojové PDF se změnilo. Otevřete je znovu.")
		return
	}
	input := mergeInput{
		Path: path, Password: source.password, SourceSize: info.Size(), SourceMtime: info.ModTime().UnixNano(),
		View: &viewRequest{Order: slices.Clone(source.order), Rotations: slices.Clone(source.rotations), Pages: pageNumbers(positions)},
	}
	a.selectDocumentTab(to)
	v := a.view
	if a.activeTab != to || !v.ready() || before < 0 || before > len(v.order) {
		return
	}
	req := v.request("insert_pages")
	req.InsertPages = &pageInsertion{Source: input, Before: before}
	a.comparison.canvasFocused = false
	setFocus.Call(a.window)
	a.runDocumentDraft(req, "Vkládám stránky do cílového PDF…", 0)
}
