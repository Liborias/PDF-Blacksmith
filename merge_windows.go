package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unsafe"
)

const (
	idMergeTab = 220 + iota
	idMergeAdd
	idMergeAddOpen
	idMergeAddSelected
	idMergeList
	idMergeUp
	idMergeDown
	idMergeRemove
	idMergeInspect
	idMergeRun
)

type mergeInput struct {
	Path        string       `json:"path"`
	Password    string       `json:"pdf_password,omitempty"`
	SourceSize  int64        `json:"source_size"`
	SourceMtime int64        `json:"source_mtime_ns"`
	View        *viewRequest `json:"view,omitempty"`
}

type mergeSourceInfo struct {
	Pages   int    `json:"pages"`
	Message string `json:"message"`
}

type mergeEntry struct {
	input     mergeInput
	info      mergeSourceInfo
	pageRange string
}

type mergePanel struct {
	entries                                   []mergeEntry
	active                                    int
	probing                                   []int
	add, addOpen, addSelected, list           uintptr
	up, down, remove, sourceStatus            uintptr
	pageRange, password, inspect, run, report uintptr
}

func (m *mergePanel) selected() *mergeEntry {
	if m.active < 0 || m.active >= len(m.entries) {
		return nil
	}
	return &m.entries[m.active]
}

func (a *app) buildMergePanel() {
	m := &a.merge
	m.active = -1
	a.beginPropertySection("Zdroje ke sloučení", 16, "Přidejte soubory nebo vybrané stránky a upravte pořadí jejich vložení.")
	m.add = a.button("Přidat PDF soubory…", idMergeAdd, 16, 42, 298, 0)
	m.addOpen = a.button("Přidat otevřené PDF", idMergeAddOpen, 16, 80, 145, 0)
	m.addSelected = a.button("Přidat výběr stran", idMergeAddSelected, 169, 80, 145, 0)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100
	m.list = a.add("LISTBOX", "Soubory ke sloučení", idMergeList, 16, 122, 298, 96, wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(m.list, 0x01a0, 0, uintptr(a.scale(24))) // LB_SETITEMHEIGHT
	a.addScrollPane(m.list, a.scale(24), true)
	m.up = a.button("Nahoru", idMergeUp, 16, 224, 94, 0)
	m.down = a.button("Dolů", idMergeDown, 118, 224, 94, 0)
	m.remove = a.button("Odebrat", idMergeRemove, 220, 224, 94, 0)
	section := a.addingSection
	a.addingSection = 0
	m.sourceStatus = a.add("STATIC", "Strany k přidání", 0, 16, 266, 298, 22, 0x4000, 0, a.normalFont)
	a.addingSection = section
	m.pageRange = a.edit(16, 292, 298, false)
	sendMessage.Call(m.pageRange, 0x1501, 1, uintptr(unsafe.Pointer(wideString("Všechny strany · např. 1, 3–5"))))
	a.label("Heslo k otevření vybraného souboru", 16, 335, 298, false)
	m.password = a.edit(16, 359, 145, true)
	m.inspect = a.button("Načíst počet stran", idMergeInspect, 169, 357, 145, 0)
	a.beginPropertySection("Výstup", 404, "Sloučený dokument se uloží jako nový soubor.")
	m.run = a.button("Uložit sloučené PDF…", idMergeRun, 16, 432, 298, 0)
	a.primaryButtons[m.run] = true
	m.report = a.readout(16, 475, 298, 122)
}

func (a *app) updateMergeActions() {
	m := &a.merge
	entry := m.selected()
	editable := !a.busy && entry != nil
	setEnabled(m.add, !a.busy)
	a.updateMergeInputButtons()
	setEnabled(m.list, !a.busy)
	setEnabled(m.up, editable && m.active > 0)
	setEnabled(m.down, editable && m.active+1 < len(m.entries))
	setEnabled(m.remove, editable)
	setEnabled(m.pageRange, editable && entry.info.Pages > 0)
	setEnabled(m.password, editable && entry.input.View == nil)
	setEnabled(m.inspect, editable && entry.input.View == nil)
	ready := !a.busy && len(m.entries) > 1
	for _, source := range m.entries {
		if source.info.Pages == 0 {
			ready = false
			break
		}
	}
	setEnabled(m.run, ready)
}

func (a *app) updateMergeInputButtons() {
	ready := !a.busy && a.view.ready()
	setEnabled(a.merge.addOpen, ready)
	setEnabled(a.merge.addSelected, ready && len(a.view.pageSelection) > 0)
}

func (a *app) rememberMergeFields() {
	m := &a.merge
	if entry := m.selected(); entry != nil {
		entry.pageRange = getText(m.pageRange)
		entry.input.Password = getText(m.password)
	}
}

func (a *app) selectMergeEntry(index int) {
	m := &a.merge
	m.active = index
	entry := m.selected()
	status, pages, password := "Strany k přidání", "", ""
	if entry != nil {
		pages, password = entry.pageRange, entry.input.Password
		if entry.info.Pages > 0 {
			status = fmt.Sprintf("Strany k přidání · ve zdroji %d", entry.info.Pages)
		} else if entry.info.Message != "" {
			status = entry.info.Message
		}
	}
	setText(m.sourceStatus, status)
	setText(m.pageRange, pages)
	setText(m.password, password)
	a.updateMergeActions()
}

func (a *app) refreshMergeList() {
	m := &a.merge
	sendMessage.Call(m.list, 0x000b, 0, 0) // WM_SETREDRAW
	sendMessage.Call(m.list, 0x0184, 0, 0) // LB_RESETCONTENT
	for _, entry := range m.entries {
		sendMessage.Call(m.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(entry.input.Path))))
	}
	sendMessage.Call(m.list, 0x0186, uintptr(m.active), 0)
	sendMessage.Call(m.list, 0x000b, 1, 0)
	pane := a.scrollPanes[m.list]
	if !pane.shown {
		showControl(m.list, false)
	}
	pane.reflow()
	invalidate(m.list)
	a.selectMergeEntry(m.active)
}

func (a *app) drawMergeRow(item *buttonDraw) bool {
	if item.controlType != 2 || item.handle != a.merge.list {
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, panelColor)
	if int(item.itemID) >= len(a.merge.entries) {
		return true
	}
	entry := a.merge.entries[item.itemID]
	color := textColor
	if item.state&1 != 0 {
		color = selectionColor
	}
	d := a.scale
	bounds.left += int32(d(7))
	bounds.right -= int32(d(6))
	if entry.info.Pages > 0 {
		drawText(item.dc, fmt.Sprintf("%d str.", entry.info.Pages), bounds, a.normalFont, color, 2)
		bounds.right -= int32(d(63))
	}
	drawText(item.dc, filepath.Base(entry.input.Path), bounds, a.normalFont, color, 0x8000)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (a *app) addMergeFiles(paths []string) {
	if a.busy || len(paths) == 0 {
		return
	}
	m := &a.merge
	a.rememberMergeFields()
	var added []int
	for _, path := range paths {
		absolute, pathErr := filepath.Abs(path)
		if pathErr != nil {
			a.report("Vyberte soubor s platnou cestou.")
			continue
		}
		info, err := os.Stat(absolute)
		if err != nil || !info.Mode().IsRegular() {
			a.report("Soubor nelze otevřít.\n" + absolute)
			continue
		}
		added = append(added, len(m.entries))
		m.entries = append(m.entries, mergeEntry{input: mergeInput{
			Path: absolute, SourceSize: info.Size(), SourceMtime: info.ModTime().UnixNano(),
		}})
	}
	if len(added) > 0 {
		m.active = added[0]
		a.refreshMergeList()
		a.inspectMergeSources(added)
	}
}

func (a *app) addOpenPDFToMerge(selectionOnly bool) {
	v := a.view
	if a.busy || !v.ready() || !v.sourceUnchanged() || (selectionOnly && len(v.pageSelection) == 0) {
		return
	}
	a.rememberMergeFields()
	req := v.request("merge")
	v.includeEdits(&req)
	path, info := v.readSource()
	entry := mergeEntry{input: mergeInput{Path: path, Password: v.password, View: req.View,
		SourceSize: info.Size(), SourceMtime: info.ModTime().UnixNano()},
		info: mergeSourceInfo{Pages: len(v.order)}}
	if selectionOnly {
		entry.pageRange = pageRanges(pageNumbers(v.pageSelection))
	}
	m := &a.merge
	m.active = len(m.entries)
	m.entries = append(m.entries, entry)
	a.refreshMergeList()
}

func (a *app) inspectMergeSources(indices []int) {
	m := &a.merge
	req := request{Action: "merge_probe", MergeInputs: make([]mergeInput, 0, len(indices))}
	for _, index := range indices {
		req.MergeInputs = append(req.MergeInputs, m.entries[index].input)
	}
	if a.beginJob(req, "Načítám počet stran…") {
		m.probing = indices
		a.launchJob(func() response { return runOperation(req) })
	}
}

func (a *app) finishMergeProbe(result response) {
	m := &a.merge
	if result.Status == "ok" && result.View != nil && len(result.View.MergeInfo) == len(m.probing) {
		for i, index := range m.probing {
			m.entries[index].info = result.View.MergeInfo[i]
		}
		a.reportStatus("")
	} else {
		setText(m.report, multiline(result.Message))
		a.reportStatus(result.Message)
	}
	m.probing = nil
	a.refreshMergeList()
}

func (a *app) mergeCommand(id, notification int) bool {
	m := &a.merge
	if id == idMergeList && notification == 1 {
		if !a.busy {
			a.rememberMergeFields()
			index, _, _ := sendMessage.Call(m.list, 0x0188, 0, 0) // LB_GETCURSEL
			a.selectMergeEntry(int(index))
		}
		return true
	}
	if notification != 0 || id < idMergeAdd || id > idMergeRun {
		return false
	}
	if a.busy {
		return true
	}
	switch id {
	case idMergeAdd:
		a.addMergeFiles(a.fileDialogPaths(pdfFilter, "", "", false, true))
	case idMergeAddOpen, idMergeAddSelected:
		a.addOpenPDFToMerge(id == idMergeAddSelected)
	case idMergeUp, idMergeDown, idMergeRemove:
		if m.selected() != nil {
			a.rememberMergeFields()
			if id == idMergeRemove {
				m.entries = slices.Delete(m.entries, m.active, m.active+1)
				m.active = min(m.active, len(m.entries)-1)
			} else {
				destination := m.active + 1
				if id == idMergeUp {
					destination = m.active - 1
				}
				if destination >= 0 && destination < len(m.entries) {
					m.entries[m.active], m.entries[destination] = m.entries[destination], m.entries[m.active]
					m.active = destination
				}
			}
			a.refreshMergeList()
		}
	case idMergeInspect:
		if entry := m.selected(); entry != nil && entry.input.View == nil {
			a.rememberMergeFields()
			a.inspectMergeSources([]int{m.active})
		}
	case idMergeRun:
		a.saveMergedPDF()
	}
	return true
}

func (a *app) saveMergedPDF() {
	m := &a.merge
	if len(m.entries) < 2 {
		return
	}
	a.rememberMergeFields()
	req := request{Action: "merge", Input: m.entries[0].input.Path}
	paths := make([]string, 0, len(m.entries))
	for index, entry := range m.entries {
		input := entry.input
		if entry.info.Pages == 0 {
			notice(a.window, "Nejprve načtěte počet stran.\n"+filepath.Base(input.Path), 0x30)
			return
		}
		view := viewRequest{}
		if input.View != nil {
			view = *input.View
		}
		if pages := strings.TrimSpace(entry.pageRange); pages != "" {
			var err error
			view.Pages, err = parseExportPages(pages, entry.info.Pages)
			if err != nil {
				m.active = index
				a.refreshMergeList()
				notice(a.window, filepath.Base(input.Path)+"\n"+err.Error(), 0x30)
				setFocus.Call(m.pageRange)
				return
			}
		}
		input.View = &view
		req.MergeInputs = append(req.MergeInputs, input)
		paths = append(paths, input.Path)
	}
	if a.chooseOutput(&req, "-slouceno", paths[1:]...) {
		setText(m.report, "")
		a.run(req)
	}
}
