package main

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

const (
	idReaderPrevious = 300 + iota
	idReaderNext
	idReaderGo
	idReaderFit
	idReaderZoomOut
	idReaderZoomIn
	idReaderFlow
	idReaderScreenBox
	idReaderProofBox
	idReaderProof
	idReaderPlateList
	idReaderAll
	idReaderNone
	idReaderScan
	idReaderPageList
	idReaderRotateLeft
	idReaderRotateRight
	idReaderMoveBack
	idReaderMoveAhead
	idReaderUndo
	idReaderSave
)

const wmReaderFinished = 0x8002

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

func (v *documentView) drawPlateRow(item *buttonDraw) bool {
	if item.controlType != 2 || item.handle != v.list { // ODT_LISTBOX
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, inputColor)
	if int(item.itemID) >= len(v.plates) {
		return true
	}
	plate := v.plates[item.itemID]
	left, size := int32(v.app.scale(7)), int32(v.app.scale(16))
	top := bounds.top + (bounds.bottom-bounds.top-size)/2
	check := rect{bounds.left + left, top, bounds.left + left + size, top + size}
	foreground := mutedColor
	if plate.Active {
		foreground = textColor
	}
	if item.state&1 != 0 {
		rounded(item.dc, check, v.app.scale(5), paletteColor{113, 42, 70}, accentColor)
		drawCheck(item.dc, check)
	} else {
		rounded(item.dc, check, v.app.scale(5), inputColor, divider)
	}
	bounds.left += int32(v.app.scale(30))
	drawText(item.dc, plateRow(plate), bounds, v.app.normalFont, foreground, 0x8000)
	if item.state&16 != 0 {
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&bounds)))
	}
	return true
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

func (v *documentView) showScan(result *viewResponse) {
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

type pageEdit struct {
	from, to, angle int
}

type readerFrame struct {
	image  *printBitmap
	proof  *viewResponse
	pixels int
}

type readerJob struct {
	generation   int
	page, pixels int
	action       string
	result       response
	frame        readerFrame
}

type documentView struct {
	app                                                                        *app
	path, password                                                             string
	original                                                                   os.FileInfo
	sizes                                                                      []map[string][2]float64
	intent                                                                     string
	order, rotations, savedOrder, savedRotations                               []int
	history                                                                    []pageEdit
	page, reference                                                            int
	continuous                                                                 bool
	fitMode                                                                    int
	zoom                                                                       float64
	viewport                                                                   rect
	spread                                                                     readingLayout
	scrollX, scrollY                                                           float64
	dragging                                                                   bool
	dragX, dragY                                                               int32
	wheelRemainder                                                             float64
	generation                                                                 int
	running, describing, invalidSource                                         bool
	frames                                                                     map[int]readerFrame
	failed                                                                     map[int]bool
	jobs                                                                       chan readerJob
	profiles                                                                   []string
	hidden                                                                     map[string]bool
	plates                                                                     []plateInfo
	proofPage                                                                  int
	previous, next, jump, pageInput, location, zoomOut, zoomIn, zoomLabel, fit uintptr
	flow, screenBox, proofBox, proof, list, pageList, documentInfo             uintptr
	outputInfo, sample, scanReport, all, none                                  uintptr
	left, right, back, ahead, undoButton, save                                 uintptr
}

func newDocumentView(a *app) *documentView {
	return &documentView{app: a, continuous: true, fitMode: 1, proofPage: -1,
		frames: make(map[int]readerFrame), failed: make(map[int]bool), hidden: make(map[string]bool),
		jobs: make(chan readerJob, 1)}
}

func (v *documentView) buildToolbar() {
	a := v.app
	v.previous = a.button("‹", idReaderPrevious, 0, 0, 35, 0)
	v.pageInput = a.edit(0, 0, 48, false)
	v.next = a.button("›", idReaderNext, 0, 0, 35, 0)
	v.location = a.label("", 0, 0, 57, false)
	v.jump = a.button("Přejít", idReaderGo, 0, 0, 65, 0)
	v.zoomOut = a.button("−", idReaderZoomOut, 0, 0, 36, 0)
	v.zoomLabel = a.label("", 0, 0, 56, false)
	v.zoomIn = a.button("+", idReaderZoomIn, 0, 0, 36, 0)
	v.fit = a.combo(idReaderFit, 0, 0, 158, []string{"Celá stránka", "Šířka stránky", "Vlastní zoom"})
}

func (v *documentView) layoutToolbar() {
	d := v.app.scale
	y := v.app.height - d(40)
	place(v.previous, d(12), y, d(35), d(32))
	place(v.pageInput, d(53), y+d(3), d(46), d(28))
	place(v.location, d(108), y+d(7), d(61), d(22))
	place(v.next, d(173), y, d(35), d(32))
	place(v.jump, d(216), y, d(63), d(32))
	place(v.zoomOut, d(304), y, d(36), d(32))
	place(v.zoomLabel, d(350), y+d(7), d(56), d(22))
	place(v.zoomIn, d(408), y, d(36), d(32))
	place(v.fit, d(461), y+d(3), d(156), d(28))
}

func (v *documentView) showToolbar(show uintptr) {
	for _, h := range []uintptr{v.previous, v.next, v.jump, v.pageInput, v.location, v.zoomOut, v.zoomIn, v.zoomLabel, v.fit} {
		showWindow.Call(h, show)
	}
}

func (v *documentView) buildPagePanel() {
	a := v.app
	a.label("Pořadí stránek", 16, 18, 298, true)
	v.pageList = a.add("LISTBOX", "", idReaderPageList, 16, 52, 298, 269, wsVscroll|wsTabstop|0x0101, 0, a.normalFont)
	v.left = a.button("Otočit vlevo", idReaderRotateLeft, 16, 343, 145, 0)
	v.right = a.button("Otočit vpravo", idReaderRotateRight, 169, 343, 145, 0)
	v.back = a.button("Posunout před", idReaderMoveBack, 16, 386, 145, 0)
	v.ahead = a.button("Posunout za", idReaderMoveAhead, 169, 386, 145, 0)
	v.undoButton = a.button("Vrátit změnu", idReaderUndo, 16, 429, 298, 0)
	v.save = a.button("Uložit upravenou kopii…", idReaderSave, 16, 482, 298, 0)
	a.primaryButtons[v.save] = true
}

func (v *documentView) buildProofPanel(catalog []profileChoice) {
	a := v.app
	a.label("Náhledový profil tisku", 16, 18, 298, true)
	names := []string{"Podle PDF / výchozí CMYK"}
	v.profiles = []string{""}
	for _, profile := range catalog {
		names = append(names, profile.Name)
		v.profiles = append(v.profiles, profile.Key)
	}
	v.proof = a.combo(idReaderProof, 16, 43, 298, names)
	a.label("Tisková oblast", 16, 83, 298, false)
	v.proofBox = a.combo(idReaderProofBox, 16, 104, 298, []string{"Celý arch", "Ořez zobrazení", "Čistý formát", "Spadávka"})
	v.outputInfo = a.readout(16, 142, 298, 78)
	a.label("Tiskové desky na stránce", 16, 240, 298, true)
	const listStyle = 0x0001 | 0x0008 | 0x0010 | 0x0040 | 0x0100
	v.list = a.add("LISTBOX", "", idReaderPlateList, 16, 265, 298, 167,
		wsVscroll|wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(v.list, 0x01a0, 0, uintptr(a.scale(24)))
	v.all = a.button("Všechny desky", idReaderAll, 16, 441, 145, 0)
	v.none = a.button("Vypnout všechny", idReaderNone, 169, 441, 145, 0)
	v.sample = a.label("", 16, 486, 298, false)
	a.scanButton = a.button("Zkontrolovat celý dokument", idReaderScan, 16, 519, 298, 0)
	v.scanReport = a.readout(16, 562, 298, 133)
}

func (v *documentView) ready() bool { return len(v.order) > 0 && !v.invalidSource }

func (v *documentView) changed() bool {
	return !slices.Equal(v.order, v.savedOrder) || !slices.Equal(v.rotations, v.savedRotations)
}

func (v *documentView) markSaved() {
	v.savedOrder, v.savedRotations = slices.Clone(v.order), slices.Clone(v.rotations)
}

func (v *documentView) includeEdits(req *request) {
	req.View.Order, req.View.Rotations = slices.Clone(v.order), slices.Clone(v.rotations)
}

func unchanged(path string, original os.FileInfo) bool {
	info, err := os.Stat(path)
	return err == nil && original != nil && os.SameFile(original, info) &&
		info.Size() == original.Size() && info.ModTime().Equal(original.ModTime())
}

func (v *documentView) sourceUnchanged() bool { return unchanged(v.path, v.original) }

func (v *documentView) open(path, password string, original os.FileInfo) {
	v.path, v.password, v.original = path, password, original
	v.generation++
	v.sizes, v.order, v.rotations, v.savedOrder, v.savedRotations, v.history = nil, nil, nil, nil, nil, nil
	v.intent = ""
	v.page, v.reference, v.fitMode = 0, 0, 1
	v.scrollX, v.scrollY, v.wheelRemainder = 0, 0, 0
	v.describing, v.invalidSource = true, false
	v.hidden = make(map[string]bool)
	v.clearFrames()
	chooseCombo(v.fit, 0)
	setText(v.documentInfo, "")
	setText(v.scanReport, "")
	v.rebuild(false)
	v.updatePageList()
	v.app.report("Načítám dokument…")
	v.app.updateActions()
	v.schedule()
}

func (v *documentView) proofMode() bool { return v.app.mode == idProofTab }

func (v *documentView) box() string {
	if v.proofMode() {
		boxes := [...]string{"media", "crop", "trim", "bleed"}
		return boxes[max(0, comboIndex(v.proofBox))]
	}
	boxes := [...]string{"crop", "media", "trim", "bleed"}
	return boxes[max(0, comboIndex(v.screenBox))]
}

func (v *documentView) request(action string) request {
	opts := v.app.export
	req := request{Action: action, Input: v.path, PDFPassword: v.password, Export: &opts,
		View: &viewRequest{Box: v.box()}}
	if index := comboIndex(v.proof); index >= 0 && index < len(v.profiles) {
		req.View.ProofProfile = v.profiles[index]
	}
	if v.proofMode() {
		for color, hidden := range v.hidden {
			if hidden {
				req.View.HidePlates = append(req.View.HidePlates, color)
			}
		}
	}
	return req
}

func (v *documentView) clearFrames() {
	v.frames = make(map[int]readerFrame)
	v.failed = make(map[int]bool)
	v.proofPage = -1
	v.plates = nil
	sendMessage.Call(v.list, 0x0184, 0, 0)
	setText(v.sample, "")
}

func (v *documentView) reframe(geometry bool) {
	v.generation++
	v.clearFrames()
	setText(v.scanReport, "")
	if geometry {
		v.reference = v.page
		v.rebuild(true)
	}
	v.changedViewport()
}

func (v *documentView) sizesInOrder() []pageSize {
	box := v.box()
	result := make([]pageSize, len(v.order))
	for index, page := range v.order {
		d := v.sizes[page-1][box]
		if d[0] <= 0 || d[1] <= 0 {
			d = [2]float64{595, 842}
		}
		if v.rotations[page-1]%180 != 0 {
			d[0], d[1] = d[1], d[0]
		}
		result[index] = pageSize{d[0], d[1]}
	}
	return result
}

func (v *documentView) viewportSize() (float64, float64) {
	return float64(v.viewport.right - v.viewport.left), float64(v.viewport.bottom - v.viewport.top)
}

func (v *documentView) rebuild(keep bool) {
	width, height := v.viewportSize()
	anchor, have := v.spread.anchor(v.scrollX+width/2, v.scrollY+height/2)
	v.buildLayout()
	if keep && have {
		v.restore(anchor, width/2, height/2)
	} else {
		v.centerPage(false)
	}
}

func (v *documentView) buildLayout() {
	sizes := v.sizesInOrder()
	width, height := v.viewportSize()
	if len(sizes) == 0 || width <= 0 || height <= 0 {
		v.spread = readingLayout{}
		return
	}
	gap := float64(v.app.scale(18))
	ref := sizes[min(max(v.reference, 0), len(sizes)-1)]
	if v.fitMode != 0 || v.zoom <= 0 {
		v.zoom = max(0.02, (width-2*gap)/ref.width)
		if v.fitMode != 2 {
			v.zoom = min(v.zoom, max(0.02, (height-2*gap)/ref.height))
		}
	}
	v.spread = layoutPages(sizes, v.page, v.continuous, v.zoom, gap)
}

func (v *documentView) resize(bounds rect) {
	if v.viewport == bounds {
		return
	}
	oldWidth, oldHeight := v.viewportSize()
	anchor, ok := v.spread.anchor(v.scrollX+oldWidth/2, v.scrollY+oldHeight/2)
	v.viewport = bounds
	v.buildLayout()
	if ok {
		width, height := v.viewportSize()
		v.restore(anchor, width/2, height/2)
	} else {
		v.centerPage(false)
	}
	v.changedViewport()
}

func (v *documentView) clamp() {
	w, h := v.viewportSize()
	v.scrollX = clampReadingOffset(v.scrollX, v.spread.width, w)
	v.scrollY = clampReadingOffset(v.scrollY, v.spread.height, h)
}

func (v *documentView) restore(anchor readingAnchor, x, y float64) {
	if newX, newY, ok := v.spread.restore(anchor, x, y); ok {
		v.scrollX, v.scrollY = newX, newY
	}
	v.clamp()
}

func (v *documentView) centerPage(atEnd bool) {
	width, height := v.viewportSize()
	page, ok := v.spread.position(v.page)
	if !ok {
		v.scrollX, v.scrollY = 0, 0
		return
	}
	v.scrollX = (v.spread.width - width) / 2
	if v.spread.height <= height {
		v.scrollY = (v.spread.height - height) / 2
	} else if atEnd && !v.continuous {
		v.scrollY = v.spread.height - height
	} else if v.continuous {
		v.scrollY = page.top - float64(v.app.scale(18))
	} else {
		v.scrollY = 0
	}
	v.clamp()
}

func (v *documentView) gotoPage(index int, atEnd bool) {
	if index < 0 || index >= len(v.order) {
		return
	}
	v.page = index
	if !v.continuous {
		v.reference = index
		v.buildLayout()
	}
	v.centerPage(atEnd)
	v.changedViewport()
}

func (v *documentView) scroll(delta float64) {
	if !v.ready() {
		return
	}
	_, height := v.viewportSize()
	if !v.continuous {
		moveTo := v.scrollY + delta
		if delta > 0 && (v.spread.height <= height || v.scrollY >= v.spread.height-height-0.5) {
			v.gotoPage(v.page+1, false)
			return
		}
		if delta < 0 && (v.spread.height <= height || v.scrollY <= 0.5) {
			v.gotoPage(v.page-1, true)
			return
		}
		v.scrollY = moveTo
	} else {
		v.scrollY += delta
	}
	v.clamp()
	v.updatePageFromCenter()
	v.changedViewport()
}

func (v *documentView) updatePageFromCenter() {
	if !v.continuous {
		return
	}
	width, height := v.viewportSize()
	if anchor, ok := v.spread.anchor(v.scrollX+width/2, v.scrollY+height/2); ok {
		v.page = anchor.index
	}
}

func (v *documentView) zoomAt(delta, x, y float64) {
	if !v.ready() || v.zoom <= 0 {
		return
	}
	anchor, ok := v.spread.anchor(v.scrollX+x, v.scrollY+y)
	if !ok {
		return
	}
	base := float64(v.app.dpi) / 72
	next := min(max(v.zoom*math.Pow(1.2, delta), min(v.zoom, base*0.16)), max(v.zoom, base*8))
	if math.Abs(next-v.zoom) < 1e-7 {
		return
	}
	v.zoom, v.fitMode = next, 0
	chooseCombo(v.fit, 2)
	v.buildLayout()
	v.restore(anchor, x, y)
	v.updatePageFromCenter()
	v.changedViewport()
}

func (v *documentView) wheel(delta float64, control bool, cursor point) {
	if control {
		v.zoomAt(delta, float64(cursor.x-v.viewport.left), float64(cursor.y-v.viewport.top))
		return
	}
	step := float64(v.app.scale(105))
	if !v.continuous {
		// Precision trackpads emit partial wheel deltas; do not flip a page on each event.
		_, height := v.viewportSize()
		atEdge := v.spread.height <= height || (delta < 0 && v.scrollY >= v.spread.height-height-0.5) ||
			(delta > 0 && v.scrollY <= 0.5)
		if atEdge {
			v.wheelRemainder += delta
			if math.Abs(v.wheelRemainder) < 1 {
				return
			}
			delta, v.wheelRemainder = math.Copysign(1, v.wheelRemainder), 0
		} else {
			v.wheelRemainder = 0
		}
	}
	v.scroll(-delta * step)
}

func (v *documentView) key(key uintptr) bool {
	_, height := v.viewportSize()
	switch key {
	case 0x25:
		v.gotoPage(v.page-1, false)
	case 0x27:
		v.gotoPage(v.page+1, false)
	case 0x26:
		v.scroll(-float64(v.app.scale(65)))
	case 0x28:
		v.scroll(float64(v.app.scale(65)))
	case 0x21:
		v.scroll(-height * 0.88)
	case 0x22, 0x20:
		v.scroll(height * 0.88)
	case 0x24:
		v.gotoPage(0, false)
	case 0x23:
		v.gotoPage(len(v.order)-1, false)
	default:
		return false
	}
	return true
}

func contains(bounds rect, x, y int32) bool {
	return x >= bounds.left && x < bounds.right && y >= bounds.top && y < bounds.bottom
}

func (v *documentView) pointer(msg uint32, param uintptr) bool {
	x, y := int32(int16(param)), int32(int16(param>>16))
	switch msg {
	case 0x0201:
		if contains(v.viewport, x, y) {
			setFocus.Call(v.app.window)
			if v.path == "" {
				v.app.openDialog()
				return true
			}
			if v.ready() {
				v.dragging, v.dragX, v.dragY = true, x, y
				user32.NewProc("SetCapture").Call(v.app.window)
			}
			return true
		}
	case 0x0200:
		if v.dragging {
			v.scrollX += float64(v.dragX - x)
			v.scrollY += float64(v.dragY - y)
			v.dragX, v.dragY = x, y
			v.clamp()
			v.updatePageFromCenter()
			v.changedViewport()
			return true
		}
	case 0x0202:
		if v.dragging {
			v.dragging = false
			user32.NewProc("ReleaseCapture").Call()
			return true
		}
	case 0x0215:
		v.dragging = false
	}
	return false
}

func (v *documentView) changedViewport() {
	v.updateActions()
	v.syncProofPanel()
	invalidate(v.app.window)
	v.schedule()
}

func (v *documentView) wantedPixels(page sheet) int {
	if v.proofMode() {
		return 1850
	}
	size := max(page.width, page.height)
	switch {
	case size > 2050:
		return 3000
	case size > 1330:
		return 2200
	case size > 660:
		return 1450
	default:
		return 780
	}
}

func (v *documentView) schedule() {
	if v.running || v.path == "" || v.invalidSource {
		return
	}
	if v.describing {
		v.launch(v.request("describe"), 0, 0)
		return
	}
	if !v.ready() {
		return
	}
	_, height := v.viewportSize()
	visible := v.spread.visible(v.scrollY, height)
	candidates := slices.Clone(visible)
	if current, ok := v.spread.position(v.page); ok {
		candidates = append([]sheet{current}, candidates...)
	}
	if v.continuous && !v.proofMode() && len(visible) > 0 {
		if next, ok := v.spread.position(visible[len(visible)-1].index + 1); ok {
			candidates = append(candidates, next)
		}
	}
	for _, item := range candidates {
		number := v.order[item.index]
		frame := v.frames[number]
		pixels := v.wantedPixels(item)
		if v.failed[number] || (frame.image != nil && frame.pixels >= pixels) {
			continue
		}
		action := "screen"
		if v.proofMode() {
			action = "preview"
		}
		req := v.request(action)
		req.View.Page, req.View.Rotation, req.View.Pixels = number, v.rotations[number-1], pixels
		v.launch(req, number, pixels)
		return
	}
}

func (v *documentView) launch(req request, number, pixels int) {
	v.running = true
	generation, original := v.generation, v.original
	go func() {
		job := readerJob{generation: generation, page: number, pixels: pixels, action: req.Action}
		if !unchanged(req.Input, original) {
			job.result = response{Status: "changed", Message: "Vstupní PDF se změnilo. Otevřete je znovu."}
		} else {
			job.result = runOperation(req)
		}
		if job.result.Status == "ok" && number > 0 && job.result.View != nil {
			var err error
			job.frame.image, err = decodePrintBitmap(job.result.View.Image)
			if err != nil {
				job.result = operationError("Obraz náhledu se nepodařilo načíst.")
			} else {
				job.frame.proof, job.frame.pixels = job.result.View, pixels
			}
			if job.frame.proof != nil {
				job.frame.proof.Image = nil
			}
		}
		v.jobs <- job
		postMessage.Call(v.app.window, wmReaderFinished, 0, 0)
	}()
}

func (v *documentView) finished(job readerJob) {
	v.running = false
	if job.generation != v.generation {
		v.schedule()
		return
	}
	if job.result.Status != "ok" {
		v.describing = false
		v.invalidSource = job.result.Status == "changed"
		if job.page != 0 {
			v.failed[job.page] = true
		}
		if job.page == 0 {
			setText(v.documentInfo, multiline(job.result.Message))
			if !v.app.full {
				v.app.activate(idBrowseTab)
			}
		}
		v.app.report(job.result.Message)
	} else if job.action == "describe" {
		data := job.result.View
		v.describing = false
		if data == nil || len(data.Sizes) == 0 {
			v.app.report("Rozměry stránek nelze načíst.")
			return
		}
		v.sizes, v.intent = data.Sizes, data.OutputIntent
		v.order = make([]int, len(v.sizes))
		v.rotations = make([]int, len(v.sizes))
		for i := range v.order {
			v.order[i] = i + 1
		}
		v.markSaved()
		intent := v.intent
		if intent == "" {
			intent = "Neuveden"
		}
		setText(v.documentInfo, multiline(fmt.Sprintf("%d stran\nVýstupní záměr PDF\n%s", len(v.order), intent)))
		v.app.report("")
		v.updatePageList()
		v.rebuild(false)
	} else if job.frame.image != nil {
		v.frames[job.page] = job.frame
		v.trimFrames(job.page)
	}
	v.app.updateActions()
	v.changedViewport()
}

func (v *documentView) trimFrames(latest int) {
	const budget = 120 << 20
	bytes := 0
	for _, frame := range v.frames {
		bytes += len(frame.image.pixels)
	}
	_, height := v.viewportSize()
	visible := make(map[int]bool)
	for _, page := range v.spread.visible(v.scrollY, height) {
		visible[v.order[page.index]] = true
	}
	for len(v.frames) > 1 && (len(v.frames) > 24 || bytes > budget) {
		victim, distance := 0, -1
		for number := range v.frames {
			if number == latest || number == v.order[v.page] || visible[number] {
				continue
			}
			position := slices.Index(v.order, number)
			gap := position - v.page
			if gap < 0 {
				gap = -gap
			}
			if gap > distance {
				victim, distance = number, gap
			}
		}
		if victim == 0 {
			break
		}
		bytes -= len(v.frames[victim].image.pixels)
		delete(v.frames, victim)
	}
}

func (v *documentView) syncProofPanel() {
	if !v.proofMode() || !v.ready() {
		return
	}
	number := v.order[v.page]
	if v.proofPage == number {
		return
	}
	frame, exists := v.frames[number]
	v.plates = nil
	sendMessage.Call(v.list, 0x0184, 0, 0)
	setText(v.sample, "")
	if !exists || frame.proof == nil {
		v.proofPage = -1
		return
	}
	v.proofPage = number
	result := frame.proof
	v.plates = result.Plates
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
	setText(v.outputInfo, multiline("Výstupní záměr\n"+intent+"\nSimulace  "+result.Simulation))
	setText(v.sample, fmt.Sprintf("Str. %d · plocha / maximální tón · %d dpi", v.page+1, result.DPI))
}

func (v *documentView) updatePageList() {
	sendMessage.Call(v.pageList, 0x000b, 0, 0)
	sendMessage.Call(v.pageList, 0x0184, 0, 0)
	for i, n := range v.order {
		item := fmt.Sprintf("Stránka %d", i+1)
		if i+1 != n {
			item += fmt.Sprintf("   ·   původní %d", n)
		}
		sendMessage.Call(v.pageList, 0x0180, 0, uintptr(unsafe.Pointer(wideString(item))))
	}
	sendMessage.Call(v.pageList, 0x000b, 1, 0)
	invalidate(v.pageList)
}

func (v *documentView) updateActions() {
	loaded := v.ready()
	editing := loaded && !v.app.busy
	for _, control := range []uintptr{v.next, v.previous, v.jump, v.pageInput, v.zoomIn, v.zoomOut, v.fit, v.pageList} {
		setEnabled(control, loaded)
	}
	setEnabled(v.previous, loaded && v.page > 0)
	setEnabled(v.next, loaded && v.page+1 < len(v.order))
	setEnabled(v.left, editing)
	setEnabled(v.right, editing)
	setEnabled(v.back, editing && v.page > 0)
	setEnabled(v.ahead, editing && v.page+1 < len(v.order))
	setEnabled(v.undoButton, editing && len(v.history) > 0)
	setEnabled(v.save, editing && v.changed())
	for _, control := range []uintptr{v.list, v.all, v.none} {
		setEnabled(control, len(v.plates) > 0)
	}
	if loaded {
		if getText(v.pageInput) != strconv.Itoa(v.page+1) {
			focus, _, _ := user32.NewProc("GetFocus").Call()
			if focus != v.pageInput {
				setText(v.pageInput, strconv.Itoa(v.page+1))
			}
		}
		setText(v.location, fmt.Sprintf("z %d", len(v.order)))
		setText(v.zoomLabel, fmt.Sprintf("%.0f %%", 100*v.zoom/(float64(v.app.dpi)/72)))
		current, _, _ := sendMessage.Call(v.pageList, 0x0188, 0, 0)
		if int(int32(current)) != v.page {
			sendMessage.Call(v.pageList, 0x0186, uintptr(v.page), 0)
		}
	} else {
		setText(v.pageInput, "")
		setText(v.location, "")
		setText(v.zoomLabel, "")
	}
}

func (v *documentView) rotate(angle int) {
	number := v.order[v.page] - 1
	v.history = append(v.history, pageEdit{v.page, v.page, v.rotations[number]})
	v.rotations[number] = (v.rotations[number] + 360 + angle) % 360
	v.reframe(true)
}

func (v *documentView) move(destination int) {
	if destination < 0 || destination >= len(v.order) {
		return
	}
	v.history = append(v.history, pageEdit{from: v.page, to: destination})
	v.order[v.page], v.order[destination] = v.order[destination], v.order[v.page]
	v.updatePageList()
	setText(v.scanReport, "")
	v.page = destination
	v.reference = destination
	v.buildLayout()
	v.gotoPage(destination, false)
}

func (v *documentView) undo() {
	index := len(v.history) - 1
	if index < 0 {
		return
	}
	edit := v.history[index]
	v.history = v.history[:index]
	if edit.from == edit.to {
		v.rotations[v.order[edit.from]-1] = edit.angle
		v.reframe(true)
	} else {
		v.order[edit.from], v.order[edit.to] = v.order[edit.to], v.order[edit.from]
		v.updatePageList()
		setText(v.scanReport, "")
	}
	v.page, v.reference = edit.from, edit.from
	v.buildLayout()
	v.gotoPage(edit.from, false)
}

func (v *documentView) command(id, notification int) bool {
	if notification == 1 {
		switch id {
		case idReaderFlow:
			x, y := v.viewportSize()
			anchor, ok := v.spread.anchor(v.scrollX+x/2, v.scrollY+y/2)
			v.continuous = comboIndex(v.flow) == 0
			if ok {
				v.page, v.reference = anchor.index, anchor.index
			}
			v.buildLayout()
			if ok {
				v.restore(anchor, x/2, y/2)
			}
			v.changedViewport()
		case idReaderFit:
			value := comboIndex(v.fit)
			if value >= 0 && value < 2 {
				v.reference, v.fitMode = v.page, value+1
				v.rebuild(false)
				v.changedViewport()
			}
		case idReaderScreenBox, idReaderProofBox, idReaderProof:
			v.reframe(true)
		case idReaderPageList:
			number, _, _ := sendMessage.Call(v.pageList, 0x0188, 0, 0)
			v.gotoPage(int(int32(number)), false)
		case idReaderPlateList:
			for index, plate := range v.plates {
				on, _, _ := sendMessage.Call(v.list, 0x0187, uintptr(index), 0)
				v.hidden[plate.Name] = on == 0
			}
			v.reframe(false)
		default:
			return false
		}
		return true
	}
	if notification != 0 || !v.ready() {
		return false
	}
	switch id {
	case idReaderPrevious:
		v.gotoPage(v.page-1, false)
	case idReaderNext:
		v.gotoPage(v.page+1, false)
	case idReaderGo:
		index, err := strconv.Atoi(strings.TrimSpace(getText(v.pageInput)))
		if err != nil || index < 1 || index > len(v.order) {
			v.app.report("Zadejte číslo existující stránky.")
			return true
		}
		v.gotoPage(index-1, false)
	case idReaderZoomOut, idReaderZoomIn:
		w, h := v.viewportSize()
		dir := 1.0
		if id == idReaderZoomOut {
			dir = -1
		}
		v.zoomAt(dir, w/2, h/2)
	case idReaderAll, idReaderNone:
		for _, plate := range v.plates {
			v.hidden[plate.Name] = id == idReaderNone
		}
		v.reframe(false)
	case idReaderScan:
		if !v.app.busy {
			req := v.request("scan")
			req.View.Order = slices.Clone(v.order)
			v.app.run(req)
		}
	case idReaderRotateLeft, idReaderRotateRight, idReaderMoveBack, idReaderMoveAhead, idReaderUndo, idReaderSave:
		if v.app.busy {
			return true
		}
		switch id {
		case idReaderRotateLeft:
			v.rotate(-90)
		case idReaderRotateRight:
			v.rotate(90)
		case idReaderMoveBack:
			v.move(v.page - 1)
		case idReaderMoveAhead:
			v.move(v.page + 1)
		case idReaderUndo:
			v.undo()
		case idReaderSave:
			if v.changed() {
				req := v.request("edit_pages")
				v.includeEdits(&req)
				if v.app.chooseOutput(&req, "-upraveno") {
					v.app.run(req)
				}
			}
		}
	default:
		return false
	}
	return true
}

func (v *documentView) paint(dc uintptr) {
	viewport := v.viewport
	fillColor(dc, viewport, inputColor)
	if len(v.spread.sheets) == 0 {
		text := "Otevřít PDF"
		if v.describing {
			text = "Načítám dokument…"
		}
		if v.app.full {
			text += "   ·   Esc zpět"
		}
		drawText(dc, text, viewport, v.app.titleFont, mutedColor, 1)
		return
	}
	_, height := v.viewportSize()
	saved, _, _ := gdi32.NewProc("SaveDC").Call(dc)
	gdi32.NewProc("IntersectClipRect").Call(dc, uintptr(viewport.left), uintptr(viewport.top), uintptr(viewport.right), uintptr(viewport.bottom))
	gdi32.NewProc("SetStretchBltMode").Call(dc, 4)
	for _, page := range v.spread.visible(v.scrollY, height) {
		x := int(viewport.left) + int(math.Round(page.left-v.scrollX))
		y := int(viewport.top) + int(math.Round(page.top-v.scrollY))
		width, height := max(1, int(math.Round(page.width))), max(1, int(math.Round(page.height)))
		number := v.order[page.index]
		bounds := rect{int32(x), int32(y), int32(x + width), int32(y + height)}
		shadow := rect{bounds.left + 3, bounds.top + 4, bounds.right + 5, bounds.bottom + 6}
		fillColor(dc, shadow, paletteColor{10, 10, 12})
		bitmap := v.frames[number].image
		if bitmap == nil {
			fillColor(dc, bounds, paletteColor{235, 234, 232})
			status := fmt.Sprintf("Načítám stránku %d…", page.index+1)
			if v.failed[number] {
				status = "Stránku nelze vykreslit."
			}
			drawText(dc, status, bounds, v.app.normalFont, canvasColor, 1)
			continue
		}
		drawBitmap(dc, bitmap, x, y, width, height)
	}
	gdi32.NewProc("RestoreDC").Call(dc, saved)
	if v.app.full {
		w, h := int(viewport.right), int(viewport.bottom)
		badge := rect{int32(w - v.app.scale(209)), int32(h - v.app.scale(39)), int32(w - v.app.scale(12)), int32(h - v.app.scale(10))}
		fillColor(dc, badge, panelColor)
		drawText(dc, fmt.Sprintf("%d / %d   ·   Esc zpět", v.page+1, len(v.order)), badge, v.app.normalFont, textColor, 1)
	}
}
