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
	handle                uintptr
	frames                []inlineTextFrame
	active                int
	original              string
	initialStyle          inlineTextStyle
	sourceBounds, bounds  searchRect
	firstBaseline         float64
	leftLimit, rightLimit float64
	fixedWidth, loading   bool
	nativeWrap            bool
	visible               bool
	foreground            uintptr
	background            uintptr
	backdrop              *printBitmap
	backdropRegion        screenRegion
	backdropRequested     bool
	capture               viewportSurface
	serial                uint64
	rect, pageArea, clip  rect
	dpi, pixels           int
	previewStyle          inlineTextStyle
	mixedFormat           uint32
	layoutPosted          bool
	textPixels            float64
	fontHandles           map[string]uintptr
	glyphPositions        []inlineGlyphPosition
	pristine, caretHidden bool
	mouseSelecting        bool
	mouseAnchor           int
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
	_                                       uint16 // CHARFORMATW tail padding before the extension
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
			wsChild|wsTabstop|0x1000|0x0044|0x0100, 0, 0, 1, 1, a.window, idInlineText, a.instance, 0)
	}
	if e.handle != 0 {
		sendMessage.Call(e.handle, 0x0459, 2|8, 0)          // EM_SETTEXTMODE, TM_RICHTEXT | TM_MULTILEVELUNDO
		sendMessage.Call(e.handle, 0x0452, 256, 0)          // EM_SETUNDOLIMIT
		sendMessage.Call(e.handle, 0x0445, 0, 1|0x00080000) // EM_SETEVENTMASK, ENM_CHANGE | ENM_SELCHANGE
		sendMessage.Call(e.handle, 0x0443, 0, 0x00ffffff)   // EM_SETBKGNDCOLOR
		sendMessage.Call(e.handle, 0x0435, 0, 8000)         // EM_EXLIMITTEXT
		e.background = 0xffffff
		sendMessage.Call(e.handle, 0x0448, 0, 1) // Disable soft wrapping until a fixed width is chosen.
		comctl32.NewProc("SetWindowSubclass").Call(e.handle, inlinePaintSubclass, 1, 0)
	}
}

func (a *app) inlineStyle() inlineTextStyle {
	p := &a.textTool
	size, _ := a.inlineMetric(p.size, inlineSize, 1, 600, p.inline.initialStyle.size)
	leading, _ := readTextNumber(p.leading, .7, 5)
	tracking, _ := a.inlineMetric(p.tracking, inlineTracking, -72, 72, p.inline.initialStyle.tracking)
	width, _ := readTextNumber(p.width, 1, 20000)
	return inlineTextStyle{p.fontKeys[max(0, comboIndex(p.font))], size, leading, tracking, width, p.alignment,
		a.checked(p.bold), a.checked(p.italic), a.checked(p.underline)}
}

func (a *app) inlineMetric(control uintptr, flag uint32, low, high, fallback float64) (float64, bool) {
	if a.textTool.inline.mixedFormat&flag != 0 && getText(control) == mixedTextNumber {
		return fallback, true
	}
	return readTextNumber(control, low, high)
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
	modified, _, _ := sendMessage.Call(e.handle, 0x00c6, 0, 0) // EM_CANUNDO
	style := a.inlineStyle()
	return a.inlineValue() != e.original || modified != 0 || e.fixedWidth ||
		style.leading != e.initialStyle.leading || style.alignment != e.initialStyle.alignment
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
	e.backdrop, e.backdropRequested = nil, false
	e.serial++
	e.capture.close()
	e.visible, e.pristine, e.mouseSelecting = false, false, false
	e.mixedFormat = 0
	e.glyphPositions = nil
	if e.handle != 0 {
		showWindow.Call(e.handle, 0)
	}

	e.previewStyle = inlineTextStyle{}
	e.rect, e.clip = rect{}, rect{}
	a.clearInlineFonts()
}

func (a *app) setInlineFrames(flow textFlow, stored *storedTextBlock, pick textSelection) {
	e := &a.textTool.inline
	a.clearInlineFrames()
	if flow.text == "" {
		return
	}
	groups := []textFlow{flow}
	if len(flow.paragraphs) > 1 {
		switch {
		case pick.change != replaceSelection:
			groups = make([]textFlow, 0, len(flow.paragraphs))
			for _, paragraph := range flow.paragraphs {
				section := (textFlow{paragraphs: [][]textLine{paragraph}}).glyphs()
				groups = append(groups, layoutText(section))
			}
		case pick.region():
			groups = inlineTextColumns(flow)
		}
	}
	for _, group := range groups {
		glyphs := group.glyphs()
		counts := make(map[int]int)
		for _, glyph := range glyphs {
			counts[glyph.run]++
		}
		editable := len(glyphs) <= 8000
		for run, count := range counts {
			object := a.textTool.model.Objects[run]
			if !object.Editable || object.WholeRunOnly && count != len(object.Glyphs) {
				editable = false
				break
			}
		}
		e.frames = append(e.frames, inlineTextFrame{group, stored, editable})
	}
	for i, frame := range e.frames {
		if frame.editable {
			a.focusInlineFrame(i, pick.point, !pick.region() && pick.mode != textBlock && pick.change == replaceSelection)
			return
		}
	}
	if len(e.frames) != 0 {
		a.focusInlineFrame(0, pick.point, false)
	}
}

func inlineTextColumns(flow textFlow) []textFlow {
	var columns []textFlow
	for _, paragraph := range flow.paragraphs {
		group := textFlow{paragraphs: [][]textLine{paragraph}}
		for _, line := range paragraph {
			group.bounds = mergeTextBounds(group.bounds, line.bounds)
			group.text += joinGlyphs(line) + "\n"
		}
		group.text = strings.TrimSuffix(group.text, "\n")
		if group.text == "" {
			continue
		}
		for i, column := range columns {
			overlap := min(column.bounds[2], group.bounds[2]) - max(column.bounds[0], group.bounds[0])
			width := min(column.bounds[2]-column.bounds[0], group.bounds[2]-group.bounds[0])
			separateRows := group.bounds[1] >= column.bounds[3]-paragraph[0].em*.2
			if separateRows && overlap >= width*.65 {
				columns[i].bounds = mergeTextBounds(column.bounds, group.bounds)
				columns[i].text += "\n" + group.text
				columns[i].paragraphs = append(columns[i].paragraphs, paragraph)
				group.text = ""
				break
			}
		}
		if group.text != "" {
			columns = append(columns, group)
		}
	}
	return columns
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
	e.mixedFormat = 0
	e.backdrop, e.backdropRequested = nil, false
	e.serial++
	e.sourceBounds = frame.flow.bounds
	e.firstBaseline = frame.flow.paragraphs[0][0].baseline
	first := p.model.Objects[frame.flow.glyphs()[0].run]
	e.foreground = paletteColor{first.Fill[0], first.Fill[1], first.Fill[2]}.win()
	a.loadInlineFonts(frame)
	font := a.inlineFontForRun(first)
	chooseCombo(p.font, max(0, slices.Index(p.fontKeys, font)))
	original := p.fontNames[font]
	if original == "" {
		original = first.Font
	}
	setText(p.original, original)
	setText(p.size, formatTextNumber(first.Size*math.Hypot(first.Matrix[2], first.Matrix[3])))
	leading := 1.2
	if lines := frame.flow.paragraphs[0]; len(lines) >= 2 {
		size, _ := readTextNumber(p.size, 1, 600)
		if size > 0 {
			leading = min(5, max(.7, (lines[1].baseline-lines[0].baseline)/size))
		}
	}
	setText(p.leading, formatTextNumber(leading))
	setText(p.tracking, formatTextNumber(first.Tracking))
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
	e.initialStyle = a.inlineStyle()
	e.previewStyle = inlineTextStyle{}
	content, runs := a.initialInlineContent(frame)
	e.original = content
	sendMessage.Call(e.handle, 0x00cf, 0, 0)
	setText(e.handle, multiline(e.original))
	for _, run := range runs {
		a.setInlineCharacters(run.Start, run.End, run, inlineBaseMask, true)
	}
	a.formatInlineParagraph(e.initialStyle)
	sendMessage.Call(e.handle, 0x00b1, 0, 0)
	sendMessage.Call(e.handle, 0x00cf, uintptr(boolFlag(!frame.editable)), 0) // EM_SETREADONLY
	sendMessage.Call(e.handle, 0x00cd, 0, 0)                                  // EM_EMPTYUNDOBUFFER
	sendMessage.Call(e.handle, 0x00b9, 0, 0)                                  // EM_SETMODIFY
	e.loading = false
	a.updateTextActions()
	a.layoutInlineEditor()
	if e.visible {
		setFocus.Call(e.handle)
		if !useCaret && (p.pick.region() || p.pick.mode == textBlock || p.pick.change != replaceSelection) {
			selectInline(e.handle, 0, -1)
		}
		if useCaret {
			at := screenTextPoint(e.pageArea, caret)
			index := a.originalInlinePosition(at.x, at.y)
			selectInline(e.handle, index, index)
		}
	}
	a.syncInlineControls(0)
	a.view.invalidateCanvas()
	a.view.schedule()
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
	if frame == nil || !frame.editable || e.loading || e.handle == 0 {
		if e.visible && frame != nil && !frame.editable {
			showWindow.Call(e.handle, 0)
			e.visible = false
		}
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
	e.pristine = !a.inlineDirty()
	scale := float64(area.right-area.left) / p.dim[0]
	style := a.inlineStyle()
	if style.size <= 0 || style.leading <= 0 || e.fixedWidth && style.width <= 0 {
		return
	}
	px := max(1, min(1900, int(math.Round(style.size*scale))))
	a.zoomInlineEditor(scale)
	if e.nativeWrap != e.fixedWidth {
		e.nativeWrap = e.fixedWidth
		sendMessage.Call(e.handle, 0x0448, 0, uintptr(boolFlag(!e.nativeWrap))) // EM_SETTARGETDEVICE
	}
	e.pixels = px
	desired := style.width
	if !e.fixedWidth {
		desired = max(e.sourceBounds[2]-e.sourceBounds[0], a.measureInlineWidth(style, scale))
	}
	left, width := inlineHorizontal(e.sourceBounds, e.leftLimit, e.rightLimit, max(1, desired), style.alignment)
	top := max(0., min(e.sourceBounds[1], e.firstBaseline-.9*e.initialStyle.size))
	padding := int32(max(1, a.scale(1)))
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
	lines = max(lines, uintptr(strings.Count(a.inlineValue(), "\n")+1))
	wantedHeight := int32(math.Ceil((1.25+float64(max(1, int(lines))-1)*style.leading)*float64(px))) + padding*2 + 4
	wantedHeight = max(wantedHeight, int32(math.Ceil((e.sourceBounds[3]-top)*scale))+padding*3)
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
		e.visible = true
		showWindow.Call(e.handle, 5)
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

func (a *app) measureInlineWidth(style inlineTextStyle, scale float64) float64 {
	e := &a.textTool.inline
	dc, _, _ := user32.NewProc("GetDC").Call(e.handle)
	if dc == 0 {
		return e.sourceBounds[2] - e.sourceBounds[0]
	}
	defer user32.NewProc("ReleaseDC").Call(e.handle, dc)
	font, _, _ := gdi32.NewProc("CreateFontW").Call(uintptr(-e.pixels), 0, 0, 0, 400,
		0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wideString(style.font))))
	if font == 0 {
		return e.sourceBounds[2] - e.sourceBounds[0]
	}
	defer gdiDelete.Call(font)
	old, _, _ := gdiSelect.Call(dc, font)
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
		if !e.layoutPosted {
			e.layoutPosted = true
			postMessage.Call(a.window, wmInlineLayout, handle, 0)
		}
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
			if !e.pristine {
				a.paintInlineBackdrop(dc, area, frame.flow.highlightBounds(), toPixel)
			}
			invalidate(e.handle)
		} else {
			for _, bounds := range frame.flow.highlightBounds() {
				box := toPixel(bounds)
				fillColor(dc, rect{box.left, box.bottom - 1, box.right, box.bottom + 1}, mutedColor)
			}
		}
	}
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
