package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	registerClass = user32.NewProc("RegisterClassExW")
	createWindow  = user32.NewProc("CreateWindowExW")
	defWindowProc = user32.NewProc("DefWindowProcW")
	sendMessage   = user32.NewProc("SendMessageW")
	postMessage   = user32.NewProc("PostMessageW")
	showWindow    = user32.NewProc("ShowWindow")
	enableWindow  = user32.NewProc("EnableWindow")
	setWindowText = user32.NewProc("SetWindowTextW")
	getTextLen    = user32.NewProc("GetWindowTextLengthW")
	getWindowText = user32.NewProc("GetWindowTextW")
	setFocus      = user32.NewProc("SetFocus")
	messageBox    = user32.NewProc("MessageBoxW")
	quitMessage   = user32.NewProc("PostQuitMessage")
	destroyWindow = user32.NewProc("DestroyWindow")
	getMessage    = user32.NewProc("GetMessageW")
	translateMsg  = user32.NewProc("TranslateMessage")
	dispatchMsg   = user32.NewProc("DispatchMessageW")
	isDialogMsg   = user32.NewProc("IsDialogMessageW")
	chooseOpen    = comdlg32.NewProc("GetOpenFileNameW")
	chooseSave    = comdlg32.NewProc("GetSaveFileNameW")
	dialogError   = comdlg32.NewProc("CommDlgExtendedError")
)

const (
	wsChild       = 0x40000000
	wsVisible     = 0x10000000
	wsTabstop     = 0x00010000
	wsGroup       = 0x00020000
	esPassword    = 0x0020
	esAutoHscroll = 0x0080
	esReadonly    = 0x0800
	wsExEdge      = 0x0200
	bsAutoCheck   = 0x0003
	wmCommand     = 0x0111
	wmClose       = 0x0010
	wmDestroy     = 0x0002
	wmSetFont     = 0x0030
	wmJobFinished = 0x8001
)

type rect struct{ left, top, right, bottom int32 }

type windowClass struct {
	size, style                                  uint32
	windowProc                                   uintptr
	classExtra, windowExtra                      int32
	instance, icon, cursor, background, menuName uintptr
	className, smallIcon                         uintptr
}

type winMessage struct {
	window         uintptr
	message        uint32
	wParam, lParam uintptr
	time           uint32
	x, y           int32
	private        uint32
}

type openFileName struct {
	size                                   uint32
	owner, instance, filter, customFilter  uintptr
	maxCustomFilter, filterIndex           uint32
	file                                   uintptr
	maxFile                                uint32
	fileTitle                              uintptr
	maxFileTitle                           uint32
	initialDir, title                      uintptr
	flags                                  uint32
	fileOffset, fileExtension              uint16
	defExt, customData, hook, templateName uintptr
	reserved                               uintptr
	reservedFlags, flagsEx                 uint32
}

//go:embed profiles.json
var printCatalog []byte

//go:embed export_presets.json
var exportCatalog []byte

type profileChoice struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type app struct {
	window, instance, normalFont, titleFont, labelFont uintptr
	dpi, width, height                                 int
	toolsWidth, preferredToolsWidth, settingsWidth      int
	tool, mode                                         int
	sidebar, thumbnailSidebar, bookmarkSidebar, controlParent uintptr
	detailsArea, thumbnailArea, bookmarkArea             rect
	panelItems                                         map[int][]panelItem
	panelScroll                                        map[int]int
	sideScroll                                         *panelScrollbar
	scrollbars                                         map[uintptr]*panelScrollbar
	scrollPanes                                        map[uintptr]*scrollPane
	addingTab, addingSection                           int
	propertySections                                  []propertySection
	propertyLayoutQueued                              bool
	toolButtons, detailButtons                         map[int]uintptr
	documentTools                                      bool
	headerBrand, windowButtonStart                     int
	tabBounds                                          rect
	leftDock, rightDock                                 dockVisibility
	full                                               bool
	placement                                          windowPlacement
	savedStyle                                         uintptr
	viewport                                           rect
	canvasBrush, panelBrush, inputBrush, logo          uintptr
	brand                                              *printBitmap
	windowButtons                                      [3]uintptr
	historyButtons                                     [2]uintptr
	commandBar                                        commandBar
	hoveredNavigation                                  uintptr
	tabs                                               []*documentTab
	activeTab                                          *documentTab
	tabList                                             uintptr
	closing                                             *documentClose
	workDir                                             string
	retiredDrafts                                       []string
	choices                                            map[uintptr]*choiceControl
	choiceCommands                                     map[int]*choiceControl
	popup                                              []popupItem
	popupWidth                                         int
	fileMenu                                           fileMenu
	options, checkedOptions, panelLabels, panelButtons map[uintptr]bool
	readOnlyFields, primaryButtons                     map[uintptr]bool
	view                                               *documentView
	toolsToggle, detailsToggle, sidebarTitle, status   uintptr
	thumbnailsToggle, bookmarksToggle, railFont        uintptr
	leftPanelsToggle, rightPanelsToggle                 uintptr
	recent                                             recentPanel
	exportScope, exportRange                           uintptr
	splitScope, splitRange, splitButton, splitReport   uintptr
	exportFormat, exportWarning, exportColorLabel, exportColorHelp uintptr
	profiles, pressProfiles, customICC, customBrowse   uintptr
	grayProfileCombo, grayCustomICC, grayBrowse, kOutputProfile uintptr
	proofApply, grayProfileLabel, kOutputLabel uintptr
	outline, decrypt, raster, visible                  uintptr
	password, cert, certPass, trust                    uintptr
	exportReport, signReport                           uintptr
	runButton, scanButton, verifyButton, signButton    uintptr
	profileChoices                                     []profileChoice
	exportPresets                                      []exportPreset
	exportOptionControls                               [6]uintptr
	exportManualStart                                  int
	export                                             exportOptions
	lastOutput, lastPassword                           string
	merge                                              mergePanel
	comparison                                         documentComparison
	ocr                                                ocrPanel
	textTool textPanel
	objects  objectEditor
	thumbnailScopes map[uintptr]bool
	busy                                               bool
	waitingOverwrite                                   bool
	cancelTextJob                                      context.CancelFunc
	closeAfterTextJob                                  bool
	pending                                            request
	results                                            chan response
}

var window *app

func main() {
	runtime.LockOSThread()
	instance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	user32.NewProc("SetProcessDPIAware").Call()
	dpi := uintptr(96)
	getDPI := user32.NewProc("GetDpiForSystem")
	if getDPI.Find() == nil {
		if value, _, _ := getDPI.Call(); value != 0 {
			dpi = value
		}
	}
	screenWidth, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	screenHeight, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	if screenWidth > 100 && screenHeight > 100 {
		dpi = min(dpi, (screenWidth-24)*96/minimumWorkspaceWidth, (screenHeight-72)*96/550)
	}
	window = &app{
		thumbnailScopes: make(map[uintptr]bool), instance: instance, dpi: int(dpi), tool: idFileTab, mode: idFileDetails,
		rightDock: dockVisibility{auxiliaryCollapsed: true},
		panelItems: make(map[int][]panelItem), panelScroll: make(map[int]int),
		scrollbars: make(map[uintptr]*panelScrollbar), scrollPanes: make(map[uintptr]*scrollPane),
		toolButtons: make(map[int]uintptr), detailButtons: make(map[int]uintptr), results: make(chan response, 1),
		options: make(map[uintptr]bool), checkedOptions: make(map[uintptr]bool),
		panelLabels: make(map[uintptr]bool), panelButtons: make(map[uintptr]bool),
		primaryButtons: make(map[uintptr]bool), readOnlyFields: make(map[uintptr]bool),
		export:  loadExportOptions(),
		choices: make(map[uintptr]*choiceControl), choiceCommands: make(map[int]*choiceControl),
	}
	if !window.create() {
		notice(0, "Okno aplikace PDF Blacksmith se nepodařilo vytvořit.", 0x10)
		return
	}
	defer window.cleanupDrafts()
	defer window.view.closeReaders()
	defer window.comparison.view.closeReaders()
	defer func() {
		if task := window.view.search.task; task != nil {
			task.cancel()
			<-task.done
		}
	}()
	showWindow.Call(window.window, 3) // SW_SHOWMAXIMIZED retains the normal window placement.
	setFocus.Call(window.toolButtons[idFileTab])
	for _, path := range os.Args[1:] {
		window.openDocument(path, "")
	}
	var msg winMessage
	for {
		result, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) <= 0 {
			break
		}
		if window.handleInput(&msg) {
			continue
		}
		handled, _, _ := isDialogMsg.Call(window.window, uintptr(unsafe.Pointer(&msg)))
		if handled == 0 {
			translateMsg.Call(uintptr(unsafe.Pointer(&msg)))
			dispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
		} else if msg.message == 0x0100 && msg.wParam == 9 {
			window.revealFocusedControl()
		}
	}
}

func wideString(s string) *uint16 {
	return syscall.StringToUTF16Ptr(strings.ReplaceAll(s, "\x00", ""))
}

func (a *app) scale(value int) int {
	return value * a.dpi / 96
}

func (a *app) font(size, weight int) uintptr {
	return a.orientedFont(size, weight, 0)
}

func (a *app) orientedFont(size, weight, angle int) uintptr {
	height := -a.scale(size)
	font, _, _ := gdi32.NewProc("CreateFontW").Call(
		uintptr(height), 0, uintptr(angle), uintptr(angle), uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wideString("Segoe UI"))),
	)
	return font
}

var (
	shell32       = syscall.NewLazyDLL("shell32.dll")
	getClientRect = user32.NewProc("GetClientRect")
	moveWindow    = user32.NewProc("MoveWindow")
)

func (a *app) create() bool {
	const style = 0x02cf0000 // Resize frame, caption, minimize/maximize, clip children.
	name := wideString("PDFBlaksmithWindow")
	icon, _, _ := user32.NewProc("LoadIconW").Call(a.instance, appIconID)
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	class := windowClass{style: 8, instance: a.instance, icon: icon, smallIcon: icon, cursor: cursor,
		className: uintptr(unsafe.Pointer(name)), windowProc: syscall.NewCallback(wndProc)}
	class.size = uint32(unsafe.Sizeof(class))
	if ok, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class))); ok == 0 {
		return false
	}
	width, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	height, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	bounds := rect{right: int32(min(a.scale(1200), max(a.scale(850), int(width)-55))),
		bottom: int32(min(a.scale(790), max(a.scale(550), int(height)-110)))}
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&bounds)), style, 0, 0)
	a.window, _, _ = createWindow.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(wideString(appName))),
		style, 0x80000000, 0x80000000, uintptr(bounds.right-bounds.left), uintptr(bounds.bottom-bounds.top), 0, 0, a.instance, 0)
	if a.window == 0 {
		return false
	}
	a.normalFont, a.titleFont, a.labelFont = a.font(15, 400), a.font(20, 650), a.font(15, 600)
	a.applyTheme()
	if !a.setupWorkspace() {
		return false
	}
	a.createInlineTextEditor()
	shell32.NewProc("DragAcceptFiles").Call(a.window, 1)
	a.layout()
	return true
}

func (a *app) add(kind, caption string, id, x, y, width, height int, style, exStyle, font uintptr) uintptr {
	parent := a.controlParent
	if parent == 0 {
		parent = a.window
	}
	handle, _, _ := createWindow.Call(
		exStyle, uintptr(unsafe.Pointer(wideString(kind))), uintptr(unsafe.Pointer(wideString(caption))), style|wsChild|wsVisible,
		uintptr(a.scale(x)), uintptr(a.scale(y)), uintptr(a.scale(width)), uintptr(a.scale(height)),
		parent, uintptr(id), a.instance, 0,
	)
	sendMessage.Call(handle, wmSetFont, font, 1)
	if a.addingTab != 0 {
		a.panelItems[a.addingTab] = append(a.panelItems[a.addingTab], panelItem{handle, x, y, width, height, a.addingSection})
		a.panelLabels[handle], a.panelButtons[handle] = true, true
	}
	return handle
}

func (a *app) label(caption string, x, y, width int, strong bool) uintptr {
	font := a.normalFont
	if strong {
		font = a.labelFont
	}
	return a.add("STATIC", caption, 0, x, y, width, 22, 0, 0, font)
}

func (a *app) edit(x, y, width int, password bool) uintptr {
	style := uintptr(wsTabstop | esAutoHscroll)
	if password {
		style |= esPassword
	}
	h := a.add("EDIT", "", 0, x, y, width, 28, style, 0, a.normalFont)
	sendMessage.Call(h, 0x00d3, 3, uintptr(a.scale(6))|uintptr(a.scale(6))<<16)
	return h
}

func (a *app) button(caption string, id, x, y, width int, style uintptr) uintptr {
	option := style&0xf == bsAutoCheck
	height := 32
	if option {
		height = 28
	}
	h := a.add("BUTTON", caption, id, x, y, width, height,
		style&^0xf|bsOwnerDraw|wsTabstop, 0, a.labelFont)
	a.options[h] = option
	return h
}

func (a *app) readout(x, y, width, height int) uintptr {
	h := a.add("EDIT", "", 0, x, y, width, height,
		wsTabstop|0x0044|esReadonly, 0, a.normalFont)
	a.readOnlyFields[h] = true
	sendMessage.Call(h, 0x00d3, 3, uintptr(a.scale(5))<<16)
	a.addScrollPane(h, a.textLineHeight(h), false)
	return h
}

func place(h uintptr, x, y, width, height int) {
	if h != 0 {
		moveWindow.Call(h, uintptr(x), uintptr(y), uintptr(max(1, width)), uintptr(max(1, height)), 1)
	}
}

func setEnabled(handle uintptr, enabled bool) {
	var value uintptr
	if enabled {
		value = 1
	}
	enableWindow.Call(handle, value)
}

func multiline(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\n", "\r\n")
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	a := window
	if a.view != nil && a.pointer(msg, lParam) {
		return 0
	}
	switch msg {
	case 0x0083: // WM_NCCALCSIZE
		a.nonClientSize(hwnd, lParam)
		return 0
	case 0x0084: // WM_NCHITTEST
		return a.hitTest(hwnd, lParam)
	case 0x0085: // WM_NCPAINT
		return 0
	case 0x0086: // WM_NCACTIVATE
		return 1
	case 0x0006: // WM_ACTIVATE
		if wParam&0xffff == 0 {
			a.hideFileMenu(false)
			a.cancelPageDrag()
			a.cancelObjectGesture()
			a.view.cancelMiddleSelection()
			a.comparison.view.cancelMiddleSelection()
		}
	case 0x002c: // WM_MEASUREITEM
		if a.measurePopup(lParam) {
			return 1
		}
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		if a.view != nil {
			a.paint(hwnd)
			return 0
		}
	case 0x0005: // WM_SIZE
		if a.view != nil && wParam != 1 {
			a.layout()
		}
		return 0
	case 0x0024: // WM_GETMINMAXINFO
		if lParam != 0 {
			a.minimumSize(lParam)
			return 0
		}
	case wmCtlColorEdit, wmCtlColorList, wmCtlColorText:
		return a.colorControl(msg, wParam, lParam)
	case wmDrawItem:
		if a.drawControl(lParam) {
			return 1
		}
	case wmCommand:
		if a.view != nil {
			if wParam>>16&0xffff == 0x0300 { // EN_CHANGE
				a.inlineControlChange(lParam)
				if scope, field := a.scopeControls(); field != 0 && field == lParam {
					delete(a.thumbnailScopes, scope)
				}
			}
			if wParam>>16&0xffff == 0 && a.tabCommand(int(wParam&0xffff), lParam) {
				return 0
			}
			a.command(int(wParam&0xffff), int(wParam>>16&0xffff))
			return 0
		}
	case 0x004e: // WM_NOTIFY
		if lParam != 0 && a.textTool.inline.handle != 0 {
			notice := (*struct {
				from, id uintptr
				code uint32
			})(*(*unsafe.Pointer)(unsafe.Pointer(&lParam)))
			if notice.from == a.textTool.inline.handle && notice.code == 0x0702 {
				a.syncInlineControls()
				invalidate(a.textTool.inline.handle)
				return 0
			}
		}
	case wmInlineCommit:
		a.commitInlineOnBlur(wParam)
		return 0
	case wmInlineLayout:
		a.finishInlineLayout(wParam)
		return 0
	case wmPropertyReadoutLayout:
		a.propertyLayoutQueued = false
		a.layoutPanel()
		return 0
	case wmJobFinished:
		a.finish(<-a.results)
		return 0
	case wmReaderFinished:
		v := a.view
		if wParam != 0 {
			v = a.comparison.view
		}
		v.finished(<-v.jobs)
		a.pruneDrafts()
		return 0
	case wmSearchUpdate:
		a.view.receiveSearch(<-a.view.search.events)
		a.pruneDrafts()
		return 0
	case wmThumbnailUpdate:
		v := a.view
		if wParam != 0 {
			v = a.comparison.view
		}
		v.schedule()
		return 0
	case 0x0113: // WM_TIMER
		if wParam == readerZoomTimer && a.view != nil {
			a.view.finishZoom()
			return 0
		}
		if wParam == readerComparisonZoomTimer && a.comparison.view != nil {
			a.comparison.view.finishZoom()
			return 0
		}
		if wParam == readerMemoryTimer && a.view.path != "" {
			a.view.refreshCache()
			if a.comparisonVisible() {
				a.comparison.view.refreshCache()
			}
			return 0
		}
	case 0x0233: // WM_DROPFILES
		a.openDrop(wParam)
		return 0
	case wmClose:
		a.closeWindow()
		return 0
	case wmDestroy:
		a.cancelPageDrag()
		a.cancelObjectGesture()
		a.clearInlineFrames()
		if a.view != nil {
			a.view.canvas.close()
		}
		if a.comparison.view != nil {
			a.comparison.view.canvas.close()
		}
		a.disposeTheme()
		quitMessage.Call(0)
		return 0
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}

func (a *app) drawControl(param uintptr) bool {
	if param == 0 {
		return false
	}
	var item buttonDraw
	copyDrawItem.Call(uintptr(unsafe.Pointer(&item)), param, unsafe.Sizeof(item))
	return a.drawPopup(&item) || a.drawFileMenuButton(&item) || a.drawButton(&item) || a.drawMergeRow(&item) || a.drawRecentRow(&item) ||
		(a.view != nil && (a.view.drawSearchRow(&item) || a.view.drawListRow(&item))) ||
		(a.comparison.view != nil && a.comparison.view.drawListRow(&item))
}

func setText(handle uintptr, value string) {
	if window != nil && window.readOnlyFields[handle] {
		old := getText(handle)
		if old == value {
			return
		}
		if (old == "") != (value == "") && !window.propertyLayoutQueued {
			result, _, _ := postMessage.Call(window.window, wmPropertyReadoutLayout, 0, 0)
			window.propertyLayoutQueued = result != 0
		}
	}
	setWindowText.Call(handle, uintptr(unsafe.Pointer(wideString(value))))
}

func getText(handle uintptr) string {
	length, _, _ := getTextLen.Call(handle)
	buffer := make([]uint16, length+1)
	getWindowText.Call(handle, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return syscall.UTF16ToString(buffer)
}

func (a *app) checked(handle uintptr) bool {
	return a.checkedOptions[handle]
}

func (a *app) setChecked(handle uintptr, checked bool) {
	a.checkedOptions[handle] = checked
	invalidate(handle)
}

func notice(owner uintptr, text string, flags uintptr) uintptr {
	code, _, _ := messageBox.Call(owner, uintptr(unsafe.Pointer(wideString(text))), uintptr(unsafe.Pointer(wideString(appName))), flags)
	return code
}

func (a *app) fileDialog(filter, suggestion, ext string, save bool) string {
	paths := a.fileDialogPaths(filter, suggestion, ext, save, false)
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (a *app) fileDialogPaths(filter, suggestion, ext string, save, multiple bool) []string {
	buffer := make([]uint16, 32768)
	copy(buffer, utf16.Encode([]rune(suggestion)))
	filterWords := utf16.Encode([]rune(filter + "\x00\x00"))
	dialog := openFileName{
		owner: a.window, filter: uintptr(unsafe.Pointer(&filterWords[0])), filterIndex: 1,
		file: uintptr(unsafe.Pointer(&buffer[0])), maxFile: uint32(len(buffer)),
		flags: 0x80000 | 0x00000800 | 0x00000008,
	}
	dialog.size = uint32(unsafe.Sizeof(dialog))
	proc := chooseOpen
	extension := syscall.StringToUTF16Ptr(ext)
	if save {
		proc = chooseSave
		dialog.defExt = uintptr(unsafe.Pointer(extension))
	} else {
		dialog.flags |= 0x00001000
		if multiple {
			dialog.flags |= 0x00000200 // OFN_ALLOWMULTISELECT
		}
	}
	ok, _, _ := proc.Call(uintptr(unsafe.Pointer(&dialog)))
	runtime.KeepAlive(buffer)
	runtime.KeepAlive(filterWords)
	runtime.KeepAlive(extension)
	if ok == 0 {
		code, _, _ := dialogError.Call()
		if code != 0 {
			notice(a.window, fmt.Sprintf("Výběr souboru se nezdařil (%d).", code), 0x10)
		}
		return nil
	}
	first := syscall.UTF16ToString(buffer)
	if !multiple {
		return []string{first}
	}
	remaining := buffer[len(utf16.Encode([]rune(first)))+1:]
	if len(remaining) == 0 || remaining[0] == 0 {
		return []string{first}
	}
	var paths []string
	for len(remaining) > 0 && remaining[0] != 0 {
		name := syscall.UTF16ToString(remaining)
		paths = append(paths, filepath.Join(first, name))
		remaining = remaining[len(utf16.Encode([]rune(name)))+1:]
	}
	return paths
}

func (a *app) chooseInput(target uintptr, filter string) {
	if selected := a.fileDialog(filter, "", "", false); selected != "" {
		setText(target, selected)
	}
}

const pdfFilter = "Soubory PDF (*.pdf)\x00*.pdf\x00Všechny soubory\x00*.*"

func (a *app) chooseOutput(req *request, suffix string, otherInputs ...string) bool {
	base := req.Input
	if base == a.view.draft {
		base = a.view.path
	}
	defaultPath := strings.TrimSuffix(base, filepath.Ext(base)) + suffix + ".pdf"
	req.Output = a.fileDialog(pdfFilter, defaultPath, "pdf", true)
	if req.Output == "" {
		return false
	}
	info, err := os.Stat(req.Output)
	if tab := a.openedFile(req.Output, info); tab != nil {
		if tab != a.activeTab {
			notice(a.window, "Cílové PDF je otevřené v jiné kartě. Vyberte jiné umístění.", 0x30)
			return false
		}
		if a.view.draft != "" {
			notice(a.window, "Chcete-li přepsat původní PDF, použijte Soubor → Uložit.", 0x30)
			return false
		}
	}
	if err == nil {
		for _, input := range append([]string{req.Input}, otherInputs...) {
			if original, statErr := os.Stat(input); statErr == nil && os.SameFile(info, original) {
				message := "Vyberte jiný výstupní soubor. Pro uložení úprav do původního PDF použijte Soubor → Uložit."
				if req.Action == "merge" {
					message = "Jako výstup vyberte jiný soubor než vstupní PDF."
				}
				notice(a.window, message, 0x30)
				return false
			}
		}
		if notice(a.window, "Chcete přepsat tento výstupní soubor?\n\n"+req.Output, 0x134) != 6 {
			return false
		}
		req.Overwrite = true
	}
	return true
}
