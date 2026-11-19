package main

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"unsafe"
)

type pageInsertion struct {
	Source  mergeInput `json:"source"`
	Before  int        `json:"before"`
	Convert bool       `json:"convert,omitempty"`
}

type insertionColors struct {
	Target   string
	Reason   string
	Warnings []string
	Convert  int
	Gray     int
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
	req := v.request("insert_probe")
	req.InsertPages = &pageInsertion{Source: input, Before: before}
	if v.hasPageEdits() {
		v.includeEdits(&req)
	}
	a.comparison.canvasFocused = false
	setFocus.Call(a.window)
	if a.beginJob(req, "Zjišťuji barevné prostory vkládaných stran…") {
		a.launchJob(func() response { return runOperation(req) })
	}
}

const (
	insertKeep    = 100
	insertConvert = 101
)

func (a *app) finishInsertProbe(req request, result response) {
	if result.Status != "ok" || result.View == nil || result.View.InsertColor == nil {
		a.reportStatus("Barvy vkládaných stran se nepodařilo ověřit.")
		if result.Message != "" {
			notice(a.window, result.Message, 0x30)
		}
		return
	}
	colors := result.View.InsertColor
	switch a.askInsertionColors(colors) {
	case insertKeep:
	case insertConvert:
		req.InsertPages.Convert = true
	default:
		a.reportStatus("Vložení stránek zrušeno.")
		return
	}
	req.Action = "insert_pages"
	status := "Vkládám stránky do cílového PDF…"
	if req.InsertPages.Convert {
		status = "Převádím barvy vkládaných stran…"
	}
	a.runDocumentDraft(req, status, 0)
}

func (a *app) askInsertionColors(colors *insertionColors) int32 {
	var details []string
	if colors.Target != "" {
		details = append(details, "Cílový prostor: "+colors.Target)
	}
	if colors.Convert > 0 {
		details = append(details, fmt.Sprintf("Stran k převodu: %d. Šedé a pouze K: %d.", colors.Convert, colors.Gray))
		for _, warning := range colors.Warnings {
			details = append(details, "• "+warning)
		}
	} else if colors.Reason != "" {
		details = append(details, colors.Reason)
	}
	content := strings.Join(details, "\n")
	if chosen, ok := a.insertionTaskDialog(content, colors.Convert > 0); ok {
		return chosen
	}
	if colors.Convert > 0 {
		switch notice(a.window, content+"\n\nAno – převést do prostoru cíle.\nNe – zachovat původní barvy.", 0x223) {
		case 6:
			return insertConvert
		case 7:
			return insertKeep
		}
	} else if notice(a.window, content+"\n\nVložit stránky se zachováním barev?", 0x31) == 1 {
		return insertKeep
	}
	return 0
}

func (a *app) insertionTaskDialog(content string, canConvert bool) (int32, bool) {
	proc := comctl32.NewProc("TaskDialogIndirect")
	if proc.Find() != nil {
		return 0, false
	}
	// commctrl.h specifies byte alignment for TASKDIALOGCONFIG and TASKDIALOG_BUTTON.
	type address [unsafe.Sizeof(uintptr(0))]byte
	packAddress := func(value uintptr) (packed address) {
		for i := range packed {
			packed[i] = byte(value >> (8 * i))
		}
		return packed
	}
	packNumber := func(value uint32) (packed [4]byte) {
		binary.LittleEndian.PutUint32(packed[:], value)
		return packed
	}
	var pin runtime.Pinner
	defer pin.Unpin()
	packText := func(text string) address {
		utf16 := wideString(text)
		pin.Pin(utf16)
		return packAddress(uintptr(unsafe.Pointer(utf16)))
	}
	type taskButton struct {
		id   [4]byte
		text address
	}
	type taskDialog struct {
		size                                                [4]byte
		parent, instance                                    address
		flags, common                                       [4]byte
		title, icon, heading, content                       address
		count                                               [4]byte
		buttons                                             address
		defaultButton, radioCount                           [4]byte
		radioButtons                                        address
		defaultRadio                                        [4]byte
		verification, expanded, expandedText, collapsedText address
		footerIcon, footer, callback, callbackData          address
		width                                               [4]byte
	}
	buttons := []taskButton{{packNumber(insertKeep), packText("Zachovat původní barvy")}}
	if canConvert {
		buttons = append(buttons, taskButton{packNumber(insertConvert), packText("Převést do prostoru cíle")})
	}
	pin.Pin(&buttons[0])
	config := taskDialog{parent: packAddress(a.window), flags: packNumber(8), common: packNumber(8),
		title: packText(appName), heading: packText("Barvy vkládaných stránek"), content: packText(content),
		count: packNumber(uint32(len(buttons))), buttons: packAddress(uintptr(unsafe.Pointer(&buttons[0]))),
		defaultButton: packNumber(insertKeep), width: packNumber(274)}
	if canConvert {
		config.icon = packAddress(0xffff) // TD_WARNING_ICON
	}
	config.size = packNumber(uint32(unsafe.Sizeof(config)))
	var chosen int32
	status, _, _ := proc.Call(uintptr(unsafe.Pointer(&config)), uintptr(unsafe.Pointer(&chosen)), 0, 0)
	return chosen, uint32(status) == 0
}
