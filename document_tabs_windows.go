package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	idSelectDocumentTab = 450 + iota
	idListDocumentTabs
)

type documentTab struct {
	button, close                            uintptr
	content                                  documentContent
	redo                                     []documentContent
	tool, mode, screenBox, proofBox, profile int
	anchor                                   readingAnchor
	hasAnchor                                bool
	branches                                 map[int]bool
	selection, thumbnailTop                  int
	query, status, documentInfo              string
	openViewPending                          bool
}

type documentClose struct {
	tabs  []*documentTab
	saves []*documentTab
	next  int
	exit  bool
}

func (a *app) contentOf(tab *documentTab) *documentContent {
	if tab == a.activeTab {
		return &a.view.documentContent
	}
	if tab == a.comparison.tab {
		return &a.comparison.view.documentContent
	}
	return &tab.content
}

func (a *app) openedFile(path string, info os.FileInfo) *documentTab {
	for _, tab := range a.tabs {
		content := a.contentOf(tab)
		if strings.EqualFold(content.path, path) || (content.original != nil && info != nil && os.SameFile(info, content.original)) {
			return tab
		}
	}
	return nil
}

func (a *app) storeActiveTab() {
	if tab := a.activeTab; tab != nil {
		v := a.view
		tab.content = v.documentContent
		tab.tool, tab.mode = a.tool, a.mode
		tab.screenBox, tab.proofBox, tab.profile = comboIndex(v.screenBox), comboIndex(v.proofBox), comboIndex(v.proof)
		v.storeReadingPosition(tab)
		tab.branches, tab.selection = v.bookmarkPanel.collapsed, v.bookmarkPanel.selected
		tab.query, tab.status, tab.documentInfo = getText(v.search.input), getText(a.status), getText(v.documentInfo)
	}
}

func (a *app) selectDocumentTab(tab *documentTab) {
	if a.busy || tab == a.activeTab {
		return
	}
	if a.saveInlineBefore(func() { a.selectDocumentTab(tab) }) {
		return
	}
	a.storeActiveTab()
	a.layoutAids.temporary = false
	a.rightAux.syncSearch(false, &a.rightDock)
	if a.comparisonKeeps(tab) {
		a.setComparisonTab(nil)
	} else if a.comparison.tabs[0] != nil {
		a.endComparison()
	}
	a.activeTab = tab
	v := a.view
	v.resetDocument()
	a.clearDocumentControls()
	if tab == nil {
		a.tool, a.mode = idFileTab, idFileDetails
		setText(a.window, appName)
		a.reportStatus("")
	} else {
		v.documentContent = tab.content
		tab.content = documentContent{}
		a.tool, a.mode = tab.tool, tab.mode
		chooseCombo(v.screenBox, tab.screenBox)
		chooseCombo(v.proofBox, tab.proofBox)
		chooseCombo(v.proof, tab.profile)
		v.syncReadingControls()
		setText(a.window, filepath.Base(v.path)+" · "+appName)
		setText(a.password, v.password)
		v.setSearchInputs(tab.query)
		setText(v.documentInfo, tab.documentInfo)
		a.reportStatus(tab.status)
		if v.ready() && !v.sourceUnchanged() {
			v.invalidSource = true
			a.reportStatus("Vstupní PDF se změnilo. Otevřete je znovu.")
		}
		v.bookmarkPanel.collapsed = tab.branches
		v.resumeReader()
		v.refreshBookmarks(tab.selection)
		v.updatePageList()
	}
	a.syncSearchAuxiliary()
	a.updateActions()
	a.layout()
	if scope, _ := a.scopeControls(); comboIndex(scope) == exportAllPages {
		a.applyThumbnailScope()
	}
	v.rebuild(false)
	v.restoreReadingPosition(tab)
	v.updateDocumentInfo()
	v.changedViewport()
	setFocus.Call(a.window)
}

func (v *documentView) syncReadingControls() {
	flow := 1
	if v.continuous {
		flow = 0
	}
	chooseCombo(v.flow, flow)
	v.syncZoomChoice()
}

func (a *app) addDocumentTab(path, password string, info os.FileInfo) {
	if tab := a.openedFile(path, info); tab != nil {
		a.selectDocumentTab(tab)
		return
	}
	content := emptyDocument()
	content.path, content.password, content.original, content.describing = path, password, info, true
	tab := &documentTab{content: content, tool: idFileTab, mode: idFileDetails, selection: -1, status: "Načítám dokument…", openViewPending: true}
	tab.button = a.button(filepath.Base(path), idSelectDocumentTab, 0, 0, 1, 0x04000000)
	tab.close = a.button("Zavřít PDF", idClosePDF, 0, 0, 1, 0x04000000)
	a.tabs = append(a.tabs, tab)
	a.selectDocumentTab(tab)
}

func (a *app) stepDocumentTab(step int) {
	if index := slices.Index(a.tabs, a.activeTab); index >= 0 && len(a.tabs) > 1 {
		a.selectDocumentTab(a.tabs[(index+step+len(a.tabs))%len(a.tabs)])
	}
}

func (a *app) showDocumentList() {
	if a.busy {
		return
	}
	items := make([]popupItem, len(a.tabs))
	for index, tab := range a.tabs {
		items[index] = popupItem{label: a.contentOf(tab).path, checked: a.activeTab == tab}
	}
	if index := a.trackPopup(a.tabList, 480, items); index >= 0 {
		a.selectDocumentTab(a.tabs[index])
	}
}

func (a *app) tabCommand(id int, handle uintptr) bool {
	if id == idListDocumentTabs {
		a.showDocumentList()
		return true
	}
	if id != idSelectDocumentTab && id != idClosePDF {
		return false
	}
	for _, tab := range a.tabs {
		if handle == tab.button {
			a.selectDocumentTab(tab)
			break
		}
		if handle == tab.close {
			a.confirmCloseDocuments([]*documentTab{tab}, false)
			break
		}
	}
	return true
}

func (a *app) closeDocument() {
	if a.activeTab != nil {
		a.confirmCloseDocuments([]*documentTab{a.activeTab}, false)
	}
}

func (a *app) closeWindow() {
	if a.cancelTextJob != nil {
		if !a.closeAfterTextJob && notice(a.window, "Přerušit právě probíhající úpravu textu a pokračovat k zavření?", 0x134) == 6 {
			if a.cancelTextJob != nil {
				a.closeAfterTextJob = true
				a.cancelTextJob()
			} else {
				a.closeWindow()
			}
		}
		return
	}
	if a.busy || a.closing != nil {
		notice(a.window, "Počkejte na dokončení probíhající úlohy.", 0x40)
		return
	}
	if len(a.tabs) > 1 {
		switch notice(a.window, "Zavřít všechny otevřené karty a aplikaci?\n\nAno zavře všechny karty. Ne zavře pouze aktivní kartu.", 0x223) {
		case 7:
			a.closeDocument()
			return
		case 6:
		default:
			return
		}
	}
	a.confirmCloseDocuments(slices.Clone(a.tabs), true)
}

func (a *app) confirmCloseDocuments(tabs []*documentTab, exit bool) {
	if a.busy || a.closing != nil {
		return
	}
	plan := &documentClose{tabs: tabs, exit: exit}
	discardInline := false
	for _, tab := range tabs {
		content := a.contentOf(tab)
		dirty := tab == a.activeTab && a.inlineDirty()
		if !dirty && !content.changed() {
			continue
		}
		switch notice(a.window, "Uložit změny do původního PDF před zavřením?\n\n"+content.path, 0x223) {
		case 6:
			plan.saves = append(plan.saves, tab)
		case 7:
			discardInline = discardInline || dirty
		default:
			return
		}
	}
	// Ask about every tab before committing or discarding the current editor.
	if discardInline {
		a.clearTextSelection()
	} else if slices.Contains(plan.saves, a.activeTab) && a.saveInlineBefore(func() {
		a.closing = plan
		a.continueCloseDocuments()
	}) {
		return
	}
	a.closing = plan
	a.continueCloseDocuments()
}

func (a *app) continueCloseDocuments() {
	plan := a.closing
	if plan.next < len(plan.saves) {
		tab := plan.saves[plan.next]
		a.selectDocumentTab(tab)
		if !a.view.ready() || !unchanged(a.view.path, a.view.original) {
			a.closing = nil
			notice(a.window, "Původní PDF už není dostupné v původní podobě.\nZměny zůstaly v otevřené kartě.", 0x30)
			return
		}
		plan.next++
		if !a.saveOriginal() {
			a.closing = nil
		}
		return
	}
	a.closing = nil
	if plan.exit {
		destroyWindow.Call(a.window)
		return
	}
	for _, tab := range plan.tabs {
		a.removeDocumentTab(tab)
	}
}

func (a *app) removeDocumentTab(tab *documentTab) {
	if a.objects.owner == tab {
		a.objects.owner, a.objects.clipboard = nil, nil
	}
	index := slices.Index(a.tabs, tab)
	if index < 0 {
		return
	}
	if a.comparisonKeeps(tab) {
		a.endComparison()
	}
	a.dropRedo(tab)
	a.retireDocumentDrafts(*a.contentOf(tab))
	a.tabs = slices.Delete(a.tabs, index, index+1)
	destroyWindow.Call(tab.button)
	destroyWindow.Call(tab.close)
	if tab == a.activeTab {
		a.activeTab = nil
		if len(a.tabs) > 0 {
			a.selectDocumentTab(a.tabs[min(index, len(a.tabs)-1)])
		} else {
			a.view.resetDocument()
			a.clearDocumentControls()
			a.tool, a.mode = idFileTab, idFileDetails
			setText(a.window, appName)
			a.reportStatus("")
			a.updateActions()
			a.layout()
		}
	} else {
		a.layout()
	}
	a.pruneDrafts()
}


// A new tab keeps the recent-files panel until a real page image is ready.
// The flag belongs to the tab, so a different tab's reader cannot consume it.
func (a *app) cancelPendingOpenView() {
	if tab := a.activeTab; tab != nil && tab.openViewPending {
		tab.openViewPending = false
		if a.view.ready() && getText(a.status) == "Načítám dokument…" {
			a.reportStatus("")
		}
	}
}

func (v *documentView) completePendingOpenView(page int) {
	if v.secondary || !v.ready() || v.page < 0 || v.page >= len(v.order) || page != v.order[v.page] {
		return
	}
	a := v.app
	tab := a.activeTab
	if tab == nil || !tab.openViewPending {
		return
	}
	tab.openViewPending = false
	if getText(a.status) == "Načítám dokument…" {
		a.reportStatus("")
	}
	if !a.busy && a.tool == idFileTab && a.mode == idFileDetails {
		a.activate(idBrowseTab)
	}
}
