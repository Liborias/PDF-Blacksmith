package main

import (
	_ "embed"
	"encoding/json"
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
	shell32  = syscall.NewLazyDLL("shell32.dll")
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
	shellExecute  = shell32.NewProc("ShellExecuteW")
)

const (
	wsChild       = 0x40000000
	wsVisible     = 0x10000000
	wsTabstop     = 0x00010000
	wsGroup       = 0x00020000
	wsVscroll     = 0x00200000
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

const (
	idRun      = 1
	idInBrowse = 100 + iota
	idEditMode
	idSignMode
	idVerifyMode
	idICCProfile
	idICCBrowse
	idCertBrowse
	idTrustBrowse
	idShowOutput
	idOutline
	idDecrypt
	idRaster
	idVisible
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

type profileChoice struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type app struct {
	window, instance, normalFont, titleFont, labelFont  uintptr
	dpi                                                 int
	mode                                                int
	busy                                                bool
	controls                                            []uintptr
	modeControls                                        map[int][]uintptr
	modes                                               map[int]uintptr
	input, report, outputButton, runButton              uintptr
	outline, decrypt, raster, visible                   uintptr
	profiles, customICC, customBrowse, modeHint         uintptr
	profileChoices                                      []profileChoice
	canvasBrush, panelBrush, inputBrush, logo           uintptr
	options, checkedOptions, panelLabels, panelButtons  map[uintptr]bool
	editPassword, verifyPassword, cert, certPass, trust uintptr
	lastOutput                                          string
	pending                                             request
	results                                             chan response
}

var window *app

func main() {
	runtime.LockOSThread()
	instance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	user32.NewProc("SetProcessDPIAware").Call()
	var dpi uintptr
	getDPI := user32.NewProc("GetDpiForSystem")
	if getDPI.Find() == nil {
		dpi, _, _ = getDPI.Call()
	}
	if dpi == 0 {
		dpi = 96
	}
	screenWidth, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	screenHeight, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	if screenWidth > 100 && screenHeight > 100 {
		dpi = min(dpi, (screenWidth-30)*96/windowWidth, (screenHeight-95)*96/windowHeight)
	}
	window = &app{
		instance: instance, dpi: int(dpi), mode: idEditMode,
		modeControls: make(map[int][]uintptr), modes: make(map[int]uintptr),
		results: make(chan response, 1),
		options: make(map[uintptr]bool), checkedOptions: make(map[uintptr]bool),
		panelLabels: make(map[uintptr]bool), panelButtons: make(map[uintptr]bool),
	}
	if !window.create() {
		notice(0, "Okno aplikace PDF Blaksmith se nepodařilo vytvořit.", 0x10)
		return
	}
	showWindow.Call(window.window, 5)
	setFocus.Call(window.input)
	var msg winMessage
	for {
		result, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) <= 0 {
			break
		}
		handled, _, _ := isDialogMsg.Call(window.window, uintptr(unsafe.Pointer(&msg)))
		if handled == 0 {
			translateMsg.Call(uintptr(unsafe.Pointer(&msg)))
			dispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
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
	height := -a.scale(size)
	font, _, _ := gdi32.NewProc("CreateFontW").Call(
		uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wideString("Segoe UI"))),
	)
	return font
}

func (a *app) create() bool {
	const baseStyle = 0x02ca0000 // Caption, system menu, minimise button and clipped child windows.
	const exStyle = 0x00080000   // Layered top-level window for subtle transparency.
	class := syscall.StringToUTF16Ptr("PDFBlaksmithWindow")
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	icon, _, _ := user32.NewProc("LoadIconW").Call(a.instance, appIconID)
	if icon == 0 {
		icon, _, _ = user32.NewProc("LoadIconW").Call(0, 32512)
	}
	wc := windowClass{
		windowProc: syscall.NewCallback(wndProc),
		instance:   a.instance, icon: icon, cursor: cursor, background: 0,
		className: uintptr(unsafe.Pointer(class)), smallIcon: icon,
	}
	wc.size = uint32(unsafe.Sizeof(wc))
	atom, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return false
	}
	bounds := rect{right: int32(a.scale(windowWidth)), bottom: int32(a.scale(windowHeight))}
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&bounds)), baseStyle, 0, exStyle)
	a.window, _, _ = createWindow.Call(
		exStyle, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(wideString(appName))),
		baseStyle, 0x80000000, 0x80000000, uintptr(bounds.right-bounds.left),
		uintptr(bounds.bottom-bounds.top), 0, 0, a.instance, 0,
	)
	if a.window == 0 {
		return false
	}
	a.normalFont, a.titleFont, a.labelFont = a.font(16, 400), a.font(38, 650), a.font(16, 600)
	a.applyTheme(bounds.right-bounds.left, bounds.bottom-bounds.top)
	a.setupControls()
	return true
}

func (a *app) add(kind, caption string, id, x, y, width, height int, style, exStyle, font uintptr) uintptr {
	handle, _, _ := createWindow.Call(
		exStyle, uintptr(unsafe.Pointer(wideString(kind))), uintptr(unsafe.Pointer(wideString(caption))), style|wsChild|wsVisible,
		uintptr(a.scale(x)), uintptr(a.scale(y)), uintptr(a.scale(width)), uintptr(a.scale(height)),
		a.window, uintptr(id), a.instance, 0,
	)
	sendMessage.Call(handle, wmSetFont, font, 1)
	return handle
}

func (a *app) label(caption string, x, y, width int, strong bool) uintptr {
	font := a.normalFont
	if strong {
		font = a.labelFont
	}
	h := a.add("STATIC", caption, 0, x, y, width, 22, 0, 0, font)
	a.panelLabels[h] = y >= 248 && y < 481
	return h
}

func (a *app) edit(x, y, width int, password bool) uintptr {
	style := uintptr(wsTabstop | esAutoHscroll)
	if password {
		style |= esPassword
	}
	h := a.add("EDIT", "", 0, x, y, width, 31, style, 0, a.normalFont)
	sendMessage.Call(h, 0x00d3, 3, uintptr(a.scale(8))|uintptr(a.scale(8))<<16)
	a.controls = append(a.controls, h)
	return h
}

func (a *app) button(caption string, id, x, y, width int, style uintptr) uintptr {
	option := style&0xf == bsAutoCheck
	height := 35
	if option {
		height = 30
	}
	h := a.add("BUTTON", caption, id, x, y, width, height,
		style&^0xf|bsOwnerDraw|wsTabstop, 0, a.labelFont)
	a.options[h] = option
	a.panelButtons[h] = y >= 248 && y < 481
	a.controls = append(a.controls, h)
	return h
}

func (a *app) addMode(mode int, handles ...uintptr) {
	a.modeControls[mode] = append(a.modeControls[mode], handles...)
}

func (a *app) setupControls() {
	a.label("Vstupní PDF", 32, 120, 710, true)
	a.input = a.edit(32, 145, 566, false)
	a.button("Vybrat PDF…", idInBrowse, 610, 143, 138, 0)

	a.modes[idEditMode] = a.button("Upravit pro tisk", idEditMode, 32, 195, 228, wsGroup)
	a.modes[idSignMode] = a.button("Podepsat", idSignMode, 276, 195, 228, 0)
	a.modes[idVerifyMode] = a.button("Ověřit podpisy", idVerifyMode, 520, 195, 228, 0)

	var catalog []profileChoice
	if err := json.Unmarshal(printCatalog, &catalog); err != nil {
		panic(err)
	}
	a.profileChoices = append([]profileChoice{{"", "Zachovat barvy"}, {"k", "Stupně šedi · pouze deska K"}}, catalog...)
	a.profileChoices = append(a.profileChoices, profileChoice{"custom", "Vlastní výstupní profil ICC"})
	a.profiles = a.add("COMBOBOX", "", idICCProfile, 43, 291, 693, 295,
		wsTabstop|wsGroup|wsVscroll|3, 0, a.normalFont)
	a.controls = append(a.controls, a.profiles)
	sendMessage.Call(a.profiles, 0x0153, ^uintptr(0), uintptr(a.scale(26)))
	for _, choice := range a.profileChoices {
		sendMessage.Call(a.profiles, 0x0143, 0, uintptr(unsafe.Pointer(wideString(choice.Name))))
	}
	sendMessage.Call(a.profiles, 0x014e, 1, 0)
	a.outline = a.button("Převést text na křivky", idOutline, 43, 331, 360, bsAutoCheck)
	a.decrypt = a.button("Pouze odstranit omezení heslem", idDecrypt, 43, 365, 360, bsAutoCheck)
	a.raster = a.button("Povolit rastrovou opravu K (600 dpi)", idRaster, 43, 398, 365, bsAutoCheck)
	a.editPassword = a.edit(420, 365, 316, true)
	a.customICC = a.edit(43, 433, 565, false)
	a.customBrowse = a.button("Vybrat ICC…", idICCBrowse, 620, 431, 116, 0)
	a.modeHint = a.label("", 43, 440, 693, false)
	a.addMode(idEditMode, a.profiles, a.outline, a.decrypt, a.raster, a.editPassword, a.customICC, a.customBrowse, a.modeHint,
		a.label("Barevný výstup", 43, 266, 693, true),
		a.label("Heslo k otevření (pokud je známé)", 420, 341, 316, false),
	)

	a.cert = a.edit(43, 303, 568, false)
	a.certPass = a.edit(43, 373, 568, true)
	a.visible = a.button("Viditelný podpis na první straně u nového podpisového pole", idVisible, 43, 425, 693, bsAutoCheck)
	a.addMode(idSignMode, a.cert, a.certPass, a.visible,
		a.label("Osobní certifikát se soukromým klíčem (.p12 / .pfx)", 43, 275, 690, true),
		a.button("Vybrat…", idCertBrowse, 625, 301, 111, 0),
		a.label("Heslo k souboru certifikátu", 43, 348, 685, false),
	)

	a.verifyPassword = a.edit(43, 303, 568, true)
	a.trust = a.edit(43, 373, 568, false)
	a.addMode(idVerifyMode, a.verifyPassword, a.trust,
		a.label("Stávající heslo k PDF (pokud je třeba)", 43, 275, 693, false),
		a.label("Doplňkový důvěryhodný kořenový certifikát (volitelné)", 43, 348, 693, true),
		a.button("Vybrat…", idTrustBrowse, 625, 371, 111, 0),
		a.label("Ověření probíhá offline s kořenovými certifikáty Windows.", 43, 435, 693, false),
	)

	a.runButton = a.button("Uložit upravené PDF…", idRun, 32, 505, 268, wsGroup)
	a.outputButton = a.button("Otevřít výstupní PDF", idShowOutput, 313, 505, 224, 0)
	enableWindow.Call(a.outputButton, 0)
	a.label("Výsledek", 32, 561, 716, true)
	a.report = a.add("EDIT", "", 0, 32, 587, 716, 96,
		wsTabstop|wsVscroll|esReadonly|0x0004|0x0040, 0, a.normalFont)
	sendMessage.Call(a.report, 0x00d3, 3, uintptr(a.scale(9))|uintptr(a.scale(9))<<16)
	a.setMode(idEditMode)
}

func (a *app) setMode(mode int) {
	a.mode = mode
	buttons := map[int]string{
		idEditMode:   "Uložit upravené PDF…",
		idSignMode:   "Uložit podepsané PDF…",
		idVerifyMode: "Ověřit podpisy",
	}
	for candidate, controls := range a.modeControls {
		var shown uintptr
		if candidate == mode {
			shown = 5
		}
		invalidate(a.modes[candidate])
		for _, control := range controls {
			showWindow.Call(control, shown)
		}
	}
	setText(a.runButton, buttons[mode])
	a.updatePrintControls()
}

func (a *app) profileKey() string {
	index, _, _ := sendMessage.Call(a.profiles, 0x0147, 0, 0)
	if index >= uintptr(len(a.profileChoices)) {
		return ""
	}
	return a.profileChoices[index].Key
}

func (a *app) updatePrintControls() {
	if a.mode != idEditMode {
		return
	}
	key := a.profileKey()
	var raster, custom, hint uintptr
	if key == "k" && !a.checked(a.outline) {
		raster = 5
	}
	if key == "custom" {
		custom = 5
	} else {
		hint = 5
	}
	showWindow.Call(a.raster, raster)
	showWindow.Call(a.customICC, custom)
	showWindow.Call(a.customBrowse, custom)
	showWindow.Call(a.modeHint, hint)
	switch {
	case a.checked(a.decrypt):
		setText(a.modeHint, "Omezení úprav lze odstranit bez hesla vlastníka, pokud se PDF otevře.")
	case key == "k" && raster == 5:
		setText(a.modeHint, "Rastrová oprava zasáhne jen strany s barvou mimo desku K.")
	case key != "" && key != "k" && key != "custom":
		setText(a.modeHint, "Profil ECI se při prvním použití stáhne z oficiálního zdroje.")
	default:
		setText(a.modeHint, "")
	}
}

func setText(handle uintptr, value string) {
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
		return ""
	}
	return syscall.UTF16ToString(buffer)
}

func (a *app) chooseInput(target uintptr, filter string) {
	if selected := a.fileDialog(filter, "", "", false); selected != "" {
		setText(target, selected)
	}
}

func (a *app) openExternal(path string) {
	result, _, _ := shellExecute.Call(a.window, uintptr(unsafe.Pointer(wideString("open"))), uintptr(unsafe.Pointer(wideString(path))), 0, 0, 1)
	if result <= 32 {
		notice(a.window, "Soubor ani odkaz se nepodařilo otevřít.", 0x10)
	}
}

const pdfFilter = "Soubory PDF (*.pdf)\x00*.pdf\x00Všechny soubory\x00*.*"

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		window.paint(hwnd)
		return 0
	case wmCtlColorEdit, wmCtlColorList, wmCtlColorText:
		return window.colorControl(message, wParam, lParam)
	case wmDrawItem:
		if lParam != 0 {
			var item buttonDraw
			copyDrawItem.Call(uintptr(unsafe.Pointer(&item)), lParam, unsafe.Sizeof(item))
			if window.drawButton(&item) {
				return 1
			}
		}
	case wmCommand:
		if int(wParam&0xffff) == idICCProfile && wParam>>16&0xffff == 1 {
			if window.profileKey() != "" {
				window.setChecked(window.decrypt, false)
			}
			window.updatePrintControls()
			return 0
		}
		if wParam>>16&0xffff == 0 && window != nil && !window.busy {
			switch int(wParam & 0xffff) {
			case idEditMode, idSignMode, idVerifyMode:
				window.setMode(int(wParam & 0xffff))
			case idInBrowse:
				window.chooseInput(window.input, pdfFilter)
			case idICCBrowse:
				window.chooseInput(window.customICC, "Barevné profily ICC nebo ICM\x00*.icc;*.icm")
			case idCertBrowse:
				window.chooseInput(window.cert, "Osobní certifikát PKCS#12\x00*.p12;*.pfx")
			case idTrustBrowse:
				window.chooseInput(window.trust, "Kořenové certifikáty\x00*.pem;*.cer;*.crt;*.der\x00Všechny soubory\x00*.*")
			case idShowOutput:
				if window.lastOutput != "" {
					window.openExternal(window.lastOutput)
				}
			case idOutline:
				window.setChecked(window.outline, !window.checked(window.outline))
				if window.checked(window.outline) {
					window.setChecked(window.decrypt, false)
				}
				window.updatePrintControls()
			case idDecrypt:
				window.setChecked(window.decrypt, !window.checked(window.decrypt))
				if window.checked(window.decrypt) {
					sendMessage.Call(window.profiles, 0x014e, 0, 0)
					window.setChecked(window.outline, false)
				}
				window.updatePrintControls()
			case idRaster:
				window.setChecked(window.raster, !window.checked(window.raster))
			case idVisible:
				window.setChecked(window.visible, !window.checked(window.visible))
			case idRun:
				window.start()
			}
		}
	case wmJobFinished:
		window.finish(<-window.results)
		return 0
	case wmClose:
		if window.busy {
			notice(hwnd, "Počkejte, až se dokončí zpracování PDF.", 0x40)
		} else {
			destroyWindow.Call(hwnd)
		}
		return 0
	case wmDestroy:
		window.disposeTheme()
		quitMessage.Call(0)
		return 0
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(message), wParam, lParam)
	return result
}

func pathText(value string) string {
	return strings.Trim(strings.TrimSpace(value), "\"")
}

func (a *app) start() {
	req := request{Input: pathText(getText(a.input))}
	if info, err := os.Stat(req.Input); err != nil || !info.Mode().IsRegular() {
		notice(a.window, "Vyberte existující soubor PDF.", 0x30)
		return
	}
	var suffix string
	switch a.mode {
	case idEditMode:
		req.Action = "transform"
		key := a.profileKey()
		req.Gray = key == "k"
		req.Outline, req.Decrypt = a.checked(a.outline), a.checked(a.decrypt)
		req.AllowRaster = req.Gray && !req.Outline && a.checked(a.raster)
		req.PDFPassword = getText(a.editPassword)
		if !req.Gray {
			req.Profile = key
		}
		if req.Profile == "custom" {
			req.CustomICC = pathText(getText(a.customICC))
			if info, err := os.Stat(req.CustomICC); err != nil || !info.Mode().IsRegular() {
				notice(a.window, "Vyberte soubor s výstupním profilem CMYK ICC nebo ICM.", 0x30)
				return
			}
		}
		if !req.Gray && !req.Outline && !req.Decrypt && req.Profile == "" {
			notice(a.window, "Vyberte alespoň jednu úpravu nebo výstupní profil.", 0x30)
			return
		}
		switch {
		case req.Gray && req.Outline:
			suffix = "-K-krivky"
		case req.Gray:
			suffix = "-K"
		case req.Profile != "":
			suffix = "-" + req.Profile
			if req.Outline {
				suffix += "-krivky"
			}
		case req.Outline:
			suffix = "-krivky"
		default:
			suffix = "-bez-hesla"
		}
	case idSignMode:
		req.Action, suffix = "sign", "-podepsano"
		req.Certificate = pathText(getText(a.cert))
		req.CertPassword, req.Visible = getText(a.certPass), a.checked(a.visible)
		if info, err := os.Stat(req.Certificate); err != nil || !info.Mode().IsRegular() {
			notice(a.window, "Vyberte osobní certifikát (.p12 nebo .pfx).", 0x30)
			return
		}
	case idVerifyMode:
		req.Action = "verify"
		req.PDFPassword, req.Trust = getText(a.verifyPassword), pathText(getText(a.trust))
	}
	if req.Action != "verify" {
		defaultPath := strings.TrimSuffix(req.Input, filepath.Ext(req.Input)) + suffix + ".pdf"
		req.Output = a.fileDialog(pdfFilter, defaultPath, "pdf", true)
		if req.Output == "" {
			return
		}
		_, err := os.Stat(req.Output)
		if err == nil {
			if notice(a.window, "Chcete přepsat tento výstupní soubor?\n\n"+req.Output, 0x134) != 6 {
				return
			}
			req.Overwrite = true
		}
	}
	a.run(req)
}

func (a *app) run(req request) {
	a.pending = req
	a.busy = true
	for _, control := range a.controls {
		enableWindow.Call(control, 0)
	}
	if req.Gray {
		a.updateReport("Převádím PDF a kontroluji tiskové separace. U vícestránkových souborů může kontrola chvíli trvat.")
	} else if req.Profile != "" && req.Profile != "custom" {
		a.updateReport("Převádím do tiskového profilu. Při prvním použití se profil stáhne z webu ECI.")
	} else {
		a.updateReport("Zpracovávám PDF…")
	}
	go func() {
		a.results <- runOperation(req)
		postMessage.Call(a.window, wmJobFinished, 0, 0)
	}()
}

func (a *app) finish(result response) {
	req := a.pending
	a.pending = request{}
	if result.Status == "signed_pdf" {
		if notice(a.window, result.Message+"\n\nChcete tuto kopii uložit bez původních podpisů?", 0x134) == 6 {
			req.ClearSignatures = true
			a.run(req)
			return
		}
		result.Message = "Operace zrušena."
	}
	a.busy = false
	a.lastOutput = ""
	for _, control := range a.controls {
		enableWindow.Call(control, 1)
	}
	setText(a.editPassword, "")
	setText(a.verifyPassword, "")
	setText(a.certPass, "")
	if result.Status == "ok" && req.Output != "" {
		a.lastOutput = req.Output
		result.Message += "\n\n" + req.Output
	} else {
		enableWindow.Call(a.outputButton, 0)
	}
	a.updateReport(result.Message)
	setFocus.Call(a.report)
}

func (a *app) updateReport(value string) {
	setText(a.report, strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\n", "\r\n"))
	sendMessage.Call(a.report, 0x00b1, 0, 0) // Select the start of the report.
	sendMessage.Call(a.report, 0x00b7, 0, 0) // Scroll the caret into view.
}
