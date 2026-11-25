package main

import (
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
	lbGetAnchorIndex = 0x019d
	lbSetCaretIndex  = 0x019e
	lbGetCaretIndex  = 0x019f
)

func (v *documentView) buildPagePanel() {
	a := v.app
	a.beginPropertySection(uiMovePages, 16, uiMovePagesHelp)
	v.moveLabel = a.label("", 16, 42, 298, false)
	a.setUIMessage(v.moveLabel, uiTargetPosition)
	v.moveTarget = a.edit(16, 68, 76, false)
	v.moveButton = a.uiButton(uiMove, idReaderMoveTo, 100, 66, 214, 0)
	v.back = a.uiButton(uiMoveBefore, idReaderMoveBack, 16, 106, 145, 0)
	v.ahead = a.uiButton(uiMoveAfter, idReaderMoveAhead, 169, 106, 145, 0)
	a.beginPropertySection(uiRotateAndDelete, 151, uiRotateAndDeleteHelp)
	v.left = a.uiButton(uiRotateLeft, idReaderRotateLeft, 16, 180, 145, 0)
	v.right = a.uiButton(uiRotateRight, idReaderRotateRight, 169, 180, 145, 0)
	v.deleteButton = a.uiButton(uiDeleteSelection, idReaderDeletePages, 16, 220, 145, 0)
	v.undoButton = a.uiButton(uiUndoChange, idReaderUndo, 169, 220, 145, 0)
	a.addingSection = 0
	v.save = a.uiButton(uiSaveEditedCopy, idReaderSave, 16, 260, 298, 0)
	a.primaryButtons[v.save] = true
}

func (v *documentView) updatePageEditActions(editing bool) {
	selected := len(v.selectedPagePositions())
	editing = editing && selected > 0
	last := len(v.order) - selected + 1
	a := v.app
	message := localizedMessage{key: uiTargetPosition}
	if last > 1 {
		message = localizedMessage{key: uiTargetPositionRange, args: []any{last}}
	}
	if getText(v.moveLabel) != a.uiLocale.formatMessage(message) {
		a.setUIMessage(v.moveLabel, message.key, message.args...)
	}
	setEnabled(v.moveTarget, editing && last > 1)
	setEnabled(v.moveButton, editing && last > 1)
	setEnabled(v.deleteButton, editing && selected < len(v.order))
}

func (v *documentView) selectedPagePositions() []int {
	if v.overview || len(v.pageSelection) > 0 {
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
	if v.overview || len(v.pageSelection) > 0 {
		_, found := slices.BinarySearch(v.pageSelection, index)
		return found
	}
	return index == v.page
}

func (v *documentView) resetPageListSelection() {
	if len(v.order) == 0 {
		return
	}
	selected := v.selectedPagePositions()
	if len(selected) == len(v.order) {
		sendMessage.Call(v.pageList, lbSetSel, 1, ^uintptr(0))
	} else {
		sendMessage.Call(v.pageList, lbSetSel, 0, ^uintptr(0))
		for _, index := range selected {
			sendMessage.Call(v.pageList, lbSetSel, 1, uintptr(index))
		}
	}
	anchor := v.page
	if v.overview {
		anchor = min(max(v.overviewAnchor, 0), len(v.order)-1)
	}
	sendMessage.Call(v.pageList, lbSetAnchorIndex, uintptr(anchor), 0)
	sendMessage.Call(v.pageList, lbSetCaretIndex, uintptr(v.page), 1)
}

func (v *documentView) syncPageListSelection() {
	if v.overview || len(v.pageSelection) > 0 {
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
	if v.overview {
		anchor, _, _ := sendMessage.Call(v.pageList, lbGetAnchorIndex, 0, 0)
		v.overviewAnchor = min(max(int(int32(anchor)), 0), len(v.order)-1)
	}
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
			if _, found := slices.BinarySearch(selected, index); !found && !v.overview {
				index = selected[0]
			}
			v.pageSelection = selected
		}
	}
	if !v.secondary {
		v.app.applyThumbnailScope()
	}
	if v.overview || index != v.page {
		v.gotoPage(index, false)
	} else if !v.secondary {
		v.updateActions()
	}
}

func (v *documentView) selectOnlyPage(index int) {
	v.pageSelection = []int{index}
	if v.overview {
		v.overviewAnchor = index
	}
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
	selected := len(v.selectedPagePositions())
	if selected == 0 {
		return
	}
	limit := len(v.order) - selected + 1
	if limit <= 1 {
		return
	}
	position, err := strconv.Atoi(strings.TrimSpace(getText(v.moveTarget)))
	if err != nil || position < 1 || position > limit {
		notice(v.app.window, v.app.uiLocale.formatMessage(localizedMessage{key: uiTargetPositionError, args: []any{limit}}), 0x30)
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
	if !v.ready() || v.app.busy || len(selected) == 0 || len(selected) >= len(v.order) {
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
