package main

import "slices"

type canvasSelectionState struct {
	view       *documentView
	generation int
	tool, mode int
	textCursor bool
	paragraph  bool
	pick       textSelection
	mixed      mixedCanvasSelection
	page       *objectPage
	objects    []pdfObject
	text       *textPage
	flow       textFlow
	stored     *storedTextBlock
	dim        textPoint
	rotation   int
}

type canvasSelectionHistory struct {
	undo, redo []canvasSelectionState
}

func (v *documentView) captureCanvasSelection() canvasSelectionState {
	a := v.app
	state := canvasSelectionState{view: v, generation: v.generation, tool: a.mode, textCursor: a.canvasTextCursor}
	if v != a.view {
		return state
	}
	switch a.mode {
	case idObjectsTab:
		state.mode = a.objects.mode
		if a.objects.pick.generation == v.generation && len(a.objects.items) != 0 {
			state.pick, state.page, state.objects = a.objects.pick, a.objects.page, a.objects.items
		}
	case idTextTab:
		p := &a.textTool
		state.mode, state.paragraph = p.mode, p.explicitParagraph
		if p.pick.generation == v.generation && p.pick.source != 0 && (p.flow.text != "" || p.pick.region()) {
			state.pick, state.text, state.flow, state.dim, state.rotation = p.pick, p.model, p.flow, p.dim, p.rotation
			if frame := a.inlineCurrent(); frame != nil {
				state.stored = frame.stored
			}
		}
	}
	if v.mixed.pick.generation == v.generation && v.mixed.pick.source != 0 {
		state.pick, state.mixed = v.mixed.pick, v.mixed
		state.mixed.pick.history = nil
	}
	state.pick.history = nil
	return state
}

func (s canvasSelectionState) sameContent(next canvasSelectionState) bool {
	if s.pick.source != next.pick.source || s.text != next.text || s.page != next.page ||
		s.mixed.text != next.mixed.text || s.mixed.objects != next.mixed.objects {
		return false
	}
	sameObjects := func(a, b pdfObject) bool { return a.Token == b.Token }
	sameGlyphs := func(a, b layoutGlyph) bool { return glyphKey(a) == glyphKey(b) }
	return slices.EqualFunc(s.objects, next.objects, sameObjects) &&
		slices.EqualFunc(s.mixed.items, next.mixed.items, sameObjects) &&
		slices.EqualFunc(s.mixed.glyphs, next.mixed.glyphs, sameGlyphs) &&
		slices.EqualFunc(s.flow.glyphs(), next.flow.glyphs(), sameGlyphs)
}

func (v *documentView) rememberCanvasSelection(previous *canvasSelectionState) {
	v.app.syncLayerSelection()
	if previous == nil || previous.view != v || previous.generation != v.generation ||
		previous.sameContent(v.captureCanvasSelection()) {
		return
	}
	steps := &v.selectionHistory
	if len(steps.undo) >= 32 {
		steps.undo = slices.Delete(steps.undo, 0, 1)
	}
	steps.undo = append(steps.undo, *previous)
	steps.redo = nil
	v.app.updateHistoryButtons()
	v.app.updateObjectActions()
}

func (v *documentView) canvasHistoryStack(command documentCommand) *[]canvasSelectionState {
	if command == docRedo {
		return &v.selectionHistory.redo
	}
	return &v.selectionHistory.undo
}

func (v *documentView) canCanvasHistory(command documentCommand) bool {
	stack := *v.canvasHistoryStack(command)
	return v == v.app.view && !v.overview && len(stack) > 0 && stack[len(stack)-1].generation == v.generation
}

func (v *documentView) stepCanvasHistory(command documentCommand) bool {
	if !v.canCanvasHistory(command) {
		return false
	}
	a := v.app
	from, to := v.canvasHistoryStack(command), &v.selectionHistory.redo
	if command == docRedo {
		to = &v.selectionHistory.undo
	}
	*to = append(*to, v.captureCanvasSelection())
	last := len(*from) - 1
	state := (*from)[last]
	*from = slices.Delete(*from, last, last+1)

	a.cancelCanvasSelections()
	a.cancelCanvasSelectionProbes()
	v.clearMixedSelection()
	a.clearTextSelection()
	a.clearObjectSelection()
	a.activate(state.tool)
	a.canvasTextCursor = state.textCursor
	switch state.tool {
	case idObjectsTab:
		a.objects.mode = state.mode
	case idTextTab:
		a.textTool.mode, a.textTool.explicitParagraph = state.mode, state.paragraph
		a.syncTextSelectionModes()
	}
	switch {
	case state.mixed.pick.source != 0:
		v.mixed = state.mixed
		v.reportMixedSelection()
	case state.page != nil:
		v.applyObjectSelection(state.page, state.pick, state.objects)
	case state.text != nil:
		a.applyTextSelection(v, state.text, state.pick, state.flow, state.stored, state.dim, state.rotation)
	}
	a.updateObjectActions()
	a.updateHistoryButtons()
	a.updateCommandBar()
	v.invalidateCanvas()
	return true
}
