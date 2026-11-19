package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	idTextTab = 280 + iota
	idTextMode
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

type textRun struct {
	Token, Family           int
	Text, Font              string
	Size                    float64
	Bold, Italic, Underline bool
	Bounds                  searchRect
}

type textPage struct {
	Objects  []textRun
	Boxes    map[string]searchRect
	Rotation int
}

type textReplacement struct {
	Token int    `json:"token"`
	Text  string `json:"text"`
}

type textEditRequest struct {
	Family    int               `json:"family"`
	Size      float64           `json:"size"`
	Bold      bool              `json:"bold"`
	Italic    bool              `json:"italic"`
	Underline bool              `json:"underline"`
	Runs      []textReplacement `json:"runs"`
}

type textSelection struct {
	source, generation, mode int
	point, tolerance         textPoint
	polygon                  []textPoint
	box                      string
}

type textPanel struct {
	mode, guide, input, font, size, original, report uintptr
	bold, italic, underline, apply, copy, undo, ocr  uintptr
	model                                            *textPage
	source, generation                               int
	pick                                             textSelection
	selected                                         []int
	drag                                             bool
	gesture                                          textSelection
}

func (a *app) buildTextPanel() {
	p := &a.textTool
	a.label("Výběr textu", 16, 16, 298, true)
	p.mode = a.combo(idTextMode, 16, 42, 298, []string{"Textový objekt", "Odstavec", "Obdélník", "Mnohoúhelníkové laso"})
	p.guide = a.add("STATIC", "", 0, 16, 81, 298, 40, 0, 0, a.normalFont)
	a.label("Vybraný text", 16, 127, 298, true)
	p.input = a.add("EDIT", "", 0, 16, 153, 298, 96, wsTabstop|0x0044|0x1000|esReadonly, 0, a.normalFont)
	sendMessage.Call(p.input, 0x00d3, 3, uintptr(a.scale(6))|uintptr(a.scale(6))<<16)
	p.original = a.add("STATIC", "", 0, 16, 258, 298, 21, 0x4000, 0, a.normalFont)
	a.label("Písmo", 16, 287, 205, false)
	a.label("Velikost (pt)", 236, 287, 78, false)
	p.font = a.combo(idTextFont, 16, 310, 210, []string{"Liberation Sans", "Liberation Serif", "Liberation Mono"})
	p.size = a.edit(236, 310, 78, false)
	p.bold = a.button("Tučné", idTextBold, 16, 350, 90, bsAutoCheck)
	p.italic = a.button("Kurzíva", idTextItalic, 110, 350, 90, bsAutoCheck)
	p.underline = a.button("Podtržené", idTextUnderline, 204, 350, 110, bsAutoCheck)
	p.apply = a.button("Použít úpravu", idTextApply, 16, 393, 145, 0)
	a.primaryButtons[p.apply] = true
	p.copy = a.button("Kopírovat text", idTextCopy, 169, 393, 145, 0)
	p.undo = a.button("Vrátit úpravu", idTextUndo, 16, 433, 145, 0)
	p.ocr = a.button("OCR výběru", idTextOCR, 169, 433, 145, 0)
	p.report = a.readout(16, 481, 298, 82)
	a.textGuide()
}

func (a *app) textGuide() {
	setText(a.textTool.guide, [...]string{
		"Klikněte na text v dokumentu.",
		"Klikněte do odstavce, který chcete vybrat.",
		"Tažením vyznačte oblast na stránce.",
		"Klikáním přidávejte body. Výběr dokončíte klávesou Enter nebo kliknutím na první bod.",
	}[max(0, comboIndex(a.textTool.mode))])
}

func (v *documentView) lastDraftIs(action string) bool {
	return len(v.history) > 0 && v.history[len(v.history)-1].draftAction == action
}

func (a *app) clearTextSelection() {
	p := &a.textTool
	if p.input == 0 {
		return
	}
	if p.drag {
		p.drag = false
		user32.NewProc("ReleaseCapture").Call()
	}
	p.model, p.selected = nil, nil
	p.pick, p.gesture = textSelection{}, textSelection{}
	setText(p.input, "")
	setText(p.original, "")
	setText(p.size, "")
	for _, option := range []uintptr{p.bold, p.italic, p.underline} {
		a.setChecked(option, false)
	}
	setText(p.report, "")
	a.updateTextActions()
	a.view.invalidateCanvas()
}

func (a *app) updateTextActions() {
	p := &a.textTool
	if p.input == 0 {
		return
	}
	valid := a.view.ready() && p.pick.generation == a.view.generation && p.pick.source != 0
	hasText := valid && p.model != nil && len(p.selected) > 0
	editable := hasText && len(p.selected) <= 64
	for _, index := range p.selected {
		if p.model.Objects[index].Token == 0 {
			editable = false
			break
		}
	}
	ready := !a.busy && a.view.ready()
	for _, h := range []uintptr{p.apply, p.font, p.size, p.bold, p.italic, p.underline} {
		setEnabled(h, ready && editable)
	}
	readonly := uintptr(1)
	if ready && editable {
		readonly = 0
	}
	sendMessage.Call(p.input, 0x00cf, readonly, 0) // EM_SETREADONLY
	setEnabled(p.copy, ready && hasText)
	setEnabled(p.ocr, ready && valid && p.model != nil && !hasText && len(p.pick.polygon) >= 3)
	setEnabled(p.undo, ready && a.view.lastDraftIs("text_edit"))
	setEnabled(p.mode, ready)
}

func (a *app) textCommand(id, notification int) bool {
	p := &a.textTool
	if id == idTextMode && notification == 1 {
		a.clearTextSelection()
		a.textGuide()
		return true
	}
	if notification != 0 || a.mode != idTextTab {
		return false
	}
	switch id {
	case idTextBold, idTextItalic, idTextUnderline:
		control := map[int]uintptr{idTextBold: p.bold, idTextItalic: p.italic, idTextUnderline: p.underline}[id]
		a.setChecked(control, !a.checked(control))
	case idTextApply:
		a.startTextEdit()
	case idTextCopy:
		sendMessage.Call(p.input, 0x00b1, 0, ^uintptr(0)) // EM_SETSEL
		sendMessage.Call(p.input, 0x0301, 0, 0)           // WM_COPY
	case idTextUndo:
		if !a.busy && a.view.lastDraftIs("text_edit") {
			a.view.undo()
		}
	case idTextOCR:
		a.startTextOCR()
	default:
		return false
	}
	return true
}

func normalizedTextPoint(area rect, x, y int32) textPoint {
	return textPoint{
		max(0, min(1, float64(x-area.left)/float64(area.right-area.left))),
		max(0, min(1, float64(y-area.top)/float64(area.bottom-area.top))),
	}
}

func (a *app) textSheetAt(x, y int32, source int) (rect, int) {
	v := a.view
	_, h := v.viewportSize()
	for _, page := range v.spread.visible(v.scrollY, h) {
		number := v.order[page.index]
		area := v.pageBounds(page)
		if number == source || (source == 0 && contains(area, x, y)) {
			return area, number
		}
	}
	return rect{}, 0
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
	if a.mode != idTextTab || a.busy || !v.ready() || a.full {
		return false
	}
	switch msg {
	case 0x0201: // WM_LBUTTONDOWN
		if !contains(v.viewport, x, y) {
			return false
		}
		setFocus.Call(a.window)
		area, source := a.textSheetAt(x, y, 0)
		if source == 0 {
			return false
		}
		pos := normalizedTextPoint(area, x, y)
		padding := float64(a.scale(5))
		tolerance := textPoint{padding / float64(area.right-area.left), padding / float64(area.bottom-area.top)}
		pick := textSelection{source: source, generation: v.generation, mode: comboIndex(p.mode), point: pos, tolerance: tolerance, box: v.box()}
		switch pick.mode {
		case textRectangle:
			pick.polygon = rectangleSelection(pos, pos)
			p.drag, p.gesture = true, pick
			user32.NewProc("SetCapture").Call(a.window)
			v.invalidateCanvas()
		case textLasso:
			gesture := &p.gesture
			if gesture.source != source || gesture.generation != v.generation || gesture.mode != textLasso {
				p.gesture = pick
			}
			if len(gesture.polygon) >= 3 && math.Abs(pos[0]-gesture.polygon[0][0]) <= tolerance[0]*2 &&
				math.Abs(pos[1]-gesture.polygon[0][1]) <= tolerance[1]*2 {
				a.completeTextGesture()
			} else if len(gesture.polygon) < 64 {
				gesture.polygon = append(gesture.polygon, pos)
				v.invalidateCanvas()
			}
		default:
			a.selectText(pick)
		}
		return true
	case 0x0200, 0x0202: // WM_MOUSEMOVE, WM_LBUTTONUP
		if p.drag {
			pick := &p.gesture
			area, source := a.textSheetAt(x, y, pick.source)
			if source != 0 {
				pick.polygon = rectangleSelection(pick.point, normalizedTextPoint(area, x, y))
			}
			if msg == 0x0202 {
				p.drag = false
				user32.NewProc("ReleaseCapture").Call()
				if source != 0 {
					end := pick.polygon[2]
					if math.Abs(end[0]-pick.point[0]) < pick.tolerance[0] || math.Abs(end[1]-pick.point[1]) < pick.tolerance[1] {
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
	selection := p.gesture
	p.gesture = textSelection{}
	a.selectText(selection)
}

func (a *app) selectText(selection textSelection) {
	p := &a.textTool
	p.pick, p.selected = selection, nil
	setText(p.input, "")
	setText(p.original, "")
	setText(p.report, "")
	a.view.invalidateCanvas()
	if p.model != nil && p.source == selection.source && p.generation == selection.generation {
		a.matchTextSelection()
		return
	}
	p.model = nil
	req := a.view.request("text_probe")
	req.View.Page = selection.source
	if a.run(req) {
		setText(p.report, "Načítám textovou vrstvu stránky…")
	}
}

func (a *app) finishTextProbe(result response) {
	p := &a.textTool
	if p.pick.source == 0 || p.pick.generation != a.view.generation || a.mode != idTextTab {
		return
	}
	if result.Status != "ok" || result.View == nil || result.View.TextPage == nil {
		a.report(result.Message)
		return
	}
	p.model, p.source, p.generation = result.View.TextPage, p.pick.source, p.pick.generation
	a.matchTextSelection()
}

func (a *app) textBounds() []searchRect {
	p := &a.textTool
	bounds := make([]searchRect, len(p.model.Objects))
	box := p.model.Boxes[p.pick.box]
	rotation := p.model.Rotation + a.view.rotations[p.pick.source-1]
	for i, run := range p.model.Objects {
		bounds[i], _ = searchDisplayRect(run.Bounds, box, rotation)
	}
	return bounds
}

func (a *app) matchTextSelection() {
	p := &a.textTool
	bounds := a.textBounds()
	if len(p.pick.polygon) >= 3 {
		for i, b := range bounds {
			if b[2] > b[0] && b[3] > b[1] &&
				pointInTextRegion(textPoint{(b[0] + b[2]) / 2, (b[1] + b[3]) / 2}, p.pick.polygon) {
				p.selected = append(p.selected, i)
			}
		}
		sortTextSelection(p.selected, bounds)
	} else if first := textHit(bounds, p.pick.point, p.pick.tolerance); first >= 0 {
		p.selected = []int{first}
		if p.pick.mode == textBlock {
			p.selected = textParagraph(bounds, first)
		}
	}
	message := "V místě kliknutí nebyl nalezen text. Pro OCR vyznačte oblast obdélníkem nebo lasem."
	setText(p.size, "")
	if len(p.selected) == 0 {
		if len(p.pick.polygon) >= 3 {
			message = "Oblast neobsahuje textovou vrstvu. Tlačítkem OCR výběru můžete rozpoznat sken nebo křivky."
		}
	} else {
		first := p.model.Objects[p.selected[0]]
		values := make([]string, 0, len(p.selected))
		editable := len(p.selected) <= 64
		for _, index := range p.selected {
			run := p.model.Objects[index]
			values = append(values, run.Text)
			editable = editable && run.Token > 0
		}
		setText(p.input, multiline(strings.Join(values, "\n")))
		setText(p.original, fmt.Sprintf("Původní písmo  %s", first.Font))
		setText(p.size, strings.ReplaceAll(strconv.FormatFloat(first.Size, 'f', 2, 64), ".", ","))
		chooseCombo(p.font, first.Family)
		a.setChecked(p.bold, first.Bold)
		a.setChecked(p.italic, first.Italic)
		a.setChecked(p.underline, first.Underline)
		switch {
		case !editable && len(p.selected) > 64:
			message = "Pro úpravu vyberte nejvýš 64 textových objektů."
		case !editable:
			message = "Výběr lze kopírovat. Obsahuje text ve skupině, souvislé sazbě nebo ve vrstvě OCR."
		case len(p.selected) > 1:
			message = "Každý řádek odpovídá jednomu objektu. Formátování se použije na celý výběr."
		default:
			message = ""
		}
	}
	setText(p.report, multiline(message))
	a.updateTextActions()
	a.view.invalidateCanvas()
}

func (a *app) startTextEdit() {
	p, v := &a.textTool, a.view
	if a.busy || p.model == nil || p.pick.generation != v.generation || len(p.selected) == 0 || len(p.selected) > 64 {
		return
	}
	size, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(getText(p.size)), ",", "."), 64)
	if err != nil || math.IsNaN(size) || size < 1 || size > 600 {
		a.report("Zadejte velikost písma od 1 do 600 bodů.")
		return
	}
	values := strings.Split(strings.ReplaceAll(getText(p.input), "\r\n", "\n"), "\n")
	if len(values) != len(p.selected) {
		a.report("Ponechte stejný počet řádků. Smazáním textu na řádku odstraníte příslušný objekt.")
		return
	}
	edit := &textEditRequest{Family: comboIndex(p.font), Size: size,
		Bold: a.checked(p.bold), Italic: a.checked(p.italic), Underline: a.checked(p.underline)}
	for i, index := range p.selected {
		token := p.model.Objects[index].Token
		if token <= 0 {
			return
		}
		edit.Runs = append(edit.Runs, textReplacement{token, values[i]})
	}
	req := v.request("text_edit")
	req.View.Page, req.TextEdit = p.pick.source, edit
	a.runDocumentDraft(req, "Upravuji text v otevřené kartě…", p.report)
}

func (a *app) startTextOCR() {
	p, v := &a.textTool, a.view
	if a.busy || p.model == nil || p.pick.generation != v.generation || len(p.selected) != 0 || len(p.pick.polygon) < 3 {
		return
	}
	position := slices.Index(v.order, p.pick.source)
	if position < 0 {
		return
	}
	box := p.model.Boxes[p.pick.box]
	rotation := p.model.Rotation + v.rotations[p.pick.source-1]
	req := v.request("ocr")
	req.View.Pages = []int{position + 1}
	req.OCRLanguage = [...]string{"ces+eng", "ces", "eng"}[max(0, comboIndex(a.ocr.language))]
	for _, point := range p.pick.polygon {
		req.OCRRegion = append(req.OCRRegion, textPointToPDF(point, box, rotation))
	}
	a.runDocumentDraft(req, "Rozpoznávám vybranou oblast…", p.report)
}

func (a *app) paintTextSelection(dc uintptr, source int, area rect) {
	p := &a.textTool
	if a.mode != idTextTab || a.full ||
		(p.pick.source != source && p.gesture.source != source) {
		return
	}
	pen, _, _ := gdi32.NewProc("CreatePen").Call(0, uintptr(max(1, a.scale(2))), paletteColor{174, 53, 94}.win())
	old, _, _ := gdi32.NewProc("SelectObject").Call(dc, pen)
	if p.pick.source == source && p.pick.generation == a.view.generation {
		drawTextOutline(dc, area, p.pick.polygon, true)
		if p.model != nil && len(p.selected) != 0 {
			for _, index := range p.selected {
				run := p.model.Objects[index]
				if b, ok := searchDisplayRect(run.Bounds, p.model.Boxes[p.pick.box], p.model.Rotation+a.view.rotations[source-1]); ok {
					drawTextOutline(dc, area, rectangleSelection(textPoint{b[0], b[1]}, textPoint{b[2], b[3]}), true)
				}
			}
		}
	}
	if p.gesture.source == source && p.gesture.generation == a.view.generation {
		drawTextOutline(dc, area, p.gesture.polygon, p.gesture.mode == textRectangle)
	}
	gdi32.NewProc("SelectObject").Call(dc, old)
	gdi32.NewProc("DeleteObject").Call(pen)
}

func drawTextOutline(dc uintptr, area rect, polygon []textPoint, closed bool) {
	if len(polygon) == 0 {
		return
	}
	vertices := make([]point, 0, len(polygon)+1)
	for _, pos := range polygon {
		vertices = append(vertices, point{area.left + int32(pos[0]*float64(area.right-area.left)),
			area.top + int32(pos[1]*float64(area.bottom-area.top))})
	}
	if closed {
		vertices = append(vertices, vertices[0])
	}
	if len(vertices) == 1 {
		position := vertices[0]
		fillColor(dc, rect{position.x - 3, position.y - 3, position.x + 4, position.y + 4}, paletteColor{174, 53, 94})
		return
	}
	gdi32.NewProc("Polyline").Call(dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
}
