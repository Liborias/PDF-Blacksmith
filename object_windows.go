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
)

type objectEditRequest struct {
	Operation string     `json:"operation,omitempty"`
	Token     int        `json:"token,omitempty"`
	Delta     textPoint  `json:"delta,omitempty"`
	At        *textPoint `json:"at,omitempty"`
	Clipboard []byte     `json:"clipboard,omitempty"`
}

type pdfObject struct {
	Token  int
	Kind   string
	Bounds searchRect
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
	probe                                      bool
	selected                                   pdfObject
	snapshot, snapshotComplete                 bool
	image                                      *printBitmap
	capsule                                    []byte
	copyWaiting                                bool
	clipboard                                  []byte
	owner                                      *documentTab
	start, cursor                              point
	pressed, moving                            bool
	delta                                      textPoint
	preview                                    dragPreview
}

func (a *app) buildObjectPanel() {
	e := &a.objects
	a.label("Výběr objektu", 16, 16, 298, true)
	a.add("STATIC", "Kliknutím vyberte objekt. Tažením jej přesunete na stejné straně.", 0, 16, 47, 298, 45, 0, 0, a.normalFont)
	e.info = a.label("Vyberte objekt na stránce.", 16, 102, 298, false)
	e.copy = a.button("Kopírovat · Ctrl+C", idObjectCopy, 16, 141, 145, 0)
	e.paste = a.button("Vložit · Ctrl+V", idObjectPaste, 169, 141, 145, 0)
	e.duplicate = a.button("Duplikovat · Ctrl+D", idObjectDuplicate, 16, 182, 145, 0)
	e.remove = a.button("Smazat · Delete", idObjectDelete, 169, 182, 145, 0)
	e.undo = a.button("Vrátit změnu · Ctrl+Z", idObjectUndo, 16, 223, 298, 0)
}

func (a *app) updateObjectActions() {
	e := &a.objects
	if e.info == 0 {
		return
	}
	ready := a.mode == idObjectsTab && a.view.ready() && !a.busy
	selected := ready && e.selected.Token != 0
	setEnabled(e.copy, selected)
	setEnabled(e.paste, ready && e.owner == a.activeTab && len(e.clipboard) != 0)
	setEnabled(e.duplicate, selected)
	setEnabled(e.remove, selected)
	setEnabled(e.undo, ready && len(a.view.history) > 0)
}

func (a *app) clearObjectSelection() {
	e := &a.objects
	a.cancelObjectGesture()
	e.page, e.source, e.pick, e.probe = nil, 0, textSelection{}, false
	e.selected, e.image, e.capsule = pdfObject{}, nil, nil
	e.snapshot, e.snapshotComplete, e.copyWaiting = false, false, false
	if e.info != 0 {
		setText(e.info, "Vyberte objekt na stránce.")
		a.updateObjectActions()
	}
}

func (e *objectEditor) endGesture() {
	e.preview.close()
	e.pressed, e.moving, e.delta = false, false, textPoint{}
}

func (a *app) cancelObjectGesture() {
	e := &a.objects
	if e.pressed {
		e.endGesture()
		user32.NewProc("ReleaseCapture").Call()
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
		space, _, _ := user32.NewProc("GetKeyState").Call(0x20)
		pick := v.textAt(x, y)
		if space&0x8000 != 0 || pick.source == 0 {
			return false
		}
		e.selected, e.image, e.capsule = pdfObject{}, nil, nil
		e.snapshot, e.snapshotComplete, e.copyWaiting = false, false, false
		e.pick, e.start, e.cursor, e.pressed = pick, point{x, y}, point{x, y}, true
		setFocus.Call(a.window)
		user32.NewProc("SetCapture").Call(a.window)
		if e.page != nil && e.source == pick.source {
			a.pickObject()
		} else {
			setText(e.info, "Zjišťuji objekty na stránce…")
			v.schedule()
		}
		a.updateObjectActions()
		v.invalidateCanvas()
		return true
	case 0x0200:
		if e.pressed {
			e.cursor = point{x, y}
			a.moveObjectPreview()
			return true
		}
	case 0x0202:
		if e.pressed {
			e.cursor = point{x, y}
			a.moveObjectPreview()
			delta, moved := e.delta, e.moving
			target := v.textAt(x, y)
			a.cancelObjectGesture()
			if moved && target.source == e.pick.source && math.Hypot(delta[0], delta[1]) > .01 {
				a.runObjectEdit(e.pick.source, objectEditRequest{Operation: "move", Token: e.selected.Token, Delta: delta})
			}
			return true
		}
	}
	return false
}

func (a *app) pickObject() {
	v, e := a.view, &a.objects
	rotation := e.rotation(v)
	box := e.page.Boxes[e.pick.box]
	p := textPointToPDF(e.pick.point, box, rotation)
	tolerance := max(e.pick.tolerance[0], e.pick.tolerance[1]) * min(box[2]-box[0], box[3]-box[1])
	for i := len(e.page.Objects) - 1; i >= 0; i-- {
		obj := e.page.Objects[i]
		b := obj.Bounds
		if p[0] >= b[0]-tolerance && p[0] <= b[2]+tolerance && p[1] >= b[1]-tolerance && p[1] <= b[3]+tolerance {
			e.selected = obj
			label := map[string]string{"path": "Vektorový tvar", "image": "Obrázek", "form": "Skupina PDF"}[obj.Kind]
			setText(e.info, label)
			break
		}
	}
	if e.selected.Token == 0 {
		setText(e.info, "Na tomto místě není upravitelný objekt.")
	}
	a.updateObjectActions()
	v.invalidateCanvas()
	a.moveObjectPreview()
	v.schedule()
}

func (e *objectEditor) rotation(v *documentView) int {
	return (e.page.Rotation + v.rotations[e.source-1] + 360) % 360
}

func (v *documentView) scheduleObjects() bool {
	a, e := v.app, &v.app.objects
	if v.secondary || a.mode != idObjectsTab || a.busy || e.pick.source == 0 || e.pick.generation != v.generation {
		return false
	}
	if (e.page == nil || e.source != e.pick.source) && !e.probe {
		e.probe = true
		req := v.request("object_probe")
		req.View.Page = e.pick.source
		v.launch(req, 0, 0, false, nil)
		return true
	}
	if e.selected.Token != 0 && !e.snapshot {
		e.snapshot = true
		req := v.request("object_snapshot")
		req.View.Page, req.View.Rotation = e.source, v.rotations[e.source-1]
		req.ObjectEdit = &objectEditRequest{Token: e.selected.Token}
		v.launch(req, e.source, 0, false, nil)
		return true
	}
	return false
}

func (a *app) finishObjects(job readerJob) {
	e, v := &a.objects, a.view
	if job.generation != v.generation || a.mode != idObjectsTab || e.pick.source == 0 {
		return
	}
	if job.action == "object_probe" {
		e.probe = false
		if job.objectPage != e.pick.source {
			return
		}
		if job.result.Status != "ok" || job.result.View == nil || job.result.View.ObjectPage == nil {
			setText(e.info, "Objekty na této straně nelze vybrat.")
			e.pick = textSelection{}
			return
		}
		e.page, e.source = job.result.View.ObjectPage, job.objectPage
		a.pickObject()
		return
	}
	if job.objectPage != e.pick.source || job.objectToken != e.selected.Token {
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
	if !e.pressed || e.selected.Token == 0 || max(absTextDrag(e.start.x-e.cursor.x), absTextDrag(e.start.y-e.cursor.y)) <= int32(a.scale(4)) && !e.moving {
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
	if a.mode != idObjectsTab || e.selected.Token == 0 || e.source != source || a.full {
		return
	}
	b := e.selected.Bounds
	if e.moving {
		b = searchRect{b[0] + e.delta[0], b[1] + e.delta[1], b[2] + e.delta[0], b[3] + e.delta[1]}
	}
	bounds, show := searchDisplayRect(b, e.page.Boxes[v.box()], e.rotation(v))
	if show {
		pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), selectionColor.win())
		old, _, _ := gdiSelect.Call(dc, pen)
		drawTextOutline(dc, area, rectangleSelection(textPoint{bounds[0], bounds[1]}, textPoint{bounds[2], bounds[3]}), true)
		gdiSelect.Call(dc, old)
		gdiDelete.Call(pen)
	}
}

func (a *app) runObjectEdit(source int, edit objectEditRequest) {
	v := a.view
	req := v.request("object_edit")
	req.View.Page, req.View.Rotation = source, v.rotations[source-1]
	req.ObjectEdit = &edit
	a.runDocumentDraft(req, "Upravuji objekt…", 0)
}

func (a *app) copySelectedObject() {
	e := &a.objects
	if len(e.capsule) == 0 {
		e.copyWaiting = false
		a.reportStatus("Objekt nelze přenést schránkou.")
		return
	}
	e.clipboard, e.owner, e.copyWaiting = e.capsule, a.activeTab, false
	a.reportStatus("Objekt byl zkopírován.")
	a.updateObjectActions()
}

func (a *app) objectCommand(id, notification int) bool {
	if id < idObjectCopy || id > idObjectUndo || notification != 0 || a.mode != idObjectsTab {
		return false
	}
	e, v := &a.objects, a.view
	if a.busy || !v.ready() {
		return true
	}
	setFocus.Call(a.window)
	switch id {
	case idObjectCopy:
		if e.selected.Token != 0 {
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
			a.runObjectEdit(source, objectEditRequest{Operation: "paste", At: at, Clipboard: e.clipboard})
		}
	case idObjectUndo:
		v.undo()
	case idObjectDuplicate, idObjectDelete:
		if e.selected.Token != 0 {
			edit := objectEditRequest{Operation: "delete", Token: e.selected.Token}
			if id == idObjectDuplicate {
				edit.Operation = "duplicate"
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
		a.cancelObjectGesture()
		a.clearObjectSelection()
		a.view.invalidateCanvas()
		return true
	case 0x2e:
		return a.objectCommand(idObjectDelete, 0)
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
			return a.objectCommand(idObjectUndo, 0)
		}
	}
	return false
}
