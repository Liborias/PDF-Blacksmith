package main

import (
	"math"
	"slices"
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
	idObjectCropRectangle
	idObjectCropSquare
	idObjectCropOval
	idObjectCropCircle
	idObjectResetCrop
	idObjectCropMoveMode
)

const (
	idObjectCropZoomOut = 527 + iota
	idObjectCropZoomIn
)

type objectEditRequest struct {
	Operation  string      `json:"operation,omitempty"`
	Grouping   string      `json:"grouping,omitempty"`
	Token      int         `json:"token,omitempty"`
	Tokens     []int       `json:"tokens,omitempty"`
	Delta      textPoint   `json:"delta,omitempty"`
	At         *textPoint  `json:"at,omitempty"`
	Clipboard  []byte      `json:"clipboard,omitempty"`
	Crop       *searchRect `json:"crop,omitempty"`
	CropShape  string      `json:"crop_shape,omitempty"`
	MaskDelta  *textPoint  `json:"mask_delta,omitempty"`
	ImageDelta *textPoint  `json:"image_delta,omitempty"`
	ImageScale float64     `json:"image_scale,omitempty"`
	CapsuleOnly bool        `json:"capsule_only,omitempty"`
}

type pdfObject struct {
	Token     int
	Kind      string
	Bounds    searchRect
	Outline   *objectOutline
	ImageMask *pdfImageMask `json:"image_mask"`
}

type pdfImageMask struct {
	Shape         string
	SourceBounds  searchRect `json:"source_bounds"`
	Outline       []textPoint
	ClipOutline   []textPoint `json:"clip_outline"`
	SourceOutline []textPoint `json:"source_outline"`
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
	capsule                                    []byte
	copyWaiting                                bool
	copyGrouping, clipboardGrouping            string
	clipboard                                  []byte
	owner                                      *documentTab
	start, cursor                              point
	pressed, released                          bool
	crop                                       imageCrop
	cropTool, cropApply, cropCancel, cropHint  uintptr
	cropReset, cropMoveMode                    uintptr
	cropScaleLabel, cropZoomOut, cropZoomIn    uintptr
	cropRevision                               uint64
	cropRatioLabel, cropRatio                  uintptr
	cropShapes                                 [4]uintptr
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
	a.buildCropOptions()
	e.cropApply = a.uiButton(uiConfirmImageCrop, idObjectApplyCrop, 16, 454, 145, 0)
	e.cropCancel = a.uiButton(uiCancelImageCrop, idObjectCancelCrop, 169, 454, 145, 0)
	e.cropReset = a.uiButton(uiResetImageCrop, idObjectResetCrop, 16, 495, 298, 0)
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
		a.setChecked(handle, e.mode == i || i == 2 && e.mode == 3)
	}
	a.updatePolygonButton(e.modes[2], e.mode == 3, uiPolygonLassoKey)
	artwork := selected && !e.hasWidgets()
	setEnabled(e.copy, artwork)
	setEnabled(e.paste, ready && e.owner == a.activeTab && len(e.clipboard) != 0)
	setEnabled(e.duplicate, artwork)
	setEnabled(e.remove, artwork)
	setEnabled(e.undo, ready && a.canDocumentCommand(a.view, docUndo, false))
	canCrop := a.canCropImage()
	setEnabled(e.cropTool, canCrop)
	a.setChecked(e.cropTool, e.crop.active)
	_, canApplyCrop := a.cropEditRequest()
	cropOptions := canCrop && !e.crop.dragging
	if a.hasImageCropTarget() {
		hint := uiChooseImageCrop
		if e.crop.active {
			hint = uiCropImageHint
			if e.selected.ImageMask != nil {
				hint = uiEditImageCropHint
				if a.cropMoveMode() == moveCropImage {
					hint = uiCropMoveImageHint
				}
			}
		}
		a.setUIMessage(e.cropHint, hint)
	}
	for i, handle := range e.cropShapes {
		setEnabled(handle, cropOptions)
		a.setChecked(handle, e.crop.active && e.crop.shape == imageCropShape(i))
	}
	setEnabled(e.cropRatio, cropOptions && !e.crop.shape.locksRatio())
	setEnabled(e.cropApply, canApplyCrop)
	setEnabled(e.cropCancel, ready && e.crop.active)
	setEnabled(e.cropReset, cropOptions && e.selected.ImageMask != nil)
	setEnabled(e.cropMoveMode, cropOptions && e.crop.active && e.selected.ImageMask != nil)
	setEnabled(e.cropZoomOut, cropOptions && a.canZoomCropImage())
	setEnabled(e.cropZoomIn, cropOptions && a.canZoomCropImage())
	setText(e.cropScaleLabel, a.uiLocale.formatMessage(localizedMessage{key: uiCropImageScale, args: []any{math.Round(100*e.crop.motion.shifts.imageScale())}}))
}

func (a *app) clearObjectSelection() {
	e := &a.objects
	showedCrop := e.selected.Kind == "image"
	a.cancelImageCrop()
	a.cancelObjectGesture()
	e.page, e.source, e.pick = nil, 0, textSelection{}
	e.resetSelection()
	e.snapshot, e.snapshotComplete, e.copyWaiting = false, false, false
	if e.info != 0 {
		a.setUIMessage(e.info, uiSelectPageObject)
		a.updateObjectActions()
		if showedCrop {
			a.layoutPanel()
		}
	}
}

func (e *objectEditor) endGesture() {
	e.pressed, e.released, e.dragReady = false, false, false
	e.pressPick = textSelection{}
}

func (a *app) cancelObjectGesture() {
	e := &a.objects
	if e.pressed || e.crop.dragging {
		e.endGesture()
		e.crop.endDrag()
		user32.NewProc("ReleaseCapture").Call()
		a.updateObjectActions()
		a.view.invalidateCanvas()
	}
}

func (v *documentView) objectPointer(msg uint32, x, y int32) bool {
	a, e := v.app, &v.app.objects
	if msg == 0x0215 || msg == 0x001f { // Capture changed or cancelled.
		if e.pressed && !e.released {
			e.endGesture()
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
	case 0x0200, 0x0202:
		if !e.pressed || e.released {
			return false
		}
		e.cursor = point{x, y}
		release := msg == 0x0202
		click := max(absTextDrag(x-e.start.x), absTextDrag(y-e.start.y)) <= int32(a.scale(4))
		if click && release {
			a.cancelObjectGesture()
			v.selectionClick(x, y, false)
			return true
		}
		if !click {
			e.released = release
			v.prepareObjectDrag()
		}
		if e.pressed && release {
			user32.NewProc("ReleaseCapture").Call()
		}
		return true
	}
	return false
}

func (v *documentView) prepareObjectDrag() {
	e := &v.app.objects
	if e.dragReady || e.pressPick.source == 0 || e.pressPick.change != replaceSelection {
		return
	}
	e.dragReady = true
	v.requestObjects(e.pressPick)
}

func (v *documentView) continueProbedDrag() {
	e := &v.app.objects
	if !e.pressed || !e.dragReady {
		return
	}
	pick, origin, cursor, released := e.pressPick, e.start, e.cursor, e.released
	e.endGesture()
	if !v.beginCanvasContentMove(pick, origin) {
		if !released {
			user32.NewProc("ReleaseCapture").Call()
		}
		return
	}
	message := uint32(0x0200)
	if released {
		message = 0x0202
	}
	v.canvasContentMovePointer(message, cursor.x, cursor.y)
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
	if e.crop.motion.requested != e.cropRevision && e.crop.motion.shifts.changed() {
		if edit, ok := a.cropEditRequest(); ok {
			e.crop.motion.requested = e.cropRevision
			req := v.request("object_crop_motion_preview")
			req.View.Page, req.View.Rotation = e.source, v.rotations[e.source-1]
			req.ObjectEdit = &edit
			v.launch(req, e.source, 0, false, nil)
			return true
		}
	}
	if e.crop.active && !e.crop.previewRequested {
		e.crop.previewRequested = true
		req := v.request("object_crop_preview")
		req.View.Page, req.View.Rotation = e.source, v.rotations[e.source-1]
		req.ObjectEdit = &objectEditRequest{Token: e.selected.Token}
		v.launch(req, e.source, 0, false, nil)
		return true
	}
	if e.selected.Token != 0 && e.copyWaiting && !e.snapshot {
		e.snapshot = true
		req := v.request("object_snapshot")
		req.View.Page, req.View.Rotation = e.source, v.rotations[e.source-1]
		req.ObjectEdit = &objectEditRequest{Token: e.selected.Token, Tokens: e.tokens(), CapsuleOnly: true}
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
			if e.pressed && e.dragReady {
				a.cancelObjectGesture()
			}
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
	if job.action == "object_crop_motion_preview" {
		if e.crop.active && e.cropRevision == job.cropRevision && job.result.Status == "ok" {
			e.crop.motion.preview, e.crop.motion.coverage = job.frame.image, job.cropCoverage
			e.crop.motion.rendered = job.cropRevision
			v.invalidateCanvas()
		}
		return
	}
	if job.action == "object_crop_preview" {
		if e.crop.active && job.result.Status == "ok" {
			e.crop.preview = job.frame.image
			v.invalidateCanvas()
		}
		return
	}
	e.snapshot, e.snapshotComplete = true, true
	if job.result.Status == "ok" && job.result.View != nil {
		e.capsule = job.result.View.ObjectClipboard
	}
	if e.copyWaiting {
		a.copySelectedObject()
	}
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
		bounds, show := searchDisplayRect(item.Bounds, e.page.Boxes[v.box()], e.rotation(v))
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
	a.runDamagedDraft(req, "Upravuji objekt…", 0, a.objectEditDamage(source, edit))
}

func (a *app) objectEditDamage(source int, edit objectEditRequest) *pageDamage {
	e, v := &a.objects, a.view
	if e.page == nil || e.source != source || len(e.items) == 0 {
		return nil
	}
	var affected searchRect
	for _, item := range e.items {
		switch edit.Operation {
		case "delete":
			affected = unionPageRects(affected, item.Bounds)
		case "duplicate":
			affected = unionPageRects(affected, moveCropBounds(item.Bounds, edit.Delta))
		case "replace_crop", "reset_crop":
			affected = unionPageRects(affected, item.cropSourceBounds())
		case "translate_crop":
			affected = unionPageRects(affected, item.Bounds)
			if edit.Crop != nil {
				affected = unionPageRects(affected, *edit.Crop)
			} else if edit.MaskDelta != nil {
				affected = unionPageRects(affected, moveCropBounds(item.Bounds, *edit.MaskDelta))
			}
		default:
			return nil
		}
	}
	// Leave room for the antialiasing footprint and minor rounding differences.
	const pad = 3.
	affected = searchRect{affected[0] - pad, affected[1] - pad, affected[2] + pad, affected[3] + pad}
	box := v.box()
	bounds, visible := searchDisplayRect(affected, e.page.Boxes[box], e.rotation(v))
	if !visible {
		return nil
	}
	return &pageDamage{source, box, bounds}
}

func (e *objectEditor) hasWidgets() bool {
	return slices.ContainsFunc(e.items, func(item pdfObject) bool { return item.Kind == "widget" })
}

func (a *app) chooseObjectGrouping() (string, bool) {
	e := &a.objects
	if len(e.items) < 2 || e.hasWidgets() {
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
	if a.mode == idObjectsTab && id == idObjectCropMoveMode && notification == 1 {
		setFocus.Call(a.window)
		a.updateObjectActions()
		a.layoutPanel()
		a.view.invalidateCanvas()
		return true
	}
	if a.mode == idObjectsTab && id == idObjectCropRatio && notification == 1 {
		setFocus.Call(a.window)
		preset := comboIndex(a.objects.cropRatio)
		if !a.objects.crop.active {
			a.beginImageCrop()
			chooseCombo(a.objects.cropRatio, preset)
		}
		a.updateObjectActions()
		a.invalidateCropMotion()
		return true
	}
	if a.mode == idObjectsTab && notification == 0 {
		switch id {
		case idObjectCropZoomOut:
			a.zoomCropImage(-1, nil)
			return true
		case idObjectCropZoomIn:
			a.zoomCropImage(1, nil)
			return true
		}
	}
	if id < idObjectCopy || id > idObjectResetCrop || notification != 0 || a.mode != idObjectsTab {
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
	case idObjectResetCrop:
		if a.canCropImage() && !e.crop.dragging && e.selected.ImageMask != nil {
			a.runObjectEdit(e.source, objectEditRequest{Operation: "reset_crop", Token: e.selected.Token})
		}
	case idObjectCropRectangle, idObjectCropSquare, idObjectCropOval, idObjectCropCircle:
		a.chooseImageCropShape(imageCropShape(id - idObjectCropRectangle))
	case idObjectPoint, idObjectRectangle, idObjectLasso:
		a.cancelCanvasSelections()
		previous := e.mode
		e.mode = id - idObjectPoint
		if e.mode == 2 && previous == 2 {
			e.mode = 3
		}
		if e.mode == 0 {
			a.canvasTextCursor = false
		}
		a.cancelCanvasSelectionProbes()
		v.invalidateCanvas()
		a.updateObjectActions()
		a.updateCommandBar()
	case idObjectCopy:
		if e.selected.Token != 0 && !e.hasWidgets() {
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
		a.doDocumentCommand(v, docUndo, false)
	case idObjectDuplicate, idObjectDelete:
		if e.selected.Token != 0 && !e.hasWidgets() {
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
		previous := a.view.captureCanvasSelection()
		a.cancelObjectGesture()
		a.clearObjectSelection()
		a.view.rememberCanvasSelection(&previous)
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
