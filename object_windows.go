package main

import (
	"math"
	"unsafe"
)

const (
	idObjectsTab = 500 + iota
	idObjectCopy
	idObjectPaste
	idObjectDuplicate
	idObjectDelete
	idObjectUndo
	idObjectPoint
	idObjectRectangle
	idObjectLasso
	idObjectCrop
	idObjectApplyCrop
	idObjectCancelCrop
	idObjectCropRatio
)

type objectEditRequest struct {
	Operation string      `json:"operation,omitempty"`
	Grouping  string      `json:"grouping,omitempty"`
	Token     int         `json:"token,omitempty"`
	Tokens    []int       `json:"tokens,omitempty"`
	Delta     textPoint   `json:"delta,omitempty"`
	At        *textPoint  `json:"at,omitempty"`
	Clipboard []byte      `json:"clipboard,omitempty"`
	Crop      *searchRect `json:"crop,omitempty"`
}

type pdfObject struct {
	Token   int
	Kind    string
	Bounds  searchRect
	Outline *objectOutline
}

type objectOutline struct {
	Contours  [][]textPoint
	Margin    textPoint
	Rectangle bool
}

type objectPage struct {
	Objects  []pdfObject
	Boxes    map[string]searchRect
	Rotation int
	Unit     float64
}

type objectEditor struct {
	info, copy, paste, duplicate, remove, undo uintptr
	page                                       *objectPage
	source                                     int
	pick                                       textSelection
	pressPick                                  textSelection
	dragReady                                  bool
	selected                                   pdfObject
	items                                      []pdfObject
	mode                                       int
	modes                                      [3]uintptr
	serial                                     uint64
	snapshot, snapshotComplete                 bool
	image                                      *printBitmap
	capsule                                    []byte
	copyWaiting                                bool
	copyGrouping, clipboardGrouping            string
	clipboard                                  []byte
	owner                                      *documentTab
	start, cursor                              point
	pressed, moving                            bool
	delta                                      textPoint
	preview                                    dragPreview
	crop                                       imageCrop
	cropTool, cropApply, cropCancel, cropHint  uintptr
	cropRatioLabel, cropRatio                  uintptr
}

func (a *app) buildObjectPanel() {
	e := &a.objects
	a.beginPropertySection(uiObjectSelection, 16, uiObjectSelectionHelp)
	for i, name := range []uiMessage{uiSelectByClick, uiRectangleSelection, uiPolygonLassoKey} {
		e.modes[i] = a.uiTextIconButton(name, idObjectPoint+i, 16+i*40, 42, 34)
	}
	a.addingSection = 0
	e.info = a.label("", 16, 85, 298, false)
	a.setUIMessage(e.info, uiSelectPageObject)
	a.beginPropertySection(uiObjectActions, 124, uiObjectActionsHelp)
	e.copy = a.uiButton(uiObjectCopyKey, idObjectCopy, 16, 165, 145, 0)
	e.paste = a.uiButton(uiObjectPasteKey, idObjectPaste, 169, 165, 145, 0)
	e.duplicate = a.uiButton(uiObjectDuplicateKey, idObjectDuplicate, 16, 206, 145, 0)
	e.remove = a.uiButton(uiObjectRemoveKey, idObjectDelete, 169, 206, 145, 0)
	e.undo = a.uiButton(uiObjectUndoKey, idObjectUndo, 16, 247, 298, 0)
	e.cropTool = a.uiTextIconButton(uiCropImage, idObjectCrop, 16, 288, 34)
	e.cropHint = a.readout(58, 288, 256, 48)
	a.buildCropRatio()
	e.cropApply = a.uiButton(uiApply, idObjectApplyCrop, 16, 379, 145, 0)
	e.cropCancel = a.uiButton(uiCancel, idObjectCancelCrop, 169, 379, 145, 0)
}

func (a *app) updateObjectActions() {
	e := &a.objects
	if e.info == 0 {
		return
	}
	ready := a.mode == idObjectsTab && a.view.ready() && !a.busy
	selected := ready && e.selected.Token != 0
	for i, handle := range e.modes {
		setEnabled(handle, ready)
		a.setChecked(handle, e.mode == i)
	}
	setEnabled(e.copy, selected)
	setEnabled(e.paste, ready && e.owner == a.activeTab && len(e.clipboard) != 0)
	setEnabled(e.duplicate, selected)
	setEnabled(e.remove, selected)
	setEnabled(e.undo, ready && len(a.view.history) > 0)
	setEnabled(e.cropTool, a.canCropImage())
	a.setChecked(e.cropTool, e.crop.active)
	_, canApplyCrop := a.imageCropBounds()
	setEnabled(e.cropRatio, ready && e.crop.active && !e.crop.dragging)
	setEnabled(e.cropApply, canApplyCrop)
	setEnabled(e.cropCancel, ready && e.crop.active)
}

func (a *app) clearObjectSelection() {
	e := &a.objects
	a.cancelImageCrop()
	a.cancelObjectGesture()
	e.page, e.source, e.pick = nil, 0, textSelection{}
	e.resetSelection()
	e.snapshot, e.snapshotComplete, e.copyWaiting = false, false, false
	if e.info != 0 {
		a.setUIMessage(e.info, uiSelectPageObject)
		a.updateObjectActions()
	}
}

func (e *objectEditor) endGesture() {
	e.preview.close()
	e.pressed, e.moving, e.dragReady, e.delta = false, false, false, textPoint{}
	e.pressPick = textSelection{}
}

func (a *app) cancelObjectGesture() {
	e := &a.objects
	if e.pressed || e.crop.dragging {
		e.endGesture()
		e.crop.dragging = false
		user32.NewProc("ReleaseCapture").Call()
		a.updateObjectActions()
		a.view.invalidateCanvas()
	}
}

func (v *documentView) objectPointer(msg uint32, x, y int32) bool {
	a, e := v.app, &v.app.objects
	if msg == 0x0215 || msg == 0x001f { // Capture changed or cancelled.
		if e.pressed {
			e.endGesture()
			v.invalidateCanvas()
		}
		return false
	}
	if a.mode != idObjectsTab || a.busy || !v.ready() || a.full {
		return false
	}
	switch msg {
	case 0x0201:
		pick := v.textAt(x, y)
		if pick.source == 0 || keyDown(0x20) || e.mode != 0 &&
			(pick.change != replaceSelection || !e.selectionContains(pick, v)) {
			return false
		}
		pick.mode = textObject
		e.pressPick, e.start, e.cursor, e.pressed = pick, point{x, y}, point{x, y}, true
		setFocus.Call(a.window)
		user32.NewProc("SetCapture").Call(a.window)
		return true
	case 0x0200:
		if e.pressed {
			e.cursor = point{x, y}
			v.prepareObjectDrag()
			a.moveObjectPreview()
			return true
		}
	case 0x0202:
		if e.pressed {
			e.cursor = point{x, y}
			a.moveObjectPreview()
			delta, moved := e.delta, e.moving
			click := max(absTextDrag(x-e.start.x), absTextDrag(y-e.start.y)) <= int32(a.scale(4))
			target := v.textAt(x, y)
			a.cancelObjectGesture()
			if moved && target.source == e.pick.source && math.Hypot(delta[0], delta[1]) > .01 {
				if grouping, accepted := a.chooseObjectGrouping(); accepted {
					a.runObjectEdit(e.pick.source, objectEditRequest{
						Operation: "move", Grouping: grouping, Token: e.selected.Token, Tokens: e.tokens(), Delta: delta,
					})
				}
			} else if click {
				v.selectionClick(x, y, false)
			}
			return true
		}
	}
	return false
}

func (v *documentView) prepareObjectDrag() {
	e := &v.app.objects
	if e.dragReady || e.pressPick.source == 0 || e.pressPick.change != replaceSelection ||
		max(absTextDrag(e.start.x-e.cursor.x), absTextDrag(e.start.y-e.cursor.y)) <= int32(v.app.scale(4)) {
		return
	}
	e.dragReady = true
	pick := e.pressPick
	if e.selectionContains(pick, v) {
		e.pick = pick
	} else {
		v.requestObjects(pick)
	}
}

func (e *objectEditor) rotation(v *documentView) int {
	return (e.page.Rotation + v.rotations[e.source-1] + 360) % 360
}

func (v *documentView) scheduleObjects() bool {
	a, e, state := v.app, &v.app.objects, &v.objectProbe
	if state.next.source != 0 && !a.busy {
		pick := state.next
		state.next = textSelection{}
		if pick.generation == v.generation && pick.tool == a.mode {
			state.flight = pick
			req := v.request("object_probe")
			req.View.Page = pick.source
			v.launch(req, 0, 0, false, nil)
			return true
		}
	}
	if v.secondary || a.mode != idObjectsTab || a.busy || e.pick.source == 0 || e.pick.generation != v.generation {
		return false
	}
	if e.selected.Token != 0 && !e.snapshot {
		e.snapshot = true
		req := v.request("object_snapshot")
		req.View.Page, req.View.Rotation = e.source, v.rotations[e.source-1]
		req.ObjectEdit = &objectEditRequest{Token: e.selected.Token, Tokens: e.tokens()}
		v.launch(req, e.source, 0, false, nil)
		return true
	}
	return false
}

func (v *documentView) finishObjects(job readerJob) {
	a, e := v.app, &v.app.objects
	if job.generation != v.generation {
		return
	}
	if job.action == "object_probe" {
		state := &v.objectProbe
		pick := state.flight
		state.flight = textSelection{}
		if pick.source != job.objectPage || pick.generation != v.generation {
			return
		}
		if state.next.source != 0 {
			if state.next.source != pick.source || state.next.generation != pick.generation {
				return
			}
			pick, state.next = state.next, textSelection{}
		}
		if job.result.Status != "ok" || job.result.View == nil || job.result.View.ObjectPage == nil {
			if pick.combined {
				v.failCanvasRegion()
			} else if a.mode == idObjectsTab {
				a.setUIMessage(e.info, uiObjectsUnavailable)
			}
			return
		}
		state.model, state.source, state.generation = job.result.View.ObjectPage, pick.source, v.generation
		v.matchObjectSelection(state.model, pick)
		return
	}
	if v.secondary || a.mode != idObjectsTab || job.objectPage != e.pick.source || job.objectSerial != e.serial {
		return
	}
	e.snapshot, e.snapshotComplete = true, true
	if job.result.Status == "ok" && job.result.View != nil {
		e.image, e.capsule = job.frame.image, job.result.View.ObjectClipboard
		if e.moving && e.image != nil {
			a.showObjectPreview()
		}
	}
	if e.copyWaiting {
		a.copySelectedObject()
	}
}

func (a *app) moveObjectPreview() {
	e, v := &a.objects, a.view
	if !e.pressed || !e.dragReady || e.selected.Token == 0 || e.pick.source != e.pressPick.source || max(absTextDrag(e.start.x-e.cursor.x), absTextDrag(e.start.y-e.cursor.y)) <= int32(a.scale(4)) && !e.moving {
		return
	}
	area, source := v.textSheetAt(0, 0, e.pick.source)
	if source == 0 {
		return
	}
	box, angle := e.page.Boxes[e.pick.box], e.rotation(v)
	size := textPoint{float64(area.right - area.left), float64(area.bottom - area.top)}
	at := textPoint{e.pick.point[0] + float64(e.cursor.x-e.start.x)/size[0],
		e.pick.point[1] + float64(e.cursor.y-e.start.y)/size[1]}
	p, start := textPointToPDF(at, box, angle), textPointToPDF(e.pick.point, box, angle)
	e.delta = textPoint{p[0] - start[0], p[1] - start[1]}
	if !e.moving {
		e.moving = true
		a.showObjectPreview()
	}
	e.preview.move(a.objectScreenCursor())
	v.invalidateCanvas()
}

func (a *app) objectScreenCursor() point {
	cursor := a.objects.cursor
	user32.NewProc("ClientToScreen").Call(a.window, uintptr(unsafe.Pointer(&cursor)))
	return cursor
}

func (a *app) showObjectPreview() {
	e := &a.objects
	d, image := a.scale, e.image
	w, h := d(96), d(78)
	if image != nil {
		scale := min(float64(d(112))/float64(image.width), float64(d(100))/float64(image.height))
		w, h = max(d(28), int(float64(image.width)*scale)), max(d(28), int(float64(image.height)*scale))
	}
	e.preview.show(a.window, a.objectScreenCursor(), point{int32(w), int32(h)}, int32(d(13)), int32(d(6)), func(dc uintptr) {
		r := rect{right: int32(w), bottom: int32(h)}
		fillColor(dc, r, panelColor)
		if image != nil {
			gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
			pad := d(4)
			fit := min(float64(w-2*pad)/float64(image.width), float64(h-2*pad)/float64(image.height))
			iw, ih := max(1, int(float64(image.width)*fit)), max(1, int(float64(image.height)*fit))
			drawBitmap(dc, image, (w-iw)/2, (h-ih)/2, iw, ih)
		} else {
			drawText(dc, getText(e.info), r, a.normalFont, textColor, 1)
		}
	})
}

func (a *app) paintObjectSelection(dc uintptr, source int, area rect) {
	e, v := &a.objects, a.view
	if a.mode != idObjectsTab || a.full {
		return
	}
	if e.selected.Token == 0 || e.source != source {
		return
	}
	if e.crop.active {
		a.paintImageCrop(dc, area)
		return
	}
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	defer gdiDelete.Call(pen)
	defer gdiSelect.Call(dc, old)
	for _, item := range e.items {
		b := item.Bounds
		if e.moving {
			b = searchRect{b[0] + e.delta[0], b[1] + e.delta[1], b[2] + e.delta[0], b[3] + e.delta[1]}
		}
		bounds, show := searchDisplayRect(b, e.page.Boxes[v.box()], e.rotation(v))
		if show {
			drawTextOutline(dc, area, rectangleSelection(textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}), true)
		}
	}
}

func (a *app) runObjectEdit(source int, edit objectEditRequest) {
	v := a.view
	req := v.request("object_edit")
	req.View.Page, req.View.Rotation = source, v.rotations[source-1]
	req.ObjectEdit = &edit
	a.runDocumentDraft(req, "Upravuji objekt…", 0)
}

func (a *app) chooseObjectGrouping() (string, bool) {
	e := &a.objects
	if len(e.items) < 2 {
		return "", true
	}
	tab, generation, selection := a.activeTab, a.view.generation, e.serial
	var cursor point
	user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&cursor)))
	index := a.trackPopupAt(cursor, 256, []popupItem{
		{label: "Ponechat objekty samostatně"},
		{label: "Seskupit do jednoho objektu"},
	})
	if index < 0 || a.busy || a.mode != idObjectsTab || a.activeTab != tab ||
		a.view.generation != generation || e.serial != selection {
		return "", false
	}
	return [...]string{"separate", "group"}[index], true
}

func (a *app) copySelectedObject() {
	e := &a.objects
	if len(e.capsule) == 0 {
		e.copyWaiting = false
		a.reportStatus("Objekt nelze přenést schránkou.")
		return
	}
	e.clipboard, e.clipboardGrouping, e.owner, e.copyWaiting = e.capsule, e.copyGrouping, a.activeTab, false
	switch e.clipboardGrouping {
	case "group":
		a.reportStatus("Výběr byl zkopírován jako skupina.")
	case "separate":
		a.reportStatus("Objekty byly zkopírovány samostatně.")
	default:
		a.reportStatus("Objekt byl zkopírován.")
	}
	a.updateObjectActions()
}

func (a *app) objectCommand(id, notification int) bool {
	if a.mode == idObjectsTab && id == idObjectCropRatio && notification == 1 {
		setFocus.Call(a.window)
		a.updateObjectActions()
		a.view.invalidateCanvas()
		return true
	}
	if id < idObjectCopy || id > idObjectCancelCrop || notification != 0 || a.mode != idObjectsTab {
		return false
	}
	e, v := &a.objects, a.view
	if a.busy || !v.ready() {
		return true
	}
	setFocus.Call(a.window)
	if id < idObjectCrop {
		a.cancelImageCrop()
	}
	switch id {
	case idObjectCrop:
		a.beginImageCrop()
	case idObjectApplyCrop:
		a.applyImageCrop()
	case idObjectCancelCrop:
		a.cancelImageCrop()
	case idObjectPoint, idObjectRectangle, idObjectLasso:
		a.cancelCanvasSelections()
		e.mode = id - idObjectPoint
		if e.mode == 0 {
			a.canvasTextCursor = false
		}
		a.cancelCanvasRegionProbes()
		if e.mode == 0 || v.mixed.pick.source == 0 {
			v.clearMixedSelection()
			a.clearObjectSelection()
		}
		v.invalidateCanvas()
		a.updateCommandBar()
	case idObjectCopy:
		if e.selected.Token != 0 {
			grouping, accepted := a.chooseObjectGrouping()
			if !accepted {
				return true
			}
			e.copyGrouping = grouping
			if !e.snapshotComplete {
				e.copyWaiting = true
				a.reportStatus("Připravuji objekt do schránky…")
				v.schedule()
			} else {
				a.copySelectedObject()
			}
		}
	case idObjectPaste:
		if len(e.clipboard) != 0 && e.owner == a.activeTab {
			source := v.order[v.page]
			var at *textPoint
			if e.pick.source != 0 && e.pick.generation == v.generation {
				if _, visible := v.textSheetAt(0, 0, e.pick.source); visible != 0 {
					source = e.pick.source
					position := e.pick.point
					at = &position
				}
			}
			a.runObjectEdit(source, objectEditRequest{
				Operation: "paste", Grouping: e.clipboardGrouping, At: at, Clipboard: e.clipboard,
			})
		}
	case idObjectUndo:
		v.undo()
	case idObjectDuplicate, idObjectDelete:
		if e.selected.Token != 0 {
			edit := objectEditRequest{Operation: "delete", Token: e.selected.Token, Tokens: e.tokens()}
			if id == idObjectDuplicate {
				grouping, accepted := a.chooseObjectGrouping()
				if !accepted {
					return true
				}
				edit.Operation, edit.Grouping = "duplicate", grouping
				box, angle := e.page.Boxes[v.box()], e.rotation(v)
				offset := 12 / max(.01, e.page.Unit)
				w, h := box[2]-box[0], box[3]-box[1]
				if angle%180 != 0 {
					w, h = h, w
				}
				p, origin := textPointToPDF(textPoint{offset / w, offset / h}, box, angle), textPointToPDF(textPoint{}, box, angle)
				edit.Delta = textPoint{p[0] - origin[0], p[1] - origin[1]}
			}
			a.runObjectEdit(e.source, edit)
		}
	}
	return true
}

func (a *app) objectKey(msg *winMessage) bool {
	if a.mode != idObjectsTab || msg.window != a.window || a.busy || a.full {
		return false
	}
	switch msg.wParam {
	case 0x1b:
		if a.cancelImageCrop() {
			return true
		}
		a.cancelObjectGesture()
		a.clearObjectSelection()
		a.view.invalidateCanvas()
		return true
	case 0x2e:
		return a.objectCommand(idObjectDelete, 0)
	case 0x0d:
		if a.objects.crop.active && !keyDown(0x11) && !keyDown(0x12) {
			a.applyImageCrop()
			return true
		}
	}
	ctrl, _, _ := user32.NewProc("GetKeyState").Call(0x11)
	if ctrl&0x8000 != 0 {
		switch msg.wParam {
		case 'C':
			return a.objectCommand(idObjectCopy, 0)
		case 'V':
			return a.objectCommand(idObjectPaste, 0)
		case 'D':
			return a.objectCommand(idObjectDuplicate, 0)
		case 'Z':
			return false
		}
	}
	return false
}
