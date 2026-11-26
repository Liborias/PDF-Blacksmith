package main

import (
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
	_
	idTextCopy
	idTextUndo
	idTextOCR
)

const (
	textObject = iota
	textBlock
	textRectangle
	textLasso
	textFrame
	textCrossing
)

var textAlignments = []string{"left", "center", "right", "justify"}

type textSelection struct {
	history                        *canvasSelectionState
	inventory                      inventoryTarget
	source, generation, mode, tool int
	point, tolerance               textPoint
	polygon                        []textPoint
	box                            string
	automatic, paragraphRequested  bool
	combined, insertPoint          bool
	change                         selectionChange
}

type textPanel struct {
	font, size, original, report                      uintptr
	width, leading, tracking                          uintptr
	regular, bold, italic, underline, copy, undo, ocr uintptr
	modes                                             [5]uintptr
	alignments                                        [4]uintptr
	typography                                        [6]uintptr
	mode, alignment                                   int
	fonts, fontKeys                                   []string
	fontNames                                         map[string]string
	model                                             *textPage
	pick                                              textSelection
	flow                                              textFlow
	dim                                               textPoint
	rotation                                          int
	explicitParagraph                                 bool
	inline                                            inlineTextEditor
	afterEdit                                         func()
}

func (a *app) buildTextPanel() {
	p := &a.textTool
	section := a.beginPropertySection(uiTextSelection, 16, uiTextSelectionHelp)
	for index, mode := range [...]struct {
		value int
		label uiMessage
	}{
		{textObject, uiTextCursorSelection},
		{textRectangle, uiTextRectangleSelection},
		{textLasso, uiTextLassoSelection},
		{textFrame, uiTextParagraphFrame},
	} {
		id := idTextSelectObject + mode.value
		if mode.value == textFrame {
			id = idTextNewFrame
		}
		p.modes[mode.value] = a.uiTextIconButton(mode.label, id, 16+index*40, 42, 34)
		watchNavigationHover(p.modes[mode.value])
	}
	section.instructions = uiTextSelectionInstructions
	a.beginPropertySection(uiTextFormat, 91, uiTextFormatHelp)
	p.original = a.add("STATIC", "", 0, 16, 137, 158, 21, 0x4000, 0, a.normalFont)
	a.uiLabel(uiFont, 16, 161, 205, false)
	a.uiLabel(uiFontSize, 236, 161, 78, false)
	p.fonts = a.installedTextFonts()
	p.fontKeys = p.fonts
	p.font = a.combo(idTextFont, 16, 183, 210, p.fonts)
	p.size = a.edit(236, 183, 54, false)
	p.typography[0] = a.uiTextIconButton(uiSmallerType, idTextSmaller, 180, 128, 32)
	p.typography[1] = a.uiTextIconButton(uiLargerType, idTextLarger, 216, 128, 32)
	p.typography[2] = a.uiTextIconButton(uiTypeSizes, idTextSizes, 291, 183, 23)
	p.regular = a.uiTextIconButton(uiRegular, idTextRegular, 16, 222, 32)
	p.bold = a.uiTextIconButton(uiBold, idTextBold, 52, 222, 32)
	p.italic = a.uiTextIconButton(uiItalic, idTextItalic, 88, 222, 32)
	p.underline = a.uiTextIconButton(uiUnderline, idTextUnderline, 124, 222, 32)
	for i, name := range []uiMessage{uiAlignLeft, uiAlignCenter, uiAlignRight, uiJustify} {
		p.alignments[i] = a.uiTextIconButton(name, idTextAlignLeft+i, 174+i*36, 222, 32)
	}
	a.uiLabel(uiTextWidth, 16, 266, 93, false)
	a.uiLabel(uiLineSpacing, 118, 266, 94, false)
	a.uiLabel(uiTracking, 220, 266, 94, false)
	p.width = a.edit(16, 288, 70, false)
	p.leading = a.edit(118, 288, 70, false)
	p.tracking = a.edit(220, 288, 70, false)
	descriptions := [...]uiMessage{
		uiTextWidthHelp,
		uiLineSpacingHelp,
		uiTrackingHelp,
	}
	for i, control := range []uintptr{p.width, p.leading, p.tracking} {
		a.bindUIHelp(control, descriptions[i])
		p.typography[i+3] = a.uiTextIconButton(descriptions[i], idTextWidths+i, 87+i*102, 288, 23)
	}
	a.addingSection = 0
	p.copy = a.uiTextIconButton(uiCopyText, idTextCopy, 16, 329, 34)
	p.undo = a.uiTextIconButton(uiUndoTextEdit, idTextUndo, 56, 329, 34)
	p.ocr = a.uiTextIconButton(uiOCRSelection, idTextOCR, 96, 329, 34)
	p.report = a.readout(16, 372, 298, 82)
	a.syncTextSelectionModes()
}

func (a *app) syncTextSelectionModes() {
	p := &a.textTool
	for i, h := range p.modes {
		if h != 0 {
			checked := i == p.mode || i == textLasso && p.mode == textCrossing
			if i == textObject {
				checked = a.canvasTextCursor && (p.mode == textObject || p.mode == textBlock)
			}
			a.setChecked(h, checked)
		}
	}
	a.updatePolygonButton(p.modes[textLasso], p.mode == textCrossing, uiTextLassoSelection)
	if len(a.commandBar.controls) != 0 {
		a.updateCommandBar()
	}
}

func (v *documentView) lastDraftIs(action string) bool {
	return len(v.history) > 0 && v.history[len(v.history)-1].draftAction == action
}

func (a *app) clearTextSelection() {
	p := &a.textTool
	if p.modes[0] == 0 {
		return
	}
	p.model, p.flow = nil, textFlow{}
	a.clearInlineFrames()
	p.pick = textSelection{}
	a.view.nativeText.next = textSelection{}
	a.view.nativeText.flight = textSelection{}
	for _, control := range []uintptr{p.original, p.size, p.width, p.leading, p.tracking, p.report} {
		setText(control, "")
	}
	for _, option := range []uintptr{p.bold, p.italic, p.underline} {
		a.setChecked(option, false)
	}
	a.updateTextActions()
	a.syncLayerSelection()
	a.view.invalidateCanvas()
}

func (a *app) updateTextActions() {
	p := &a.textTool
	if p.modes[0] == 0 {
		return
	}
	ready := !a.busy && a.view.ready()
	valid := ready && p.pick.generation == a.view.generation && p.pick.source != 0
	frame := a.inlineCurrent()
	hasText := valid && (p.flow.text != "" || frame != nil && frame.newBounds != nil && a.inlineValue() != "")
	edit := valid && frame != nil && frame.editable
	if p.inline.handle != 0 {
		sendMessage.Call(p.inline.handle, 0x00cf, uintptr(boolFlag(!edit)), 0)
	}
	for _, h := range []uintptr{p.font, p.size, p.width, p.leading, p.tracking, p.regular, p.bold, p.italic, p.underline} {
		setEnabled(h, edit)
	}
	for _, h := range p.typography {
		setEnabled(h, edit)
	}
	for _, h := range p.modes {
		if h != 0 {
			setEnabled(h, ready)
		}
	}
	for i, h := range p.alignments {
		setEnabled(h, edit)
		a.setChecked(h, i == p.alignment && edit)
	}
	a.setChecked(p.regular, edit && p.inline.mixedFormat&(inlineBold|inlineItalic) == 0 &&
		!a.checked(p.bold) && !a.checked(p.italic))
	setEnabled(p.copy, hasText)
	setEnabled(p.ocr, valid && p.model != nil && p.pick.mode != textFrame && !hasText && p.pick.region())
	setEnabled(p.undo, ready && a.canDocumentCommand(a.view, docUndo, false))
}

func (a *app) textCommand(id, notification int) bool {
	if a.mode != idTextTab {
		return false
	}
	if id == idTextFont && notification == 1 && !a.textTool.inline.loading {
		a.changeInlineCharacters(inlineFace)
		a.finishInlineLayout(0)
		a.focusInlineEditor()
		return true
	}
	if notification != 0 {
		return false
	}
	if a.typographyCommand(id) {
		return true
	}
	p := &a.textTool
	switch {
	case id >= idTextSelectObject && id <= idTextSelectLasso || id == idTextNewFrame:
		if a.saveInlineBefore(func() { a.textCommand(id, notification) }) {
			return true
		}
		a.cancelCanvasSelections()
		previous := p.mode
		p.mode = id - idTextSelectObject
		if p.mode == textLasso && previous == textLasso {
			p.mode = textCrossing
		}
		if id == idTextNewFrame {
			p.mode = textFrame
		}
		p.explicitParagraph = p.mode == textBlock
		if p.mode < textRectangle {
			a.canvasTextCursor = true
		}
		a.cancelCanvasSelectionProbes()
		if p.mode < textRectangle || p.mode == textFrame {
			a.view.clearMixedSelection()
			a.clearTextSelection()
		}
		a.syncTextSelectionModes()
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
		case idTextCopy:
			a.copyTextSelection()
		case idTextUndo:
			a.doDocumentCommand(a.view, docUndo, false)
		case idTextOCR:
			a.startTextOCR()
		default:
			return false
		}
	}
	switch id {
	case idTextRegular:
		a.changeInlineCharacters(inlineBold | inlineItalic)
	case idTextBold:
		a.changeInlineCharacters(inlineBold)
	case idTextItalic:
		a.changeInlineCharacters(inlineItalic)
	case idTextUnderline:
		a.changeInlineCharacters(inlineLine)
	case idTextAlignLeft, idTextAlignCenter, idTextAlignRight, idTextAlignJustify:
		p.inline.loading = true
		a.formatInlineParagraph(a.inlineStyle())
		p.inline.loading = false
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
		float64(x-area.left) / float64(area.right-area.left),
		float64(y-area.top) / float64(area.bottom-area.top),
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
	pick := v.canvasPick()
	pick.source, pick.point = source, normalizedTextPoint(area, x, y)
	pick.tolerance = textPoint{padding / float64(area.right-area.left), padding / float64(area.bottom-area.top)}
	return pick
}

func (v *documentView) canvasPick() textSelection {
	return textSelection{generation: v.generation, tool: v.app.mode, mode: v.app.textTool.mode,
		change: canvasSelectionChange(), box: v.box()}
}

func (a *app) textKey(msg *winMessage) bool {
	if a.mode != idTextTab || a.full || a.busy || msg.wParam != 0x1b || a.textTool.pick.source == 0 {
		return false
	}
	if frame := a.inlineCurrent(); frame != nil && frame.newBounds != nil {
		return true
	}
	previous := a.view.captureCanvasSelection()
	if !a.saveInlineBefore(a.clearTextSelection) {
		a.clearTextSelection()
		a.view.rememberCanvasSelection(&previous)
	}
	return true
}

func (a *app) selectText(pick textSelection) {
	a.view.dispatchCanvasSelection(pick)
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
	_, trackingOK := a.inlineMetric(p.tracking, inlineTracking, -72, 72, e.initialStyle.tracking)
	if style.size <= 0 || e.fixedWidth && style.width <= 0 || style.leading <= 0 || !trackingOK {
		a.reportUI(uiInvalidTextMetrics)
		return
	}
	a.layoutInlineEditor()
	b := e.bounds
	top := textPoint{b[0] / p.dim[0], b[1] / p.dim[1]}
	baseline := e.firstBaseline - b[1]
	if frame.newBounds != nil {
		baseline = .9 * style.size
	}
	content := a.inlineValue()
	if !a.inlineDirty() {
		return
	}
	block := textBlockEdit{Text: content, Runs: a.inlineRuns(content), Wrap: e.fixedWidth,
		Align: textAlignments[style.alignment], Width: b[2] - b[0], Height: p.dim[1] - b[1],
		Leading: style.leading, Tracking: style.tracking, Baseline: &baseline,
		Origin: textPointToPDF(top, p.model.Boxes[p.pick.box], p.rotation), Rotation: p.rotation}
	if style.leading == e.initialStyle.leading && !e.fixedWidth {
		for _, group := range frame.flow.paragraphs {
			for _, line := range group {
				left := 0.
				if style.alignment == e.initialStyle.alignment && style.alignment == 0 {
					first := line.glyphs[0]
					for _, glyph := range line.glyphs {
						if strings.TrimSpace(glyph.text) != "" {
							first = glyph
							break
						}
					}
					origin := p.model.Objects[first.run].Glyphs[first.index].Origin
					left = pdfPointToText(origin, p.model.Boxes[p.pick.box], p.rotation)[0]*p.dim[0] - b[0]
				}
				block.Lines = append(block.Lines, textPoint{left, line.baseline - b[1]})
			}
		}
	}
	block.Selection = a.inlineGlyphSelection(frame)
	req := v.request("text_edit")
	req.View.Page = p.pick.source
	req.TextEdit = &textEditRequest{Font: style.font, Size: style.size, Insert: frame.newBounds != nil,
		Bold: style.bold, Italic: style.italic, Underline: style.underline, Block: block}
	a.runDocumentDraft(req, "Upravuji text v otevřené kartě…", p.report)
}

func (a *app) inlineGlyphSelection(frame *inlineTextFrame) []glyphSelection {
	return glyphSelectionFor(a.textTool.model, frame.flow.glyphs())
}

func glyphSelectionFor(model *textPage, glyphs []layoutGlyph) []glyphSelection {
	var selection []glyphSelection
	indexed := map[int]int{}
	for _, glyph := range glyphs {
		token := model.Objects[glyph.run].BlockToken
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
	if a.busy || p.model == nil || p.pick.generation != v.generation || p.flow.text != "" ||
		p.pick.mode == textFrame || !p.pick.region() {
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
	if a.mode != idTextTab || a.full || p.pick.source != source || p.pick.generation != a.view.generation {
		return
	}
	if p.flow.text == "" && p.pick.mode != textFrame {
		a.drawSelectionRegion(dc, area, p.pick)
	}
	a.paintInlineFrames(dc, area)
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
	drawScreenOutline(dc, vertices)
}

func drawScreenOutline(dc uintptr, vertices []point) {
	if len(vertices) >= 2 {
		gdi32.NewProc("Polyline").Call(dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
	}
}
