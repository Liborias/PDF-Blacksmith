package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	idSearchTab = 260 + iota
	idSearchRun
	idSearchClear
	idSearchPrevious
	idSearchNext
	idSearchList
	idToggleSearchResults
	idCloseSearchResults
)

const wmSearchUpdate = 0x8004

const (
	uiResults      uiMessage = "ui.vysledky.83e98906cc"
	uiSearchPrompt uiMessage = "ui.zadejte.hledany.vyraz.ac8cfc6638"
)

type searchRequest struct {
	Input    string `json:"input"`
	Password string `json:"pdf_password,omitempty"`
	Query    string `json:"query"`
	Order    []int  `json:"order"`
}

type searchHit struct {
	Text  string       `json:"text"`
	Rects []searchRect `json:"rects"`
}

type searchPage struct {
	Number, Position, Rotation int
	Boxes                      map[string]searchRect
	Hits                       []searchHit
}

type searchUpdate struct {
	Status, Message string
	Scanned, Total  int
	Page            *searchPage
}

type searchRow struct {
	page  *searchPage
	index int
}

type searchTask struct {
	context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

type searchPanel struct {
	input, start, clear, previous, next, status uintptr
	heading, close, term, resultStatus, list    uintptr
	events                                      chan searchUpdate
	task                                        *searchTask
	query, pending                              string
	widest                                      int
	pages                                       map[int]*searchPage
	rows                                        []searchRow
}

func (v *documentView) buildSearchPanel() {
	a, p := v.app, &v.search
	p.events = make(chan searchUpdate, 8)
	a.beginPropertySection(uiDocumentText, 15, uiDocumentTextHelp)
	p.input = a.edit(16, 41, 298, false)
	sendMessage.Call(p.input, 0x00c5, 256, 0) // EM_LIMITTEXT
	p.start = a.uiButton(uiSearch, idSearchRun, 16, 79, 145, 0)
	a.primaryButtons[p.start] = true
	p.clear = a.uiButton(uiClear, idSearchClear, 169, 79, 145, 0)
	p.previous = a.uiButton(uiPreviousMatch, idSearchPrevious, 16, 119, 145, 0)
	p.next = a.uiButton(uiNextMatch, idSearchNext, 169, 119, 145, 0)
	p.status = a.readout(16, 161, 298, 52)
}

func (v *documentView) buildSearchResults() {
	a, p := v.app, &v.search
	p.heading = a.uiLabel(uiResults, 0, 0, 1, true)
	a.panelLabels[p.heading] = true
	p.close = a.button("×", idCloseSearchResults, 0, 0, 1, 0)
	a.panelButtons[p.close] = true
	a.bindUIHelp(p.close, uiClose)
	p.term = a.readout(0, 0, 1, 1)
	p.resultStatus = a.readout(0, 0, 1, 1)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100
	p.list = a.add("LISTBOX", "", idSearchList, 0, 0, 1, 1, wsTabstop|listStyle, 0, a.normalFont)
	a.bindUILabel(p.list, uiSearchResults)
	sendMessage.Call(p.list, 0x01a0, 0, uintptr(a.scale(26))) // LB_SETITEMHEIGHT
	a.addScrollPane(p.list, a.scale(26), true)
	a.setUIMessage(p.term, uiSearchPrompt)
}

func (v *documentView) preferredSearchWidth() int {
	d := v.app.scale
	return min(d(330), max(d(216), v.search.widest+d(36)))
}

func (v *documentView) layoutSearchResults() {
	a, p := v.app, &v.search
	area := a.rightAuxArea
	width, height := int(area.right-area.left), int(area.bottom-area.top)
	show := width > 0 && a.visibleRightAuxiliary() == auxiliaryResults
	term, status, list := a.scrollPanes[p.term], a.scrollPanes[p.resultStatus], a.scrollPanes[p.list]
	if !show {
		term.arrange(rect{}, false)
		status.arrange(rect{}, false)
		list.arrange(rect{}, false)
		showControl(a.searchSidebar, false)
		return
	}
	place(a.searchSidebar, int(area.left), int(area.top), width, height)
	d := a.scale
	margin := min(d(12), width/10)
	available := width - 2*margin
	place(p.heading, margin, d(8), available-d(34), d(22))
	place(p.close, width-margin-d(26), d(6), d(26), d(24))
	left, right := int32(margin), int32(width-margin)
	term.arrange(rect{left, int32(d(37)), right, int32(d(82))}, true)
	status.arrange(rect{left, int32(d(86)), right, int32(d(133))}, p.query != "")
	list.arrange(rect{left, int32(d(137)), right, int32(height - margin)}, len(p.rows) > 0)
	showControl(a.searchSidebar, true)
	user32.NewProc("RedrawWindow").Call(a.searchSidebar, 0, 0, 0x0085) // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN
}

func (a *app) syncSearchAuxiliary() {
	a.rightAux.syncSearch(a.mode == idSearchTab && a.view.ready(), &a.rightDock)
}

func (v *documentView) setSearchStatus(key uiMessage, args ...any) {
	for _, handle := range [...]uintptr{v.search.status, v.search.resultStatus} {
		if key == "" {
			setText(handle, "")
		} else {
			v.app.setUIMessage(handle, key, args...)
		}
		v.app.scrollPanes[handle].reflow()
	}
}

func (a *app) localizeHeaderSearch() {
	handle := a.commandBar.searchEdit
	if handle == 0 {
		return
	}
	cue := a.uiText(uiMessage("ui.najit.v.pdf.b75578bbb9"))
	sendMessage.Call(handle, 0x1501, 0, uintptr(unsafe.Pointer(wideString(cue)))) // EM_SETCUEBANNER
	a.setTooltip(handle, a.uiText(uiFind)+" (Ctrl+F)")
}

// The two fields represent one query. Guard synchronous EN_CHANGE notifications
// during programmatic changes, including tab restore and document reset.
func (v *documentView) setSearchInputs(query string) {
	a := v.app
	old := a.commandBar.searchSync
	a.commandBar.searchSync = true
	defer func() { a.commandBar.searchSync = old }()
	handles := []uintptr{v.search.input}
	if !v.secondary {
		handles = append(handles, a.commandBar.searchEdit)
	}
	for _, handle := range handles {
		if handle != 0 && getText(handle) != query {
			setText(handle, query)
		}
	}
}

func (a *app) searchInputChanged(handle uintptr) bool {
	if a.view == nil || handle == 0 ||
		(handle != a.commandBar.searchEdit && handle != a.view.search.input) {
		return false
	}
	if a.commandBar.searchSync || !a.view.ready() || a.busy {
		return true
	}
	query := getText(handle)
	tab := a.activeTab
	run := func() {
		if a.activeTab != tab || !a.view.ready() || a.busy || getText(handle) != query {
			return
		}
		if a.mode != idSearchTab {
			a.activate(idSearchTab)
		}
		if a.mode != idSearchTab {
			return
		}
		a.view.setSearchInputs(query)
		// Avoid relaunching the worker for whitespace-only formatting changes.
		if strings.Join(strings.Fields(query), " ") != a.view.search.query {
			a.view.queueSearch(query)
		}
	}
	if !a.saveInlineBefore(run) {
		run()
	}
	return true
}

func (a *app) showSearch() {
	if a.saveInlineBefore(a.showSearch) {
		return
	}
	if !a.view.ready() {
		return
	}
	if a.full {
		a.toggleFullscreen()
	}
	// A second Ctrl+F returns to the query without reopening a manually closed Results panel.
	if a.mode == idSearchTab && !a.rightDock.showsPrimary() {
		a.rightDock.revealPrimary()
	}
	a.activate(idSearchTab)
	input := a.view.search.input
	if visible, _, _ := user32.NewProc("IsWindowVisible").Call(a.commandBar.searchEdit); visible != 0 && !a.busy {
		input = a.commandBar.searchEdit
	} else {
		a.revealPropertyControl(input)
	}
	setFocus.Call(input)
	sendMessage.Call(input, 0x00b1, 0, ^uintptr(0)) // EM_SETSEL
	if input == a.view.search.input {
		a.revealFocusedControl()
	}
}

func (v *documentView) updateSearchActions() {
	p := &v.search
	setEnabled(p.start, v.ready() && !v.app.busy)
	setEnabled(p.input, v.ready() && !v.app.busy)
	if !v.secondary && v.app.commandBar.searchEdit != 0 {
		setEnabled(v.app.commandBar.searchEdit, v.ready() && !v.app.busy)
	}
	setEnabled(p.previous, len(p.rows) > 0)
	setEnabled(p.next, len(p.rows) > 0)
}

func (v *documentView) clearSearchResults() {
	p := &v.search
	p.rows, p.pages = nil, nil
	p.widest = 0
	sendMessage.Call(p.list, 0x0184, 0, 0) // LB_RESETCONTENT
	v.app.scrollPanes[p.list].reflow()
	invalidate(p.list)
	v.invalidateCanvas()
	v.updateSearchActions()
}

func (v *documentView) cancelSearch() {
	p := &v.search
	p.pending = ""
	if p.task != nil {
		p.task.cancel()
		v.setSearchStatus(uiSearchStopped)
	}
}

func (v *documentView) resetSearch() {
	p := &v.search
	v.cancelSearch()
	p.query = ""
	v.setSearchInputs("")
	v.setSearchStatus("")
	v.app.setUIMessage(p.term, uiSearchPrompt)
	v.clearSearchResults()
	v.layoutSearchResults()
}

func (v *documentView) queueSearch(query string) {
	if !v.ready() || v.app.busy {
		return
	}
	v.setSearchInputs(query)
	query = strings.Join(strings.Fields(query), " ")
	if query == "" {
		v.resetSearch()
		return
	}
	p := &v.search
	v.cancelSearch()
	p.query, p.pending = query, query
	v.clearSearchResults()
	setText(p.term, query)
	v.app.scrollPanes[p.term].reflow()
	v.setSearchStatus(uiSearchingDocument)
	v.layoutSearchResults()
	v.launchSearch()
}

func (v *documentView) launchSearch() {
	p := &v.search
	if p.task != nil || p.pending == "" || !v.ready() || v.app.busy {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	task := &searchTask{ctx, cancel, make(chan struct{})}
	p.task, p.pending = task, ""
	path, original := v.readSource()
	req := searchRequest{path, v.password, p.query, slices.Clone(v.order)}
	owner, events := v.app.window, p.events
	go func() {
		result := runPDFSearch(ctx, req, original, func(progress searchUpdate) {
			select {
			case events <- progress:
				postMessage.Call(owner, wmSearchUpdate, 0, 0)
			case <-ctx.Done():
			}
		})
		close(task.done)
		events <- result
		postMessage.Call(owner, wmSearchUpdate, 0, 0)
	}()
}

func runPDFSearch(ctx context.Context, req searchRequest, original os.FileInfo, report func(searchUpdate)) searchUpdate {
	failure := searchUpdate{Status: "error", Message: "Komunikace s vyhledáváním v PDF se nezdařila."}
	changed := searchUpdate{Status: "changed", Message: "Vstupní PDF se změnilo. Otevřete je znovu."}
	if !unchanged(req.Input, original) {
		return changed
	}
	cmd, err := engineCommandContext(ctx, "--search")
	if err != nil {
		failure.Message = err.Error()
		return failure
	}
	cmd.SysProcAttr.CreationFlags |= 0x00004000 // BELOW_NORMAL_PRIORITY_CLASS
	payload, err := json.Marshal(req)
	if err != nil {
		return failure
	}
	cmd.Stdin = bytes.NewReader(payload)
	stdout, err := cmd.StdoutPipe()
	if err != nil || cmd.Start() != nil {
		return failure
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	result := failure
	for scanner.Scan() {
		var event searchUpdate
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			cmd.Process.Kill()
			break
		}
		if event.Status == "progress" {
			report(event)
		} else {
			result = event
		}
	}
	if scanner.Err() != nil {
		cmd.Process.Kill()
	}
	if cmd.Wait() != nil {
		return failure
	}
	if !unchanged(req.Input, original) {
		return changed
	}
	return result
}

func (v *documentView) receiveSearch(event searchUpdate) {
	p := &v.search
	if event.Status == "progress" {
		if p.task.Err() == nil && v.ready() {
			if event.Page != nil {
				v.addSearchPage(event.Page)
			}
			v.setSearchStatus(uiSearchProgress, event.Scanned, len(v.order), event.Total)
		}
		return
	}
	cancelled := p.task.Err() != nil
	p.task.cancel()
	p.task = nil
	if !cancelled {
		switch event.Status {
		case "ok":
			switch {
			case event.Total == 0:
				v.setSearchStatus(uiNoSearchMatch)
			case event.Total > len(p.rows):
				v.setSearchStatus(uiSearchLimit, event.Total, len(p.rows))
			default:
				v.setSearchStatus(uiSearchCount, event.Total)
			}
		case "error", "changed":
			v.clearSearchResults()
			for _, handle := range [...]uintptr{p.status, p.resultStatus} {
				setText(handle, multiline(event.Message))
				v.app.scrollPanes[handle].reflow()
			}
		}
		if p.widest == 0 && v.app.visibleRightAuxiliary() == auxiliaryResults {
			v.app.layout()
		}
	}
	v.launchSearch()
	v.app.startOverwrite()
}

func (v *documentView) addSearchPage(page *searchPage) {
	p := &v.search
	firstResult := len(p.rows) == 0
	oldWidth := v.preferredSearchWidth()
	if p.pages == nil {
		p.pages = make(map[int]*searchPage)
	}
	current := p.pages[page.Number]
	if current == nil {
		current = &searchPage{Number: page.Number, Position: page.Position, Rotation: page.Rotation, Boxes: page.Boxes}
		p.pages[page.Number] = current
	}
	first := len(current.Hits)
	current.Hits = append(current.Hits, page.Hits...)
	sendMessage.Call(p.list, 0x000b, 0, 0) // WM_SETREDRAW
	for index, hit := range page.Hits {
		p.rows = append(p.rows, searchRow{current, first + index})
		label := fmt.Sprintf("%d · %s", page.Position, hit.Text)
		if p.widest < v.app.scale(294) {
			p.widest = max(p.widest, min(v.app.scale(294), v.app.captionWidth(label, v.app.normalFont)))
		}
		sendMessage.Call(p.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(label)))) // LB_ADDSTRING
	}
	sendMessage.Call(p.list, 0x000b, 1, 0)
	pane := v.app.scrollPanes[p.list]
	if !pane.shown {
		showControl(p.list, false)
	}
	pane.reflow()
	invalidate(p.list)
	v.updateSearchActions()
	if v.app.visibleRightAuxiliary() == auxiliaryResults && (firstResult || oldWidth != v.preferredSearchWidth()) {
		v.app.layout()
	}
	if firstResult && v.app.mode == idSearchTab {
		v.jumpToSearch(0)
	} else {
		v.invalidateCanvas()
	}
}

func (p *searchPanel) selected() int {
	index, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0) // LB_GETCURSEL
	if index >= uintptr(len(p.rows)) {
		return -1
	}
	return int(index)
}

func (v *documentView) jumpToSearch(row int) {
	p := &v.search
	if !v.ready() || row < 0 || row >= len(p.rows) {
		return
	}
	item := p.rows[row]
	position := item.page.Position - 1
	if position < 0 || position >= len(v.order) || v.order[position] != item.page.Number {
		return
	}
	if v.app.saveInlineBefore(func() {
		if position < len(v.order) && v.order[position] == item.page.Number {
			v.gotoPage(position, false)
		}
	}) {
		return
	}
	sendMessage.Call(p.list, 0x0186, uintptr(row), 0) // LB_SETCURSEL
	box := item.page.Boxes[v.box()]
	angle := item.page.Rotation + v.rotations[item.page.Number-1]
	for _, bounds := range item.page.Hits[item.index].Rects {
		if r, ok := searchDisplayRect(bounds, box, angle); ok {
			anchor := readingAnchor{position, (r[0] + r[2]) / 2, (r[1] + r[3]) / 2}
			v.gotoPageAt(position, false, &anchor)
			return
		}
	}
	v.gotoPage(position, false)
}

func (v *documentView) stepSearch(step int) {
	p := &v.search
	if len(p.rows) == 0 {
		return
	}
	current := p.selected()
	if current < 0 && step < 0 {
		current = 0
	}
	v.jumpToSearch((current + step + len(p.rows)) % len(p.rows))
}

func (v *documentView) searchCommand(id, notification int) bool {
	p := &v.search
	if id == idSearchList && notification == 1 { // LBN_SELCHANGE
		v.jumpToSearch(p.selected())
		return true
	}
	if notification != 0 {
		return false
	}
	switch id {
	case idSearchRun:
		v.queueSearch(getText(p.input))
	case idSearchClear:
		v.resetSearch()
		v.app.layout()
		setFocus.Call(p.input)
	case idSearchPrevious:
		v.stepSearch(-1)
	case idSearchNext:
		v.stepSearch(1)
	default:
		return false
	}
	return true
}

func (v *documentView) drawSearchRow(item *buttonDraw) bool {
	p := &v.search
	if item.controlType != 2 || item.handle != p.list {
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, panelColor)
	if item.itemID >= uint32(len(p.rows)) {
		return true
	}
	row := p.rows[item.itemID]
	foreground := textColor
	if item.state&1 != 0 {
		foreground = selectionColor
	}
	bounds.left += int32(v.app.scale(7))
	bounds.right -= int32(v.app.scale(6))
	drawText(item.dc, fmt.Sprintf("%d · %s", row.page.Position, row.page.Hits[row.index].Text), bounds, v.app.normalFont, foreground, 0x8000)
	if item.state&16 != 0 {
		v.app.drawFocus(item.dc, item.bounds)
	}
	return true
}

var alphaBlend = syscall.NewLazyDLL("msimg32.dll").NewProc("AlphaBlend")

func (v *documentView) paintSearchHighlights(dc uintptr, number int, bounds rect) {
	p := &v.search
	page := p.pages[number]
	if page == nil || (v.app.mode != idSearchTab && v.app.visibleRightAuxiliary() != auxiliaryResults) {
		return
	}
	mask, _, _ := createSurfaceDC.Call(dc)
	if mask == 0 {
		return
	}
	defer deleteSurfaceDC.Call(mask)
	bitmap, _, _ := createSurfaceBitmap.Call(dc, 1, 1)
	if bitmap == 0 {
		return
	}
	defer gdiDelete.Call(bitmap)
	old, _, _ := gdiSelect.Call(mask, bitmap)
	if old == 0 || old == ^uintptr(0) {
		return
	}
	defer gdiSelect.Call(mask, old)
	box, angle := page.Boxes[v.box()], page.Rotation+v.rotations[number-1]
	w, h := float64(bounds.right-bounds.left), float64(bounds.bottom-bounds.top)
	selected := -1
	if index := p.selected(); index >= 0 && p.rows[index].page == page {
		selected = p.rows[index].index
	}
	paint := func(hit searchHit, opacity uintptr) {
		for _, raw := range hit.Rects {
			if r, ok := searchDisplayRect(raw, box, angle); ok {
				x, y := bounds.left+int32(math.Floor(w*r[0])), bounds.top+int32(math.Floor(h*r[1]))
				right, bottom := bounds.left+int32(math.Ceil(w*r[2])), bounds.top+int32(math.Ceil(h*r[3]))
				alphaBlend.Call(dc, uintptr(x), uintptr(y), uintptr(right-x), uintptr(bottom-y), mask, 0, 0, 1, 1, opacity<<16)
			}
		}
	}
	fillColor(mask, rect{right: 1, bottom: 1}, paletteColor{255, 200, 78})
	for index, hit := range page.Hits {
		if index != selected {
			paint(hit, 103)
		}
	}
	if selected >= 0 {
		fillColor(mask, rect{right: 1, bottom: 1}, selectionColor)
		paint(page.Hits[selected], 133)
	}
}
