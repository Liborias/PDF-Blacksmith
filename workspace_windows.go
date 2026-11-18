package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	idOpen = 100 + iota
	idShowOutput
	idTogglePanel
	idFullscreen
	idBrowseTab
	idPagesTab
	idProofTab
	idExportTab
	idSignTab
	idReload
	idICCProfile
	idICCBrowse
	idOutline
	idDecrypt
	idRaster
	idRun
	idPreset
	idResetPreset
	idCertBrowse
	idTrustBrowse
	idSign
	idVerify
	idVisible
	idClosePDF
	idFileTab
	idSaveDocument
	idSaveDocumentAs
	idPrintDocument
	idExportScope
)

const (
	exportAllPages = iota
	exportCurrentPage
	exportPageRanges
)

const (
	panelWidth   = 350
	headerHeight = 50
	footerHeight = 48
)

var workspaceTabs = []struct {
	id, width int
	name      string
	needsPDF  bool
}{
	{idFileTab, 82, "Soubor", false},
	{idBrowseTab, 109, "Zobrazení", false},
	{idPagesTab, 94, "Stránky", false},
	{idProofTab, 106, "Separace", true},
	{idSignTab, 146, "Digitální podpis", true},
}

type panelItem struct {
	handle              uintptr
	x, y, width, height int
}

type point struct{ x, y int32 }

type windowPlacement struct {
	length, flags, show uint32
	minimum, maximum    point
	normal              rect
}

type minMaxInfo struct {
	reserved, maxSize, maxPosition, minimumTrack, maximumTrack point
}

type monitorInfo struct {
	size          uint32
	monitor, work rect
	flags         uint32
}

type scrollInfo struct {
	size, mask       uint32
	minimum, maximum int32
	page             uint32
	position, track  int32
}

func (a *app) setupWorkspace() bool {
	className := wideString("PDFBlaksmithSidebar")
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	class := windowClass{instance: a.instance, className: uintptr(unsafe.Pointer(className)),
		windowProc: syscall.NewCallback(sidebarProc), cursor: cursor}
	class.size = uint32(unsafe.Sizeof(class))
	if ok, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class))); ok == 0 {
		return false
	}
	a.sidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|wsVscroll|0x02000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	if a.sidebar == 0 {
		return false
	}

	var catalog []profileChoice
	if json.Unmarshal(printCatalog, &catalog) != nil {
		return false
	}
	a.profileChoices = append([]profileChoice{{"", "Zachovat barvy"}, {"k", "Stupně šedi · pouze deska K"}}, catalog...)
	a.profileChoices = append(a.profileChoices, profileChoice{"custom", "Vlastní výstupní profil ICC"})

	a.panelButton = a.button("Skrýt panel", idTogglePanel, 0, 0, 118, 0)
	a.fullButton = a.button("Celá obrazovka", idFullscreen, 0, 0, 134, 0)
	a.windowButtons[0] = a.button("Minimalizovat", idMinimize, 0, 0, 36, 0)
	a.windowButtons[1] = a.button("Maximalizovat nebo obnovit", idMaximize, 0, 0, 36, 0)
	a.windowButtons[2] = a.button("Zavřít", idCloseWindow, 0, 0, 36, 0)
	a.status = a.add("STATIC", "", 0, 0, 0, 160, 22, 0x4000, 0, a.normalFont)
	for index, tab := range workspaceTabs {
		style := uintptr(0)
		if index == 0 {
			style = wsGroup
		}
		a.modes[tab.id] = a.button(tab.name, tab.id, 0, 0, tab.width, style)
	}

	a.view = newDocumentView(a)
	a.view.buildToolbar()
	a.beginPanel(idBrowseTab)
	a.label("Zobrazení stránek", 16, 18, 290, true)
	a.label("Rozvržení", 16, 55, 290, false)
	a.view.flow = a.combo(idReaderFlow, 16, 77, 298, []string{"Souvisle", "Po stránkách"})
	chooseCombo(a.view.flow, 1)
	a.label("Rámeček PDF v náhledu", 16, 120, 290, false)
	a.view.screenBox = a.combo(idReaderScreenBox, 16, 142, 298, []string{"Ořez zobrazení", "Celý arch", "Čistý formát", "Spadávka"})
	a.view.documentInfo = a.readout(16, 187, 298, 82)
	a.label("Heslo k otevření PDF", 16, 292, 290, false)
	a.password = a.edit(16, 315, 298, true)
	a.button("Načíst znovu s heslem", idReload, 16, 353, 298, 0)
	a.label("Ctrl + kolečko  ·  zoom pod kurzorem", 16, 408, 298, false)
	a.label("F11  ·  celá obrazovka", 16, 433, 298, false)

	a.beginPanel(idPagesTab)
	a.view.buildPagePanel()
	a.beginPanel(idProofTab)
	a.view.buildProofPanel(catalog)
	a.beginPanel(idExportTab)
	a.buildExportPanel()
	a.beginPanel(idSignTab)
	a.buildSignPanel()
	a.addingTab, a.controlParent = 0, 0
	a.layout()
	a.updateActions()
	return true
}

func (a *app) beginPanel(tab int) {
	a.addingTab, a.controlParent = tab, a.sidebar
}

func (a *app) buildExportPanel() {
	a.runButton = a.button("Uložit exportované PDF…", idRun, 16, 17, 298, 0)
	a.primaryButtons[a.runButton] = true
	a.label("Strany k exportu", 16, 64, 298, true)
	a.exportScope = a.combo(idExportScope, 16, 86, 145, []string{"Všechny strany", "Aktuální strana", "Vlastní výběr"})
	a.exportRange = a.edit(169, 86, 145, false)
	sendMessage.Call(a.exportRange, 0x1501, 1, uintptr(unsafe.Pointer(wideString("Např. 1, 3–5, 8")))) // EM_SETCUEBANNER
	a.label("Barevný výstup", 16, 124, 298, true)
	names := make([]string, 0, len(a.profileChoices))
	for _, p := range a.profileChoices {
		names = append(names, p.Name)
	}
	a.profiles = a.combo(idICCProfile, 16, 146, 298, names)
	chooseCombo(a.profiles, 1)
	a.customICC = a.edit(16, 180, 187, false)
	a.customBrowse = a.button("Vybrat ICC…", idICCBrowse, 212, 177, 102, 0)
	a.outline = a.button("Převést text na křivky", idOutline, 16, 213, 298, bsAutoCheck)
	a.decrypt = a.button("Pouze odstranit omezení heslem", idDecrypt, 16, 243, 298, bsAutoCheck)
	a.raster = a.button(fmt.Sprintf("Povolit rastrovou opravu K (%d dpi)", a.export.RepairDPI), idRaster, 16, 273, 298, bsAutoCheck)
	a.exportReport = a.readout(16, 313, 298, 105)
	a.label("Parametry exportu", 16, 439, 298, true)
	labels := []string{"Záměr reprodukce", "Černý bod", "Obrázky", "Komprese", "Kontrola desek", "Oprava K"}
	choices := [][]string{
		{"Podle PDF", "Perceptuální", "Relativní", "Saturace", "Absolutní"},
		{"Automaticky", "Zapnout", "Vypnout"},
		{"Beze změny", "150 dpi", "300 dpi", "600 dpi"},
		{"Automatická", "Bezeztrátová"},
		{"75 dpi", "150 dpi", "300 dpi"},
		{"300 dpi", "600 dpi"},
	}
	for index, label := range labels {
		x := 16 + (index%2)*153
		y := 474 + (index/2)*63
		a.label(label, x, y, 145, false)
		a.presetControls[index] = a.combo(idPreset, x, y+21, 145, choices[index])
	}
	a.button("Obnovit výchozí parametry", idResetPreset, 16, 669, 298, 0)
	a.applyExportOptions(a.export)
}

func (a *app) buildSignPanel() {
	a.label("Podepsat dokument", 16, 18, 298, true)
	a.label("Certifikát PKCS#12 (.p12 nebo .pfx)", 16, 53, 298, false)
	a.cert = a.edit(16, 76, 194, false)
	a.button("Vybrat…", idCertBrowse, 219, 74, 95, 0)
	a.label("Heslo k certifikátu", 16, 122, 298, false)
	a.certPass = a.edit(16, 144, 298, true)
	a.visible = a.button("Viditelný podpis v novém poli", idVisible, 16, 183, 298, bsAutoCheck)
	a.signButton = a.button("Uložit podepsané PDF…", idSign, 16, 224, 298, 0)
	a.label("Ověřit podpisy", 16, 284, 298, true)
	a.label("Doplňkový kořenový certifikát", 16, 318, 298, false)
	a.trust = a.edit(16, 341, 194, false)
	a.button("Vybrat…", idTrustBrowse, 219, 339, 95, 0)
	a.verifyButton = a.button("Ověřit podpisy", idVerify, 16, 384, 298, 0)
	a.signReport = a.readout(16, 431, 298, 179)
}

func (a *app) applyExportOptions(value exportOptions) {
	indexes := []int{
		defaultIndex(intents, value.Intent, defaults.Intent),
		defaultIndex(blackPoints, value.BlackPoint, defaults.BlackPoint),
		defaultIndex(imageDPIs, value.ImageDPI, defaults.ImageDPI),
		defaultIndex(compressions, value.Compression, defaults.Compression),
		defaultIndex(proofDPIs, value.ProofDPI, defaults.ProofDPI),
		defaultIndex(repairDPIs, value.RepairDPI, defaults.RepairDPI),
	}
	for index, handle := range a.presetControls {
		chooseCombo(handle, indexes[index])
	}
}

func (a *app) updatePresets() {
	c := a.presetControls
	a.export = exportOptions{
		intents[max(0, comboIndex(c[0]))], blackPoints[max(0, comboIndex(c[1]))],
		imageDPIs[max(0, comboIndex(c[2]))], compressions[max(0, comboIndex(c[3]))],
		proofDPIs[max(0, comboIndex(c[4]))], repairDPIs[max(0, comboIndex(c[5]))],
	}
	if err := saveExportOptions(a.export); err != nil {
		a.report("Předvolby nelze uložit. " + err.Error())
	}
	setText(a.raster, fmt.Sprintf("Povolit rastrovou opravu K (%d dpi)", a.export.RepairDPI))
}

func (a *app) profileKey() string {
	index := comboIndex(a.profiles)
	if index < 0 || index >= len(a.profileChoices) {
		return ""
	}
	return a.profileChoices[index].Key
}

func (a *app) activate(id int) {
	if (id == idProofTab || id == idSignTab || id == idExportTab) && !a.documentTools {
		return
	}
	old := a.mode
	a.mode, a.panelHidden = id, false
	if (old == idProofTab) != (id == idProofTab) {
		a.view.reframe()
	}
	a.layout()
}

func (a *app) layout() {
	var size rect
	if ok, _, _ := getClientRect.Call(a.window, uintptr(unsafe.Pointer(&size))); ok == 0 {
		return
	}
	a.width, a.height = int(size.right), int(size.bottom)
	if a.width == 0 || a.height == 0 {
		return
	}
	d := a.scale
	if a.full {
		a.viewport = size
	} else {
		right := a.width
		if !a.panelHidden {
			right -= d(panelWidth)
		}
		a.viewport = rect{0, int32(d(headerHeight)), int32(right), int32(max(d(headerHeight+1), a.height-d(footerHeight)))}
		a.layoutHeader()
		place(a.status, a.width-d(231), a.height-d(33), d(218), d(22))
		a.view.layoutToolbar()
		place(a.sidebar, a.width-d(panelWidth), d(headerHeight), d(panelWidth), max(1, a.height-d(headerHeight+footerHeight)))
	}
	chrome := uintptr(5)
	if a.full {
		chrome = 0
	}
	for _, control := range []uintptr{a.panelButton, a.fullButton, a.status} {
		showWindow.Call(control, chrome)
	}
	for _, tab := range workspaceTabs {
		show := chrome
		if tab.needsPDF && !a.documentTools {
			show = 0
		}
		handle := a.modes[tab.id]
		showWindow.Call(handle, show)
		invalidate(handle)
	}
	for _, handle := range a.windowButtons {
		showWindow.Call(handle, chrome)
	}
	a.view.showToolbar(chrome)
	sidebar := chrome
	if a.panelHidden {
		sidebar = 0
	}
	showWindow.Call(a.sidebar, sidebar)
	setText(a.panelButton, map[bool]string{true: "Zobrazit panel", false: "Skrýt panel"}[a.panelHidden])
	a.layoutPanel()
	a.view.resize(a.viewport)
	invalidate(a.window)
}

func (a *app) panelMetrics() (int, int) {
	height := 1
	for _, item := range a.panelItems[a.mode] {
		height = max(height, item.y+item.height+18)
	}
	available := max(1, (a.height-a.scale(headerHeight+footerHeight))*96/a.dpi)
	return height, available
}

func (a *app) layoutPanel() {
	if a.full || a.panelHidden {
		return
	}
	height, available := a.panelMetrics()
	offset := min(max(0, a.panelScroll[a.mode]), max(0, height-available))
	a.panelScroll[a.mode] = offset
	info := scrollInfo{size: uint32(unsafe.Sizeof(scrollInfo{})), mask: 7, maximum: int32(height - 1), page: uint32(available), position: int32(offset)}
	user32.NewProc("SetScrollInfo").Call(a.sidebar, 1, uintptr(unsafe.Pointer(&info)), 1)
	key := a.profileKey()
	hasPageRange := comboIndex(a.exportScope) == exportPageRanges
	for tab, items := range a.panelItems {
		for _, item := range items {
			show := tab == a.mode
			if item.handle == a.exportRange {
				show = show && hasPageRange
			}
			if item.handle == a.customICC || item.handle == a.customBrowse {
				show = show && key == "custom"
			}
			if item.handle == a.raster {
				show = show && key == "k" && !a.checked(a.outline)
			}
			if show {
				place(item.handle, a.scale(item.x), a.scale(item.y-offset), a.scale(item.width), a.scale(item.height))
				showWindow.Call(item.handle, 5)
			} else {
				showWindow.Call(item.handle, 0)
			}
		}
	}
	invalidate(a.sidebar)
}

func (a *app) revealFocusedControl() {
	if a.panelHidden || a.full {
		return
	}
	focus, _, _ := user32.NewProc("GetFocus").Call()
	_, height := a.panelMetrics()
	offset := a.panelScroll[a.mode]
	for _, item := range a.panelItems[a.mode] {
		if item.handle != focus {
			continue
		}
		if item.y < offset || item.y+item.height > offset+height {
			a.panelScroll[a.mode] = max(0, min(item.y, item.y+item.height-height))
			a.layoutPanel()
		}
		return
	}
}

func sidebarProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	a := window
	switch msg {
	case wmPaint:
		var paint paintInfo
		var bounds rect
		getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
		dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		fillColor(dc, bounds, panelColor)
		user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		return 0
	case wmEraseBkgnd:
		return 1
	case wmDrawItem:
		if a.drawControl(lParam) {
			return 1
		}
	case wmCtlColorEdit, wmCtlColorList, wmCtlColorText:
		return a.colorControl(msg, wParam, lParam)
	case wmCommand:
		a.command(int(wParam&0xffff), int(wParam>>16&0xffff))
		return 0
	case 0x020a: // WM_MOUSEWHEEL
		a.panelScroll[a.mode] -= int(int16(wParam>>16)) * 62 / 120
		a.layoutPanel()
		return 0
	case 0x0115: // WM_VSCROLL
		total, height := a.panelMetrics()
		value := a.panelScroll[a.mode]
		switch wParam & 0xffff {
		case 0:
			value -= 25
		case 1:
			value += 25
		case 2:
			value -= height
		case 3:
			value += height
		case 4, 5:
			info := scrollInfo{size: uint32(unsafe.Sizeof(scrollInfo{})), mask: 0x10}
			user32.NewProc("GetScrollInfo").Call(hwnd, 1, uintptr(unsafe.Pointer(&info)))
			value = int(info.track)
		case 6:
			value = 0
		case 7:
			value = total
		}
		a.panelScroll[a.mode] = value
		a.layoutPanel()
		return 0
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}

func (a *app) minimumSize(pointer uintptr) {
	var info minMaxInfo
	copyDrawItem.Call(uintptr(unsafe.Pointer(&info)), pointer, unsafe.Sizeof(info))
	info.minimumTrack = point{int32(a.scale(850)), int32(a.scale(545))}
	copyDrawItem.Call(pointer, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))
}

func (a *app) toggleFullscreen() {
	getStyle := user32.NewProc("GetWindowLongW")
	setStyle := user32.NewProc("SetWindowLongW")
	position := user32.NewProc("SetWindowPos")
	const styleIndex = ^uintptr(15)
	if a.full {
		a.full = false
		setStyle.Call(a.window, styleIndex, a.savedStyle)
		user32.NewProc("SetWindowPlacement").Call(a.window, uintptr(unsafe.Pointer(&a.placement)))
		position.Call(a.window, 0, 0, 0, 0, 0, 0x37)
	} else {
		monitor, _, _ := user32.NewProc("MonitorFromWindow").Call(a.window, 2)
		info := monitorInfo{size: uint32(unsafe.Sizeof(monitorInfo{}))}
		if ok, _, _ := user32.NewProc("GetMonitorInfoW").Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
			return
		}
		a.placement = windowPlacement{length: uint32(unsafe.Sizeof(windowPlacement{}))}
		user32.NewProc("GetWindowPlacement").Call(a.window, uintptr(unsafe.Pointer(&a.placement)))
		a.savedStyle, _, _ = getStyle.Call(a.window, styleIndex)
		a.full = true
		setStyle.Call(a.window, styleIndex, (a.savedStyle&^0x00c40000)|0x80000000)
		r := info.monitor
		position.Call(a.window, 0, uintptr(r.left), uintptr(r.top), uintptr(r.right-r.left), uintptr(r.bottom-r.top), 0x24)
	}
	a.layout()
	setFocus.Call(a.window)
}

func (a *app) openDrop(handle uintptr) {
	proc := shell32.NewProc("DragQueryFileW")
	defer shell32.NewProc("DragFinish").Call(handle)
	if a.busy {
		return
	}
	length, _, _ := proc.Call(handle, 0, 0, 0)
	if length == 0 {
		return
	}
	buf := make([]uint16, length+1)
	proc.Call(handle, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	a.openDocument(syscall.UTF16ToString(buf), "")
}

func (a *app) openDialog() {
	if a.busy {
		return
	}
	if path := a.fileDialog(pdfFilter, "", "", false); path != "" {
		a.openDocument(path, "")
	}
}

func (a *app) mayReplaceDocument() bool {
	return !a.view.changed() || notice(a.window, "Zahodit neuložené změny stránek?", 0x134) == 6
}

func (a *app) closeDocument() {
	if a.busy || a.view.path == "" || !a.mayReplaceDocument() {
		return
	}
	a.view.resetDocument()
	setText(a.window, appName)
	setText(a.password, "")
	setText(a.signReport, "")
	setText(a.exportReport, "")
	a.reportStatus("")
	a.mode, a.panelHidden = idBrowseTab, true
	a.updateActions()
	a.layout()
}

func (a *app) openDocument(path, password string) {
	if a.busy {
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		a.report("Cestu k dokumentu nelze otevřít.")
		return
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		a.report("Vyberte existující soubor PDF.")
		return
	}
	if !a.mayReplaceDocument() {
		return
	}
	a.mode, a.panelHidden = idBrowseTab, true
	a.loadDocument(path, password, info)
}

func (a *app) loadDocument(path, password string, info os.FileInfo) {
	setText(a.password, password)
	setText(a.window, filepath.Base(path)+" · "+appName)
	setText(a.signReport, "")
	setText(a.exportReport, "")
	chooseCombo(a.exportScope, exportAllPages)
	setText(a.exportRange, "")
	a.view.open(path, password, info)
	a.layout()
}

func (a *app) command(id, notification int) {
	if choice := a.choiceCommands[id]; choice != nil && notification == 0 {
		a.openChoice(choice)
		return
	}
	if a.view.command(id, notification) {
		return
	}
	if notification == 1 {
		switch id {
		case idPreset:
			a.updatePresets()
		case idExportScope:
			a.layoutPanel()
			if comboIndex(a.exportScope) == exportPageRanges {
				setFocus.Call(a.exportRange)
			}
		case idICCProfile:
			if a.profileKey() != "" {
				a.setChecked(a.decrypt, false)
			}
			a.layoutPanel()
		}
		return
	}
	if notification != 0 {
		return
	}
	switch id {
	case idFileTab:
		a.openFileMenu()
	case idBrowseTab, idPagesTab, idProofTab, idExportTab, idSignTab:
		a.activate(id)
	case idOpen:
		a.openDialog()
	case idClosePDF:
		a.closeDocument()
	case idSaveDocument, idSaveDocumentAs:
		a.saveDocument(id == idSaveDocumentAs)
	case idPrintDocument:
		a.printDocument()
	case idShowOutput:
		if !a.busy && a.lastOutput != "" {
			a.openDocument(a.lastOutput, a.lastPassword)
		}
	case idTogglePanel:
		a.panelHidden = !a.panelHidden
		a.layout()
	case idFullscreen:
		a.toggleFullscreen()
	case idMinimize:
		showWindow.Call(a.window, 6)
	case idMaximize:
		a.toggleMaximized()
	case idCloseWindow:
		postMessage.Call(a.window, wmClose, 0, 0)
	case idReload:
		if !a.busy && a.view.path != "" {
			a.openDocument(a.view.path, getText(a.password))
		}
	case idICCBrowse:
		a.chooseInput(a.customICC, "Barevné profily ICC nebo ICM\x00*.icc;*.icm")
	case idCertBrowse:
		a.chooseInput(a.cert, "Osobní certifikát PKCS#12\x00*.p12;*.pfx")
	case idTrustBrowse:
		a.chooseInput(a.trust, "Kořenové certifikáty\x00*.pem;*.cer;*.crt;*.der")
	case idOutline:
		a.setChecked(a.outline, !a.checked(a.outline))
		if a.checked(a.outline) {
			a.setChecked(a.decrypt, false)
		}
		a.layoutPanel()
	case idDecrypt:
		a.setChecked(a.decrypt, !a.checked(a.decrypt))
		if a.checked(a.decrypt) {
			chooseCombo(a.profiles, 0)
			a.setChecked(a.outline, false)
		}
		a.layoutPanel()
	case idVisible, idRaster:
		h := a.visible
		if id == idRaster {
			h = a.raster
		}
		a.setChecked(h, !a.checked(h))
	case idResetPreset:
		a.applyExportOptions(defaults)
		a.updatePresets()
	case idRun, idSign, idVerify:
		a.startAction(id)
	}
}

func (a *app) startAction(id int) {
	if a.busy || !a.view.ready() {
		return
	}
	req := a.view.request("")
	var suffix string
	switch id {
	case idRun:
		req.Action = "transform"
		switch comboIndex(a.exportScope) {
		case exportCurrentPage:
			req.View.Pages = []int{a.view.page + 1}
		case exportPageRanges:
			pages, err := parseExportPages(getText(a.exportRange), len(a.view.order))
			if err != nil {
				a.report(err.Error())
				setFocus.Call(a.exportRange)
				return
			}
			req.View.Pages = pages
		}
		key := a.profileKey()
		req.Outline, req.Decrypt = a.checked(a.outline), a.checked(a.decrypt)
		req.Gray = key == "k"
		req.AllowRaster = req.Gray && !req.Outline && a.checked(a.raster)
		if !req.Gray {
			req.Profile = key
		}
		if req.Profile == "custom" {
			req.CustomICC = strings.Trim(strings.TrimSpace(getText(a.customICC)), "\"")
			if info, err := os.Stat(req.CustomICC); err != nil || !info.Mode().IsRegular() {
				a.report("Vyberte soubor s výstupním profilem CMYK ICC nebo ICM.")
				return
			}
		}
		if key == "" && !req.Outline && !req.Decrypt && len(req.View.Pages) == 0 {
			a.report("Vyberte strany k exportu, úpravu nebo výstupní profil.")
			return
		}
		switch {
		case req.Gray:
			suffix = "-K"
		case key != "":
			suffix = "-" + key
		case req.Decrypt:
			suffix = "-bez-hesla"
		}
		if req.Outline {
			suffix += "-krivky"
		}
		if len(req.View.Pages) > 0 {
			suffix += "-vyber"
		}
	case idSign:
		req.Action, suffix = "sign", "-podepsano"
		req.Certificate = getText(a.cert)
		req.CertPassword, req.Visible = getText(a.certPass), a.checked(a.visible)
		if info, err := os.Stat(req.Certificate); err != nil || !info.Mode().IsRegular() {
			a.report("Vyberte osobní certifikát.")
			return
		}
	case idVerify:
		req.Action, req.Trust = "verify", getText(a.trust)
	}
	if req.Action != "verify" {
		if !a.chooseOutput(&req, suffix) {
			return
		}
		if a.view.changed() {
			a.view.includeEdits(&req)
		}
	}
	a.run(req)
}

func (a *app) updateActions() {
	documentReady := a.view.ready()
	if a.documentTools != documentReady {
		a.documentTools = documentReady
		if !documentReady && (a.mode == idProofTab || a.mode == idSignTab || a.mode == idExportTab) {
			a.activate(idBrowseTab)
		} else {
			a.layout()
		}
	}
	ready := !a.busy && documentReady
	for _, button := range []uintptr{a.runButton, a.signButton, a.verifyButton, a.scanButton} {
		setEnabled(button, ready)
	}
	a.view.updateActions()
}

func (a *app) reportStatus(message string) {
	setText(a.status, strings.SplitN(message, "\n", 2)[0])
}

func (a *app) report(message string) {
	a.reportStatus(message)
	var target uintptr
	switch a.mode {
	case idExportTab:
		target = a.exportReport
	case idSignTab:
		target = a.signReport
	}
	if target != 0 {
		setText(target, multiline(message))
	}
}

func (a *app) run(req request) {
	if req.ReplaceSource {
		if a.beginJob(req, "Ukládám dokument…") {
			a.waitingOverwrite = true
			a.startOverwrite()
		}
		return
	}
	if a.beginJob(req, "Zpracovávám dokument…") {
		a.launchJob(func() response { return runOperation(req) })
	}
}

func (a *app) beginJob(req request, status string) bool {
	if a.busy {
		return false
	}
	if !a.view.sourceUnchanged() {
		a.report("Vstupní PDF se změnilo. Otevřete je znovu.")
		return false
	}
	a.pending, a.busy = req, true
	a.updateActions()
	a.reportStatus(status)
	return true
}

func (a *app) startOverwrite() {
	// Ghostscript may still have the original open. Wait for both readers to
	// finish before replacing the file on Windows.
	if a.waitingOverwrite && !a.view.running && a.view.prefetch.page == 0 {
		a.waitingOverwrite = false
		req := a.pending
		a.launchJob(func() response { return runOperation(req) })
	}
}

func (a *app) launchJob(operation func() response) {
	go func() {
		a.results <- operation()
		postMessage.Call(a.window, wmJobFinished, 0, 0)
	}()
}

func (a *app) finish(result response) {
	req := a.pending
	a.busy = false
	a.pending = request{}
	if result.Status == "signed_pdf" {
		message := result.Message
		question := "Uložit kopii bez původních podpisů?"
		if req.ReplaceSource {
			message = "Původní PDF obsahuje digitální podpis."
			question = "Přepsat původní soubor a odstranit jeho stávající podpisy?"
		}
		if notice(a.window, message+"\n\n"+question, 0x134) == 6 {
			req.ClearSignatures = true
			a.run(req)
			return
		}
		result.Message = "Uložení zrušeno."
	}
	if result.Status == "ok" && req.Output != "" && !req.ReplaceSource {
		a.lastOutput, a.lastPassword = req.Output, ""
		if req.Action == "save_document" ||
			(req.Action == "transform" && !req.Gray && !req.Outline && req.Profile == "" && !req.Decrypt) {
			a.lastPassword = req.PDFPassword
		}
	}
	if result.Status == "ok" && req.Action == "save_document" {
		a.reopenSaved(req)
	}
	if result.Status == "ok" && req.Action == "scan" && result.View != nil {
		a.view.showScan(result.View)
		a.reportStatus("Kontrola tiskových desek dokončena.")
	} else {
		message := result.Message
		if result.Status == "ok" && req.Output != "" {
			message += "\n" + req.Output
		}
		a.reportStatus(message)
		if req.Action == "sign" || req.Action == "verify" {
			setText(a.signReport, multiline(message))
		}
		if req.Action == "transform" {
			setText(a.exportReport, multiline(message))
		}
	}
	if req.Action == "sign" {
		setText(a.certPass, "")
	}
	if result.Status == "error" && (req.Action == "save_document" || req.Action == "print_page") {
		notice(a.window, result.Message, 0x10)
	}
	a.updateActions()
	a.view.schedule()
}

func (a *app) handleInput(msg *winMessage) bool {
	if msg.message == 0x020a { // WM_MOUSEWHEEL
		pos := point{int32(int16(msg.lParam)), int32(int16(msg.lParam >> 16))}
		user32.NewProc("ScreenToClient").Call(a.window, uintptr(unsafe.Pointer(&pos)))
		if contains(a.viewport, pos.x, pos.y) {
			a.view.wheel(float64(int16(msg.wParam>>16))/120, msg.wParam&8 != 0, pos)
			return true
		}
	}
	if msg.message != 0x0100 {
		return false
	}
	ctrl, _, _ := user32.NewProc("GetKeyState").Call(0x11)
	if ctrl&0x8000 != 0 {
		switch msg.wParam {
		case 'O':
			a.openDialog()
			return true
		case 'W':
			a.closeDocument()
			return true
		case 'S':
			shift, _, _ := user32.NewProc("GetKeyState").Call(0x10)
			a.saveDocument(shift&0x8000 != 0)
			return true
		case 'P':
			a.printDocument()
			return true
		}
	}
	switch msg.wParam {
	case 0x7a: // F11
		a.toggleFullscreen()
		return true
	case 0x1b: // Escape
		if a.full {
			a.toggleFullscreen()
			return true
		}
	}
	if msg.window == a.window || a.full {
		return a.view.key(msg.wParam)
	}
	return false
}
