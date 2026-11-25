package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"unsafe"
)

const (
	idCompareOpen = 455 + iota
	idCompareTitleLeft
	idCompareTitleRight
	idCompareThumbsLeft
	idCompareThumbsRight
	idComparePrevLeft
	idComparePrevRight
	idCompareNextLeft
	idCompareNextRight
)

type comparisonPane struct {
	title, thumbs, previous, next, location uintptr
	hiddenThumbs                            bool
}

type documentComparison struct {
	tabs          [2]*documentTab
	panes         [2]comparisonPane
	view          *documentView
	tab           *documentTab
	button        uintptr
	thumbnailArea rect
	canvasFocused bool
}

func (a *app) buildComparison(className *uint16) bool {
	c := &a.comparison
	v := newDocumentView(a)
	v.secondary = true
	v.thumbnailSidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000|0x04000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	if v.thumbnailSidebar == 0 {
		return false
	}
	c.view = v
	v.overviewBar = a.addPanelScrollbar(a.window, v.scrollOverview)
	a.controlParent = v.thumbnailSidebar
	v.buildThumbnails()
	a.controlParent = 0
	for i := range c.panes {
		p := &c.panes[i]
		p.title = a.button("", idCompareTitleLeft+i, 0, 0, 1, 0)
		p.thumbs = a.shellButton(uiThumbnails, idCompareThumbsLeft+i, 1, 0)
		p.previous = a.button("‹", idComparePrevLeft+i, 0, 0, 1, 0)
		p.next = a.button("›", idCompareNextLeft+i, 0, 0, 1, 0)
		p.location = a.label("", 0, 0, 1, false)
	}
	return true
}

func (a *app) comparisonVisible() bool {
	return !a.full && a.comparison.tabs[0] != nil
}

func (a *app) chooseComparison() {
	if a.busy {
		return
	}
	c := &a.comparison
	if c.tabs[0] != nil {
		a.endComparison()
		a.layout()
		a.updateActions()
		return
	}
	if !a.view.ready() {
		return
	}
	var tabs []*documentTab
	var items []popupItem
	for _, tab := range a.tabs {
		if tab != a.activeTab {
			tabs = append(tabs, tab)
			items = append(items, popupItem{label: a.contentOf(tab).path})
		}
	}
	if len(tabs) == 0 {
		return
	}
	selected := 0
	if len(tabs) > 1 {
		selected = a.trackPopup(c.button, 480, items)
	}
	if selected >= 0 {
		c.tabs = [2]*documentTab{a.activeTab, tabs[selected]}
		a.layout()
		a.updateActions()
	}
}

func (a *app) endComparison() {
	a.setComparisonTab(nil)
	c := &a.comparison
	c.tabs = [2]*documentTab{}
	c.canvasFocused = false
	for i := range c.panes {
		c.panes[i].hiddenThumbs = false
	}
}

func (a *app) setComparisonTab(tab *documentTab) {
	c := &a.comparison
	if c.tab == tab {
		return
	}
	v := c.view
	a.cancelPageDrag()
	if c.tab != nil {
		v.storeReadingPosition(c.tab)
		c.tab.content = v.documentContent
		if v.readerError != "" {
			c.tab.status, c.tab.documentInfo = v.readerError, v.readerError
		}
	}
	c.tab = tab
	c.canvasFocused = false
	v.resetRendering()
	v.documentContent = emptyDocument()
	v.scrollX, v.scrollY = 0, 0
	if tab != nil {
		v.documentContent = tab.content
		v.secondaryBox = max(0, tab.screenBox)
		if !v.sourceUnchanged() {
			v.invalidSource = true
		}
		v.resumeReader()
	}
	v.updatePageList()
	v.rebuild(false)
}

func (v *documentView) storeReadingPosition(tab *documentTab) {
	w, h := v.viewportSize()
	tab.anchor, tab.hasAnchor = v.spread.anchor(v.scrollX+w/2, v.scrollY+h/2)
	pane := v.app.scrollPanes[v.pageList]
	_, _, tab.thumbnailTop = pane.lines()
}

func (v *documentView) restoreReadingPosition(tab *documentTab) {
	if tab == nil || !v.ready() {
		return
	}
	if tab.hasAnchor {
		w, h := v.viewportSize()
		v.restore(tab.anchor, w/2, h/2)
	}
	pane := v.app.scrollPanes[v.pageList]
	pane.scrollTo(tab.thumbnailTop)
	pane.syncScrollbar()
}

func (a *app) layoutComparison() rect {
	c := &a.comparison
	show := a.comparisonVisible()
	for _, p := range c.panes {
		for _, handle := range []uintptr{p.title, p.thumbs, p.previous, p.next, p.location} {
			showControl(handle, show)
		}
	}
	c.thumbnailArea = rect{}
	if !show {
		showControl(c.view.thumbnailSidebar, false)
		showControl(c.view.overviewBar.handle, false)
		return a.viewport
	}

	other := c.tabs[0]
	if other == a.activeTab {
		other = c.tabs[1]
	}
	attached := c.tab != other
	a.setComparisonTab(other)
	activeArea := a.viewport
	mid := (a.viewport.left + a.viewport.right) / 2
	left, right := a.viewport, a.viewport
	left.right, right.left = mid-1, mid+1
	d := a.scale
	thumbButtonWidth := max(d(70), a.captionWidth(a.uiText(uiThumbnails), a.labelFont)+d(12))
	for i, pane := range [...]rect{left, right} {
		p, tab := &c.panes[i], c.tabs[i]
		x, y, w := int(pane.left), int(pane.top), int(pane.right-pane.left)
		th := min(thumbButtonWidth, max(1, w-d(32)))
		a.setComparisonLabel(p.title, filepath.Base(a.contentOf(tab).path))
		a.primaryButtons[p.title] = tab == a.activeTab
		place(p.title, x+d(4), y+d(3), max(1, w-th-d(12)), d(26))
		place(p.thumbs, int(pane.right)-th-d(4), y+d(3), th, d(26))
		pane.top += int32(d(33))
		pane.bottom -= int32(d(32))
		bottom := int(pane.bottom) + d(3)
		place(p.previous, x+d(4), bottom, d(29), d(25))
		place(p.next, int(pane.right)-d(33), bottom, d(29), d(25))
		place(p.location, x+d(39), bottom+d(3), max(1, w-d(78)), d(22))
		var thumbs rect
		if !p.hiddenThumbs {
			width := int32(max(d(51), min(d(115), w/4)))
			thumbs = rect{pane.left, pane.top, pane.left + width, pane.bottom}
			pane.left += width
		}
		if tab == a.activeTab {
			a.thumbnailArea, activeArea = thumbs, pane
		} else {
			c.thumbnailArea = thumbs
			a.layoutComparisonThumbnails(thumbs)
			c.view.resize(pane)
			c.view.updateOverviewBar()
			if attached {
				c.view.restoreReadingPosition(tab)
				c.view.changedViewport()
			}
		}
	}
	a.updateComparisonNavigation()
	return activeArea
}

func (a *app) layoutComparisonThumbnails(bounds rect) {
	v := a.comparison.view
	show := bounds.right > bounds.left && v.ready()
	if show {
		w, h := int(bounds.right-bounds.left), int(bounds.bottom-bounds.top)
		place(v.thumbnailSidebar, int(bounds.left), int(bounds.top), w, h)
		v.layoutThumbnails(w, h)
	}
	showControl(v.thumbnailSidebar, show)
}

func (a *app) setComparisonLabel(handle uintptr, value string) {
	if getText(handle) != value {
		setText(handle, value)
	}
}

func (a *app) updateComparisonNavigation() {
	c := &a.comparison
	if c.view == nil {
		return
	}
	label := a.uiText(uiStartSideBySide)
	if c.tabs[0] != nil {
		label = a.uiText(uiStopSideBySide)
	}
	a.setComparisonLabel(c.button, label)
	setEnabled(c.button, !a.busy && (c.tabs[0] != nil || a.view.ready() && len(a.tabs) > 1))
	if !a.comparisonVisible() || c.tab == nil {
		return
	}
	for i, tab := range c.tabs {
		v, p := c.view, &c.panes[i]
		if tab == a.activeTab {
			v = a.view
		}
		setEnabled(p.title, !a.busy)
		setEnabled(p.previous, v.ready() && v.page > 0)
		setEnabled(p.next, v.ready() && v.page+1 < len(v.order))
		setEnabled(p.thumbs, v.ready())
		setEnabled(v.pageList, v.ready())
		a.primaryButtons[p.thumbs] = !p.hiddenThumbs
		location := ""
		if v.ready() {
			location = fmt.Sprintf("%d / %d   ·   %.0f %%", v.page+1, len(v.order), v.zoomPercent())
		}
		a.setComparisonLabel(p.location, location)
	}
}

func (a *app) comparisonCommand(id, notification int) bool {
	if notification != 0 {
		return false
	}
	if id == idCompareOpen {
		a.chooseComparison()
		return true
	}
	if id < idCompareTitleLeft || id > idCompareNextRight || !a.comparisonVisible() {
		return false
	}
	c := &a.comparison
	i := (id - idCompareTitleLeft) % 2
	tab := c.tabs[i]
	v := c.view
	if tab == a.activeTab {
		v = a.view
	}
	switch id {
	case idCompareTitleLeft, idCompareTitleRight:
		a.selectDocumentTab(tab)
	case idCompareThumbsLeft, idCompareThumbsRight:
		c.panes[i].hiddenThumbs = !c.panes[i].hiddenThumbs
		a.layout()
		v.schedule()
	case idComparePrevLeft, idComparePrevRight:
		v.gotoPage(v.page-1, false)
	case idCompareNextLeft, idCompareNextRight:
		v.gotoPage(v.page+1, false)
	}
	return true
}

func (a *app) thumbnailView(handle uintptr) *documentView {
	if v := a.view; v != nil && v.pageList == handle {
		return v
	}
	if v := a.comparison.view; v != nil && v.pageList == handle {
		return v
	}
	return nil
}

func (a *app) pointer(msg uint32, param uintptr) bool {
	if a.comparisonVisible() {
		v := a.comparison.view
		if a.view.selection.active() || a.objects.pressed || a.view.dragging {
			if a.view.pointer(msg, param) {
				return true
			}
			return v.pointer(msg, param)
		}
		if msg == 0x0201 || msg == 0x0207 {
			x, y := int32(int16(param)), int32(int16(param>>16))
			a.comparison.canvasFocused = contains(v.viewport, x, y)
		}
		if v.pointer(msg, param) {
			return true
		}
	}
	return a.view.pointer(msg, param)
}

func (a *app) comparisonKeeps(tab *documentTab) bool {
	return a.comparison.tabs[0] != nil && slices.Contains(a.comparison.tabs[:], tab)
}

func (v *documentView) showComparisonReaderError(message string) {
	v.readerError = message
	v.canvas.close()
	v.app.layoutComparisonThumbnails(v.app.comparison.thumbnailArea)
	v.app.updateComparisonNavigation()
	v.invalidateCanvas()
}
