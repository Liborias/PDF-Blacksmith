package main

import (
	"syscall"
	"unsafe"
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

func (a *app) typographyCommand(id int) bool {
	if id < idTextSmaller || id > idTextTracking {
		return false
	}
	p := &a.textTool
	if frame := a.inlineCurrent(); frame == nil || !frame.editable || a.busy {
		return true
	}
	if id == idTextSmaller || id == idTextLarger {
		current, _ := readTextNumber(p.size, 1, 600)
		for i, value := range typographicSizes {
			if id == idTextSmaller {
				value = typographicSizes[len(typographicSizes)-1-i]
			}
			if id == idTextSmaller && value < current || id == idTextLarger && value > current {
				setText(p.size, formatTextNumber(value))
				break
			}
		}
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

func (a *app) syncInlineControls() {
	p, e := &a.textTool, &a.textTool.inline
	if a.inlineCurrent() == nil || e.loading || e.layoutPosted || a.busy {
		return
	}
	format := inlineCharFormat{size: uint32(unsafe.Sizeof(inlineCharFormat{}))}
	sendMessage.Call(e.handle, 0x043a, 1, uintptr(unsafe.Pointer(&format)))
	e.loading = true
	defer func() { e.loading = false }()
	if format.mask&inlineFace != 0 {
		for i, name := range p.fonts {
			if name == syscall.UTF16ToString(format.face[:]) {
				chooseCombo(p.font, i)
				break
			}
		}
	}
	if format.mask&inlineSize != 0 && format.height > 0 {
		setText(p.size, formatTextNumber(float64(format.height)/20))
	}
	setText(p.tracking, formatTextNumber(float64(format.spacing)/20))
	for flag, control := range map[uint32]uintptr{inlineBold: p.bold, inlineItalic: p.italic, inlineLine: p.underline} {
		a.setChecked(control, format.mask&flag != 0 && format.effects&flag != 0)
	}
	a.setChecked(p.regular, format.mask&(inlineBold|inlineItalic) == inlineBold|inlineItalic &&
		format.effects&(inlineBold|inlineItalic) == 0)
}
