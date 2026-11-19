package main

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	idTextTab = 280 + iota
	_
	idTextFont
	idTextBold
	idTextItalic
	idTextUnderline
	idTextApply
	idTextCopy
	idTextUndo
	idTextOCR
)

const (
	textObject = iota
	textBlock
	textRectangle
	textLasso
)

var textAlignments = []string{"left", "center", "right", "justify"}

type textSelection struct {
	source, generation, mode, tool int
	point, tolerance               textPoint
	polygon                        []textPoint
	box                            string
	automatic                      bool
}

type textPanel struct {
	guide, font, size, original, report                      uintptr
	width, leading, tracking                                 uintptr
	regular, bold, italic, underline, apply, copy, undo, ocr uintptr
	modes, alignments                                        [4]uintptr
	mode, alignment                                          int
	fonts                                                    []string
	tooltip                                                  uintptr
	tooltipNames                                             []*uint16
	model                                                    *textPage
	pick, gesture                                            textSelection
	flow                                                     textFlow
	dim                                                      textPoint
	rotation                                                 int
	drag, hasHover                                           bool
	hover                                                    textPoint
	inline                                                   inlineTextEditor
	afterEdit                                                func()
}

func (a *app) buildTextPanel() {
	p := &a.textTool
	controls := [2]uint32{8, 4} // ICC_BAR_CLASSES includes native tooltips.
	comctl32.NewProc("InitCommonControlsEx").Call(uintptr(unsafe.Pointer(&controls)))
	a.label("Výběr textu", 16, 16, 298, true)
	for i, name := range []string{"Textový objekt", "Odstavec", "Obdélník", "Mnohoúhelníkové laso"} {
		p.modes[i] = a.textIconButton(name, idTextSelectObject+i, 16+i*40, 42, 34)
	}
	p.guide = a.add("STATIC", "", 0, 16, 82, 298, 39, 0, 0, a.normalFont)
	p.original = a.add("STATIC", "", 0, 16, 137, 298, 21, 0x4000, 0, a.normalFont)
	a.label("Písmo", 16, 161, 205, false)
	a.label("Velikost (pt)", 236, 161, 78, false)
	p.fonts = a.installedTextFonts()
	p.font = a.combo(idTextFont, 16, 183, 210, p.fonts)
	p.size = a.edit(236, 183, 78, false)
	p.regular = a.textIconButton("Běžný řez (Regular)", idTextRegular, 16, 222, 32)
	p.bold = a.textIconButton("Tučné", idTextBold, 52, 222, 32)
	p.italic = a.textIconButton("Kurzíva", idTextItalic, 88, 222, 32)
	p.underline = a.textIconButton("Podtržené", idTextUnderline, 124, 222, 32)
	for i, name := range []string{"Zarovnat vlevo", "Zarovnat na střed", "Zarovnat vpravo", "Zarovnat do bloku"} {
		p.alignments[i] = a.textIconButton(name, idTextAlignLeft+i, 174+i*36, 222, 32)
	}
	a.label("Šířka (pt)", 16, 266, 93, false)
	a.label("Řádkování (×)", 118, 266, 94, false)
	a.label("Proklad (pt)", 220, 266, 94, false)
	p.width = a.edit(16, 288, 94, false)
	p.leading = a.edit(118, 288, 94, false)
	p.tracking = a.edit(220, 288, 94, false)
	p.apply = a.textIconButton("Použít úpravu", idTextApply, 16, 329, 34)
	p.copy = a.textIconButton("Kopírovat text", idTextCopy, 56, 329, 34)
	p.undo = a.textIconButton("Vrátit úpravu", idTextUndo, 96, 329, 34)
	p.ocr = a.textIconButton("OCR výběru", idTextOCR, 136, 329, 34)
	p.report = a.readout(16, 372, 298, 82)
	a.textGuide()
}

func (a *app) textGuide() {
	p := &a.textTool
	setText(p.guide, [...]string{
		"Klikněte na text a pište přímo do rámečku.",
		"Klikněte do odstavce a upravte ho v dokumentu.",
		"Vyznačte oblast a klikněte do rámečku.",
		"Klikáním přidávejte body. Dokončete Enterem nebo kliknutím na první bod.",
	}[p.mode])
	for i, h := range p.modes {
		a.setChecked(h, i == p.mode)
	}
}

func (v *documentView) lastDraftIs(action string) bool {
	return len(v.history) > 0 && v.history[len(v.history)-1].draftAction == action
}

func (a *app) clearTextSelection() {
	p := &a.textTool
	if p.guide == 0 {
		return
	}
	if p.drag {
		p.drag = false
		user32.NewProc("ReleaseCapture").Call()
	}
	p.model, p.flow, p.hasHover = nil, textFlow{}, false
	a.clearInlineFrames()
	p.pick, p.gesture = textSelection{}, textSelection{}
	a.view.nativeText.next = textSelection{}
	a.view.nativeText.flight = textSelection{}
	for _, control := range []uintptr{p.original, p.size, p.width, p.leading, p.tracking, p.report} {
		setText(control, "")
	}
	for _, option := range []uintptr{p.bold, p.italic, p.underline} {
		a.setChecked(option, false)
	}
	a.updateTextActions()
	a.view.invalidateCanvas()
}

func (a *app) updateTextActions() {
	p := &a.textTool
	if p.guide == 0 {
		return
	}
	ready := !a.busy && a.view.ready()
	valid := ready && p.pick.generation == a.view.generation && p.pick.source != 0
	hasText := valid && p.flow.text != ""
	frame := a.inlineCurrent()
	edit := hasText && frame != nil && frame.editable
	if p.inline.handle != 0 {
		sendMessage.Call(p.inline.handle, 0x00cf, uintptr(boolFlag(!edit)), 0)
	}
	for _, h := range []uintptr{p.apply, p.font, p.size, p.width, p.leading, p.tracking, p.regular, p.bold, p.italic, p.underline} {
		setEnabled(h, edit)
	}
	for _, h := range p.modes {
		setEnabled(h, ready)
	}
	for i, h := range p.alignments {
		setEnabled(h, edit)
		a.setChecked(h, i == p.alignment && edit)
	}
	a.setChecked(p.regular, edit && !a.checked(p.bold) && !a.checked(p.italic))
	setEnabled(p.copy, hasText)
	setEnabled(p.ocr, valid && p.model != nil && !hasText && len(p.pick.polygon) >= 3)
	setEnabled(p.undo, ready && a.view.lastDraftIs("text_edit"))
}

func (a *app) textCommand(id, notification int) bool {
	if notification != 0 || a.mode != idTextTab {
		return false
	}
	p := &a.textTool
	switch {
	case id >= idTextSelectObject && id <= idTextSelectLasso:
		if a.saveInlineBefore(func() { a.textCommand(id, notification) }) {
			return true
		}
		p.mode = id - idTextSelectObject
		a.clearTextSelection()
		a.textGuide()
	case id >= idTextAlignLeft && id <= idTextAlignJustify:
		p.alignment = id - idTextAlignLeft
		a.updateTextActions()
	default:
		switch id {
		case idTextRegular:
			a.setChecked(p.bold, false)
			a.setChecked(p.italic, false)
			a.updateTextActions()
		case idTextBold, idTextItalic, idTextUnderline:
			control := map[int]uintptr{idTextBold: p.bold, idTextItalic: p.italic, idTextUnderline: p.underline}[id]
			a.setChecked(control, !a.checked(control))
			a.updateTextActions()
		case idTextApply:
			a.startTextEdit()
		case idTextCopy:
			a.copyTextSelection()
		case idTextUndo:
			if !a.busy && a.view.lastDraftIs("text_edit") {
				a.view.undo()
			}
		case idTextOCR:
			a.startTextOCR()
		default:
			return false
		}
	}
	a.layoutInlineEditor()
	a.view.invalidateCanvas()
	switch id {
	case idTextRegular, idTextBold, idTextItalic, idTextUnderline,
		idTextAlignLeft, idTextAlignCenter, idTextAlignRight, idTextAlignJustify:
		a.focusInlineEditor()
	}
	return true
}

func (a *app) focusInlineEditor() {
	if e := &a.textTool.inline; e.visible && !a.busy {
		setFocus.Call(e.handle)
	}
}

func normalizedTextPoint(area rect, x, y int32) textPoint {
	return textPoint{
		max(0, min(1, float64(x-area.left)/float64(area.right-area.left))),
		max(0, min(1, float64(y-area.top)/float64(area.bottom-area.top))),
	}
}

func (v *documentView) textSheetAt(x, y int32, source int) (rect, int) {
	_, h := v.viewportSize()
	for _, page := range v.spread.visible(v.scrollY, h) {
		number := v.order[page.index]
		area := v.pageBounds(page)
		if number == source || source == 0 && contains(area, x, y) {
			return area, number
		}
	}
	return rect{}, 0
}

func (v *documentView) textAt(x, y int32) textSelection {
	if !contains(v.viewport, x, y) {
		return textSelection{}
	}
	area, source := v.textSheetAt(x, y, 0)
	if source == 0 {
		return textSelection{}
	}
	padding := float64(v.app.scale(3))
	return textSelection{source: source, generation: v.generation, tool: v.app.mode, mode: v.app.textTool.mode,
		point: normalizedTextPoint(area, x, y), tolerance: textPoint{padding / float64(area.right-area.left),
			padding / float64(area.bottom-area.top)}, box: v.box()}
}

func (a *app) textPointer(msg uint32, x, y int32) bool {
	p, v := &a.textTool, a.view
	if msg == 0x0215 || msg == 0x001f { // WM_CAPTURECHANGED, WM_CANCELMODE
		if p.drag {
			p.drag, p.gesture = false, textSelection{}
			v.invalidateCanvas()
		}
		return false
	}
	if msg == 0x02a3 && p.hasHover { // WM_MOUSELEAVE
		p.hasHover = false
		v.invalidateCanvas()
	}
	if a.mode != idTextTab || a.busy || !v.ready() || a.full {
		return false
	}
	switch msg {
	case 0x0201: // WM_LBUTTONDOWN
		pick := v.textAt(x, y)
		if pick.source == 0 {
			return false
		}
		if a.editFrameAt(pick) {
			return true
		}
		if a.saveInlineBefore(a.resumeTextPick(pick)) {
			return true
		}
		setFocus.Call(a.window)
		switch pick.mode {
		case textRectangle:
			pick.polygon = rectangleSelection(pick.point, pick.point)
			p.drag, p.gesture = true, pick
			user32.NewProc("SetCapture").Call(a.window)
			v.invalidateCanvas()
		case textLasso:
			gesture := &p.gesture
			if gesture.source != pick.source || gesture.generation != v.generation || gesture.mode != textLasso {
				p.gesture = pick
			}
			tolerance, pos := pick.tolerance, pick.point
			if len(gesture.polygon) >= 3 && math.Abs(pos[0]-gesture.polygon[0][0]) <= tolerance[0]*3 &&
				math.Abs(pos[1]-gesture.polygon[0][1]) <= tolerance[1]*3 {
				a.completeTextGesture()
			} else if len(gesture.polygon) < 64 {
				gesture.polygon = append(gesture.polygon, pos)
				p.hasHover = false
				v.invalidateCanvas()
			}
		default:
			a.selectText(pick)
		}
		return true
	case 0x0200, 0x0202: // WM_MOUSEMOVE, WM_LBUTTONUP
		if p.drag {
			pick := &p.gesture
			area, source := v.textSheetAt(x, y, pick.source)
			if source != 0 {
				pick.polygon = rectangleSelection(pick.point, normalizedTextPoint(area, x, y))
			}
			if msg == 0x0202 {
				p.drag = false
				user32.NewProc("ReleaseCapture").Call()
				if source != 0 {
					end := pick.polygon[2]
					if math.Abs(end[0]-pick.point[0]) < pick.tolerance[0] ||
						math.Abs(end[1]-pick.point[1]) < pick.tolerance[1] {
						pick.polygon = nil
					}
					a.completeTextGesture()
				} else {
					p.gesture = textSelection{}
				}
			}
			v.invalidateCanvas()
			return true
		}
		if msg == 0x0200 && p.mode == textLasso && len(p.gesture.polygon) > 0 {
			area, source := v.textSheetAt(x, y, p.gesture.source)
			show := source != 0 && contains(v.viewport, x, y)
			hover := textPoint{}
			if show {
				hover = normalizedTextPoint(area, x, y)
			}
			if show && !p.hasHover {
				track := struct {
					size, flags uint32
					window      uintptr
					hover       uint32
				}{flags: 2, window: a.window}
				track.size = uint32(unsafe.Sizeof(track))
				user32.NewProc("TrackMouseEvent").Call(uintptr(unsafe.Pointer(&track)))
			}
			if show != p.hasHover || hover != p.hover {
				p.hasHover, p.hover = show, hover
				v.invalidateCanvas()
			}
			return true
		}
	}
	return false
}

func (a *app) textKey(msg *winMessage) bool {
	if a.mode != idTextTab || a.full || a.busy {
		return false
	}
	p := &a.textTool
	if msg.wParam == 0x1b && (p.gesture.source != 0 || p.pick.source != 0) {
		a.clearTextSelection()
		return true
	}
	if msg.wParam == 0x0d && msg.window == a.window && p.gesture.mode == textLasso && len(p.gesture.polygon) >= 3 {
		a.completeTextGesture()
		return true
	}
	return false
}

func (a *app) completeTextGesture() {
	p := &a.textTool
	pick := p.gesture
	p.gesture, p.hasHover = textSelection{}, false
	a.selectText(pick)
}

func (a *app) selectText(pick textSelection) {
	p := &a.textTool
	p.model, p.pick, p.flow = nil, pick, textFlow{}
	a.clearInlineFrames()
	setText(p.original, "")
	setText(p.report, "")
	a.updateTextActions()
	a.view.invalidateCanvas()
	a.view.requestNativeText(pick)
}

func formatTextNumber(value float64) string {
	return strings.ReplaceAll(strconv.FormatFloat(value, 'f', 2, 64), ".", ",")
}

func readTextNumber(control uintptr, low, high float64) (float64, bool) {
	value, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(getText(control)), ",", "."), 64)
	return value, err == nil && value >= low && value <= high
}

func (a *app) startTextEdit() {
	p, v := &a.textTool, a.view
	e, frame := &p.inline, a.inlineCurrent()
	if a.busy || frame == nil || !frame.editable || p.model == nil || p.pick.generation != v.generation {
		return
	}
	style := a.inlineStyle()
	_, trackingOK := readTextNumber(p.tracking, -style.size*.18, 72)
	if style.size <= 0 || e.fixedWidth && style.width <= 0 || style.leading <= 0 || !trackingOK {
		a.report("Zkontrolujte velikost písma, šířku rámečku a řádkování.")
		return
	}
	a.layoutInlineEditor()
	b := e.bounds
	top := textPoint{b[0] / p.dim[0], b[1] / p.dim[1]}
	baseline := e.firstBaseline - b[1]
	block := textBlockEdit{Text: strings.ReplaceAll(strings.ReplaceAll(getText(e.handle), "\r\n", "\n"), "\r", "\n"),
		Align: textAlignments[style.alignment], Width: b[2] - b[0], Height: p.dim[1] - b[1],
		Leading: style.leading, Tracking: style.tracking, Baseline: &baseline,
		Origin: textPointToPDF(top, p.model.Boxes[p.pick.box], p.rotation), Rotation: p.rotation}
	block.Selection = a.inlineGlyphSelection(frame)
	req := v.request("text_edit")
	req.View.Page = p.pick.source
	req.TextEdit = &textEditRequest{Font: style.font, Size: style.size,
		Bold: style.bold, Italic: style.italic, Underline: style.underline, Block: block}
	a.runDocumentDraft(req, "Upravuji text v otevřené kartě…", p.report)
}

func (a *app) inlineGlyphSelection(frame *inlineTextFrame) []glyphSelection {
	var selection []glyphSelection
	indexed := map[int]int{}
	for _, glyph := range frame.flow.glyphs() {
		token := a.textTool.model.Objects[glyph.run].BlockToken
		index, found := indexed[token]
		if !found {
			index = len(selection)
			indexed[token] = index
			selection = append(selection, glyphSelection{Token: token})
		}
		selection[index].Glyphs = append(selection[index].Glyphs, glyph.index)
	}
	return selection
}

func (a *app) startTextOCR() {
	p, v := &a.textTool, a.view
	if a.busy || p.model == nil || p.pick.generation != v.generation || p.flow.text != "" || len(p.pick.polygon) < 3 {
		return
	}
	position := slices.Index(v.order, p.pick.source)
	if position < 0 {
		return
	}
	box := p.model.Boxes[p.pick.box]
	req := v.request("ocr")
	req.View.Pages = []int{position + 1}
	req.OCRLanguage = [...]string{"ces+eng", "ces", "eng"}[max(0, comboIndex(a.ocr.language))]
	for _, point := range p.pick.polygon {
		req.OCRRegion = append(req.OCRRegion, textPointToPDF(point, box, p.rotation))
	}
	a.runDocumentDraft(req, "Rozpoznávám vybranou oblast…", p.report)
}

func (a *app) paintTextSelection(dc uintptr, source int, area rect) {
	p := &a.textTool
	if a.mode != idTextTab || a.full || (p.pick.source != source && p.gesture.source != source) {
		return
	}
	color := paletteColor{174, 53, 94}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), color.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	if p.pick.source == source && p.pick.generation == a.view.generation {
		drawTextOutline(dc, area, p.pick.polygon, true)
		a.paintInlineFrames(dc, area)
	}
	if p.gesture.source == source && p.gesture.generation == a.view.generation {
		polygon := p.gesture.polygon
		if p.gesture.mode == textLasso {
			if p.hasHover {
				drawTextOutline(dc, area, append(slices.Clone(polygon), p.hover), false)
			} else {
				drawTextOutline(dc, area, polygon, false)
			}
			n := int32(max(2, a.scale(2)))
			for _, vertex := range polygon {
				pos := screenTextPoint(area, vertex)
				fillColor(dc, rect{pos.x - n, pos.y - n, pos.x + n + 1, pos.y + n + 1}, color)
			}
		} else {
			drawTextOutline(dc, area, polygon, true)
		}
	}
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}

func screenTextPoint(area rect, pos textPoint) point {
	return point{area.left + int32(pos[0]*float64(area.right-area.left)),
		area.top + int32(pos[1]*float64(area.bottom-area.top))}
}

func drawTextOutline(dc uintptr, area rect, polygon []textPoint, closed bool) {
	if len(polygon) < 2 {
		return
	}
	vertices := make([]point, 0, len(polygon)+1)
	for _, pos := range polygon {
		vertices = append(vertices, screenTextPoint(area, pos))
	}
	if closed {
		vertices = append(vertices, vertices[0])
	}
	gdi32.NewProc("Polyline").Call(dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
}
