package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"
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
	idReaderMoveTo
	idReaderDeletePages
	idReaderScanSelection
)

const (
	wmReaderFinished          = 0x8002
	readerZoomTimer           = 2
	readerComparisonZoomTimer = 3
	zoomSettleDelay           = 65 * time.Millisecond
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
	if rgba, ok := decoded.(*image.RGBA); ok {
		rowSize := result.width * 4
		for y := range result.height {
			src := rgba.Pix[y*rgba.Stride : y*rgba.Stride+rowSize]
			dst := result.pixels[y*rowSize : (y+1)*rowSize]
			for i := 0; i < rowSize; i += 4 {
				dst[i], dst[i+1], dst[i+2] = src[i+2], src[i+1], src[i]
			}
		}
		return result, nil
	}
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

func (v *documentView) drawListRow(item *buttonDraw) bool {
	if v.drawBookmarkRow(item) {
		return true
	}
	if item.controlType != 2 || (item.handle != v.list && item.handle != v.pageList) { // ODT_LISTBOX
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, panelColor)
	index := int(item.itemID)
	selected := item.state&1 != 0
	foreground := textColor
	if selected {
		foreground = selectionColor
	}
	var label string
	if item.handle == v.pageList {
		if index >= len(v.order) {
			return true
		}
		v.drawPageThumbnail(item, index, foreground)
		return true
	} else {
		if index >= len(v.plates) {
			return true
		}
		plate := v.plates[index]
		label = plateRow(plate)
		left, size := int32(v.app.scale(7)), int32(v.app.scale(16))
		top := bounds.top + (bounds.bottom-bounds.top-size)/2
		check := rect{bounds.left + left, top, bounds.left + left + size, top + size}
		rounded(item.dc, check, v.app.scale(5), panelColor, disabledColor)
		if selected {
			drawCheck(item.dc, check, selectionColor)
		}
		if !plate.Active {
			foreground = mutedColor
		}
		bounds.left += int32(v.app.scale(30))
	}
	drawText(item.dc, label, bounds, v.app.normalFont, foreground, 0x8000)
	if item.state&16 != 0 {
		v.app.drawFocus(item.dc, item.bounds)
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
	page             int
	order, selection []int
	rotations        []pageRotation
	bookmarks        []pdfBookmark
	bookmarkSelection int
	draftBase *documentContent
	draftAction string
}

type pageRotation struct{ number, angle int }

type readerFrame struct {
	image  *printBitmap
	proof  *viewResponse
	pixels int
}

type readerTask struct {
	generation   int
	page, pixels int
	background   bool
	tile         *screenTile
}

type readerJob struct {
	textFrame uint64
	readerTask
	action string
	result response
	frame  readerFrame
}

type documentContent struct {
	path, password, draft                        string
	original, draftInfo                          os.FileInfo
	sizes                                        []map[string][2]float64
	intent                                       string
	order, rotations, savedOrder, savedRotations  []int
	history                                      []pageEdit
	bookmarks, savedBookmarks                    []pdfBookmark
	page, reference                              int
	pageSelection                                []int
	thumbnailScale, fitMode                      int
	continuous, describing, invalidSource        bool
	zoom                                         float64
	hidden                                       map[string]bool
}

func emptyDocument() documentContent {
	return documentContent{fitMode: 1, thumbnailScale: 32, hidden: make(map[string]bool)}
}

type documentView struct {
	documentContent
	app                                                                        *app
	secondary                                                                  bool
	secondaryBox                                                               int
	thumbnailSidebar                                                           uintptr
	readerError                                                                string
	nativeText                                                                 textProbeState
	dragStart                                                                  point
	dragMoved                                                                  bool
	engine, prefetchEngine                                                     *readerEngine
	bookmarkPanel                                                              bookmarkPanel
	search                                                                     searchPanel
	pageDrag                                                                   thumbnailDrag
	thumbnailPanelPixels                                                       int
	viewport                                                                   rect
	spread                                                                     readingLayout
	scrollX, scrollY                                                           float64
	dragging                                                                   bool
	dragX, dragY                                                               int32
	wheelRemainder                                                             float64
	generation                                                                 int
	running                                                                    bool
	reading, printPreview, thumbnails                                          pageCache
	visibleDetails                                                             []screenTile
	detailFrames                                                               map[screenTile]detailFrame
	failedDetails                                                              map[screenTile]bool
	detailClock                                                                uint64
	detailBytes                                                                int64
	canvas                                                                     viewportSurface
	canvasDirty                                                                bool
	zoomDeadline                                                               time.Time
	cacheLimit                                                                 int64
	prefetch                                                                   readerTask
	prefetchFailed                                                             map[int]bool
	jobs                                                                       chan readerJob
	profiles                                                                   []string
	plates                                                                     []plateInfo
	proofPage                                                                  int
	previous, next, jump, pageInput, location, zoomOut, zoomIn, zoomLabel, fit uintptr
	flow, screenBox, proofBox, proof, list, pageList, documentInfo             uintptr
	outputInfo, sample, scanReport, all, none                                  uintptr
    convertMode, convertGrayProfile, convertGrayICC, convertGrayBrowse        uintptr
    convertKProfile, convertGrayLabel, convertKLabel, convertReport            uintptr
	thumbnailHeading, thumbnailHint, scanSelection                             uintptr
	left, right, back, ahead, undoButton, save                                 uintptr
	moveLabel, moveTarget, moveButton, deleteButton                            uintptr
}

func newDocumentView(a *app) *documentView {
	return &documentView{app: a, documentContent: emptyDocument(), proofPage: -1, jobs: make(chan readerJob, 2)}
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
		wsTabstop|listStyle, 0, a.normalFont)
	rowHeight := a.scale(24)
	sendMessage.Call(v.list, 0x01a0, 0, uintptr(rowHeight))
	a.addScrollPane(v.list, rowHeight, true)
	v.all = a.button("Všechny desky", idReaderAll, 16, 441, 145, 0)
	v.none = a.button("Vypnout všechny", idReaderNone, 169, 441, 145, 0)
	v.sample = a.label("", 16, 486, 298, false)
	a.scanButton = a.button("Zkontrolovat celý dokument", idReaderScan, 16, 519, 298, 0)
	v.scanSelection = a.button("Zkontrolovat vybrané náhledy", idReaderScanSelection, 16, 559, 298, 0)
	v.scanReport = a.readout(16, 602, 298, 133)
    a.label("Převod barev aktuálního dokumentu", 16, 763, 298, true)
    v.convertMode = a.combo(idProofColorMode, 16, 789, 298, []string{
        "Stupně šedi · DeviceGray", "Složená černá → K100", "Pouze K · CMYK C0 M0 Y0",
    })
    chooseCombo(v.convertMode, defaultIndex(colorModes, a.export.ColorMode, defaults.ColorMode))
    v.convertGrayLabel = a.label("ICC profil převodu do šedi", 16, 829, 298, false)
    v.convertGrayProfile = a.combo(idProofGrayProfile, 16, 850, 298, []string{
        "Automaticky podle dokumentu", "Gray Gamma 2.2", "Gray Gamma 1.8", "Vlastní ICC Gray…",
    })
    chooseCombo(v.convertGrayProfile, defaultIndex(grayProfiles, a.export.GrayProfile, defaults.GrayProfile))
    v.convertGrayICC = a.edit(16, 886, 187, false)
    v.convertGrayBrowse = a.button("Vybrat ICC…", idProofGrayICC, 212, 883, 102, 0)
    v.convertKLabel = a.label("Cílový tiskový profil CMYK", 16, 925, 298, false)
    cmykNames := make([]string, 0, len(catalog))
    for _, profile := range catalog { cmykNames = append(cmykNames, profile.Name) }
    v.convertKProfile = a.combo(idProofKProfile, 16, 946, 298, cmykNames)
    index := slices.IndexFunc(catalog, func(p profileChoice) bool { return p.Key == a.export.KProfile })
    chooseCombo(v.convertKProfile, max(index, 0))
    a.proofApply = a.button("Použít převod a otevřít kopii…", idProofApply, 16, 1000, 298, 0)
    a.primaryButtons[a.proofApply] = true
    v.convertReport = a.readout(16, 1047, 298, 112)
}

func (v *documentView) ready() bool { return len(v.order) > 0 && !v.invalidSource }

func (v *documentView) changed() bool {
	return v.documentContent.changed()
}

func (v *documentContent) changed() bool {
	return v.draft != "" || v.hasPageEdits()
}

func (v *documentContent) hasPageEdits() bool {
	return !slices.Equal(v.order, v.savedOrder) || !slices.Equal(v.rotations, v.savedRotations) ||
		!slices.Equal(v.bookmarks, v.savedBookmarks)
}

func (v *documentView) markSaved() {
	v.savedOrder, v.savedRotations = slices.Clone(v.order), slices.Clone(v.rotations)
	v.savedBookmarks = slices.Clone(v.bookmarks)
}

func (v *documentView) includeEdits(req *request) {
	req.View.Order, req.View.Rotations = slices.Clone(v.order), slices.Clone(v.rotations)
	req.View.Bookmarks = v.bookmarkChanges()
}

func unchanged(path string, original os.FileInfo) bool {
	info, err := os.Stat(path)
	return err == nil && original != nil && os.SameFile(original, info) &&
		info.Size() == original.Size() && info.ModTime().Equal(original.ModTime())
}

func (v *documentView) readSource() (string, os.FileInfo) {
	if v.draft != "" {
		return v.draft, v.draftInfo
	}
	return v.path, v.original
}

func (v *documentView) sourceUnchanged() bool {
	path, info := v.readSource()
	return unchanged(path, info)
}

func (v *documentView) resetRendering() {
	v.generation++
	user32.NewProc("KillTimer").Call(v.app.window, v.zoomTimer())
	v.zoomDeadline = time.Time{}
	v.readerError = ""
	v.nativeText = textProbeState{}
	if v.dragging {
		v.dragging = false
		user32.NewProc("ReleaseCapture").Call()
	}
	released := v.cachedBytes()
	v.canvas.close()
	v.reading.clear()
	v.printPreview.clear()
	v.thumbnails.clear()
	v.endPageDrag()
	v.clearDetails()
	v.closeReaders()
	v.prefetchFailed = make(map[int]bool)
	v.wheelRemainder = 0
	if released >= 32<<20 {
		go debug.FreeOSMemory()
	}
}

func (v *documentView) resetDocument() {
	v.resetSearch()
	v.app.clearTextSelection()
	v.app.resetThumbnailScopes()
	user32.NewProc("KillTimer").Call(v.app.window, readerMemoryTimer)
	v.resetRendering()
	v.documentContent = emptyDocument()
	v.bookmarkPanel.collapsed = nil
	v.refreshBookmarks(-1)
	v.scrollX, v.scrollY = 0, 0
	v.clearPlateView()
	chooseCombo(v.fit, 0)
	chooseCombo(v.flow, 1)
	setText(v.documentInfo, "")
	setText(v.scanReport, "")
	v.rebuild(false)
	v.updatePageList()
}

func (v *documentView) closeReaders() {
	if v.engine != nil {
		v.engine.close()
		v.engine = nil
	}
	v.stopPrefetchEngine()
}

func (v *documentView) stopPrefetchEngine() {
	if v.prefetchEngine != nil {
		v.prefetchEngine.close()
		v.prefetchEngine = nil
	}
}

func (v *documentView) open(path, password string, original os.FileInfo) {
	v.resetDocument()
	v.path, v.password, v.original = path, password, original
	v.describing = true
	v.resumeReader()
	v.app.report("Načítám dokument…")
	v.app.updateActions()
	v.changedViewport()
}

func (v *documentView) resumeReader() {
	v.engine = &readerEngine{}
	v.updateCacheLimit()
	user32.NewProc("SetTimer").Call(v.app.window, readerMemoryTimer, 3000, 0)
}

func (v *documentView) proofMode() bool { return !v.secondary && v.app.mode == idProofTab }

func (v *documentView) box() string {
	if v.proofMode() {
		boxes := [...]string{"media", "crop", "trim", "bleed"}
		return boxes[max(0, comboIndex(v.proofBox))]
	}
	boxes := [...]string{"crop", "media", "trim", "bleed"}
	if v.secondary {
		return boxes[v.secondaryBox]
	}
	return boxes[max(0, comboIndex(v.screenBox))]
}

func (v *documentView) request(action string) request {
	opts := v.app.export
	path, _ := v.readSource()
	req := request{Action: action, Input: path, PDFPassword: v.password, Export: &opts,
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

func (v *documentView) clearPlateView() {
	v.proofPage = -1
	v.plates = nil
	sendMessage.Call(v.list, 0x0184, 0, 0)
	v.app.scrollPanes[v.list].reflow()
	setText(v.sample, "")
	setText(v.outputInfo, "")
}

func (v *documentView) reframe() {
	v.generation++
	v.clearDetails()
	v.clearPlateView()
	setText(v.scanReport, "")
	v.reference = v.page
	v.rebuild(true)
	v.changedViewport()
}

func (v *documentView) sizesInOrder() []pageSize {
	box := v.box()
	result := make([]pageSize, len(v.order))
	for index, page := range v.order {
		result[index] = v.pageDimensions(page, box)
	}
	return result
}

func (v *documentView) pageDimensions(number int, box string) pageSize {
	d := v.sizes[number-1][box]
	if d[0] <= 0 || d[1] <= 0 {
		d = [2]float64{595, 842}
	}
	if v.rotations[number-1]%180 != 0 {
		d[0], d[1] = d[1], d[0]
	}
	return pageSize{d[0], d[1]}
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
	v.gotoPageAt(index, atEnd, nil)
}

func (v *documentView) gotoPageAt(index int, atEnd bool, anchor *readingAnchor) {
	if index < 0 || index >= len(v.order) {
		return
	}
	v.page = index
	if !v.continuous {
		v.reference = index
		v.buildLayout()
	}
	if anchor == nil {
		v.centerPage(atEnd)
	} else {
		width, height := v.viewportSize()
		v.restore(*anchor, width/2, height/2)
	}
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
	if timer, _, _ := user32.NewProc("SetTimer").Call(v.app.window, v.zoomTimer(),
		uintptr(zoomSettleDelay/time.Millisecond), 0); timer != 0 {
		v.zoomDeadline = time.Now().Add(zoomSettleDelay)
	}
	v.changedViewport()
}

func (v *documentView) zoomTimer() uintptr {
	if v.secondary {
		return readerComparisonZoomTimer
	}
	return readerZoomTimer
}

func (v *documentView) finishZoom() {
	if !v.zoomDeadline.IsZero() && time.Now().Before(v.zoomDeadline) {
		return
	}
	user32.NewProc("KillTimer").Call(v.app.window, v.zoomTimer())
	v.zoomDeadline = time.Time{}
	v.schedule()
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

func absTextDrag(value int32) int32 {
	return max(value,-value)
}

func (v *documentView) pointer(msg uint32, param uintptr) bool {
	x, y := int32(int16(param)), int32(int16(param>>16))
	if !v.secondary && v.app.textPointer(msg, x, y) {
		return true
	}
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
				v.dragStart, v.dragMoved = point{x, y}, false
				user32.NewProc("SetCapture").Call(v.app.window)
			}
			return true
		}
	case 0x0200:
		if v.dragging {
			v.dragMoved = v.dragMoved || max(absTextDrag(x-v.dragStart.x), absTextDrag(y-v.dragStart.y)) > int32(v.app.scale(3))
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
			if !v.dragMoved && max(absTextDrag(x-v.dragStart.x), absTextDrag(y-v.dragStart.y)) <= int32(v.app.scale(3)) {
				v.clickedForText(x, y)
			}
			return true
		}
	case 0x0215:
		v.dragging = false
	}
	return false
}

func (v *documentView) changedViewport() {
	if v.secondary {
		v.syncPageListSelection()
	} else {
		v.syncProofPanel()
		v.updateActions()
		v.app.layoutInlineEditor()
	}
	v.app.updateComparisonNavigation()
	v.syncDetails()
	v.invalidateCanvas()
	v.schedule()
}

func previewPixels(size float64) int {
	switch {
	case size <= 780:
		return 820
	case size <= 1020:
		return 1070
	default:
		return 1450
	}
}

func (v *documentView) schedule() {
	if v.app.busy && v.app.pending.ReplaceSource {
		v.app.startOverwrite()
		return
	}
	if v.running || v.path == "" || v.invalidSource || (v.secondary && !v.app.comparisonVisible()) {
		return
	}
	if v.describing {
		v.launch(v.request("describe"), 0, 0, false, nil)
		return
	}
	if !v.ready() || !v.zoomDeadline.IsZero() {
		return
	}
	if v.scheduleNativeText() || (!v.secondary && v.app.scheduleInlineBackground()) {
		return
	}
	_, height := v.viewportSize()
	candidates := slices.Clone(v.spread.visible(v.scrollY, height))
	if current, ok := v.spread.position(v.page); ok {
		candidates = append([]sheet{current}, candidates...)
	}
	cache, proof := v.activeCache(), v.proofMode()
	for _, item := range candidates {
		number := v.order[item.index]
		frame := cache.frames[number]
		pixels := v.fullPreviewTarget(item)
		if pixels == 0 || cache.failed[number] || frame.pixels >= pixels {
			continue
		}
		if !proof && v.prefetch.generation == v.generation && v.prefetch.page == number && v.prefetch.pixels >= pixels {
			continue
		}
		v.renderPage(number, pixels, false)
		return
	}
	for _, tile := range v.visibleDetails {
		if v.detailFrames[tile].image == nil && !v.failedDetails[tile] {
			v.renderDetail(tile)
			return
		}
	}
	if !v.app.busy && v.prefetch.page == 0 {
		if !v.scheduleThumbnails() && !proof && !v.secondary {
			v.schedulePrefetch()
		}
	}
}

func (v *documentView) schedulePrefetch() {
	v.updateCacheLimit()
	free := v.cacheLimit - v.reading.bytes - v.printPreview.bytes - v.detailBytes - v.canvas.bytes() - v.thumbnails.bytes
	if free > 0 {
		box := v.box()
		width, height := v.viewportSize()
		margin := float64(2 * v.app.scale(18))
		for offset := 1; offset < len(v.order); offset++ {
			number := v.order[(v.page+offset)%len(v.order)]
			if v.reading.frames[number].image != nil || v.reading.failed[number] || v.prefetchFailed[number] {
				continue
			}
			size := v.pageDimensions(number, box)
			long, short := max(size.width, size.height), min(size.width, size.height)
			fit := min((width-margin)/size.width, (height-margin)/size.height)
			pixels := previewPixels(fit * long)
			// Leave room for Ghostscript rounding the width and height to pixels.
			rendered := min(float64(pixels), long*64)
			estimate := int64(math.Ceil(rendered)+2) * int64(math.Ceil(rendered*short/long)+2) * 4
			if estimate <= free {
				v.renderPage(number, pixels, true)
				return
			}
		}
	}
	v.stopPrefetchEngine()
}

func (v *documentView) renderPage(number, pixels int, background bool) {
	action := "screen"
	if v.proofMode() {
		action = "preview"
	}
	req := v.request(action)
	req.View.Page, req.View.Rotation, req.View.Pixels = number, v.rotations[number-1], pixels
	v.launch(req, number, pixels, background, nil)
}

func (v *documentView) launch(req request, number, pixels int, background bool, tile *screenTile) {
	_, original := v.readSource()
	generation := v.generation
	var textFrame uint64
	if req.Action == "text_background" {
		textFrame = v.app.textTool.inline.serial
	}
	task := readerTask{generation, number, pixels, background, tile}
	engine := v.engine
	if background {
		v.prefetch = task
		if v.prefetchEngine == nil {
			v.prefetchEngine = &readerEngine{background: true}
		}
		engine = v.prefetchEngine
	} else {
		v.running = true
	}
	go func() {
		job := readerJob{readerTask: task, action: req.Action, textFrame: textFrame,
			result: response{Status: "changed", Message: "Vstupní PDF se změnilo. Otevřete je znovu."}}
		if unchanged(req.Input, original) {
			result := engine.run(req)
			if unchanged(req.Input, original) {
				job.result = result
			}
		}
		if job.result.Status == "ok" && number > 0 && job.result.View != nil {
			var err error
			view := job.result.View
			if view.RasterWidth != 0 {
				job.frame.image = &printBitmap{width: view.RasterWidth, height: view.RasterHeight, pixels: view.Image}
			} else {
				job.frame.image, err = decodePrintBitmap(view.Image)
			}
			if err != nil {
				job.result = operationError("Obraz náhledu se nepodařilo načíst.")
			} else if tile != nil && (job.frame.image.width != tile.Width || job.frame.image.height != tile.Height) {
				job.frame.image = nil
				job.result = operationError("Rozměry vykresleného výřezu neodpovídají náhledu.")
			} else {
				job.frame.proof, job.frame.pixels = view, pixels
			}
			view.Image = nil
		}
		v.jobs <- job
		postMessage.Call(v.app.window, wmReaderFinished, v.readerMessageSource(), 0)
	}()
}

func (v *documentView) hasRenderJobs() bool {
	return v.running || v.prefetch.page != 0
}

func (v *documentView) readerMessageSource() uintptr {
	if v.secondary {
		return 1
	}
	return 0
}

func (v *documentView) finished(job readerJob) {
	if job.background {
		v.finishedPrefetch(job)
		return
	}
	v.running = false
	if job.action == "text_probe" {
		v.finishNativeText(job)
		v.schedule()
		return
	}
	if job.action == "text_background" {
		v.app.finishInlineBackground(job)
		v.schedule()
		return
	}
	if job.generation != v.generation {
		v.schedule()
		return
	}
	if job.tile != nil {
		v.finishDetail(job)
		return
	}
	if job.result.Status != "ok" {
		v.describing = false
		v.invalidSource = job.result.Status == "changed"
		if job.page != 0 {
			v.activeCache().failed[job.page] = true
		}
		if job.page == 0 && !v.secondary {
			setText(v.documentInfo, multiline(job.result.Message))
			if !v.app.full {
				v.app.activate(idBrowseTab)
			}
		}
		if v.secondary {
			if job.page == 0 || v.invalidSource {
				v.showComparisonReaderError(job.result.Message)
			}
		} else {
			v.app.report(job.result.Message)
		}
	} else if job.action == "describe" {
		data := job.result.View
		v.describing = false
		if data == nil || len(data.Sizes) == 0 {
			if v.secondary {
				v.showComparisonReaderError("Rozměry stránek nelze načíst.")
			} else {
				v.app.report("Rozměry stránek nelze načíst.")
			}
			return
		}
		v.applyDescription(data)
		v.app.rememberRecentFile(v.path)
		if v.secondary {
			v.app.layoutComparisonThumbnails(v.app.comparison.thumbnailArea)
		} else {
			v.app.report("")
		}
	} else if job.frame.image != nil {
		v.activeCache().store(job.page, job.frame)
		v.trimCaches()
	}
	if !v.secondary {
		v.updateDocumentInfo()
		v.app.updateActions()
	}
	v.changedViewport()
}

func (v *documentView) applyDescription(data *viewResponse) {
	v.sizes, v.intent, v.bookmarks = data.Sizes, data.OutputIntent, data.Bookmarks
	v.order, v.rotations = make([]int, len(v.sizes)), make([]int, len(v.sizes))
	for i := range v.order {
		v.order[i] = i + 1
	}
	v.page = min(v.page, len(v.order)-1)
	v.reference = v.page
	v.markSaved()
	if !v.secondary {
		v.refreshBookmarks(-1)
	}
	v.updatePageList()
	v.rebuild(false)
}

func (v *documentView) finishedPrefetch(job readerJob) {
	v.prefetch = readerTask{}
	if job.generation == v.generation && v.ready() {
		v.updateCacheLimit()
		switch {
		case job.result.Status == "changed":
			v.invalidSource = true
			if v.secondary {
				v.showComparisonReaderError(job.result.Message)
				v.changedViewport()
			} else {
				v.app.report(job.result.Message)
				v.app.updateActions()
			}
		case job.action == "thumbnail":
			if job.frame.image == nil {
				v.thumbnails.failed[job.page] = true
			} else {
				v.thumbnails.store(job.page, job.frame)
				v.trimThumbnails(v.thumbnailBudget())
				invalidate(v.pageList)
			}
		case job.frame.image == nil:
			v.prefetchFailed[job.page] = true
		case v.reading.frames[job.page].image == nil &&
			v.reading.bytes+v.printPreview.bytes+v.detailBytes+v.canvas.bytes()+v.thumbnails.bytes+int64(len(job.frame.image.pixels)) <= v.cacheLimit:
			v.reading.store(job.page, job.frame)
			v.syncDetails()
			v.updateDocumentInfo()
			_, height := v.viewportSize()
			for _, item := range v.spread.visible(v.scrollY, height) {
				if v.order[item.index] == job.page {
					v.invalidateCanvas()
					break
				}
			}
		}
	}
	first, last := v.thumbnailRange()
	if !v.ready() || (v.proofMode() && first == last) || v.app.busy {
		v.stopPrefetchEngine()
	}
	v.schedule()
}

func (v *documentView) updateDocumentInfo() {
	if v.secondary {
		return
	}
	if !v.ready() {
		return
	}
	intent := v.intent
	if intent == "" {
		intent = "Neuveden"
	}
	setText(v.documentInfo, multiline(fmt.Sprintf("%d stran\nVýstupní záměr PDF\n%s\nV paměti %d z %d stran",
		len(v.order), intent, len(v.reading.frames), len(v.order))))
}

func (v *documentView) syncProofPanel() {
	if !v.proofMode() || !v.ready() {
		return
	}
	number := v.order[v.page]
	if v.proofPage == number {
		return
	}
	frame, exists := v.printPreview.frames[number]
	v.plates = nil
	sendMessage.Call(v.list, 0x0184, 0, 0)
	setText(v.sample, "")
	if !exists || frame.proof == nil {
		v.app.scrollPanes[v.list].reflow()
		v.proofPage = -1
		return
	}
	v.proofPage = number
	result := frame.proof
	v.plates = result.Plates
	for _, plate := range v.plates {
		sendMessage.Call(v.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(plateRow(plate)))))
	}
	v.app.scrollPanes[v.list].reflow()
	v.syncPlateSelection()
	intent := result.OutputIntent
	if intent == "" {
		intent = "Neuveden"
	}
	setText(v.outputInfo, multiline("Výstupní záměr\n"+intent+"\nSimulace  "+result.Simulation))
	setText(v.sample, fmt.Sprintf("Str. %d · plocha / maximální tón · %d dpi", v.page+1, result.DPI))
}

func (v *documentView) syncPlateSelection() {
	for index, plate := range v.plates {
		selected := uintptr(1)
		if v.hidden[plate.Name] {
			selected = 0
		}
		sendMessage.Call(v.list, 0x0185, selected, uintptr(index)) // LB_SETSEL
	}
}

func (v *documentView) platesChanged() {
	v.generation++
	v.printPreview.clear()
	v.changedViewport()
}

func (v *documentView) updatePageList() {
	sendMessage.Call(v.pageList, 0x000b, 0, 0)
	sendMessage.Call(v.pageList, 0x0184, 0, 0)
	for i := range v.order {
		sendMessage.Call(v.pageList, 0x0180, 0, uintptr(unsafe.Pointer(wideString(strconv.Itoa(i+1)))))
	}
	v.resetPageListSelection()
	sendMessage.Call(v.pageList, 0x000b, 1, 0)
	v.app.scrollPanes[v.pageList].reflow()
	invalidate(v.pageList)
}

func (v *documentView) updateActions() {
	loaded := v.ready()
	editing := loaded && !v.app.busy
	v.app.updateMergeInputButtons()
	v.updateBookmarkActions()
	v.updateSearchActions()
	for _, control := range []uintptr{v.next, v.previous, v.jump, v.pageInput, v.zoomIn, v.zoomOut, v.fit, v.pageList} {
		setEnabled(control, loaded)
	}
	setEnabled(v.previous, loaded && v.page > 0)
	setEnabled(v.next, loaded && v.page+1 < len(v.order))
	setEnabled(v.left, editing)
	setEnabled(v.right, editing)
	selected := v.selectedPagePositions()
	single := len(selected) == 1
	setEnabled(v.back, editing && single && selected[0] > 0)
	setEnabled(v.ahead, editing && single && selected[0]+1 < len(v.order))
	setEnabled(v.scanSelection, editing)
	setEnabled(v.undoButton, editing && len(v.history) > 0)
	setEnabled(v.save, editing && v.changed())
	v.updatePageEditActions(editing)
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
		v.syncPageListSelection()
	} else {
		setText(v.pageInput, "")
		setText(v.location, "")
		setText(v.zoomLabel, "")
	}
}

func (v *documentView) rotate(angle int) {
	selected := v.selectedPagePositions()
	before := make([]pageRotation, 0, len(selected))
	after := make([]pageRotation, 0, len(selected))
	for _, position := range selected {
		number := v.order[position]
		previous := v.rotations[number-1]
		before = append(before, pageRotation{number, previous})
		after = append(after, pageRotation{number, (previous + 360 + angle) % 360})
	}
	v.history = append(v.history, pageEdit{page: v.page, rotations: before})
	v.applyRotations(after)
}

func (v *documentView) applyRotations(rotations []pageRotation) {
	for _, rotation := range rotations {
		number := rotation.number
		v.rotations[number-1] = rotation.angle
		v.reading.remove(number)
		v.printPreview.remove(number)
		v.thumbnails.remove(number)
		delete(v.prefetchFailed, number)
	}
	invalidate(v.pageList)
	v.updateDocumentInfo()
	v.reframe()
}

func (v *documentView) undo() {
	index := len(v.history) - 1
	if index < 0 {
		return
	}
	edit := v.history[index]
	v.history[index] = pageEdit{}
	v.history = v.history[:index]
	if edit.draftBase != nil {
		v.app.undoDocumentDraft(*edit.draftBase, edit.draftAction)
		return
	}
	if edit.bookmarks != nil {
		v.bookmarks = edit.bookmarks
		v.refreshBookmarks(edit.bookmarkSelection)
		v.app.updateActions()
		return
	}
	if edit.order == nil {
		v.page, v.reference = edit.page, edit.page
		v.applyRotations(edit.rotations)
		v.gotoPage(edit.page, false)
		return
	}
	v.changePageOrder(edit.order, edit.selection, edit.page)
}

func (v *documentView) command(id, notification int) bool {
	if v.bookmarkCommand(id, notification) || v.searchCommand(id, notification) {
		return true
	}
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
		case idReaderScreenBox:
			v.reading.clear()
			clear(v.prefetchFailed)
			v.updateDocumentInfo()
			v.reframe()
		case idReaderProofBox, idReaderProof:
			v.printPreview.clear()
			v.reframe()
		case idReaderPageList:
			v.readPageSelection()
		case idReaderPlateList:
			for index, plate := range v.plates {
				on, _, _ := sendMessage.Call(v.list, 0x0187, uintptr(index), 0)
				v.hidden[plate.Name] = on == 0
			}
			v.platesChanged()
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
		v.syncPlateSelection()
		v.platesChanged()
	case idReaderScan, idReaderScanSelection:
		if !v.app.busy {
			req := v.request("scan")
			if id == idReaderScanSelection {
				req.View.Order = v.selectedSourcePages()
			} else {
				req.View.Order = slices.Clone(v.order)
			}
			v.app.run(req)
		}
	case idReaderRotateLeft, idReaderRotateRight, idReaderMoveBack, idReaderMoveAhead, idReaderUndo, idReaderSave,
		idReaderMoveTo, idReaderDeletePages:
		if v.app.busy {
			return true
		}
		switch id {
		case idReaderRotateLeft:
			v.rotate(-90)
		case idReaderRotateRight:
			v.rotate(90)
		case idReaderMoveBack:
			v.moveSelectedPages(v.selectedPagePositions()[0] - 1)
		case idReaderMoveAhead:
			v.moveSelectedPages(v.selectedPagePositions()[0] + 1)
		case idReaderMoveTo:
			v.moveToPosition()
		case idReaderDeletePages:
			v.deleteSelectedPages()
		case idReaderUndo:
			v.undo()
		case idReaderSave:
			if v.changed() {
				v.app.saveDocument(true)
			}
		}
	default:
		return false
	}
	return true
}

func (v *documentView) paintCanvas(dc uintptr) {
	viewport := v.viewport
	fillColor(dc, viewport, inputColor)
	if len(v.spread.sheets) == 0 {
		text := "Otevřít PDF"
		if v.describing {
			text = "Načítám dokument…"
		} else if v.readerError != "" {
			text = v.readerError
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
	cache := v.activeCache()
	for _, page := range v.spread.visible(v.scrollY, height) {
		number := v.order[page.index]
		bounds := v.pageBounds(page)
		shadow := rect{bounds.left + 3, bounds.top + 4, bounds.right + 5, bounds.bottom + 6}
		fillColor(dc, shadow, paletteColor{10, 10, 12})
		bitmap := cache.frames[number].image
		details := v.fullPreviewTarget(page) == 0
		if details {
			fillColor(dc, bounds, paletteColor{255, 255, 255})
			for _, tile := range v.visibleDetails {
				if tile.page == number && v.failedDetails[tile] {
					details = false
					break
				}
			}
		}
		if !details && bitmap == nil {
			fillColor(dc, bounds, paletteColor{235, 234, 232})
			drawText(dc, "Stránku nelze vykreslit.", bounds, v.app.normalFont, canvasColor, 1)
		} else if !details {
			drawBitmap(dc, bitmap, int(bounds.left), int(bounds.top), int(bounds.right-bounds.left), int(bounds.bottom-bounds.top))
		}
		v.paintDetails(dc, number, bounds)
		if !v.secondary {
			v.paintSearchHighlights(dc, number, bounds)
			v.app.paintTextSelection(dc, number, bounds)
		}
	}
	gdi32.NewProc("RestoreDC").Call(dc, saved)
	if v.app.full {
		w, h := int(viewport.right), int(viewport.bottom)
		badge := rect{int32(w - v.app.scale(209)), int32(h - v.app.scale(39)), int32(w - v.app.scale(12)), int32(h - v.app.scale(10))}
		fillColor(dc, badge, panelColor)
		drawText(dc, fmt.Sprintf("%d / %d   ·   Esc zpět", v.page+1, len(v.order)), badge, v.app.normalFont, textColor, 1)
	}
}
