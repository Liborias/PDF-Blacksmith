package main

import (
	"math"
	"slices"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	wmInlineLayout = 0x8019
	inlineFace     = 0x20000000
	inlineSize     = 0x80000000
	inlineColor    = 0x40000000
	inlineTracking = 0x00200000
	inlineBold     = 1
	inlineItalic   = 2
	inlineLine     = 4
	inlineBaseMask = inlineFace | inlineSize | inlineColor | inlineTracking | inlineBold | inlineItalic | inlineLine
)

type inlineTextRun struct {
	Start     int     `json:"start"`
	End       int     `json:"end"`
	Font      string  `json:"font"`
	Size      float64 `json:"size"`
	Tracking  float64 `json:"tracking"`
	Color     uint32  `json:"color"`
	Bold      bool    `json:"bold"`
	Italic    bool    `json:"italic"`
	Underline bool    `json:"underline"`
}

func (a *app) inlineValue() string {
	// RichEdit counts one UTF-16 unit per paragraph break internally.
	const capacity = 8002
	buf := make([]uint16, capacity)
	type getTextEx struct {
		bytes, flags, codePage uint32
		defaultChar, used      uintptr
	}
	request := getTextEx{bytes: capacity * 2, codePage: 1200}
	sendMessage.Call(a.textTool.inline.handle, 0x045e, uintptr(unsafe.Pointer(&request)), uintptr(unsafe.Pointer(&buf[0])))
	return strings.ReplaceAll(syscall.UTF16ToString(buf), "\r", "\n")
}

func (a *app) initialInlineContent(frame *inlineTextFrame) (string, []inlineTextRun) {
	p := &a.textTool
	var content strings.Builder
	var runs []inlineTextRun
	offset := 0
	appendText := func(value string, index int) {
		source := p.model.Objects[index]
		font := slices.IndexFunc(p.fonts, func(name string) bool { return strings.EqualFold(name, source.Font) })
		if font < 0 {
			font = min(max(source.Family, 0), len(bundledTextFonts)-1)
		}
		style := inlineTextRun{Font: p.fonts[font],
			Size:  max(1, math.Round(source.Size*math.Hypot(source.Matrix[2], source.Matrix[3])*20)/20),
			Color: uint32(paletteColor{source.Fill[0], source.Fill[1], source.Fill[2]}.win()),
			Bold:  source.Bold, Italic: source.Italic, Underline: source.Underline}
		if source.Block != nil {
			style.Tracking = source.Block.Tracking
		}
		content.WriteString(value)
		length := len(utf16.Encode([]rune(value)))
		runs = appendInlineRun(runs, style, offset, offset+length)
		offset += length
	}
	previous := -1
	for _, group := range frame.flow.paragraphs {
		for _, line := range group {
			if previous >= 0 {
				appendText("\n", previous)
			}
			writeGlyphLine(line, func(value string, index int) {
				appendText(value, index)
				previous = index
			})
		}
	}
	return content.String(), runs
}

func appendInlineRun(runs []inlineTextRun, style inlineTextRun, start, end int) []inlineTextRun {
	if len(runs) > 0 {
		last := &runs[len(runs)-1]
		style.Start, style.End = last.Start, last.End
		if style == *last {
			last.End = end
			return runs
		}
	}
	style.Start, style.End = start, end
	return append(runs, style)
}

func inlineSelection(handle uintptr) (int, int) {
	var start, end int32
	sendMessage.Call(handle, 0x00b0, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return int(start), int(end)
}

func selectInline(handle uintptr, start, end int) {
	sendMessage.Call(handle, 0x00b1, uintptr(start), uintptr(end))
}

func (a *app) setInlineCharacters(start, end int, style inlineTextRun, mask uint32, selectRange bool) {
	h := a.textTool.inline.handle
	if selectRange {
		selectInline(h, start, end)
	}
	format := inlineCharFormat{mask: mask, color: style.Color, height: int32(math.Round(style.Size * 20)), charset: 1,
		spacing: int16(math.Round(style.Tracking * 20)), effects: boolFlag(style.Bold) | boolFlag(style.Italic)<<1 | boolFlag(style.Underline)<<2}
	format.size = uint32(unsafe.Sizeof(format))
	copy(format.face[:len(format.face)-1], syscall.StringToUTF16(style.Font))
	sendMessage.Call(h, 0x0444, 1, uintptr(unsafe.Pointer(&format))) // EM_SETCHARFORMAT, SCF_SELECTION
}

func (a *app) changeInlineCharacters(mask uint32) {
	e := &a.textTool.inline
	if frame := a.inlineCurrent(); frame == nil || !frame.editable || a.busy {
		return
	}
	style := a.inlineStyle()
	if mask == inlineSize && style.size <= 0 {
		return
	}
	start, end := inlineSelection(e.handle)
	change := inlineTextRun{Font: style.font, Size: style.size, Tracking: style.tracking,
		Bold: style.bold, Italic: style.italic, Underline: style.underline}
	e.loading = true
	sendMessage.Call(e.handle, 0x0458, 0, 0) // EM_STOPGROUPTYPING
	if start == end && a.inlineValue() != "" {
		a.setInlineCharacters(0, -1, change, mask, true)
		selectInline(e.handle, start, end)
	} else {
		a.setInlineCharacters(0, 0, change, mask, false)
	}
	sendMessage.Call(e.handle, 0x0458, 0, 0)
	e.loading = false
	a.updateHistoryButtons()
}

func (a *app) inlineRuns(content string) []inlineTextRun {
	e := &a.textTool.inline
	length := len(utf16.Encode([]rune(content)))
	if length == 0 {
		return nil
	}
	start, end := inlineSelection(e.handle)
	e.loading = true
	sendMessage.Call(e.handle, 0x000b, 0, 0) // WM_SETREDRAW
	defer func() {
		selectInline(e.handle, start, end)
		sendMessage.Call(e.handle, 0x000b, 1, 0)
		e.loading = false
		invalidate(e.handle)
	}()
	var runs []inlineTextRun
	var read func(int, int)
	read = func(left, right int) {
		selectInline(e.handle, left, right)
		format := inlineCharFormat{size: uint32(unsafe.Sizeof(inlineCharFormat{}))}
		sendMessage.Call(e.handle, 0x043a, 1, uintptr(unsafe.Pointer(&format))) // EM_GETCHARFORMAT
		if format.mask&inlineBaseMask != inlineBaseMask && right-left > 1 {
			mid := (left + right) / 2
			read(left, mid)
			read(mid, right)
			return
		}
		style := inlineTextRun{Font: syscall.UTF16ToString(format.face[:]), Size: float64(format.height) / 20,
			Color: format.color & 0xffffff, Tracking: float64(format.spacing) / 20,
			Bold: format.effects&inlineBold != 0, Italic: format.effects&inlineItalic != 0,
			Underline: format.effects&inlineLine != 0}
		if style.Font == "" {
			style.Font = e.initialStyle.font
		}
		if style.Size <= 0 {
			style.Size = e.initialStyle.size
		}
		runs = appendInlineRun(runs, style, left, right)
	}
	read(0, length)
	return runs
}

func (a *app) zoomInlineEditor(scale float64) {
	e := &a.textTool.inline
	if e.textPixels == scale {
		return
	}
	if e.dpi == 0 {
		dc, _, _ := user32.NewProc("GetDC").Call(e.handle)
		dpi, _, _ := gdi32.NewProc("GetDeviceCaps").Call(dc, 90)
		user32.NewProc("ReleaseDC").Call(e.handle, dc)
		e.dpi = max(72, int(dpi))
	}
	const denominator = 1000
	ratio := max(.016, min(63., scale*72/float64(e.dpi)))
	sendMessage.Call(e.handle, 0x04e1, uintptr(math.Round(ratio*denominator)), denominator) // EM_SETZOOM
	e.textPixels = scale
}

func (a *app) formatInlineParagraph(style inlineTextStyle) {
	e := &a.textTool.inline
	if e.previewStyle.alignment == style.alignment && e.previewStyle.leading == style.leading {
		return
	}
	e.previewStyle = style
	para := inlineParaFormat{mask: 0x108, alignment: [...]uint16{1, 3, 2, 4}[style.alignment],
		spacing: int32(math.Round(style.leading * 20)), rule: 5}
	para.size = uint32(unsafe.Sizeof(para))
	start, end := inlineSelection(e.handle)
	selectInline(e.handle, 0, -1)
	sendMessage.Call(e.handle, 0x0447, 0, uintptr(unsafe.Pointer(&para)))
	selectInline(e.handle, start, end)
}

func (a *app) finishInlineLayout(control uintptr) {
	p, e := &a.textTool, &a.textTool.inline
	e.layoutPosted = false
	if e.loading || a.inlineCurrent() == nil || a.busy {
		return
	}
	switch control {
	case p.size:
		a.changeInlineCharacters(inlineSize)
	case p.tracking:
		if _, ok := readTextNumber(p.tracking, -72, 72); ok {
			a.changeInlineCharacters(inlineTracking)
		}
	case p.leading:
		if style := a.inlineStyle(); style.leading > 0 {
			e.loading = true
			a.formatInlineParagraph(style)
			e.loading = false
		}
	}
	a.layoutInlineEditor()
	a.updateHistoryButtons()
	a.view.invalidateCanvas()
}
