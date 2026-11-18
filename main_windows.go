package main

import (
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
	bsRadio       = 0x0009
	wmCommand     = 0x0111
	wmClose       = 0x0010
	wmDestroy     = 0x0002
	wmSetFont     = 0x0030
	wmJobFinished = 0x8001
	bmGetCheck    = 0x00f0
	bmSetCheck    = 0x00f1
)

const (
	idRun      = 1
	idInBrowse = 100 + iota
	idEditMode
	idSignMode
	idVerifyMode
	idGSBrowse
	idGSDownload
	idCertBrowse
	idTrustBrowse
	idShowOutput
	idGray
	idOutline
	idDecrypt
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

type app struct {
	window, instance, normalFont, titleFont, labelFont  uintptr
	dpi                                                 int
	mode                                                int
	busy                                                bool
	controls                                            []uintptr
	modeControls                                        map[int][]uintptr
	modes                                               map[int]uintptr
	input, report, outputButton, runButton              uintptr
	gray, outline, decrypt, gs, visible                 uintptr
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
	dpi, _, _ := user32.NewProc("GetDpiForSystem").Call()
	if dpi == 0 {
		dpi = 96
	}
	window = &app{
		instance: instance, dpi: int(dpi), mode: idEditMode,
		modeControls: make(map[int][]uintptr), modes: make(map[int]uintptr),
		results: make(chan response, 1),
	}
	if !window.create() {
		notice(0, "Okno aplikace PDF Tisk se nepodařilo vytvořit.", 0x10)
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
	const baseStyle = 0x00ca0000 // Caption, system menu and minimise button.
	class := syscall.StringToUTF16Ptr("PDFTiskWindow")
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	icon, _, _ := user32.NewProc("LoadIconW").Call(a.instance, 1)
	if icon == 0 {
		icon, _, _ = user32.NewProc("LoadIconW").Call(0, 32512)
	}
	wc := windowClass{
		windowProc: syscall.NewCallback(wndProc),
		instance:   a.instance, icon: icon, cursor: cursor, background: 16,
		className: uintptr(unsafe.Pointer(class)), smallIcon: icon,
	}
	wc.size = uint32(unsafe.Sizeof(wc))
	atom, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return false
	}
	bounds := rect{right: int32(a.scale(730)), bottom: int32(a.scale(606))}
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&bounds)), baseStyle, 0, 0)
	a.window, _, _ = createWindow.Call(
		0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(wideString("PDF Tisk"))),
		baseStyle, 0x80000000, 0x80000000, uintptr(bounds.right-bounds.left),
		uintptr(bounds.bottom-bounds.top), 0, 0, a.instance, 0,
	)
	if a.window == 0 {
		return false
	}
	a.normalFont, a.titleFont, a.labelFont = a.font(15, 400), a.font(33, 650), a.font(15, 600)
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
	return a.add("STATIC", caption, 0, x, y, width, 21, 0, 0, font)
}

func (a *app) edit(x, y, width int, password bool) uintptr {
	style := uintptr(wsTabstop | esAutoHscroll)
	if password {
		style |= esPassword
	}
	h := a.add("EDIT", "", 0, x, y, width, 26, style, wsExEdge, a.normalFont)
	a.controls = append(a.controls, h)
	return h
}

func (a *app) button(caption string, id, x, y, width int, style uintptr) uintptr {
	h := a.add("BUTTON", caption, id, x, y, width, 27, style|wsTabstop, 0, a.normalFont)
	a.controls = append(a.controls, h)
	return h
}

func (a *app) addMode(mode int, handles ...uintptr) {
	a.modeControls[mode] = append(a.modeControls[mode], handles...)
}

func (a *app) setupControls() {
	a.add("STATIC", "PDF Tisk", 0, 24, 17, 682, 42, 0, 0, a.titleFont)
	a.label("Tiskové separace, úprava PDF a digitální podpisy", 25, 61, 681, false)
	a.label("Vstupní PDF", 25, 96, 680, true)
	a.input = a.edit(24, 118, 563, false)
	a.button("Vybrat…", idInBrowse, 598, 117, 108, 0)
	a.modes[idEditMode] = a.button("Upravit pro tisk", idEditMode, 25, 159, 207, bsRadio|wsGroup)
	a.modes[idSignMode] = a.button("Podepsat", idSignMode, 235, 159, 207, bsRadio)
	a.modes[idVerifyMode] = a.button("Ověřit podpisy", idVerifyMode, 445, 159, 261, bsRadio)
	a.add("BUTTON", "", 0, 24, 188, 682, 194, 7|wsGroup, 0, a.normalFont)

	a.gray = a.button("Převést do šedi a ověřit desku K", idGray, 42, 205, 340, bsAutoCheck|wsGroup)
	a.outline = a.button("Převést text na křivky", idOutline, 42, 236, 340, bsAutoCheck)
	a.decrypt = a.button("Pouze uložit kopii bez hesla", idDecrypt, 42, 267, 340, bsAutoCheck)
	a.editPassword = a.edit(397, 234, 290, true)
	a.gs = a.edit(42, 334, 438, false)
	a.addMode(idEditMode, a.gray, a.outline, a.decrypt, a.editPassword, a.gs,
		a.label("Stávající heslo k PDF (pokud je třeba)", 397, 211, 288, false),
		a.label("Ghostscript pro převod do šedi a na křivky", 43, 310, 638, true),
		a.button("Vybrat…", idGSBrowse, 490, 333, 94, 0),
		a.button("Stáhnout…", idGSDownload, 593, 333, 94, 0),
	)
	setText(a.gs, ghostscriptPath())
	sendMessage.Call(a.gray, bmSetCheck, 1, 0)

	a.cert = a.edit(42, 231, 540, false)
	a.certPass = a.edit(42, 290, 540, true)
	a.visible = a.button("Viditelný podpis na první straně, pokud PDF nemá vlastní podpisové pole", 0, 42, 332, 644, bsAutoCheck)
	a.addMode(idSignMode, a.cert, a.certPass, a.visible,
		a.label("Osobní certifikát se soukromým klíčem (.p12 / .pfx)", 43, 207, 638, true),
		a.button("Vybrat…", idCertBrowse, 593, 230, 94, 0),
		a.label("Heslo k souboru certifikátu", 43, 266, 638, false),
	)

	a.verifyPassword = a.edit(42, 231, 540, true)
	a.trust = a.edit(42, 291, 540, false)
	a.addMode(idVerifyMode, a.verifyPassword, a.trust,
		a.label("Stávající heslo k PDF (pokud je třeba)", 43, 207, 638, false),
		a.label("Doplňkový důvěryhodný kořenový certifikát (volitelné)", 43, 267, 638, true),
		a.button("Vybrat…", idTrustBrowse, 593, 290, 94, 0),
		a.label("Ověření probíhá offline s kořenovými certifikáty Windows.", 43, 337, 638, false),
	)

	a.runButton = a.button("Uložit upravené PDF…", idRun, 24, 399, 225, 1|wsGroup)
	a.outputButton = a.button("Otevřít výstupní PDF", idShowOutput, 260, 399, 192, 0)
	enableWindow.Call(a.outputButton, 0)
	a.label("Výsledek", 25, 447, 680, true)
	a.report = a.add("EDIT", "", 0, 24, 470, 682, 112,
		wsTabstop|wsVscroll|esReadonly|0x0004|0x0040, wsExEdge, a.normalFont)
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
		var shown, checked uintptr
		if candidate == mode {
			shown, checked = 5, 1
		}
		sendMessage.Call(a.modes[candidate], bmSetCheck, checked, 0)
		for _, control := range controls {
			showWindow.Call(control, shown)
		}
	}
	setText(a.runButton, buttons[mode])
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
	checked, _, _ := sendMessage.Call(handle, bmGetCheck, 0, 0)
	return checked == 1
}

func notice(owner uintptr, text string, flags uintptr) uintptr {
	code, _, _ := messageBox.Call(owner, uintptr(unsafe.Pointer(wideString(text))), uintptr(unsafe.Pointer(wideString("PDF Tisk"))), flags)
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
	case wmCommand:
		if wParam>>16&0xffff == 0 && window != nil && !window.busy {
			switch int(wParam & 0xffff) {
			case idEditMode, idSignMode, idVerifyMode:
				window.setMode(int(wParam & 0xffff))
			case idInBrowse:
				window.chooseInput(window.input, pdfFilter)
			case idGSBrowse:
				window.chooseInput(window.gs, "Konzolový Ghostscript\x00gswin64c.exe;gswin32c.exe")
			case idGSDownload:
				window.openExternal("https://www.ghostscript.com/releases/gsdnld.html")
			case idCertBrowse:
				window.chooseInput(window.cert, "Osobní certifikát PKCS#12\x00*.p12;*.pfx")
			case idTrustBrowse:
				window.chooseInput(window.trust, "Kořenové certifikáty\x00*.pem;*.cer;*.crt;*.der\x00Všechny soubory\x00*.*")
			case idShowOutput:
				if window.lastOutput != "" {
					window.openExternal(window.lastOutput)
				}
			case idGray, idOutline:
				if window.checked(window.gray) || window.checked(window.outline) {
					sendMessage.Call(window.decrypt, bmSetCheck, 0, 0)
				}
			case idDecrypt:
				if window.checked(window.decrypt) {
					sendMessage.Call(window.gray, bmSetCheck, 0, 0)
					sendMessage.Call(window.outline, bmSetCheck, 0, 0)
				}
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
		for _, font := range []uintptr{window.normalFont, window.titleFont, window.labelFont} {
			gdi32.NewProc("DeleteObject").Call(font)
		}
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
		req.Gray, req.Outline, req.Decrypt = a.checked(a.gray), a.checked(a.outline), a.checked(a.decrypt)
		req.PDFPassword = getText(a.editPassword)
		req.Ghostscript = pathText(getText(a.gs))
		if !req.Gray && !req.Outline && !req.Decrypt {
			notice(a.window, "Vyberte alespoň jednu úpravu.", 0x30)
			return
		}
		if req.Gray || req.Outline {
			req.Ghostscript, _ = filepath.Abs(req.Ghostscript)
			base := strings.ToLower(filepath.Base(req.Ghostscript))
			info, err := os.Stat(req.Ghostscript)
			if err != nil || !info.Mode().IsRegular() || (base != "gswin64c.exe" && base != "gswin32c.exe") {
				notice(a.window, "Nainstalujte Ghostscript nebo vyberte gswin64c.exe či gswin32c.exe.", 0x30)
				return
			}
		}
		switch {
		case req.Gray && req.Outline:
			suffix = "-K-krivky"
		case req.Gray:
			suffix = "-K"
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
