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
)

const wmSearchUpdate = 0x8004

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
	input, start, clear, previous, next, status, list uintptr
	events                                            chan searchUpdate
	task                                              *searchTask
	query, pending                                    string
	widest                                            int
	pages                                             map[int]*searchPage
	rows                                              []searchRow
}

func (v *documentView) buildSearchPanel() {
	a, p := v.app, &v.search
	p.events = make(chan searchUpdate, 8)
	a.label("Text v dokumentu", 16, 15, 298, true)
	p.input = a.edit(16, 41, 298, false)
	sendMessage.Call(p.input, 0x00c5, 256, 0) // EM_LIMITTEXT
	p.start = a.button("Hledat", idSearchRun, 16, 79, 145, 0)
	a.primaryButtons[p.start] = true
	p.clear = a.button("Vymazat", idSearchClear, 169, 79, 145, 0)
	p.previous = a.button("Předchozí shoda", idSearchPrevious, 16, 119, 145, 0)
	p.next = a.button("Další shoda", idSearchNext, 169, 119, 145, 0)
	p.status = a.readout(16, 161, 298, 52)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100
	p.list = a.add("LISTBOX", "Výsledky hledání", idSearchList, 16, 221, 298, 116, wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(p.list, 0x01a0, 0, uintptr(a.scale(26))) // LB_SETITEMHEIGHT
	a.addScrollPane(p.list, a.scale(26), true)
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
	a.activate(idSearchTab)
	input := a.view.search.input
	setFocus.Call(input)
	sendMessage.Call(input, 0x00b1, 0, ^uintptr(0)) // EM_SETSEL
	a.revealFocusedControl()
}

func (v *documentView) updateSearchActions() {
	p := &v.search
	setEnabled(p.start, v.ready() && !v.app.busy)
	setEnabled(p.input, v.ready())
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
		setText(p.status, "Hledání zastaveno.")
	}
}

func (v *documentView) resetSearch() {
	p := &v.search
	v.cancelSearch()
	p.query = ""
	setText(p.input, "")
	setText(p.status, "")
	v.clearSearchResults()
}

func (v *documentView) queueSearch(query string) {
	if !v.ready() || v.app.busy {
		return
	}
	query = strings.Join(strings.Fields(query), " ")
	if query == "" {
		v.resetSearch()
		return
	}
	p := &v.search
	v.cancelSearch()
	p.query, p.pending = query, query
	v.clearSearchResults()
	setText(p.status, "Prohledávám dokument…")
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
			setText(p.status, fmt.Sprintf("Prohledáno %d z %d stran. Shod zatím %d.", event.Scanned, len(v.order), event.Total))
		}
		return
	}
	cancelled := p.task.Err() != nil
	p.task.cancel()
	p.task = nil
	if !cancelled {
		switch event.Status {
		case "ok":
			message := fmt.Sprintf("Nalezené shody: %d", event.Total)
			if event.Total == 0 {
				message = "Žádná shoda v textové vrstvě PDF."
			} else if event.Total > len(p.rows) {
				message += fmt.Sprintf("\nZobrazeno prvních %d shod.", len(p.rows))
			}
			setText(p.status, multiline(message))
		case "error", "changed":
			v.clearSearchResults()
			setText(p.status, multiline(event.Message))
		}
		if a := v.app; a.mode == idSearchTab && !a.full && !a.detailsCollapsed && p.widest == 0 && a.settingsWidth > a.scale(panelWidth) {
			a.layout()
		}
	}
	v.launchSearch()
	v.app.startOverwrite()
}

func (v *documentView) addSearchPage(page *searchPage) {
	p := &v.search
	firstResult := len(p.rows) == 0
	oldWidth := p.widest
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
		if p.widest < v.app.scale(374) {
			p.widest = max(p.widest, min(v.app.scale(374), v.app.captionWidth(label, v.app.normalFont)))
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
	if v.app.mode == idSearchTab {
		if oldWidth != p.widest && !v.app.detailsCollapsed && !v.app.full {
			v.app.layout()
		}
		if firstResult {
			v.jumpToSearch(0)
		} else {
			v.invalidateCanvas()
		}
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
	if page == nil || v.app.mode != idSearchTab {
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
		fillColor(mask, rect{right: 1, bottom: 1}, paletteColor{225, 96, 140})
		paint(page.Hits[selected], 133)
	}
}
