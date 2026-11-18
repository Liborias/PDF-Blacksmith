package main

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	viewerWidth      = 1100
	viewerHeight     = 730
	wmViewerFinished = 0x8002

	idViewerPrevious = 1100 + iota
	idViewerNext
	idViewerGo
	idViewerZoom
	idViewerBox
	idViewerProof
	idViewerPlateList
	idViewerAll
	idViewerNone
	idViewerScan
	idViewerRotateLeft
	idViewerRotateRight
	idViewerMoveBack
	idViewerMoveAhead
	idViewerUndo
	idViewerSave
)

type printBitmap struct {
	width, height int
	pixels        []byte
}

type bitmapHeader struct {
	size                        uint32
	width, height               int32
	planes, bits                uint16
	compression, imageSize      uint32
	xPixels, yPixels            int32
	colorsUsed, colorsImportant uint32
}

func decodePrintBitmap(data []byte) (*printBitmap, error) {
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := decoded.Bounds()
	result := &printBitmap{width: bounds.Dx(), height: bounds.Dy(),
		pixels: make([]byte, bounds.Dx()*bounds.Dy()*4)}
	for y := range result.height {
		for x := range result.width {
			r, g, b, _ := decoded.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			offset := (y*result.width + x) * 4
			result.pixels[offset], result.pixels[offset+1], result.pixels[offset+2] = byte(b>>8), byte(g>>8), byte(r>>8)
		}
	}
	return result, nil
}

type pageEdit struct {
	from, to int
	angle    int
}

type previewJob struct {
	request  request
	response response
	picture  *printBitmap
}

type viewerWindow struct {
	*toolWindow
	input, password                                                   string
	original                                                          os.FileInfo
	busy                                                              bool
	jobs                                                              chan previewJob
	order, rotations, savedOrder, savedRotations                      []int
	page                                                              int
	history                                                           []pageEdit
	profiles                                                          []string
	hidden                                                            map[string]bool
	plates                                                            []plateInfo
	picture                                                           *printBitmap
	shiftX, shiftY, dragX, dragY                                      int
	dragging                                                          bool
	previous, next, location, pageInput, jump, zoom, box, proof, list uintptr
	outputInfo, sample, scanButton, scanReport, all, none, status     uintptr
	left, right, back, ahead, undo, save                              uintptr
}

func (a *app) previewPDF(path, password string) {
	if path == "" {
		notice(a.window, "Vyberte PDF, které chcete zobrazit.", 0x30)
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		notice(a.window, "Cesta k PDF se nepodařila otevřít.", 0x30)
		return
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		notice(a.window, "Vyberte existující soubor PDF.", 0x30)
		return
	}
	tool := createTool(a, "PDF Blaksmith · Náhled a editor", viewerWidth, viewerHeight)
	if tool == nil {
		return
	}
	v := &viewerWindow{toolWindow: tool, input: path, password: password, original: info,
		jobs: make(chan previewJob, 1), hidden: make(map[string]bool)}
	v.setup()
	tool.paint, tool.command, tool.message, tool.closeAllowed = v.paintPage, v.onCommand, v.onMessage, v.canClose
	tool.drawItem = v.drawPlateRow
	tool.show()
	v.refresh()
}

func (v *viewerWindow) setup() {
	v.host.add("STATIC", filepath.Base(v.input), 0, 24, 15, 818, 33, 0x4000, 0, v.host.titleFont)
	v.save = v.button("Uložit upravenou kopii…", idViewerSave, 862, 14, 212)
	v.host.mode = idViewerSave
	v.previous = v.button("‹", idViewerPrevious, 24, 60, 48)
	v.next = v.button("›", idViewerNext, 80, 60, 48)
	v.pageInput = v.host.edit(141, 63, 63, false)
	v.location = v.label("z …", 215, 67, 84, false)
	v.jump = v.button("Přejít", idViewerGo, 301, 60, 88)
	v.zoom = toolCombo(v.host, idViewerZoom, 405, 63, 142,
		[]string{"Přizpůsobit", "125 % náhledu", "150 % náhledu", "200 % náhledu"})
	v.box = toolCombo(v.host, idViewerBox, 558, 63, 140,
		[]string{"Celý arch", "Ořez zobrazení", "Čistý formát", "Spadávka"})

	v.label("Náhledový profil tisku", 728, 57, 346, true)
	names := []string{"Podle PDF / výchozí CMYK"}
	v.profiles = []string{""}
	for _, choice := range v.parent.profileChoices {
		if choice.Key != "" && choice.Key != "k" && choice.Key != "custom" {
			names = append(names, choice.Name)
			v.profiles = append(v.profiles, choice.Key)
		}
	}
	v.proof = toolCombo(v.host, idViewerProof, 728, 82, 346, names)
	v.outputInfo = v.readout(728, 122, 346, 78)
	v.label("Tiskové desky na stránce", 728, 215, 346, true)
	v.label("Kliknutím desku zapnete nebo vypnete.", 728, 241, 346, false)
	const multiSelectList = 0x0001 | 0x0008 | 0x0010 | 0x0040 | 0x0100
	v.list = v.host.add("LISTBOX", "", idViewerPlateList, 728, 271, 346, 181,
		wsVscroll|wsTabstop|multiSelectList, wsExEdge, v.host.normalFont)
	v.host.controls = append(v.host.controls, v.list)
	sendMessage.Call(v.list, 0x01a0, 0, uintptr(v.host.scale(25)))
	v.all = v.button("Všechny desky", idViewerAll, 728, 462, 161)
	v.none = v.button("Vypnout všechny", idViewerNone, 900, 462, 174)
	v.sample = v.label("", 728, 505, 346, false)
	v.scanButton = v.button("Zkontrolovat celý dokument", idViewerScan, 728, 533, 346)
	v.scanReport = v.readout(728, 579, 346, 126)

	v.left = v.button("Otočit vlevo", idViewerRotateLeft, 24, 620, 125)
	v.right = v.button("Otočit vpravo", idViewerRotateRight, 159, 620, 130)
	v.back = v.button("Posunout před", idViewerMoveBack, 299, 620, 135)
	v.ahead = v.button("Posunout za", idViewerMoveAhead, 444, 620, 131)
	v.undo = v.button("Zpět", idViewerUndo, 585, 620, 113)
	v.status = v.readout(24, 665, 674, 40)
}

func (v *viewerWindow) changed() bool {
	return !slices.Equal(v.order, v.savedOrder) || !slices.Equal(v.rotations, v.savedRotations)
}

func (v *viewerWindow) canClose() bool {
	if v.busy && v.changed() {
		notice(v.host.window, "Počkejte, až doběhne zpracování upraveného PDF.", 0x40)
		return false
	}
	return !v.changed() || notice(v.host.window, "Zavřít náhled a zahodit neuložené úpravy stránek?", 0x134) == 6
}

func (v *viewerWindow) currentRequest(action string) request {
	page, rotation := 1, 0
	if len(v.order) > 0 {
		page = v.order[v.page]
		rotation = v.rotations[page-1]
	}
	var hidden []string
	for name, disabled := range v.hidden {
		if disabled {
			hidden = append(hidden, name)
		}
	}
	profile := ""
	if index := comboIndex(v.proof); index >= 0 && index < len(v.profiles) {
		profile = v.profiles[index]
	}
	options := v.parent.export
	boxes := [...]string{"media", "crop", "trim", "bleed"}
	return request{Action: action, Input: v.input, PDFPassword: v.password, Export: &options,
		View: &viewRequest{Page: page, Rotation: rotation, Box: boxes[max(0, comboIndex(v.box))], ProofProfile: profile, HidePlates: hidden}}
}

func (v *viewerWindow) refresh() {
	v.picture, v.plates = nil, nil
	sendMessage.Call(v.list, 0x0184, 0, 0)
	setText(v.sample, "")
	invalidate(v.host.window)
	v.run(v.currentRequest("preview"), "Vykresluji tiskový náhled a desky…")
}

func (v *viewerWindow) run(req request, message string) {
	info, err := os.Stat(v.input)
	if err != nil || !os.SameFile(v.original, info) || info.Size() != v.original.Size() || !info.ModTime().Equal(v.original.ModTime()) {
		setText(v.status, "Vstupní PDF se na disku změnilo. Otevřete nový náhled.")
		v.updateControls()
		return
	}
	v.busy = true
	for _, control := range v.host.controls {
		enableWindow.Call(control, 0)
	}
	setText(v.status, message)
	go func() {
		result := runOperation(req)
		job := previewJob{request: req, response: result}
		if result.Status == "ok" && req.Action == "preview" && result.View != nil {
			var err error
			job.picture, err = decodePrintBitmap(result.View.Image)
			if err != nil {
				job.response = operationError("Obraz náhledu se nepodařilo načíst.")
			}
			result.View.Image = nil
		}
		v.jobs <- job
		postMessage.Call(v.host.window, wmViewerFinished, 0, 0)
	}()
}

func (v *viewerWindow) finish(job previewJob) {
	v.busy = false
	r := job.response
	if r.Status == "signed_pdf" {
		if notice(v.host.window, r.Message+"\n\nUložit kopii bez stávajících podpisů?", 0x134) == 6 {
			job.request.ClearSignatures = true
			v.run(job.request, "Ukládám upravenou kopii…")
			return
		}
		r.Message = "Uložení bylo zrušeno."
	}
	for _, control := range v.host.controls {
		enableWindow.Call(control, 1)
	}
	if r.Status != "ok" {
		setText(v.status, multiline(r.Message))
		v.updateControls()
		return
	}
	switch job.request.Action {
	case "preview":
		if r.View == nil || job.picture == nil {
			setText(v.status, "Náhled nebyl vrácen.")
			break
		}
		v.picture = job.picture
		if v.order == nil {
			v.order = make([]int, r.View.Pages)
			v.rotations = make([]int, r.View.Pages)
			for index := range v.order {
				v.order[index] = index + 1
			}
			v.savedOrder, v.savedRotations = slices.Clone(v.order), slices.Clone(v.rotations)
		}
		v.fillPlates(r.View)
		setText(v.status, "")
		invalidate(v.host.window)
	case "scan":
		if r.View != nil {
			v.showScan(r.View)
		}
		setText(v.status, "")
	case "edit_pages":
		v.savedOrder, v.savedRotations = slices.Clone(v.order), slices.Clone(v.rotations)
		v.parent.lastOutput = job.request.Output
		enableWindow.Call(v.parent.outputButton, 1)
		v.parent.updateReport(r.Message + "\n\n" + job.request.Output)
		setText(v.status, multiline(r.Message+"\n"+job.request.Output))
	}
	v.updateControls()
}

func multiline(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", "\r\n")
}

func plateCaption(name string) string {
	switch name {
	case "Cyan":
		return "Cyan (C)"
	case "Magenta":
		return "Magenta (M)"
	case "Yellow":
		return "Yellow (Y)"
	case "Black":
		return "Black (K)"
	default:
		return name + " · přímá"
	}
}

func coverageText(plate plateInfo) string {
	if plate.Active && plate.Coverage < 0.01 {
		return "<0,01 %"
	}
	return strings.Replace(fmt.Sprintf("%.2f %%", plate.Coverage), ".", ",", 1)
}

func plateRow(plate plateInfo) string {
	maximum := strings.Replace(fmt.Sprintf("%.1f", plate.Maximum), ".", ",", 1)
	return fmt.Sprintf("%s    %s  /  max %s %%", plateCaption(plate.Name), coverageText(plate), maximum)
}

func (v *viewerWindow) drawPlateRow(item *buttonDraw) bool {
	if item.controlType != 2 || item.handle != v.list { // ODT_LISTBOX
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, inputColor)
	if int(item.itemID) >= len(v.plates) {
		return true
	}
	plate := v.plates[item.itemID]
	left, size := int32(v.host.scale(7)), int32(v.host.scale(16))
	top := bounds.top + (bounds.bottom-bounds.top-size)/2
	check := rect{bounds.left + left, top, bounds.left + left + size, top + size}
	foreground := mutedColor
	if plate.Active {
		foreground = textColor
	}
	if item.state&1 != 0 {
		rounded(item.dc, check, v.host.scale(5), paletteColor{113, 42, 70}, accentColor)
		drawCheck(item.dc, check)
	} else {
		rounded(item.dc, check, v.host.scale(5), inputColor, divider)
	}
	bounds.left += int32(v.host.scale(30))
	drawText(item.dc, plateRow(plate), bounds, v.host.normalFont, foreground, 0x8000)
	if item.state&16 != 0 {
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&bounds)))
	}
	return true
}

func (v *viewerWindow) fillPlates(result *viewResponse) {
	v.plates = result.Plates
	sendMessage.Call(v.list, 0x0184, 0, 0)
	for index, plate := range v.plates {
		sendMessage.Call(v.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(plateRow(plate)))))
		if !v.hidden[plate.Name] {
			sendMessage.Call(v.list, 0x0185, 1, uintptr(index))
		}
	}
	intent := result.OutputIntent
	if intent == "" {
		intent = "Neuveden"
	}
	setText(v.outputInfo, multiline("Výstupní záměr PDF\n"+intent+"\nSimulace  "+result.Simulation))
	setText(v.sample, fmt.Sprintf("Plocha / maximální tón · %d dpi", result.DPI))
}

func pageRanges(pages []int) string {
	slices.Sort(pages)
	parts := make([]string, 0, len(pages))
	for index := 0; index < len(pages); {
		start := pages[index]
		end := start
		index++
		for index < len(pages) && pages[index] == end+1 {
			end = pages[index]
			index++
		}
		if start == end {
			parts = append(parts, strconv.Itoa(start))
		} else {
			parts = append(parts, fmt.Sprintf("%d–%d", start, end))
		}
	}
	return strings.Join(parts, ", ")
}

func (v *viewerWindow) showScan(result *viewResponse) {
	index := make(map[int]int, len(v.order))
	for position, page := range v.order {
		index[page] = position + 1
	}
	lines := []string{fmt.Sprintf("%d stran · nejméně %d dpi", result.Pages, result.DPI)}
	if len(result.UsedPlates) == 0 {
		lines = append(lines, "Ve vykreslené ploše nebyl nalezen tiskový bod.")
	}
	for _, plate := range result.UsedPlates {
		pages := make([]int, 0, len(plate.Pages))
		for _, original := range plate.Pages {
			pages = append(pages, index[original])
		}
		lines = append(lines, plateCaption(plate.Name)+"   str. "+pageRanges(pages))
	}
	setText(v.scanReport, multiline(strings.Join(lines, "\n")))
}

func (v *viewerWindow) updateControls() {
	loaded := len(v.order) != 0
	setText(v.pageInput, strconv.Itoa(v.page+1))
	setText(v.location, fmt.Sprintf("z %d", len(v.order)))
	for _, control := range []uintptr{v.pageInput, v.jump, v.scanButton, v.left, v.right} {
		setEnabled(control, loaded)
	}
	setEnabled(v.previous, loaded && v.page > 0)
	setEnabled(v.next, loaded && v.page+1 < len(v.order))
	setEnabled(v.back, loaded && v.page > 0)
	setEnabled(v.ahead, loaded && v.page+1 < len(v.order))
	setEnabled(v.undo, len(v.history) > 0)
	setEnabled(v.save, loaded && v.changed())
	setEnabled(v.list, len(v.plates) > 0)
	setEnabled(v.all, len(v.plates) > 0)
	setEnabled(v.none, len(v.plates) > 0)
}

func setEnabled(handle uintptr, enabled bool) {
	var value uintptr
	if enabled {
		value = 1
	}
	enableWindow.Call(handle, value)
}

func (v *viewerWindow) goToPage(index int) {
	if index < 0 || index >= len(v.order) {
		setText(v.status, "Zadejte číslo existující stránky.")
		v.updateControls()
		return
	}
	v.page = index
	v.shiftX, v.shiftY = 0, 0
	v.refresh()
}

func (v *viewerWindow) rotate(angle int) {
	index := v.order[v.page] - 1
	v.history = append(v.history, pageEdit{from: v.page, to: v.page, angle: v.rotations[index]})
	v.rotations[index] = (v.rotations[index] + 360 + angle) % 360
	v.shiftX, v.shiftY = 0, 0
	v.refresh()
}

func (v *viewerWindow) move(destination int) {
	if destination < 0 || destination >= len(v.order) {
		return
	}
	v.history = append(v.history, pageEdit{from: v.page, to: destination})
	v.order[v.page], v.order[destination] = v.order[destination], v.order[v.page]
	v.page = destination
	setText(v.scanReport, "")
	v.updateControls()
}

func (v *viewerWindow) undoEdit() {
	last := len(v.history) - 1
	if last < 0 {
		return
	}
	edit := v.history[last]
	v.history = v.history[:last]
	if edit.from == edit.to {
		v.rotations[v.order[edit.from]-1] = edit.angle
	} else {
		v.order[edit.from], v.order[edit.to] = v.order[edit.to], v.order[edit.from]
		setText(v.scanReport, "")
	}
	v.goToPage(edit.from)
}

func (v *viewerWindow) saveCopy() {
	if !v.changed() {
		return
	}
	req := v.currentRequest("edit_pages")
	req.View.Order, req.View.Rotations = slices.Clone(v.order), slices.Clone(v.rotations)
	if v.host.chooseOutput(&req, "-upraveno") {
		v.run(req, "Ukládám upravenou kopii…")
	}
}

func (v *viewerWindow) onCommand(id, notification int) {
	if v.busy {
		return
	}
	if notification == 1 {
		switch id {
		case idViewerProof, idViewerBox:
			setText(v.scanReport, "")
			v.refresh()
		case idViewerZoom:
			v.shiftX, v.shiftY = 0, 0
			invalidate(v.host.window)
		case idViewerPlateList:
			for index, plate := range v.plates {
				state, _, _ := sendMessage.Call(v.list, 0x0187, uintptr(index), 0)
				v.hidden[plate.Name] = state == 0
			}
			v.refresh()
		}
		return
	}
	if notification != 0 {
		return
	}
	switch id {
	case idViewerPrevious:
		v.goToPage(v.page - 1)
	case idViewerNext:
		v.goToPage(v.page + 1)
	case idViewerGo:
		page, _ := strconv.Atoi(strings.TrimSpace(getText(v.pageInput)))
		v.goToPage(page - 1)
	case idViewerAll, idViewerNone:
		for _, plate := range v.plates {
			v.hidden[plate.Name] = id == idViewerNone
		}
		v.refresh()
	case idViewerScan:
		req := v.currentRequest("scan")
		req.View.Order = slices.Clone(v.order)
		v.run(req, "Kontroluji tiskové desky celého dokumentu…")
	case idViewerRotateLeft:
		v.rotate(-90)
	case idViewerRotateRight:
		v.rotate(90)
	case idViewerMoveBack:
		v.move(v.page - 1)
	case idViewerMoveAhead:
		v.move(v.page + 1)
	case idViewerUndo:
		v.undoEdit()
	case idViewerSave:
		v.saveCopy()
	}
}

func (v *viewerWindow) onMessage(msg uint32, _ uintptr, param uintptr) bool {
	const wmDown, wmUp, wmMove, wmCaptureChanged = 0x0201, 0x0202, 0x0200, 0x0215
	if msg == wmViewerFinished {
		select {
		case job := <-v.jobs:
			v.finish(job)
		default:
		}
		return true
	}
	x, y := int(int16(param)), int(int16(param>>16))
	canvas := v.canvas()
	switch msg {
	case wmDown:
		if v.picture != nil && comboIndex(v.zoom) > 0 &&
			x >= int(canvas.left) && x < int(canvas.right) && y >= int(canvas.top) && y < int(canvas.bottom) {
			v.dragging, v.dragX, v.dragY = true, x, y
			user32.NewProc("SetCapture").Call(v.host.window)
			return true
		}
	case wmMove:
		if v.dragging {
			v.shiftX += x - v.dragX
			v.shiftY += y - v.dragY
			v.dragX, v.dragY = x, y
			invalidate(v.host.window)
			return true
		}
	case wmUp:
		if v.dragging {
			v.dragging = false
			user32.NewProc("ReleaseCapture").Call()
			return true
		}
	case wmCaptureChanged:
		v.dragging = false
	}
	return false
}

func (v *viewerWindow) canvas() rect {
	return v.host.area(24, 109, 698, 604)
}

func (v *viewerWindow) paintPage(dc uintptr) {
	canvas := v.canvas()
	fillColor(dc, canvas, inputColor)
	if v.picture == nil {
		return
	}
	bitmap := v.picture
	margins := v.host.scale(12)
	spaceWidth, spaceHeight := int(canvas.right-canvas.left)-margins*2, int(canvas.bottom-canvas.top)-margins*2
	fit := min(float64(spaceWidth)/float64(bitmap.width), float64(spaceHeight)/float64(bitmap.height))
	factors := [...]float64{1, 1.25, 1.5, 2}
	scale := fit * factors[max(0, comboIndex(v.zoom))]
	width, height := max(1, int(float64(bitmap.width)*scale)), max(1, int(float64(bitmap.height)*scale))
	maxShiftX, maxShiftY := max(0, (width-spaceWidth)/2), max(0, (height-spaceHeight)/2)
	v.shiftX = min(max(v.shiftX, -maxShiftX), maxShiftX)
	v.shiftY = min(max(v.shiftY, -maxShiftY), maxShiftY)
	x := int(canvas.left+canvas.right)/2 - width/2 + v.shiftX
	y := int(canvas.top+canvas.bottom)/2 - height/2 + v.shiftY
	header := bitmapHeader{size: 40, width: int32(bitmap.width), height: -int32(bitmap.height), planes: 1, bits: 32}
	saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
	gdi32.NewProc("IntersectClipRect").Call(dc, uintptr(canvas.left), uintptr(canvas.top), uintptr(canvas.right), uintptr(canvas.bottom))
	gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
	gdi32.NewProc("StretchDIBits").Call(dc, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0,
		uintptr(bitmap.width), uintptr(bitmap.height), uintptr(unsafe.Pointer(&bitmap.pixels[0])),
		uintptr(unsafe.Pointer(&header)), 0, 0x00cc0020)
	gdi32.NewProc("RestoreDC").Call(dc, saved)
}
