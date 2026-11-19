package main

import (
	"math"
	"slices"
	"strings"
)

type textGlyph struct {
	Text   string
	Bounds searchRect
	Origin textPoint
}

type storedTextBlock struct {
	ID, Row, Paragraph       int
	Width, Leading, Tracking float64
	Alignment, Rotation      int
	Origin                   textPoint
	WordWrap                 bool `json:"word_wrap"`
}

type textRun struct {
	BlockToken              int `json:"block_token"`
	Editable                bool
	Font                    string
	Family                  int
	Text                    string
	Block                   *storedTextBlock
	Size                    float64
	Matrix                  [6]float64
	Bold, Italic, Underline bool
	Glyphs                  []textGlyph
}

type textPage struct {
	Objects  []textRun
	Boxes    map[string]searchRect
	Rotation int
}

type glyphSelection struct {
	Token  int   `json:"token"`
	Glyphs []int `json:"glyphs"`
}

type textBlockEdit struct {
	Text      string           `json:"text"`
	Align     string           `json:"align"`
	Width     float64          `json:"width"`
	Height    float64          `json:"height"`
	Leading   float64          `json:"leading"`
	Tracking  float64          `json:"tracking"`
	Origin    textPoint        `json:"origin"`
	Rotation  int              `json:"rotation"`
	Selection []glyphSelection `json:"selection"`
}

type textEditRequest struct {
	Font      string        `json:"font"`
	Size      float64       `json:"size"`
	Bold      bool          `json:"bold"`
	Italic    bool          `json:"italic"`
	Underline bool          `json:"underline"`
	Block     textBlockEdit `json:"block"`
}

type textProbeState struct {
	model              *textPage
	source, generation int
	next, flight       textSelection
}

func (v *documentView) nativeTextGeometry(model *textPage, pick textSelection) ([]layoutGlyph, textPoint, int) {
	box := model.Boxes[pick.box]
	rotation := (model.Rotation + v.rotations[pick.source-1] + 360) % 360
	dim := textPoint{box[2] - box[0], box[3] - box[1]}
	if dim[0] <= 0 || dim[1] <= 0 {
		return nil, dim, rotation
	}
	if rotation%180 != 0 {
		dim[0], dim[1] = dim[1], dim[0]
	}
	var glyphs []layoutGlyph
	for i, run := range model.Objects {
		em := max(1., run.Size*math.Hypot(run.Matrix[2], run.Matrix[3]))
		for j, glyph := range run.Glyphs {
			if b, ok := searchDisplayRect(glyph.Bounds, box, rotation); ok {
				baseline := pdfPointToText(glyph.Origin, box, rotation)[1] * dim[1]
				glyphs = append(glyphs, layoutGlyph{i, j, glyph.Text,
					searchRect{b[0] * dim[0], b[1] * dim[1], b[2] * dim[0], b[3] * dim[1]}, baseline, em})
			}
		}
	}
	return glyphs, dim, rotation
}

func (v *documentView) requestNativeText(pick textSelection) {
	p := &v.nativeText
	if p.model != nil && p.source == pick.source && p.generation == pick.generation {
		p.next, p.flight = textSelection{}, textSelection{}
		v.app.matchTextSelection(v, p.model, pick)
		return
	}
	p.next = pick
	v.schedule()
}

func (v *documentView) scheduleNativeText() bool {
	p := &v.nativeText
	pick := p.next
	if pick.source == 0 || v.app.busy {
		return false
	}
	p.next = textSelection{}
	if pick.generation != v.generation || pick.tool != v.app.mode {
		return false
	}
	p.flight = pick
	req := v.request("text_probe")
	req.View.Page = pick.source
	v.launch(req, 0, 0, false, nil)
	return true
}

func (v *documentView) finishNativeText(job readerJob) {
	p := &v.nativeText
	pick := p.flight
	p.flight = textSelection{}
	if job.generation != v.generation || pick.generation != v.generation || pick.source == 0 {
		return
	}
	if p.next.source != 0 {
		if p.next.source != pick.source || p.next.generation != pick.generation {
			return
		}
		pick, p.next = p.next, textSelection{}
	}
	if job.result.Status != "ok" || job.result.View == nil || job.result.View.TextPage == nil {
		if !pick.automatic && v.app.mode == idTextTab {
			v.app.report(job.result.Message)
		}
		return
	}
	p.model, p.source, p.generation = job.result.View.TextPage, pick.source, pick.generation
	v.app.matchTextSelection(v, p.model, pick)
}

func (v *documentView) clickedForText(x, y int32) {
	a := v.app
	if a.busy || a.full || !v.ready() || (a.mode == idTextTab && (!v.secondary || a.textTool.gesture.source != 0)) {
		return
	}
	pick := v.textAt(x, y)
	if pick.source == 0 {
		return
	}
	pick.mode, pick.automatic = textObject, true
	v.requestNativeText(pick)
}

func (a *app) matchTextSelection(v *documentView, model *textPage, pick textSelection) {
	if pick.tool != a.mode || pick.generation != v.generation {
		return
	}
	glyphs, dim, rotation := v.nativeTextGeometry(model, pick)
	var flow textFlow
	var stored *storedTextBlock
	if len(pick.polygon) >= 3 {
		selected := slices.DeleteFunc(glyphs, func(g layoutGlyph) bool {
			b := g.bounds
			return !pointInTextRegion(textPoint{(b[0] + b[2]) / (2 * dim[0]), (b[1] + b[3]) / (2 * dim[1])}, pick.polygon)
		})
		flow = layoutText(selected)
	} else {
		bounds := make([]searchRect, len(glyphs))
		for i, g := range glyphs {
			bounds[i] = g.bounds
		}
		hit := textHit(bounds, textPoint{pick.point[0] * dim[0], pick.point[1] * dim[1]},
			textPoint{pick.tolerance[0] * dim[0], pick.tolerance[1] * dim[1]})
		if hit >= 0 {
			seed := glyphs[hit]
			if pick.mode == textBlock {
				flow = layoutText(glyphs).paragraphAt(seed)
			} else if block := model.Objects[seed.run].Block; block != nil {
				flow, stored = storedBlockFlow(model, glyphs, block.ID)
			} else {
				flow = layoutText(slices.DeleteFunc(glyphs, func(g layoutGlyph) bool { return g.run != seed.run }))
			}
		}
	}
	if pick.automatic {
		if flow.text == "" || a.busy {
			return
		}
		oldBox := pick.box
		raw := textPointToPDF(pick.point, model.Boxes[oldBox], rotation)
		if v.secondary {
			a.selectDocumentTab(a.comparison.tab)
			v = a.view
		}
		a.activate(idTextTab)
		a.textTool.mode = textObject
		a.textGuide()
		pick.tool, pick.generation, pick.box = idTextTab, v.generation, v.box()
		pick.point = pdfPointToText(raw, model.Boxes[pick.box], rotation)
		v.nativeText.model, v.nativeText.source, v.nativeText.generation = model, pick.source, pick.generation
		// Refit in the box of the active view if the previous tool displayed a print box.
		if oldBox != pick.box {
			// A generation change affects drawing and page framing, the source is the same PDF page.
			pick.automatic = false
			a.matchTextSelection(v, model, pick)
			return
		}
	}
	p := &a.textTool
	p.model, p.pick, p.flow, p.dim, p.rotation = model, pick, flow, dim, rotation
	p.editable = false
	p.stored = nil
	setText(p.input, multiline(flow.text))
	setText(p.original, "")
	setText(p.report, "")
	if flow.text == "" {
		message := "Vyberte text nebo vyznačte oblast pro OCR."
		if len(pick.polygon) >= 3 {
			message = "V oblasti není textová vrstva. Použijte OCR výběru pro sken nebo křivky."
		}
		setText(p.report, message)
	} else {
		selection := flow.glyphs()
		first := model.Objects[selection[0].run]
		p.editable = len(selection) <= 8000
		for _, glyph := range selection {
			if !model.Objects[glyph.run].Editable {
				p.editable = false
				break
			}
		}
		if len(pick.polygon) >= 3 {
			p.flow.bounds[0], p.flow.bounds[1], p.flow.bounds[2] = 1e20, 1e20, 0
			for _, point := range pick.polygon {
				p.flow.bounds[0] = min(p.flow.bounds[0], point[0]*dim[0])
				p.flow.bounds[1] = min(p.flow.bounds[1], point[1]*dim[1])
				p.flow.bounds[2] = max(p.flow.bounds[2], point[0]*dim[0])
			}
		}
		setText(p.original, first.Font)
		font := slices.IndexFunc(p.fonts, func(font string) bool { return strings.EqualFold(font, first.Font) })
		if font < 0 {
			font = min(max(0, first.Family), len(bundledTextFonts)-1)
		}
		chooseCombo(p.font, font)
		setText(p.size, formatTextNumber(first.Size))
		setText(p.width, formatTextNumber(math.Floor(min(dim[0]-p.flow.bounds[0], max(10, p.flow.bounds[2]-p.flow.bounds[0]+.75))*100)/100))
		setText(p.leading, "1,2")
		setText(p.tracking, "0")
		p.alignment = 0
		if stored != nil && stored.Rotation == rotation {
			anchor := pdfPointToText(stored.Origin, model.Boxes[pick.box], rotation)
			if anchor[0] >= 0 && anchor[1] >= 0 && anchor[0] < 1 && anchor[1] < 1 &&
				stored.Width >= 1 && stored.Width <= (1-anchor[0])*dim[0]+.003 {
				p.stored = stored
				setText(p.width, formatTextNumber(math.Floor(stored.Width*100)/100))
				setText(p.leading, formatTextNumber(stored.Leading))
				setText(p.tracking, formatTextNumber(stored.Tracking))
				p.alignment = min(max(stored.Alignment, 0), len(textAlignments)-1)
			}
		}
		a.setChecked(p.bold, first.Bold)
		a.setChecked(p.italic, first.Italic)
		a.setChecked(p.underline, first.Underline)
		if !p.editable {
			setText(p.report, "Tento výběr lze kopírovat. Část sazby nelze bezpečně oddělit, nebo výběr přesahuje 8 000 znaků.")
		}
	}
	a.updateTextActions()
	v.invalidateCanvas()
}

func storedBlockFlow(model *textPage, glyphs []layoutGlyph, group int) (textFlow, *storedTextBlock) {
	selected := slices.DeleteFunc(glyphs, func(g layoutGlyph) bool {
		block := model.Objects[g.run].Block
		return block == nil || block.ID != group
	})
	counts := make(map[int]int)
	for _, g := range selected {
		counts[g.run]++
	}
	var runs []textRun
	for index, run := range model.Objects {
		if run.Block == nil || run.Block.ID != group {
			continue
		}
		if counts[index] != len(run.Glyphs) {
			return layoutText(selected), nil
		}
		runs = append(runs, run)
	}
	flow := layoutText(selected)
	if len(runs) == 0 || len(flow.paragraphs) == 0 {
		return flow, nil
	}
	slices.SortStableFunc(runs, func(a, b textRun) int {
		if a.Block.Row < b.Block.Row {
			return -1
		}
		if a.Block.Row > b.Block.Row {
			return 1
		}
		return 0
	})
	var text strings.Builder
	for i, run := range runs {
		if i > 0 {
			previous := runs[i-1].Block
			if newlines := min(64, run.Block.Paragraph-previous.Paragraph); newlines > 0 {
				text.WriteString(strings.Repeat("\n", newlines))
			} else if previous.WordWrap || run.Block.Row != previous.Row+1 {
				text.WriteByte(' ')
			}
		}
		text.WriteString(run.Text)
	}
	flow.text = text.String()
	return flow, runs[0].Block
}
