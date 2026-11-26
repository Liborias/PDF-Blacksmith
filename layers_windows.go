package main

import (
	"maps"
	"strconv"
	"unsafe"
)

const (
	idToggleLayers = 238
	idLayerTree    = 239
)

type pageInventory struct {
	Items   []inventoryItem
	Layers  []inventoryLayer
	Limited bool
}

type inventoryItem struct {
	ID, Kind, Name string
	Object, Text   int // Existing object and text probe tokens in this page revision.
	Layers         []string
	Children       []inventoryItem
}

type inventoryLayer struct {
	ID, Name, State string
	Locked          bool
}

type inventoryTarget struct{ object, text int }

type layerTreeRow struct {
	id, label string
	depth     int
	branch    bool
	item      *inventoryItem
}

type layerPageKey struct {
	tab                *documentTab
	source, generation int
}

type layersPanel struct {
	sidebar, toggle, heading, status, list uintptr
	key                                    layerPageKey
	model                                  *pageInventory
	requested, failed                      bool
	rows                                   []layerTreeRow
	folded                                 map[string]bool
	objects, text                          map[int]bool
}

func (a *app) buildLayersPanel() {
	p := &a.layers
	a.controlParent = 0
	p.toggle = a.shellButton(uiLayers, idToggleLayers, sidebarRailWidth, 0)
	a.controlParent = p.sidebar
	p.heading = a.uiLabel(uiLayersAndObjects, 0, 0, 1, true)
	a.panelLabels[p.heading] = true
	p.status = a.label("", 0, 0, 1, false)
	a.panelLabels[p.status] = true
	const style = 0x0001 | 0x0010 | 0x0040 | 0x0100 // Notify, owner draw, strings, partial rows.
	p.list = a.add("LISTBOX", "", idLayerTree, 0, 0, 1, 1, wsTabstop|style, 0, a.normalFont)
	a.bindUILabel(p.list, uiLayersAndObjects)
	a.bindUIHelp(p.list, uiLayerTreeHelp)
	sendMessage.Call(p.list, 0x01a0, 0, uintptr(a.scale(24))) // LB_SETITEMHEIGHT
	a.addScrollPane(p.list, a.scale(24), true)
	a.controlParent = 0
}

func (a *app) layoutLayers() {
	p, area := &a.layers, a.rightAuxArea
	pane := a.scrollPanes[p.list]
	if pane == nil {
		return
	}
	width, height := int(area.right-area.left), int(area.bottom-area.top)
	show := width > 0 && a.visibleRightAuxiliary() == auxiliaryLayers
	if !show {
		pane.arrange(rect{}, false)
		showControl(p.sidebar, false)
		return
	}
	place(p.sidebar, int(area.left), int(area.top), width, height)
	d := a.scale
	margin := min(d(12), width/10)
	place(p.heading, margin, d(8), width-2*margin, d(22))
	place(p.status, margin, d(34), width-2*margin, d(22))
	pane.arrange(rect{int32(margin), int32(d(62)), int32(width - margin), int32(height - margin)}, len(p.rows) > 0)
	showControl(p.sidebar, true)
}

func (a *app) syncLayersPage() {
	p, v := &a.layers, a.view
	var next layerPageKey
	if v.ready() {
		next = layerPageKey{a.activeTab, v.order[v.page], v.generation}
	}
	if p.key == next {
		return
	}
	visible := a.visibleRightAuxiliary() == auxiliaryLayers
	needsReset := visible || p.requested || len(p.rows) > 0
	p.key, p.model = next, nil
	p.requested, p.failed, p.folded = false, false, nil
	p.objects, p.text = nil, nil
	if needsReset {
		a.rebuildLayerRows(false)
	}
}

func (v *documentView) scheduleLayers() bool {
	a := v.app
	p := &a.layers
	if v.secondary || a.visibleRightAuxiliary() != auxiliaryLayers || p.requested || p.key.source == 0 {
		return false
	}
	p.requested = true
	a.setUIMessage(p.status, uiLayerReading)
	req := v.request("page_inventory")
	req.View.Page = p.key.source
	v.launch(req, p.key.source, 0, true, nil)
	return true
}

func (a *app) finishLayers(job readerJob) {
	p := &a.layers
	if p.key != (layerPageKey{a.activeTab, job.page, job.generation}) {
		return
	}
	p.failed = job.result.Status != "ok" || job.result.View == nil || job.result.View.Inventory == nil
	if !p.failed {
		p.model = job.result.View.Inventory
	}
	a.rebuildLayerRows(false)
	a.syncLayerSelection()
}

func (a *app) rebuildLayerRows(keepPosition bool) {
	p := &a.layers
	if p.list == 0 {
		return
	}
	selected := p.selected()
	_, _, top := a.scrollPanes[p.list].lines()
	p.rows = nil
	switch {
	case p.failed:
		a.setUIMessage(p.status, uiLayerReadError)
	case p.model == nil:
		a.setUIMessage(p.status, uiLayerReading)
	default:
		title := uiLayerPage
		if p.model.Limited {
			title = uiLayerPageLimited
		}
		a.setUIMessage(p.status, title, a.view.page+1)
		a.appendLayerRows()
	}
	sendMessage.Call(p.list, 0x000b, 0, 0)                                          // WM_SETREDRAW
	sendMessage.Call(p.list, 0x0184, 0, 0)                                          // LB_RESETCONTENT
	sendMessage.Call(p.list, 0x01a8, uintptr(len(p.rows)), uintptr(96*len(p.rows))) // LB_INITSTORAGE
	chosen := -1
	for index, row := range p.rows {
		sendMessage.Call(p.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(row.label)))) // LB_ADDSTRING
		if keepPosition && row.id == selected.id {
			chosen = index
		}
	}
	if chosen >= 0 {
		sendMessage.Call(p.list, 0x0186, uintptr(chosen), 0) // LB_SETCURSEL
	}
	if keepPosition {
		sendMessage.Call(p.list, 0x0197, uintptr(max(0, min(top, len(p.rows)-1))), 0) // LB_SETTOPINDEX
	}
	sendMessage.Call(p.list, 0x000b, 1, 0)
	a.layoutLayers()
	a.scrollPanes[p.list].reflow()
	invalidate(p.list)
}

func (a *app) appendLayerRows() {
	p, model := &a.layers, a.layers.model
	names := make(map[string]string, len(model.Layers))
	if len(model.Layers) > 0 {
		p.rows = append(p.rows, layerTreeRow{id: "layers", label: a.uiText(uiPDFLayers), branch: true})
	}
	for number, layer := range model.Layers {
		name := layer.Name
		if name == "" {
			name = a.uiText(uiLayers) + " " + strconv.Itoa(number+1)
		}
		names[layer.ID] = name
		if p.folded["layers"] {
			continue
		}
		state := uiLayerStateUnknown
		switch layer.State {
		case "on":
			state = uiLayerStateOn
		case "off":
			state = uiLayerStateOff
		}
		caption := name + " · " + a.uiText(state)
		if layer.Locked {
			caption += " · " + a.uiText(uiLayerLocked)
		}
		p.rows = append(p.rows, layerTreeRow{id: "ocg/" + layer.ID, label: caption, depth: 1})
	}
	p.rows = append(p.rows, layerTreeRow{id: "art", label: a.uiText(uiLayerContents), branch: len(model.Items) > 0})
	if p.folded["art"] {
		return
	}
	if len(model.Items) == 0 {
		p.rows = append(p.rows, layerTreeRow{id: "empty", label: a.uiText(uiLayerEmpty), depth: 1})
		return
	}
	counts := make(map[string]int)
	var appendItems func([]inventoryItem, int, bool)
	appendItems = func(items []inventoryItem, depth int, visible bool) {
		for index := range items {
			item := &items[index]
			counts[item.Kind]++
			row := layerTreeRow{id: item.ID, depth: depth, branch: len(item.Children) > 0, item: item}
			if visible {
				row.label = a.uiText(layerItemNames[item.Kind]) + " " + strconv.Itoa(counts[item.Kind])
				if item.Name != "" {
					row.label += " · " + item.Name
				}
				for _, id := range item.Layers {
					if name := names[id]; name != "" {
						row.label += " · " + name
					}
				}
				p.rows = append(p.rows, row)
			}
			appendItems(item.Children, depth+1, visible && !p.isFolded(row))
		}
	}
	appendItems(model.Items, 1, true)
}

var layerItemNames = map[string]uiMessage{
	"path": uiVectorShape, "image": uiLayerImage, "form": uiPDFGroup, "group": uiLayerTransparencyGroup,
	"text": uiText, "hidden_text": uiLayerHiddenText, "widget": uiLayerWidget, "shading": uiLayerShading,
}

func (p *layersPanel) selected() layerTreeRow {
	index, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0) // LB_GETCURSEL
	if index >= uintptr(len(p.rows)) {
		return layerTreeRow{}
	}
	return p.rows[index]
}

func (a *app) layerIndent(row layerTreeRow, width int) int {
	d := a.scale
	return min(d(4+row.depth*12), max(d(4), width-d(70)))
}

func (a *app) drawLayerRow(item *buttonDraw) bool {
	p := &a.layers
	if p.list == 0 || item.handle != p.list || item.controlType != 2 {
		return false
	}
	bounds, fill, ink := item.bounds, panelColor, textColor
	if item.itemID >= uint32(len(p.rows)) {
		fillColor(item.dc, bounds, fill)
		return true
	}
	row := p.rows[item.itemID]
	selected := row.item != nil && (p.objects[row.item.Object] || p.text[row.item.Text])
	if selected {
		fill, ink = uiColors["selected"], selectionColor
	} else if row.item != nil && !row.branch && row.item.Object == 0 && row.item.Text == 0 {
		ink = mutedColor
	}
	fillColor(item.dc, bounds, fill)
	d := a.scale
	if selected {
		fillColor(item.dc, rect{bounds.left, bounds.top, bounds.left + int32(d(2)), bounds.bottom}, selectionColor)
	}
	bounds.left += int32(a.layerIndent(row, int(bounds.right-bounds.left)))
	if row.branch {
		drawTreeChevron(item.dc, bounds.left+int32(d(6)), (bounds.top+bounds.bottom)/2, d(4), d(2), ink, p.isFolded(row))
	}
	font := a.normalFont
	if row.depth == 0 {
		font = a.labelFont
	}
	bounds.left += int32(d(17))
	bounds.right -= int32(d(4))
	drawText(item.dc, row.label, bounds, font, ink, 0x8000) // DT_END_ELLIPSIS
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (p *layersPanel) isFolded(row layerTreeRow) bool {
	if folded, changed := p.folded[row.id]; changed {
		return folded
	}
	return row.item != nil
}

func (a *app) foldLayerRow(row layerTreeRow, folded bool) {
	p := &a.layers
	if !row.branch || p.isFolded(row) == folded {
		return
	}
	if p.folded == nil {
		p.folded = make(map[string]bool)
	}
	p.folded[row.id] = folded
	a.rebuildLayerRows(true)
}

func (a *app) layerMouseAction(location uintptr) {
	p := &a.layers
	index, _, _ := sendMessage.Call(p.list, 0x01a9, 0, location) // LB_ITEMFROMPOINT
	if index>>16 != 0 || index >= uintptr(len(p.rows)) {
		return
	}
	row := p.rows[index]
	var bounds rect
	sendMessage.Call(p.list, 0x0198, index, uintptr(unsafe.Pointer(&bounds))) // LB_GETITEMRECT
	x, y := int32(int16(location)), int32(int16(location>>16))
	if !contains(bounds, x, y) {
		return
	}
	indent := a.layerIndent(row, int(bounds.right-bounds.left))
	if row.branch && (row.item == nil || row.item.Object == 0 || int(x) < indent+a.scale(16)) {
		a.foldLayerRow(row, !p.isFolded(row))
	} else {
		a.selectLayerItem(row)
	}
}

func (a *app) layerKey(msg *winMessage) bool {
	if msg.window != a.layers.list || keyDown(0x12) {
		return false
	}
	if keyDown(0x11) {
		command, ok := viewShortcuts[documentShortcut{msg.wParam, keyDown(0x10)}]
		if ok && (command == docUndo || command == docRedo) {
			a.doDocumentCommand(a.view, command, false)
			return true
		}
	}
	row := a.layers.selected()
	switch msg.wParam {
	case 0x0d, 0x20: // Enter or Space.
		if row.branch && (row.item == nil || row.item.Object == 0) {
			a.foldLayerRow(row, !a.layers.isFolded(row))
		} else {
			a.selectLayerItem(row)
		}
	case 0x25, 0x27: // Left or Right.
		a.foldLayerRow(row, msg.wParam == 0x25)
	default:
		return false
	}
	return true
}

func (a *app) selectLayerItem(row layerTreeRow) {
	node, v := row.item, a.view
	if node == nil || a.busy || v.overview || !v.ready() || node.Object == 0 && node.Text == 0 {
		return
	}
	pick := v.canvasPick()
	pick.source, pick.mode, pick.inventory = a.layers.key.source, textObject, inventoryTarget{node.Object, node.Text}
	pick.automatic = true
	pick.combined = pick.change != replaceSelection && v.hasCanvasSelection(pick)
	v.dispatchCanvasSelection(pick)
}

func (a *app) syncLayerSelection() {
	p, v := &a.layers, a.view
	if p.model == nil || p.key.tab != a.activeTab || p.key.generation != v.generation || a.visibleRightAuxiliary() != auxiliaryLayers {
		return
	}
	state := v.captureCanvasSelection()
	objects, texts := make(map[int]bool), make(map[int]bool)
	if state.pick.source == p.key.source {
		items, glyphs, page := state.objects, state.flow.glyphs(), state.text
		if state.mixed.pick.source != 0 {
			items, glyphs, page = state.mixed.items, state.mixed.glyphs, state.mixed.text
		}
		for _, item := range items {
			objects[item.Token] = true
		}
		for _, glyph := range glyphs {
			if token := page.Objects[glyph.run].BlockToken; token != 0 {
				texts[token] = true
			}
		}
	}
	if !maps.Equal(p.objects, objects) || !maps.Equal(p.text, texts) {
		p.objects, p.text = objects, texts
		invalidate(p.list)
	}
}
