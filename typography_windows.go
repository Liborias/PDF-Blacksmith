package main

import (
	"slices"
	"syscall"
	"unicode/utf16"
)

const (
	idTextSmaller = 521 + iota
	idTextLarger
	idTextSizes
	idTextWidths
	idTextLeadings
	idTextTracking
)

var typographicSizes = []float64{6, 7, 8, 9, 10, 11, 12, 14, 16, 18, 20, 22, 24, 28, 32, 36, 40, 48, 60, 72, 96, 120, 144, 200, 300, 450, 600}

const mixedTextNumber = "—"

func (a *app) typographyCommand(id int) bool {
	if id < idTextSmaller || id > idTextTracking {
		return false
	}
	p := &a.textTool
	if frame := a.inlineCurrent(); frame == nil || !frame.editable || a.busy {
		return true
	}
	if id == idTextSmaller || id == idTextLarger {
		a.stepInlineSizes(id == idTextLarger)
		a.focusInlineEditor()
		return true
	}
	fields := [...]uintptr{p.size, p.width, p.leading, p.tracking}
	presets := [...][]float64{
		typographicSizes, {0, 40, 60, 90, 120, 160, 200, 250, 300, 400, 500},
		{.8, 1, 1.15, 1.2, 1.3, 1.5, 1.75, 2, 2.5, 3}, {-2, -1, -.5, -.25, 0, .25, .5, 1, 2, 4},
	}
	index := id - idTextSizes
	var entries []popupItem
	for _, value := range presets[index] {
		caption := formatTextNumber(value) + " pt"
		if id == idTextWidths && value == 0 {
			caption = "Automaticky"
		} else if id == idTextLeadings {
			caption = formatTextNumber(value) + "×"
		}
		entries = append(entries, popupItem{label: caption})
	}
	chosen := a.trackPopup(p.typography[id-idTextSmaller], 144, entries)
	if chosen >= 0 {
		value := formatTextNumber(presets[index][chosen])
		if id == idTextWidths && chosen == 0 {
			value = ""
		}
		setText(fields[index], value)
	}
	a.focusInlineEditor()
	return true
}

func (a *app) syncInlineControls(edited uintptr) {
	p, e := &a.textTool, &a.textTool.inline
	if a.inlineCurrent() == nil || !e.visible || e.loading || e.layoutPosted || a.busy {
		return
	}
	e.loading = true
	defer func() { e.loading = false }()
	start, end := inlineSelection(e.handle)
	// Without a character selection the Typography commands address the active frame.
	wholeFrame := start == end && a.inlineValue() != ""
	if wholeFrame {
		sendMessage.Call(e.handle, 0x000b, 0, 0) // WM_SETREDRAW
		selectInline(e.handle, 0, -1)
	}
	format := readInlineFormat(e.handle)
	if format.mask&inlineTracking != 0 && (wholeFrame || end-start > 1) {
		document := inlineTextDocument(e.handle)
		defer document.release()
		selection := document.selection()
		defer selection.release()
		if selection.mixedSpacing() {
			format.mask &^= inlineTracking
		}
	}
	if wholeFrame {
		selectInline(e.handle, start, end)
		sendMessage.Call(e.handle, 0x000b, 1, 0)
	}
	e.mixedFormat = inlineBaseMask &^ format.mask
	choice := a.choices[p.font]
	if e.mixedFormat&inlineFace != 0 {
		if choice.placeholder == "" {
			choice.placeholder = uiMixedTypography
			setText(p.font, a.choiceLabel(choice))
			invalidate(p.font)
		}
	} else {
		index := slices.Index(p.fontKeys, syscall.UTF16ToString(format.face[:]))
		if index < 0 {
			index = choice.index
		}
		chooseCombo(p.font, index)
	}
	for _, field := range [...]struct {
		control uintptr
		mask    uint32
		value   float64
	}{
		{p.size, inlineSize, float64(format.height) / 20},
		{p.tracking, inlineTracking, float64(format.spacing) / 20},
	} {
		if field.control != edited {
			value := formatTextNumber(field.value)
			if e.mixedFormat&field.mask != 0 {
				value = mixedTextNumber
			}
			if getText(field.control) != value {
				setText(field.control, value)
			}
		}
	}
	for flag, control := range map[uint32]uintptr{inlineBold: p.bold, inlineItalic: p.italic, inlineLine: p.underline} {
		a.setChecked(control, format.mask&flag != 0 && format.effects&flag != 0)
	}
	a.setChecked(p.regular, format.mask&(inlineBold|inlineItalic) == inlineBold|inlineItalic &&
		format.effects&(inlineBold|inlineItalic) == 0)
}

func (a *app) stepInlineSizes(larger bool) {
	e := &a.textTool.inline
	start, end := inlineSelection(e.handle)
	left, right := start, end
	if left == right {
		left, right = 0, len(utf16.Encode([]rune(a.inlineValue())))
		if right == 0 {
			left, right = start, end
		}
	}
	e.loading = true
	sendMessage.Call(e.handle, 0x000b, 0, 0) // WM_SETREDRAW
	var changes []inlineTextRun
	visitInlineFormats(e.handle, left, right, inlineSize, nil, func(from, to int, format inlineCharFormat) {
		size := float64(format.height) / 20
		index, exact := slices.BinarySearch(typographicSizes, size)
		switch {
		case !larger:
			index--
		case exact:
			index++
		}
		if size > 0 && index >= 0 && index < len(typographicSizes) {
			changes = appendInlineRun(changes, inlineTextRun{Size: typographicSizes[index]}, from, to)
		}
	})
	if len(changes) > 0 {
		sendMessage.Call(e.handle, 0x0458, 0, 0) // EM_STOPGROUPTYPING
		var endCollection func()
		if len(changes) > 1 {
			endCollection = beginInlineEditCollection(e.handle)
		}
		for _, run := range changes {
			a.setInlineCharacters(run.Start, run.End, run, inlineSize, true)
		}
		if endCollection != nil {
			endCollection()
		}
		sendMessage.Call(e.handle, 0x0458, 0, 0)
	}
	selectInline(e.handle, start, end)
	sendMessage.Call(e.handle, 0x000b, 1, 0)
	e.loading = false
	a.syncInlineControls(0)
	a.layoutInlineEditor()
	a.updateHistoryButtons()
	a.view.invalidateCanvas()
}
