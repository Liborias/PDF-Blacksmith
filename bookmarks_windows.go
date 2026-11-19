package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	idToggleBookmarks = 240 + iota
	idBookmarkList
	_
	idBookmarkBranch
	idBookmarkNew
	idBookmarkCurrent
	idBookmarkApply
	idBookmarkUndo
)

type pdfBookmark struct {
	Title string `json:"title"`
	Page  int    `json:"page"`
	Depth int    `json:"depth"`
}

type bookmarkChange struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	Page  int    `json:"page,omitempty"`
}

type bookmarkPanel struct {
	heading, list, branch, add, empty   uintptr
	titleLabel, targetLabel             uintptr
	title, target, current, apply, undo uintptr
	rows                                []int
	collapsed                           map[int]bool
	selected                            int
	widest                              int
	creating                            bool
}

func (v *documentView) buildBookmarkPanel() {
	a, panel := v.app, &v.bookmarkPanel
	panel.selected = -1
	panel.heading = a.label("Záložky", 0, 0, 1, true)
	panel.add = a.button("Nová záložka…", idBookmarkNew, 0, 0, 1, 0)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100
	panel.list = a.add("LISTBOX", "Strom záložek", idBookmarkList, 0, 0, 1, 1, wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(panel.list, 0x01a0, 0, uintptr(a.scale(24))) // LB_SETITEMHEIGHT
	a.addScrollPane(panel.list, a.scale(24), true)
	panel.empty = a.label("Dokument zatím nemá záložky.", 0, 0, 1, false)
	panel.branch = a.button("Sbalit větev", idBookmarkBranch, 0, 0, 1, 0)
	panel.titleLabel = a.label("Název záložky", 0, 0, 1, true)
	panel.title = a.edit(0, 0, 1, false)
	panel.targetLabel = a.label("Cílová strana", 0, 0, 1, true)
	panel.target = a.edit(0, 0, 1, false)
	panel.current = a.button("Aktuální strana", idBookmarkCurrent, 0, 0, 1, 0)
	panel.apply = a.button("Použít úpravu", idBookmarkApply, 0, 0, 1, 0)
	panel.undo = a.button("Vrátit změnu", idBookmarkUndo, 0, 0, 1, 0)
	for _, control := range [...]uintptr{panel.heading, panel.empty, panel.titleLabel, panel.targetLabel} {
		a.panelLabels[control] = true
	}
	for _, control := range [...]uintptr{panel.add, panel.branch, panel.current, panel.apply, panel.undo} {
		a.panelButtons[control] = true
	}
	a.primaryButtons[panel.apply] = true
}

func (v *documentView) layoutBookmarkPanel() {
	a, panel := v.app, &v.bookmarkPanel
	area := a.bookmarkArea
	width, height := int(area.right-area.left), int(area.bottom-area.top)
	pane := a.scrollPanes[panel.list]
	if width <= 0 || !v.ready() || a.full {
		pane.arrange(rect{}, false)
		showControl(a.bookmarkSidebar, false)
		return
	}
	place(a.bookmarkSidebar, int(area.left), int(area.top), width, height)
	d := a.scale
	margin := min(d(12), width/10)
	available := width - 2*margin
	compact := available < d(192)
	formHeight := d(198)
	if compact {
		formHeight = d(269)
	}
	formTop := max(d(104), height-formHeight)
	place(panel.heading, margin, d(8), available, d(22))
	place(panel.add, margin, d(34), available, d(32))
	empty := len(v.bookmarks) == 0
	place(panel.empty, margin, d(75), available, d(56))
	showControl(panel.empty, empty)
	pane.arrange(rect{int32(margin), int32(d(74)), int32(width - margin), int32(formTop - d(7))}, !empty)
	place(panel.branch, margin, formTop, available, d(32))
	showControl(panel.branch, !empty)
	place(panel.titleLabel, margin, formTop+d(39), available, d(22))
	place(panel.title, margin, formTop+d(61), available, d(28))
	place(panel.targetLabel, margin, formTop+d(96), available, d(22))
	if compact {
		setText(panel.add, "Přidat záložku")
		place(panel.target, margin, formTop+d(118), available, d(28))
		place(panel.current, margin, formTop+d(151), available, d(32))
		place(panel.apply, margin, formTop+d(190), available, d(32))
		place(panel.undo, margin, formTop+d(228), available, d(32))
	} else {
		setText(panel.add, "Nová záložka…")
		half := (available - d(8)) / 2
		place(panel.target, margin, formTop+d(120), available-d(119), d(28))
		place(panel.current, width-margin-d(111), formTop+d(118), d(111), d(32))
		place(panel.apply, margin, formTop+d(157), half, d(32))
		place(panel.undo, width-margin-half, formTop+d(157), half, d(32))
	}
	showControl(a.bookmarkSidebar, true)
	// Moving siblings during resize can leave copied pixels in the owner-drawn list.
	user32.NewProc("RedrawWindow").Call(a.bookmarkSidebar, 0, 0, 0x0085) // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN
}

func (v *documentView) preferredBookmarkWidth() int {
	return min(v.app.scale(330), max(v.app.scale(216), v.bookmarkPanel.widest+v.app.scale(44)))
}

func (v *documentView) bookmarkChanges() []bookmarkChange {
	var changes []bookmarkChange
	for index, item := range v.bookmarks {
		if index >= len(v.savedBookmarks) {
			changes = append(changes, bookmarkChange{Title: item.Title, Page: item.Page})
		} else if old := v.savedBookmarks[index]; item != old {
			change := bookmarkChange{Index: index + 1, Title: item.Title}
			if item.Page != old.Page {
				change.Page = item.Page
			}
			changes = append(changes, change)
		}
	}
	return changes
}

func (v *documentView) bookmarkHasChildren(index int) bool {
	return index >= 0 && index+1 < len(v.bookmarks) && v.bookmarks[index+1].Depth > v.bookmarks[index].Depth
}

func (v *documentView) updateBookmarkActions() {
	panel := &v.bookmarkPanel
	ready := v.ready() && !v.app.busy
	selected := panel.selected >= 0 && panel.selected < len(v.bookmarks)
	editing := ready && (selected || panel.creating)
	setEnabled(panel.add, ready)
	setEnabled(panel.list, v.ready())
	setEnabled(panel.branch, selected && v.bookmarkHasChildren(panel.selected))
	for _, control := range []uintptr{panel.title, panel.target, panel.current, panel.apply} {
		setEnabled(control, editing)
	}
	setEnabled(panel.undo, ready && len(v.history) > 0)
}

func (v *documentView) showBookmark(index int) {
	panel := &v.bookmarkPanel
	panel.selected, panel.creating = index, false
	title, target := "", ""
	if index >= 0 && index < len(v.bookmarks) {
		item := v.bookmarks[index]
		title = item.Title
		if position := slices.Index(v.order, item.Page); position >= 0 {
			target = strconv.Itoa(position + 1)
		}
	}
	setText(panel.title, title)
	setText(panel.target, target)
	if panel.collapsed[index] {
		setText(panel.branch, "Rozbalit větev")
	} else {
		setText(panel.branch, "Sbalit větev")
	}
	v.updateBookmarkActions()
}

func (v *documentView) refreshBookmarks(selected int) {
	panel := &v.bookmarkPanel
	oldWidth := v.preferredBookmarkWidth()
	panel.widest = 0
	panel.rows = panel.rows[:0]
	sendMessage.Call(panel.list, 0x000b, 0, 0) // WM_SETREDRAW
	sendMessage.Call(panel.list, 0x0184, 0, 0) // LB_RESETCONTENT
	skippedDepth := -1
	for index, item := range v.bookmarks {
		if skippedDepth >= 0 && item.Depth > skippedDepth {
			continue
		}
		skippedDepth = -1
		if panel.collapsed[index] && v.bookmarkHasChildren(index) {
			skippedDepth = item.Depth
		}
		panel.rows = append(panel.rows, index)
		title := bookmarkTitle(item)
		if panel.widest < v.app.scale(286) {
			width := v.app.captionWidth(title, v.app.normalFont) + v.app.scale(min(item.Depth, 12)*13)
			panel.widest = max(panel.widest, width)
		}
		sendMessage.Call(panel.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(title))))
	}
	row := slices.Index(panel.rows, selected)
	sendMessage.Call(panel.list, 0x0186, uintptr(row), 0) // LB_SETCURSEL
	sendMessage.Call(panel.list, 0x000b, 1, 0)
	pane := v.app.scrollPanes[panel.list]
	if !pane.shown {
		showControl(panel.list, false)
	}
	pane.reflow()
	invalidate(panel.list)
	if row < 0 {
		selected = -1
	}
	v.showBookmark(selected)
	if a := v.app; v.ready() && !a.full && !a.bookmarksCollapsed && a.bookmarkArea.right > a.bookmarkArea.left && oldWidth != v.preferredBookmarkWidth() {
		a.layout()
	} else {
		v.layoutBookmarkPanel()
	}
}

func bookmarkTitle(item pdfBookmark) string {
	if strings.TrimSpace(item.Title) == "" {
		return "Bez názvu"
	}
	return item.Title
}

func (v *documentView) drawBookmarkRow(item *buttonDraw) bool {
	panel := &v.bookmarkPanel
	if item.controlType != 2 || item.handle != panel.list {
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, panelColor)
	if int(item.itemID) >= len(panel.rows) {
		return true
	}
	index := panel.rows[item.itemID]
	bookmark := v.bookmarks[index]
	color := textColor
	if item.state&1 != 0 {
		color = selectionColor
	}
	d := v.app.scale
	bounds.left += int32(v.bookmarkIndent(bookmark, int(bounds.right-bounds.left)))
	if v.bookmarkHasChildren(index) {
		x, y := bounds.left+int32(d(6)), (bounds.top+bounds.bottom)/2
		long, short := int32(max(2, d(4))), int32(max(1, d(2)))
		chevron := [...]point{{x - long, y - short}, {x, y + short}, {x + long, y - short}}
		if panel.collapsed[index] {
			chevron = [...]point{{x - short, y - long}, {x + short, y}, {x - short, y + long}}
		}
		pen, _, _ := gdiPen.Call(0, 1, color.win())
		old, _, _ := gdiSelect.Call(item.dc, pen)
		gdi32.NewProc("Polyline").Call(item.dc, uintptr(unsafe.Pointer(&chevron[0])), uintptr(len(chevron)))
		gdiSelect.Call(item.dc, old)
		gdiDelete.Call(pen)
	}
	bounds.left += int32(d(17))
	bounds.right -= int32(d(6))
	drawText(item.dc, bookmarkTitle(bookmark), bounds, v.app.normalFont, color, 0x8000)
	if item.state&16 != 0 {
		v.app.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (v *documentView) bookmarkIndent(bookmark pdfBookmark, width int) int {
	d := v.app.scale
	return min(d(7+min(bookmark.Depth, 12)*13), max(d(7), width-d(66)))
}

func (v *documentView) bookmarkAt(location uintptr) (int, bool) {
	panel := &v.bookmarkPanel
	row, _, _ := sendMessage.Call(panel.list, 0x01a9, 0, location) // LB_ITEMFROMPOINT
	if row>>16 != 0 || row >= uintptr(len(panel.rows)) {
		return -1, false
	}
	var bounds rect
	sendMessage.Call(panel.list, 0x0198, row, uintptr(unsafe.Pointer(&bounds))) // LB_GETITEMRECT
	x, y := int32(int16(location)), int32(int16(location>>16))
	if !contains(bounds, x, y) {
		return -1, false
	}
	index := panel.rows[row]
	indent := int(bounds.left) + v.bookmarkIndent(v.bookmarks[index], int(bounds.right-bounds.left))
	return index, v.bookmarkHasChildren(index) && int(x) >= indent && int(x) < indent+v.app.scale(17)
}

func (v *documentView) toggleBookmarkBranch(index int) {
	panel := &v.bookmarkPanel
	top, _, _ := sendMessage.Call(panel.list, 0x018e, 0, 0) // LB_GETTOPINDEX
	first := index
	if top < uintptr(len(panel.rows)) {
		first = panel.rows[top]
	}
	if panel.collapsed == nil {
		panel.collapsed = make(map[int]bool)
	}
	panel.collapsed[index] = !panel.collapsed[index]
	v.refreshBookmarks(panel.selected)
	if row := slices.Index(panel.rows, first); row >= 0 {
		sendMessage.Call(panel.list, 0x0197, uintptr(row), 0) // LB_SETTOPINDEX
	}
}

func (v *documentView) goToBookmark() {
	index := v.bookmarkPanel.selected
	if index >= 0 && index < len(v.bookmarks) {
		if position := slices.Index(v.order, v.bookmarks[index].Page); position >= 0 {
			v.gotoPage(position, false)
		}
	}
}

func (v *documentView) startBookmark() {
	panel := &v.bookmarkPanel
	sendMessage.Call(panel.list, 0x0186, ^uintptr(0), 0)
	v.showBookmark(-1)
	panel.creating = true
	setText(panel.title, fmt.Sprintf("Strana %d", v.page+1))
	setText(panel.target, strconv.Itoa(v.page+1))
	v.updateBookmarkActions()
	setFocus.Call(panel.title)
	sendMessage.Call(panel.title, 0x00b1, 0, ^uintptr(0)) // EM_SETSEL
}

func (v *documentView) editBookmark() {
	panel := &v.bookmarkPanel
	index := panel.selected
	if !panel.creating && (index < 0 || index >= len(v.bookmarks)) {
		return
	}
	item := pdfBookmark{Title: getText(panel.title)}
	if strings.TrimSpace(item.Title) == "" {
		notice(v.app.window, "Zadejte název záložky.", 0x30)
		setFocus.Call(panel.title)
		return
	}
	if !panel.creating {
		item.Page, item.Depth = v.bookmarks[index].Page, v.bookmarks[index].Depth
	}
	target := strings.TrimSpace(getText(panel.target))
	if target != "" || panel.creating {
		position, err := strconv.Atoi(target)
		if err != nil || position < 1 || position > len(v.order) {
			notice(v.app.window, fmt.Sprintf("Zadejte cílovou stranu od 1 do %d.", len(v.order)), 0x30)
			setFocus.Call(panel.target)
			return
		}
		item.Page = v.order[position-1]
	}
	if !panel.creating && item == v.bookmarks[index] {
		return
	}
	previous := v.bookmarks
	if previous == nil {
		previous = []pdfBookmark{}
	}
	v.app.dropRedo(v.app.activeTab)
	v.history = append(v.history, pageEdit{bookmarks: previous, bookmarkSelection: index})
	updated := slices.Clone(v.bookmarks)
	if panel.creating {
		index = len(updated)
		updated = append(updated, item)
	} else {
		updated[index] = item
	}
	v.bookmarks = updated
	v.refreshBookmarks(index)
	v.app.updateActions()
}

func (v *documentView) bookmarkCommand(id, notification int) bool {
	if id < idBookmarkList || id > idBookmarkUndo {
		return false
	}
	if !v.ready() {
		return true
	}
	panel := &v.bookmarkPanel
	if id == idBookmarkList && notification == 1 { // LBN_SELCHANGE
		row, _, _ := sendMessage.Call(panel.list, 0x0188, 0, 0)
		if row < uintptr(len(panel.rows)) {
			v.showBookmark(panel.rows[row])
		}
		return true
	}
	if notification != 0 || v.app.busy {
		return true
	}
	switch id {
	case idBookmarkBranch:
		if v.bookmarkHasChildren(panel.selected) {
			v.toggleBookmarkBranch(panel.selected)
		}
	case idBookmarkNew:
		v.startBookmark()
	case idBookmarkCurrent:
		setText(panel.target, strconv.Itoa(v.page+1))
	case idBookmarkApply:
		v.editBookmark()
	case idBookmarkUndo:
		v.undo()
	}
	return true
}
