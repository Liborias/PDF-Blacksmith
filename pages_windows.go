package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	lbSetSel         = 0x0185
	lbGetSel         = 0x0187
	lbGetSelCount    = 0x0190
	lbGetSelItems    = 0x0191
	lbSetAnchorIndex = 0x019c
	lbSetCaretIndex  = 0x019e
	lbGetCaretIndex  = 0x019f
)

func (v *documentView) buildPagePanel() {
	a := v.app
	v.moveLabel = a.label("Cílová pozice", 16, 16, 298, true)
	v.moveTarget = a.edit(16, 42, 76, false)
	v.moveButton = a.button("Přesunout", idReaderMoveTo, 100, 40, 214, 0)
	v.back = a.button("Posunout před", idReaderMoveBack, 16, 80, 145, 0)
	v.ahead = a.button("Posunout za", idReaderMoveAhead, 169, 80, 145, 0)
	v.left = a.button("Otočit vlevo", idReaderRotateLeft, 16, 120, 145, 0)
	v.right = a.button("Otočit vpravo", idReaderRotateRight, 169, 120, 145, 0)
	v.deleteButton = a.button("Smazat vybrané", idReaderDeletePages, 16, 160, 145, 0)
	v.undoButton = a.button("Vrátit změnu", idReaderUndo, 169, 160, 145, 0)
	v.save = a.button("Uložit upravenou kopii…", idReaderSave, 16, 200, 298, 0)
	a.primaryButtons[v.save] = true
}

func (v *documentView) updatePageEditActions(editing bool) {
	selected := max(1, len(v.pageSelection))
	last := len(v.order) - selected + 1
	label := "Cílová pozice"
	if last > 1 {
		label = fmt.Sprintf("Cílová pozice (1–%d)", last)
	}
	if getText(v.moveLabel) != label {
		setText(v.moveLabel, label)
	}
	setEnabled(v.moveTarget, editing && last > 1)
	setEnabled(v.moveButton, editing && last > 1)
	setEnabled(v.deleteButton, editing && selected < len(v.order))
}

func (v *documentView) selectedPagePositions() []int {
	if len(v.pageSelection) > 0 {
		return v.pageSelection
	}
	return []int{v.page}
}

func (v *documentView) selectedSourcePages() []int {
	positions := v.selectedPagePositions()
	numbers := make([]int, len(positions))
	for i, position := range positions {
		numbers[i] = v.order[position]
	}
	return numbers
}

func (v *documentView) inPageSelection(index int) bool {
	if len(v.pageSelection) > 0 {
		_, found := slices.BinarySearch(v.pageSelection, index)
		return found
	}
	return index == v.page
}

func (v *documentView) resetPageListSelection() {
	if len(v.order) == 0 {
		return
	}
	sendMessage.Call(v.pageList, lbSetSel, 0, ^uintptr(0))
	for _, index := range v.selectedPagePositions() {
		sendMessage.Call(v.pageList, lbSetSel, 1, uintptr(index))
	}
	sendMessage.Call(v.pageList, lbSetAnchorIndex, uintptr(v.page), 0)
	sendMessage.Call(v.pageList, lbSetCaretIndex, uintptr(v.page), 1)
}

func (v *documentView) syncPageListSelection() {
	if len(v.pageSelection) > 0 {
		return
	}
	count, _, _ := sendMessage.Call(v.pageList, lbGetSelCount, 0, 0)
	state, _, _ := sendMessage.Call(v.pageList, lbGetSel, uintptr(v.page), 0)
	if count != 1 || state != 1 {
		v.resetPageListSelection()
	}
}

func (v *documentView) readPageSelection() {
	if !v.ready() {
		return
	}
	count, _, _ := sendMessage.Call(v.pageList, lbGetSelCount, 0, 0)
	caret, _, _ := sendMessage.Call(v.pageList, lbGetCaretIndex, 0, 0)
	index := min(max(int(int32(caret)), 0), len(v.order)-1)
	v.pageSelection = nil
	if count > 0 && count <= uintptr(len(v.order)) {
		items := make([]int32, int(count))
		read, _, _ := sendMessage.Call(v.pageList, lbGetSelItems, count, uintptr(unsafe.Pointer(&items[0])))
		if read > 0 && read <= count {
			selected := make([]int, read)
			for i, item := range items[:read] {
				selected[i] = int(item)
			}
			slices.Sort(selected)
			if _, found := slices.BinarySearch(selected, index); !found {
				index = selected[0]
			}
			v.pageSelection = selected
		}
	}
	if !v.secondary {
		v.app.applyThumbnailScope()
	}
	if index != v.page {
		v.gotoPage(index, false)
	} else if !v.secondary {
		v.updateActions()
	}
}

func (v *documentView) selectOnlyPage(index int) {
	v.pageSelection = []int{index}
	if !v.secondary {
		v.app.applyThumbnailScope()
	}
	v.gotoPage(index, false)
	v.resetPageListSelection()
}

func (v *documentView) rememberOrder() {
	v.app.dropRedo(v.app.activeTab)
	v.history = append(v.history, pageEdit{page: v.page, order: v.order, selection: v.pageSelection})
}

func (v *documentView) changePageOrder(order, selection []int, page int) {
	v.generation++
	v.order, v.pageSelection = order, selection
	v.page, v.reference = page, page
	if v.bookmarkPanel.selected >= 0 {
		v.showBookmark(v.bookmarkPanel.selected)
	}
	v.updatePageList()
	if v.search.query != "" {
		v.queueSearch(v.search.query)
	}
	v.rebuild(false)
	setText(v.scanReport, "")
	v.updateDocumentInfo()
	v.changedViewport()
}

func (v *documentView) moveSelectedPages(first int) bool {
	selected := v.selectedPagePositions()
	order := placePages(v.order, selected, first)
	if order == nil {
		return false
	}
	v.rememberOrder()
	var selection []int
	if len(v.pageSelection) > 0 {
		selection = make([]int, len(selected))
		for i := range selection {
			selection[i] = first + i
		}
	}
	v.changePageOrder(order, selection, first)
	return true
}

func (v *documentView) moveToPosition() {
	if !v.ready() || v.app.busy {
		return
	}
	limit := len(v.order) - len(v.selectedPagePositions()) + 1
	if limit <= 1 {
		return
	}
	position, err := strconv.Atoi(strings.TrimSpace(getText(v.moveTarget)))
	if err != nil || position < 1 || position > limit {
		notice(v.app.window, fmt.Sprintf("Zadejte cílovou pozici od 1 do %d.", limit), 0x30)
		setFocus.Call(v.moveTarget)
		return
	}
	if v.moveSelectedPages(position - 1) {
		v.app.reportStatus("")
	} else {
		v.app.reportStatus("Strany už jsou na této pozici.")
	}
}

func (v *documentView) deleteSelectedPages() {
	selected := v.selectedPagePositions()
	if !v.ready() || v.app.busy || len(selected) >= len(v.order) {
		return
	}
	order := make([]int, 0, len(v.order)-len(selected))
	picked := 0
	for index, number := range v.order {
		if picked < len(selected) && selected[picked] == index {
			v.reading.remove(number)
			v.printPreview.remove(number)
			picked++
		} else {
			order = append(order, number)
		}
	}
	v.rememberOrder()
	v.changePageOrder(order, nil, min(selected[0], len(order)-1))
}
