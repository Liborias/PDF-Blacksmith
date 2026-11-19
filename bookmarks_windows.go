package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	idBookmarksTab = 240 + iota
	idBookmarkList
	idBookmarkGo
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
	list, goTo, branch, add, empty      uintptr
	title, target, current, apply, undo uintptr
	rows                                []int
	collapsed                           map[int]bool
	selected                            int
	creating                            bool
}

func (v *documentView) buildBookmarkPanel() {
	a, panel := v.app, &v.bookmarkPanel
	panel.selected = -1
	panel.add = a.button("Nová záložka…", idBookmarkNew, 16, 16, 298, 0)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100
	panel.list = a.add("LISTBOX", "Strom záložek", idBookmarkList, 16, 58, 298, 240, wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(panel.list, 0x01a0, 0, uintptr(a.scale(24))) // LB_SETITEMHEIGHT
	a.addScrollPane(panel.list, a.scale(24), true)
	panel.empty = a.label("Dokument zatím nemá záložky.", 16, 61, 298, false)
	panel.goTo = a.button("Přejít na stranu", idBookmarkGo, 16, 305, 145, 0)
	panel.branch = a.button("Sbalit větev", idBookmarkBranch, 169, 305, 145, 0)
	a.label("Název záložky", 16, 347, 298, true)
	panel.title = a.edit(16, 372, 298, false)
	a.label("Cílová strana", 16, 410, 298, true)
	panel.target = a.edit(16, 435, 145, false)
	panel.current = a.button("Aktuální strana", idBookmarkCurrent, 169, 433, 145, 0)
	panel.apply = a.button("Použít úpravu", idBookmarkApply, 16, 478, 145, 0)
	panel.undo = a.button("Vrátit změnu", idBookmarkUndo, 169, 478, 145, 0)
	a.primaryButtons[panel.apply] = true
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
	setEnabled(panel.goTo, selected && v.ready() && slices.Contains(v.order, v.bookmarks[panel.selected].Page))
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
		sendMessage.Call(panel.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(bookmarkTitle(item)))))
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
	bounds.left += int32(d(7 + min(bookmark.Depth, 12)*13))
	if v.bookmarkHasChildren(index) {
		indicator := "▾"
		if panel.collapsed[index] {
			indicator = "▸"
		}
		marker := rect{bounds.left, bounds.top, bounds.left + int32(d(15)), bounds.bottom}
		drawText(item.dc, indicator, marker, v.app.normalFont, color, 0)
	}
	bounds.left += int32(d(17))
	bounds.right -= int32(d(6))
	drawText(item.dc, bookmarkTitle(bookmark), bounds, v.app.normalFont, color, 0x8000)
	if item.state&16 != 0 {
		v.app.drawFocus(item.dc, item.bounds)
	}
	return true
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
	v.app.revealFocusedControl()
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
	v.app.layoutPanel()
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
	if id == idBookmarkGo || (id == idBookmarkList && notification == 2) { // LBN_DBLCLK
		v.goToBookmark()
		return true
	}
	if notification != 0 || v.app.busy {
		return true
	}
	switch id {
	case idBookmarkBranch:
		if v.bookmarkHasChildren(panel.selected) {
			if panel.collapsed == nil {
				panel.collapsed = make(map[int]bool)
			}
			panel.collapsed[panel.selected] = !panel.collapsed[panel.selected]
			v.refreshBookmarks(panel.selected)
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
