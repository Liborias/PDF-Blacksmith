package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	idPreview   = 150
	idSettings  = 151
	idToolSave  = 1001
	idToolReset = 1002
)

type toolWindow struct {
	host          *app
	parent        *app
	width, height int
	command       func(id, notification int)
	closeAllowed  func() bool
	paint         func(uintptr)
	drawItem      func(*buttonDraw) bool
	message       func(uint32, uintptr, uintptr) bool
}

var activeTool *toolWindow
var toolClassRegistered bool

func createTool(parent *app, title string, width, height int) *toolWindow {
	if activeTool != nil {
		return nil
	}
	className := wideString("PDFBlaksmithTool")
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	icon, _, _ := user32.NewProc("LoadIconW").Call(parent.instance, appIconID)
	if !toolClassRegistered {
		class := windowClass{windowProc: syscall.NewCallback(toolProc), instance: parent.instance,
			className: uintptr(unsafe.Pointer(className)), cursor: cursor, icon: icon, smallIcon: icon}
		class.size = uint32(unsafe.Sizeof(class))
		result, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class)))
		if result == 0 {
			return nil
		}
		toolClassRegistered = true
	}

	dpi := parent.dpi
	screenWidth, _, _ := user32.NewProc("GetSystemMetrics").Call(0)
	screenHeight, _, _ := user32.NewProc("GetSystemMetrics").Call(1)
	if screenWidth > 100 && screenHeight > 100 {
		dpi = min(dpi, int(screenWidth-28)*96/width, int(screenHeight-86)*96/height)
	}
	host := &app{instance: parent.instance, dpi: dpi, options: make(map[uintptr]bool),
		checkedOptions: make(map[uintptr]bool), panelLabels: make(map[uintptr]bool), panelButtons: make(map[uintptr]bool),
		readOnlyFields: make(map[uintptr]bool)}
	host.normalFont, host.labelFont, host.titleFont = host.font(16, 400), host.font(16, 600), host.font(25, 650)
	host.canvasBrush, host.panelBrush, host.inputBrush = makeBrush(canvasColor), makeBrush(panelColor), makeBrush(inputColor)

	tool := &toolWindow{host: host, parent: parent, width: width, height: height}
	activeTool = tool
	const style = 0x02c80000 // Caption, system menu and clipped child windows.
	bounds := host.area(0, 0, width, height)
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&bounds)), style, 0, 0)
	w, h := bounds.right-bounds.left, bounds.bottom-bounds.top
	x, y := max(0, (int(screenWidth)-int(w))/2), max(0, (int(screenHeight)-int(h))/2)
	host.window, _, _ = createWindow.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(wideString(title))),
		style, uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent.window, 0, parent.instance, 0)
	if host.window == 0 {
		host.disposeTheme()
		activeTool = nil
		return nil
	}
	dwm := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if dwm.Find() == nil {
		dark, round := uint32(1), uint32(2)
		dwm.Call(host.window, 20, uintptr(unsafe.Pointer(&dark)), 4)
		dwm.Call(host.window, 33, uintptr(unsafe.Pointer(&round)), 4)
	}
	return tool
}

func (tool *toolWindow) show() {
	enableWindow.Call(tool.parent.window, 0)
	showWindow.Call(tool.host.window, 5)
}

func (tool *toolWindow) close() {
	if tool.closeAllowed != nil && !tool.closeAllowed() {
		return
	}
	enableWindow.Call(tool.parent.window, 1)
	destroyWindow.Call(tool.host.window)
	tool.host.disposeTheme()
	activeTool = nil
	user32.NewProc("SetActiveWindow").Call(tool.parent.window)
}

func (tool *toolWindow) label(text string, x, y, width int, strong bool) uintptr {
	control := tool.host.label(text, x, y, width, strong)
	tool.host.panelLabels[control] = false
	return control
}

func (tool *toolWindow) button(text string, id, x, y, width int) uintptr {
	control := tool.host.button(text, id, x, y, width, 0)
	tool.host.panelButtons[control] = false
	return control
}

func (tool *toolWindow) readout(x, y, w, h int) uintptr {
	const esMultiReadonly = 0x0004 | 0x0040 | esReadonly
	handle := tool.host.add("EDIT", "", 0, x, y, w, h, wsTabstop|wsVscroll|esMultiReadonly, 0, tool.host.normalFont)
	// Ask colorControl to paint this read-only edit in the same colors as editable fields.
	tool.host.readOnlyFields[handle] = true
	sendMessage.Call(handle, 0x00d3, 3, uintptr(tool.host.scale(7))|uintptr(tool.host.scale(7))<<16)
	return handle
}

func toolProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	tool := activeTool
	if tool != nil {
		if tool.message != nil && tool.message(msg, wParam, lParam) {
			return 0
		}
		switch msg {
		case wmEraseBkgnd:
			return 1
		case wmPaint:
			var info paintInfo
			dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
			fillColor(dc, tool.host.area(0, 0, tool.width, tool.height), canvasColor)
			if tool.paint != nil {
				tool.paint(dc)
			}
			user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
			return 0
		case wmCtlColorEdit, wmCtlColorList, wmCtlColorText:
			return tool.host.colorControl(msg, wParam, lParam)
		case wmDrawItem:
			if lParam != 0 {
				var item buttonDraw
				copyDrawItem.Call(uintptr(unsafe.Pointer(&item)), lParam, unsafe.Sizeof(item))
				if tool.host.drawButton(&item) || (tool.drawItem != nil && tool.drawItem(&item)) {
					return 1
				}
			}
		case wmCommand:
			id, notification := int(wParam&0xffff), int(wParam>>16&0xffff)
			if id == 2 {
				tool.close()
			} else if tool.command != nil {
				tool.command(id, notification)
			}
			return 0
		case wmClose:
			tool.close()
			return 0
		}
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}

func toolCombo(host *app, id, x, y, width int, names []string) uintptr {
	h := host.add("COMBOBOX", "", id, x, y, width, 250, wsTabstop|wsVscroll|3, 0, host.normalFont)
	host.controls = append(host.controls, h)
	sendMessage.Call(h, 0x0153, ^uintptr(0), uintptr(host.scale(24)))
	for _, name := range names {
		sendMessage.Call(h, 0x0143, 0, uintptr(unsafe.Pointer(wideString(name))))
	}
	sendMessage.Call(h, 0x014e, 0, 0)
	return h
}

func comboIndex(handle uintptr) int {
	index, _, _ := sendMessage.Call(handle, 0x0147, 0, 0)
	return int(int32(index))
}

func chooseCombo(handle uintptr, index int) {
	sendMessage.Call(handle, 0x014e, uintptr(index), 0)
}

type exportOptions struct {
	Intent      int    `json:"intent"`
	BlackPoint  int    `json:"black_point"`
	ImageDPI    int    `json:"image_dpi"`
	Compression string `json:"compression"`
	ProofDPI    int    `json:"proof_dpi"`
	RepairDPI   int    `json:"repair_dpi"`
}

var defaults = exportOptions{Intent: -1, BlackPoint: -1, ImageDPI: 0, Compression: "auto", ProofDPI: 150, RepairDPI: 600}
var (
	intents      = []int{-1, 0, 1, 2, 3}
	blackPoints  = []int{-1, 1, 0}
	imageDPIs    = []int{0, 150, 300, 600}
	compressions = []string{"auto", "flate"}
	proofDPIs    = []int{75, 150, 300}
	repairDPIs   = []int{300, 600}
)

func defaultIndex[T comparable](values []T, value, fallback T) int {
	fallbackIndex := 0
	for index, candidate := range values {
		if candidate == value {
			return index
		}
		if candidate == fallback {
			fallbackIndex = index
		}
	}
	return fallbackIndex
}

func settingsFile() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, appName, "export-settings.json"), nil
}

func loadExportOptions() exportOptions {
	options := defaults
	path, err := settingsFile()
	if err != nil {
		return options
	}
	if data, err := os.ReadFile(path); err == nil {
		if err = json.Unmarshal(data, &options); err != nil {
			return defaults
		}
	}
	options.Intent = intents[defaultIndex(intents, options.Intent, defaults.Intent)]
	options.BlackPoint = blackPoints[defaultIndex(blackPoints, options.BlackPoint, defaults.BlackPoint)]
	options.ImageDPI = imageDPIs[defaultIndex(imageDPIs, options.ImageDPI, defaults.ImageDPI)]
	options.Compression = compressions[defaultIndex(compressions, options.Compression, defaults.Compression)]
	options.ProofDPI = proofDPIs[defaultIndex(proofDPIs, options.ProofDPI, defaults.ProofDPI)]
	options.RepairDPI = repairDPIs[defaultIndex(repairDPIs, options.RepairDPI, defaults.RepairDPI)]
	return options
}

func saveExportOptions(options exportOptions) error {
	path, err := settingsFile()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), "export-settings-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	writeErr := json.NewEncoder(file).Encode(options)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}

func (a *app) settings() {
	tool := createTool(a, "PDF Blaksmith · Nastavení exportu", 604, 344)
	if tool == nil {
		return
	}
	host := tool.host
	tool.label("Nastavení exportu", 26, 20, 548, true)
	tool.label("Záměr reprodukce", 26, 62, 253, false)
	tool.label("Kompenzace černého bodu", 309, 62, 268, false)
	intent := toolCombo(host, 0, 26, 85, 254, []string{"Podle PDF a Ghostscriptu", "Perceptuální", "Relativní kolorimetrický", "Saturace", "Absolutní kolorimetrický"})
	black := toolCombo(host, 0, 309, 85, 268, []string{"Podle PDF a Ghostscriptu", "Zapnout", "Vypnout"})
	tool.label("Převzorkování obrázků", 26, 132, 253, false)
	tool.label("Komprese obrázků", 309, 132, 268, false)
	resolution := toolCombo(host, 0, 26, 155, 254, []string{"Původní rozlišení", "150 dpi", "300 dpi", "600 dpi"})
	compression := toolCombo(host, 0, 309, 155, 268, []string{"Automatická", "Bezeztrátová (Flate)"})
	tool.label("Kontrola desek", 26, 202, 253, false)
	tool.label("Rastrová oprava desky K", 309, 202, 268, false)
	proof := toolCombo(host, 0, 26, 225, 254, []string{"75 dpi", "150 dpi", "300 dpi"})
	repair := toolCombo(host, 0, 309, 225, 268, []string{"300 dpi", "600 dpi"})

	apply := func(options exportOptions) {
		chooseCombo(intent, defaultIndex(intents, options.Intent, defaults.Intent))
		chooseCombo(black, defaultIndex(blackPoints, options.BlackPoint, defaults.BlackPoint))
		chooseCombo(resolution, defaultIndex(imageDPIs, options.ImageDPI, defaults.ImageDPI))
		chooseCombo(compression, defaultIndex(compressions, options.Compression, defaults.Compression))
		chooseCombo(proof, defaultIndex(proofDPIs, options.ProofDPI, defaults.ProofDPI))
		chooseCombo(repair, defaultIndex(repairDPIs, options.RepairDPI, defaults.RepairDPI))
	}
	apply(a.export)
	tool.button("Obnovit výchozí", idToolReset, 26, 286, 157)
	tool.button("Zrušit", 2, 300, 286, 110)
	tool.button("Uložit nastavení", idToolSave, 424, 286, 153)
	host.mode = idToolSave
	tool.command = func(id, notification int) {
		if notification != 0 {
			return
		}
		switch id {
		case idToolReset:
			apply(defaults)
		case idToolSave:
			options := exportOptions{intents[comboIndex(intent)], blackPoints[comboIndex(black)],
				imageDPIs[comboIndex(resolution)], compressions[comboIndex(compression)],
				proofDPIs[comboIndex(proof)], repairDPIs[comboIndex(repair)]}
			if err := saveExportOptions(options); err != nil {
				notice(host.window, fmt.Sprintf("Nastavení nelze uložit.\n%v", err), 0x30)
				return
			}
			a.export = options
			a.updatePrintControls()
			tool.close()
		}
	}
	tool.show()
	setFocus.Call(intent)
}

func mainMenu() uintptr {
	menu, _, _ := user32.NewProc("CreateMenu").Call()
	sub, _, _ := user32.NewProc("CreatePopupMenu").Call()
	file, _, _ := user32.NewProc("CreatePopupMenu").Call()
	appendMenu := user32.NewProc("AppendMenuW")
	appendMenu.Call(file, 0, idPreview, uintptr(unsafe.Pointer(wideString("Náhled a editor vstupu…"))))
	appendMenu.Call(sub, 0, idSettings, uintptr(unsafe.Pointer(wideString("Nastavení exportu…"))))
	appendMenu.Call(menu, 0x10, file, uintptr(unsafe.Pointer(wideString("&Soubor"))))
	appendMenu.Call(menu, 0x10, sub, uintptr(unsafe.Pointer(wideString("&Export"))))
	return menu
}
