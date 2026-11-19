package main

import (
	"math"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const idInlineText = 490

type inlineTextFrame struct {
	flow     textFlow
	stored   *storedTextBlock
	editable bool
}

type inlineTextEditor struct {
	handle, font          uintptr
	frames                []inlineTextFrame
	active                int
	original              string
	initialStyle          inlineTextStyle
	sourceBounds, bounds  searchRect
	leftLimit, rightLimit float64
	fixedWidth, loading   bool
	visible               bool
	foreground            uintptr
	background            uintptr
	rect, pageArea, clip  rect
	dpi, pixels           int
	previewStyle          inlineTextStyle
	textPixels            float64
}

type inlineTextStyle struct {
	font                           string
	size, leading, tracking, width float64
	alignment                      int
	bold, italic, underline        bool
}

type inlineCharFormat struct {
	size, mask, effects                     uint32
	height, offset                          int32
	color                                   uint32
	charset, pitch                          byte
	face                                    [32]uint16
	weight                                  uint16
	spacing                                 int16
	background, locale, reserved            uint32
	style                                   int16
	kerning                                 uint16
	underline, animation, author, reserved2 byte
}

type inlineParaFormat struct {
	size, mask                                                                                      uint32
	numbering, effects                                                                              uint16
	start, right, offset                                                                            int32
	alignment, tabCount                                                                             uint16
	tabs                                                                                            [32]int32
	before, after, spacing                                                                          int32
	style                                                                                           int16
	rule, outline                                                                                   byte
	shadeWeight, shadeStyle, numberStart, numberStyle, numberTab, borderSpace, borderWidth, borders uint16
}

func (a *app) createInlineTextEditor() {
	e := &a.textTool.inline
	e.active = -1
	if syscall.NewLazyDLL("Msftedit.dll").Load() == nil {
		e.handle, _, _ = createWindow.Call(0, uintptr(unsafe.Pointer(wideString("RICHEDIT50W"))), 0,
			wsChild|wsTabstop|0x00800000|0x1000|0x0044|0x0100, 0, 0, 1, 1, a.window, idInlineText, a.instance, 0)
	}
	if e.handle != 0 {
		sendMessage.Call(e.handle, 0x0459, 1|8, 0)        // EM_SETTEXTMODE, TM_PLAINTEXT | TM_MULTILEVELUNDO
		sendMessage.Call(e.handle, 0x0445, 0, 1)          // EM_SETEVENTMASK, ENM_CHANGE
		sendMessage.Call(e.handle, 0x0443, 0, 0x00ffffff) // EM_SETBKGNDCOLOR
		sendMessage.Call(e.handle, 0x0435, 0, 8000)       // EM_EXLIMITTEXT
		e.background = 0xffffff
	}
}

func (a *app) inlineStyle() inlineTextStyle {
	p := &a.textTool
	size, _ := readTextNumber(p.size, 1, 600)
	leading, _ := readTextNumber(p.leading, .7, 5)
	tracking, _ := readTextNumber(p.tracking, -size*.18, 72)
	width, _ := readTextNumber(p.width, 1, 20000)
	return inlineTextStyle{p.fonts[max(0, comboIndex(p.font))], size, leading, tracking, width, p.alignment,
		a.checked(p.bold), a.checked(p.italic), a.checked(p.underline)}
}

func (a *app) inlineCurrent() *inlineTextFrame {
	e := &a.textTool.inline
	if e.active < 0 || e.active >= len(e.frames) {
		return nil
	}
	return &e.frames[e.active]
}

func (a *app) inlineDirty() bool {
	e := &a.textTool.inline
	frame := a.inlineCurrent()
	if frame == nil || !frame.editable || e.loading {
		return false
	}
	style := a.inlineStyle()
	original := e.initialStyle
	if !e.fixedWidth {
		style.width = original.width
	}
	return getText(e.handle) != e.original || style != original
}

// Editorial commands continue after the draft was written. On any error the
// native editor and its selection stay open so the user can fix or undo the input.
func (a *app) saveInlineBefore(next func()) bool {
	if !a.inlineDirty() {
		return false
	}
	if !a.busy {
		a.textTool.afterEdit = next
		a.startTextEdit()
		if !a.busy {
			a.textTool.afterEdit = nil
		}
	}
	return true
}

func (a *app) clearInlineFrames() {
	e := &a.textTool.inline
	e.frames, e.active = nil, -1
	e.visible = false
	if e.handle != 0 {
		showWindow.Call(e.handle, 0)
	}
	if e.font != 0 {
		sendMessage.Call(e.handle, wmSetFont, a.normalFont, 0)
		gdiDelete.Call(e.font)
		e.font = 0
	}
	e.previewStyle = inlineTextStyle{}
	e.rect, e.clip = rect{}, rect{}
}

func (a *app) setInlineFrames(flow textFlow, stored *storedTextBlock, pick textSelection) {
	e := &a.textTool.inline
	a.clearInlineFrames()
	if flow.text == "" {
		return
	}
	groups := []textFlow{flow}
	if len(pick.polygon) >= 3 && len(flow.paragraphs) > 1 {
		groups = nil
		for _, paragraph := range flow.paragraphs {
			group := layoutText(textFlow{paragraphs: [][]textLine{paragraph}}.glyphs())
			if group.text != "" {
				groups = append(groups, group)
			}
		}
	}
	for _, group := range groups {
		glyphs := group.glyphs()
		editable := len(glyphs) <= 8000
		for _, glyph := range glyphs {
			if !a.textTool.model.Objects[glyph.run].Editable {
				editable = false
				break
			}
		}
		e.frames = append(e.frames, inlineTextFrame{group, stored, editable})
	}
	if len(e.frames) == 1 {
		a.focusInlineFrame(0, pick.point, len(pick.polygon) == 0)
	}
}

func (a *app) focusInlineFrame(index int, caret textPoint, useCaret bool) {
	p, e := &a.textTool, &a.textTool.inline
	if e.handle == 0 {
		a.report("Editor textu systému Windows se nepodařilo otevřít.")
		return
	}
	if index < 0 || index >= len(e.frames) {
		return
	}
	frame := &e.frames[index]
	e.active, e.loading, e.fixedWidth, e.pixels = index, true, false, 0
	e.sourceBounds = frame.flow.bounds
	first := p.model.Objects[frame.flow.glyphs()[0].run]
	e.foreground = paletteColor{first.Fill[0], first.Fill[1], first.Fill[2]}.win()
	font := slices.IndexFunc(p.fonts, func(value string) bool { return strings.EqualFold(value, first.Font) })
	if font < 0 {
		font = min(max(first.Family, 0), len(bundledTextFonts)-1)
	}
	chooseCombo(p.font, font)
	setText(p.original, first.Font)
	setText(p.size, formatTextNumber(first.Size))
	setText(p.leading, "1,2")
	setText(p.tracking, "0")
	setText(p.width, formatTextNumber(max(1, e.sourceBounds[2]-e.sourceBounds[0])))
	p.alignment = 0
	a.setChecked(p.bold, first.Bold)
	a.setChecked(p.italic, first.Italic)
	a.setChecked(p.underline, first.Underline)
	if saved := frame.stored; saved != nil && saved.Rotation == p.rotation {
		anchor := pdfPointToText(saved.Origin, p.model.Boxes[p.pick.box], p.rotation)
		x, y := anchor[0]*p.dim[0], anchor[1]*p.dim[1]
		if x >= 0 && y >= 0 && saved.Width >= 1 && x+saved.Width <= p.dim[0]+.003 {
			e.sourceBounds[0], e.sourceBounds[1], e.sourceBounds[2] = x, y, x+saved.Width
			p.alignment = min(max(saved.Alignment, 0), len(textAlignments)-1)
			setText(p.leading, formatTextNumber(saved.Leading))
			setText(p.tracking, formatTextNumber(saved.Tracking))
			setText(p.width, formatTextNumber(saved.Width))
		}
	}
	included := map[[2]int]bool{}
	for _, glyph := range frame.flow.glyphs() {
		included[[2]int{glyph.run, glyph.index}] = true
	}
	glyphs, _, _ := a.view.nativeTextGeometry(p.model, p.pick)
	e.leftLimit, e.rightLimit = inlineNeighbourEdges(e.sourceBounds, glyphs, included, p.dim[0])
	e.original = multiline(frame.flow.text)
	setText(e.handle, e.original)
	sendMessage.Call(e.handle, 0x00cf, uintptr(boolFlag(!frame.editable)), 0) // EM_SETREADONLY
	sendMessage.Call(e.handle, 0x00cd, 0, 0)                                  // EM_EMPTYUNDOBUFFER
	e.initialStyle = a.inlineStyle()
	e.loading = false
	a.updateTextActions()
	a.layoutInlineEditor()
	if e.visible {
		setFocus.Call(e.handle)
		if useCaret {
			at := screenTextPoint(e.pageArea, caret)
			local := point{at.x - e.rect.left, at.y - e.rect.top}
			i, _, _ := sendMessage.Call(e.handle, 0x00d7, 0, uintptr(unsafe.Pointer(&local))) // EM_CHARFROMPOS
			if int32(i) >= 0 {
				sendMessage.Call(e.handle, 0x00b1, i, i)
			}
		}
	}
	a.view.invalidateCanvas()
}

func boolFlag(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func (a *app) inlineFrameAt(pick textSelection) int {
	p := &a.textTool
	if p.pick.source != pick.source || p.pick.generation != pick.generation {
		return -1
	}
	x, y := pick.point[0]*p.dim[0], pick.point[1]*p.dim[1]
	for i, frame := range p.inline.frames {
		b := frame.flow.bounds
		if x >= b[0]-2 && x <= b[2]+2 && y >= b[1]-3 && y <= b[3]+3 {
			return i
		}
	}
	return -1
}

func (a *app) editFrameAt(pick textSelection) bool {
	index := a.inlineFrameAt(pick)
	if index < 0 {
		return false
	}
	if index != a.textTool.inline.active && a.saveInlineBefore(a.resumeTextPick(pick)) {
		return true
	}
	if index != a.textTool.inline.active {
		a.focusInlineFrame(index, pick.point, true)
	}
	return true
}

func (a *app) layoutInlineEditor() {
	p, e, v := &a.textTool, &a.textTool.inline, a.view
	frame := a.inlineCurrent()
	if frame == nil || e.loading || e.handle == 0 {
		return
	}
	area, source := v.textSheetAt(0, 0, p.pick.source)
	if source == 0 || v.generation != p.pick.generation || a.full || p.dim[0] <= 0 || a.mode != idTextTab {
		if e.visible {
			showWindow.Call(e.handle, 0)
			e.visible = false
		}
		return
	}
	e.pageArea = area
	scale := float64(area.right-area.left) / p.dim[0]
	style := a.inlineStyle()
	if style.size <= 0 || style.leading <= 0 || e.fixedWidth && style.width <= 0 {
		return
	}
	px := max(1, min(1900, int(math.Round(style.size*scale))))
	preview := style
	preview.width = 0
	if e.font == 0 || e.pixels != px || e.previewStyle != preview {
		a.formatInlineEditor(preview, px, scale)
	}
	desired := style.width
	if !e.fixedWidth {
		desired = max(4*style.size/5, a.measureInlineWidth(style, scale))
	}
	left, width := inlineHorizontal(e.sourceBounds, e.leftLimit, e.rightLimit, max(1, desired), style.alignment)
	top := e.sourceBounds[1]
	if frame.stored == nil || e.sourceBounds == frame.flow.bounds {
		top = max(0., top-.22*style.size)
	}
	padding := int32(max(2, a.scale(2)))
	box := rect{area.left + int32(math.Round(left*scale)) - padding,
		area.top + int32(math.Round(top*scale)) - padding, 0, 0}
	box.right = box.left + int32(math.Ceil(width*scale)) + 2*padding + 2
	// Set the width first so RichEdit reports the current number of wrapped lines.
	height := max(1, e.rect.bottom-e.rect.top)
	box.bottom = box.top + height
	if box != e.rect {
		place(e.handle, int(box.left), int(box.top), int(box.right-box.left), int(height))
	}
	lines, _, _ := sendMessage.Call(e.handle, 0x00ba, 0, 0) // EM_GETLINECOUNT
	lines = max(lines, uintptr(strings.Count(strings.ReplaceAll(getText(e.handle), "\r\n", "\n"), "\n")+1))
	wantedHeight := int32(math.Ceil((1.25+float64(max(1, int(lines))-1)*style.leading)*float64(px))) + padding*2 + 4
	box.bottom = min(area.bottom+1, box.top+wantedHeight)
	e.bounds = searchRect{left, top, left + width, min(p.dim[1], top+float64(wantedHeight)/scale)}
	clip := intersectInlineRect(box, intersectInlineRect(v.viewport, area))
	if clip.right <= clip.left || clip.bottom <= clip.top {
		if e.visible {
			showWindow.Call(e.handle, 0)
			e.visible = false
		}
		return
	}
	if e.rect != box || e.clip != clip || !e.visible {
		place(e.handle, int(box.left), int(box.top), int(box.right-box.left), int(box.bottom-box.top))
		region, _, _ := gdi32.NewProc("CreateRectRgn").Call(uintptr(clip.left-box.left), uintptr(clip.top-box.top),
			uintptr(clip.right-box.left), uintptr(clip.bottom-box.top))
		if ok, _, _ := user32.NewProc("SetWindowRgn").Call(e.handle, region, 1); ok == 0 {
			gdiDelete.Call(region)
		}
		e.rect, e.clip = box, clip
	}
	if !e.visible {
		showWindow.Call(e.handle, 5)
		e.visible = true
	}
	if !e.fixedWidth {
		value := formatTextNumber(math.Floor(width*100) / 100)
		if getText(p.width) != value {
			e.loading = true
			setText(p.width, value)
			e.loading = false
		}
	}
}

func intersectInlineRect(a, b rect) rect {
	return rect{max(a.left, b.left), max(a.top, b.top), min(a.right, b.right), min(a.bottom, b.bottom)}
}

func (a *app) formatInlineEditor(style inlineTextStyle, px int, scale float64) {
	e := &a.textTool.inline
	weight := 400
	if style.bold {
		weight = 700
	}
	font, _, _ := gdi32.NewProc("CreateFontW").Call(uintptr(-px), 0, 0, 0, uintptr(weight),
		uintptr(boolFlag(style.italic)), uintptr(boolFlag(style.underline)), 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wideString(style.font))))
	if font == 0 {
		return
	}
	old := e.font
	sendMessage.Call(e.handle, wmSetFont, font, 0)
	if old != 0 {
		gdiDelete.Call(old)
	}
	e.font, e.pixels, e.previewStyle = font, px, style
	dc, _, _ := user32.NewProc("GetDC").Call(e.handle)
	dpi, _, _ := gdi32.NewProc("GetDeviceCaps").Call(dc, 90)
	user32.NewProc("ReleaseDC").Call(e.handle, dc)
	twips := 1440 / float64(max(1, int(dpi)))
	format := inlineCharFormat{mask: 0xe8200007, color: uint32(e.foreground), height: int32(math.Round(float64(px) * twips)), charset: 1,
		spacing: int16(math.Round(style.tracking * scale * twips)), effects: boolFlag(style.bold) | boolFlag(style.italic)<<1 | boolFlag(style.underline)<<2}
	format.size = uint32(unsafe.Sizeof(format))
	copy(format.face[:], syscall.StringToUTF16(style.font))
	sendMessage.Call(e.handle, 0x0444, 4, uintptr(unsafe.Pointer(&format))) // EM_SETCHARFORMAT, SCF_ALL
	para := inlineParaFormat{mask: 0x108, alignment: uint16(1 + style.alignment), spacing: int32(math.Round(style.leading * 20)), rule: 5}
	if style.alignment == 1 {
		para.alignment = 3
	} else if style.alignment == 2 {
		para.alignment = 2
	}
	para.size = uint32(unsafe.Sizeof(para))
	var start, end int32
	sendMessage.Call(e.handle, 0x00b0, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	sendMessage.Call(e.handle, 0x00b1, 0, ^uintptr(0))
	sendMessage.Call(e.handle, 0x0447, 0, uintptr(unsafe.Pointer(&para))) // EM_SETPARAFORMAT
	sendMessage.Call(e.handle, 0x00b1, uintptr(start), uintptr(end))
}

func (a *app) measureInlineWidth(style inlineTextStyle, scale float64) float64 {
	e := &a.textTool.inline
	dc, _, _ := user32.NewProc("GetDC").Call(e.handle)
	if dc == 0 {
		return e.sourceBounds[2] - e.sourceBounds[0]
	}
	defer user32.NewProc("ReleaseDC").Call(e.handle, dc)
	old, _, _ := gdiSelect.Call(dc, e.font)
	defer gdiSelect.Call(dc, old)
	width := 0.
	for _, line := range strings.Split(strings.ReplaceAll(getText(e.handle), "\r\n", "\n"), "\n") {
		line = strings.ReplaceAll(line, "\t", "    ")
		encoded := syscall.StringToUTF16(line)
		var measure point
		gdi32.NewProc("GetTextExtentPoint32W").Call(dc, uintptr(unsafe.Pointer(&encoded[0])), uintptr(len(encoded)-1), uintptr(unsafe.Pointer(&measure)))
		extent := float64(measure.x)/scale + float64(max(0, len([]rune(line))-1))*style.tracking
		width = max(width, extent+.75)
	}
	return width
}

func (a *app) inlineControlChange(handle uintptr) {
	p, e := &a.textTool, &a.textTool.inline
	if e.loading || a.inlineCurrent() == nil {
		return
	}
	switch handle {
	case e.handle, p.size, p.leading, p.tracking, p.width:
		if handle == p.width {
			e.fixedWidth = getText(handle) != ""
		}
		a.layoutInlineEditor()
		a.view.invalidateCanvas()
	}
}

func (a *app) paintInlineFrames(dc uintptr, area rect) {
	p, e := &a.textTool, &a.textTool.inline
	if p.dim[0] <= 0 || p.dim[1] <= 0 {
		return
	}
	toPixel := func(b searchRect) rect {
		left := screenTextPoint(area, textPoint{b[0] / p.dim[0], b[1] / p.dim[1]})
		right := screenTextPoint(area, textPoint{b[2] / p.dim[0], b[3] / p.dim[1]})
		return rect{left.x, left.y, right.x, right.y}
	}
	for i, frame := range e.frames {
		if i == e.active && e.visible {
			// Paint over the previous glyphs that would otherwise reappear when
			// the editor shrinks. The neighboring frames are never erased.
			if frame.editable {
				for n, line := range frame.flow.highlightBounds() {
					box := toPixel(line)
					paper, _, _ := gdi32.NewProc("GetPixel").Call(dc, uintptr(box.left+1), uintptr(max(area.top, box.top-2)))
					if paper > 0xffffff {
						paper = 0xffffff
					}
					if n == 0 && paper != e.background {
						e.background = paper
						sendMessage.Call(e.handle, 0x0443, 0, paper)
					}
					fillColor(dc, box, paletteColor{uint8(paper), uint8(paper >> 8), uint8(paper >> 16)})
				}
			}
			drawInlineFrame(dc, rect{e.rect.left - 1, e.rect.top - 1, e.rect.right, e.rect.bottom})
		} else {
			drawInlineFrame(dc, toPixel(frame.flow.bounds))
		}
	}
}

func drawInlineFrame(dc uintptr, box rect) {
	vertices := []point{{box.left, box.top}, {box.right, box.top}, {box.right, box.bottom}, {box.left, box.bottom}, {box.left, box.top}}
	gdi32.NewProc("Polyline").Call(dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
}

func (a *app) resumeTextPick(pick textSelection) func() {
	position := slices.Index(a.view.order, pick.source)
	return func() {
		if position < 0 || position >= len(a.view.order) {
			return
		}
		pick.source, pick.generation, pick.mode = a.view.order[position], a.view.generation, textBlock
		a.selectText(pick)
	}
}
