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
	WholeRunOnly            bool `json:"whole_run_only"`
	Font                    string
	Family                  int
	Text                    string
	Block                   *storedTextBlock
	Fill                    [3]uint8
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
	Baseline  *float64         `json:"baseline_offset,omitempty"`
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
				// One PDF show instruction can cover several columns. Pick the
				// connected text at the pointer, including neighboring word runs.
				flow = textLineAt(glyphs, seed)
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
			tab := a.comparison.tab
			if a.saveInlineBefore(func() {
				if a.comparisonVisible() && a.comparison.tab == tab {
					a.matchTextSelection(v, model, pick)
				}
			}) {
				return
			}
			a.selectDocumentTab(tab)
			if a.activeTab != tab {
				return
			}
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
	setText(p.original, "")
	setText(p.report, "")
	a.setInlineFrames(flow, stored, pick)
	if flow.text == "" {
		message := "Vyberte text nebo vyznačte oblast pro OCR."
		if len(pick.polygon) >= 3 {
			message = "V oblasti není textová vrstva. Použijte OCR výběru pro sken nebo křivky."
		}
		setText(p.report, message)
	} else if len(p.inline.frames) > 1 {
		setText(p.report, "Klikněte do rámečku, který chcete upravit.")
	} else if frame := a.inlineCurrent(); frame != nil && !frame.editable {
		setText(p.report, "Tento výběr lze zkopírovat. Zkuste vybrat celý textový objekt nebo užší oblast.")
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
			} else if run.Block.Row != previous.Row && (previous.WordWrap || run.Block.Row != previous.Row+1) {
				text.WriteByte(' ')
			}
		}
		// Font fallback can produce several text operators on one row.
		// Preserve spaces belonging to the fragment before joining them.
		if i > 0 && run.Block.Row == runs[i-1].Block.Row {
			previousGlyphs := runs[i-1].Glyphs
			if len(previousGlyphs) > 0 && strings.TrimSpace(previousGlyphs[len(previousGlyphs)-1].Text) == "" {
				text.WriteByte(' ')
			} else if len(run.Glyphs) > 0 && strings.TrimSpace(run.Glyphs[0].Text) == "" {
				text.WriteByte(' ')
			}
		}
		text.WriteString(run.Text)
	}
	flow.text = text.String()
	return flow, runs[0].Block
}
