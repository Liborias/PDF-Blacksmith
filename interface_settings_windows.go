package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	interfaceSettingsFile = "interface-settings.json"
	interfaceDialogStyle  = 0x80c80880 // WS_POPUP | WS_CAPTION | WS_SYSMENU | DS_CENTER | DS_MODALFRAME
	settingsApply         = 1
	settingsCancel        = 2
	settingsAppearance    = 10
	settingsLanguage      = 11
	settingsCzech         = 12
	settingsEnglish       = 13
	settingsLight         = 14
	settingsDark          = 15
	settingsPaletteFirst  = 20
)

func loadInterfacePreferences() interfacePreferences {
	var fallback uiLocale = uiCzech
	if path, err := localAppFile("language-settings.json"); err == nil {
		data, _ := os.ReadFile(path)
		fallback = decodeUILanguage(data)
	}
	if path, err := localAppFile(interfaceSettingsFile); err == nil {
		data, _ := os.ReadFile(path)
		return decodeInterfacePreferences(data, fallback)
	}
	return decodeInterfacePreferences(nil, fallback)
}

func saveInterfacePreferences(settings interfacePreferences) error {
	path, err := localAppFile(interfaceSettingsFile)
	if err != nil {
		return err
	}
	return saveLocalJSON(path, settings)
}

type interfaceDialog struct {
	window, appearance, language, description, palettesLabel uintptr
	czech, english, light, dark, apply, cancel               uintptr
	palettes                                                 [len(uiPaletteOrder)]uintptr
	showLanguage                                             bool
}

var interfaceDialogCallback = syscall.NewCallback(interfaceDialogProc)

func (a *app) showInterfaceSettings() {
	if a.interfaceDialog != nil || a.saveInlineBefore(a.showInterfaceSettings) {
		return
	}
	previous := interfacePreferences{Locale: a.uiLocale, Theme: a.uiTheme}
	a.interfaceDialog = &interfaceDialog{}
	// Empty DLGTEMPLATE. Child controls are created in WM_INITDIALOG.
	template := struct {
		style, exStyle     uint32
		count              uint16
		x, y, cx, cy       int16
		menu, class, title uint16
	}{style: interfaceDialogStyle, cx: 250, cy: 200}
	result, _, _ := user32.NewProc("DialogBoxIndirectParamW").Call(a.instance, uintptr(unsafe.Pointer(&template)), a.window, interfaceDialogCallback, 0)
	a.interfaceDialog = nil
	if result != settingsApply {
		a.setUILanguage(previous.Locale)
		a.setUITheme(previous.Theme)
	}
	if result == ^uintptr(0) {
		notice(a.window, a.uiText(uiLanguageDialogFailed), 0x10)
	}
}

func (a *app) toggleThemeMode() {
	theme := a.uiTheme.oppositeMode()
	if err := saveInterfacePreferences(interfacePreferences{Locale: a.uiLocale, Theme: theme}); err != nil {
		notice(a.window, a.uiText(uiLanguageSaveFailed)+"\n\n"+err.Error(), 0x10)
		return
	}
	a.setUITheme(theme)
}

func interfaceDialogProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	a := window
	dialog := a.interfaceDialog
	if dialog == nil {
		return 0
	}
	switch message {
	case 0x0110: // WM_INITDIALOG
		a.buildInterfaceDialog(hwnd, dialog)
		return 1
	case wmCtlColorText: // WM_CTLCOLORSTATIC
		gdiTextColor.Call(wParam, textColor.win())
		gdiBkColor.Call(wParam, panelColor.win())
		return a.panelBrush
	case 0x0136: // WM_CTLCOLORDLG
		return a.panelBrush
	case wmEraseBkgnd:
		var bounds rect
		getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
		fillColor(wParam, bounds, panelColor)
		return 1
	case wmDrawItem:
		if lParam != 0 {
			var item buttonDraw
			copyDrawItem.Call(uintptr(unsafe.Pointer(&item)), lParam, unsafe.Sizeof(item))
			if item.controlType == 4 {
				a.drawInterfaceButton(dialog, &item)
				return 1
			}
		}
	case wmCommand:
		if wParam>>16&0xffff == 0 && a.interfaceSettingsCommand(dialog, int(wParam&0xffff)) {
			return 1
		}
	case wmClose:
		user32.NewProc("EndDialog").Call(hwnd, settingsCancel)
		return 1
	}
	return 0
}

func (a *app) interfaceSettingsCommand(dialog *interfaceDialog, id int) bool {
	theme := a.uiTheme
	switch id {
	case settingsApply:
		if err := saveInterfacePreferences(interfacePreferences{Locale: a.uiLocale, Theme: a.uiTheme}); err != nil {
			notice(dialog.window, a.uiText(uiLanguageSaveFailed)+"\n\n"+err.Error(), 0x10)
			return true
		}
		user32.NewProc("EndDialog").Call(dialog.window, settingsApply)
		return true
	case settingsCancel:
		user32.NewProc("EndDialog").Call(dialog.window, settingsCancel)
		return true
	case settingsAppearance, settingsLanguage:
		dialog.showLanguage = id == settingsLanguage
	case settingsCzech:
		a.setUILanguage(uiCzech)
	case settingsEnglish:
		a.setUILanguage(uiEnglish)
	case settingsLight:
		theme.Mode = "light"
	case settingsDark:
		theme.Mode = "dark"
	default:
		index := id - settingsPaletteFirst
		if index < 0 || index >= len(uiPaletteOrder) {
			return false
		}
		theme.Palette = uiPaletteOrder[index]
	}
	a.setUITheme(theme)
	a.updateInterfaceDialog(dialog)
	return true
}

func (a *app) buildInterfaceDialog(hwnd uintptr, dialog *interfaceDialog) {
	dialog.window = hwnd
	var owner rect
	user32.NewProc("GetWindowRect").Call(a.window, uintptr(unsafe.Pointer(&owner)))
	frame := rect{right: int32(a.scale(440)), bottom: int32(a.scale(328))}
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&frame)), interfaceDialogStyle, 0, 0)
	width, height := int(frame.right-frame.left), int(frame.bottom-frame.top)
	place(hwnd, (int(owner.left+owner.right)-width)/2, (int(owner.top+owner.bottom)-height)/2, width, height)
	a.applyWindowTheme(hwnd)
	oldParent := a.controlParent
	a.controlParent = hwnd
	button := func(caption string, id, x, y, width int) uintptr {
		handle := a.button(caption, id, x, y, width, 0)
		watchNavigationHover(handle)
		return handle
	}
	dialog.appearance = button("", settingsAppearance, 20, 14, 194)
	dialog.language = button("", settingsLanguage, 226, 14, 194)
	dialog.description = a.add("STATIC", "", 0, 20, 60, 400, 46, 0, 0, a.normalFont)
	dialog.palettesLabel = a.label("", 20, 105, 400, true)
	for i := range dialog.palettes {
		dialog.palettes[i] = button("", settingsPaletteFirst+i, 20+i%2*206, 130+i/2*36, 194)
	}
	dialog.light = button("", settingsLight, 20, 246, 194)
	dialog.dark = button("", settingsDark, 226, 246, 194)
	dialog.czech = button("Čeština", settingsCzech, 20, 130, 194)
	dialog.english = button("English", settingsEnglish, 226, 130, 194)
	dialog.apply = button("", settingsApply, 208, 284, 100)
	dialog.cancel = button("", settingsCancel, 320, 284, 100)
	a.controlParent = oldParent
	a.updateInterfaceDialog(dialog)
	sendMessage.Call(hwnd, 0x0400+1, settingsApply, 0) // DM_SETDEFID
}

func (a *app) updateInterfaceDialog(dialog *interfaceDialog) {
	setText(dialog.window, a.uiText(uiSettingsTitle))
	setText(dialog.appearance, a.uiText(uiInterface))
	setText(dialog.language, a.uiText(uiLanguage))
	setText(dialog.palettesLabel, a.uiText(uiInterfaceColors))
	setText(dialog.apply, a.uiText(uiApply))
	setText(dialog.cancel, a.uiText(uiCancel))
	setText(dialog.light, a.uiText(uiLightMode))
	setText(dialog.dark, a.uiText(uiDarkMode))
	for i, handle := range dialog.palettes {
		setText(handle, uiPalettes.Palettes[uiPaletteOrder[i]].caption(a.uiLocale))
		showControl(handle, !dialog.showLanguage)
	}
	for _, handle := range []uintptr{dialog.palettesLabel, dialog.light, dialog.dark} {
		showControl(handle, !dialog.showLanguage)
	}
	showControl(dialog.czech, dialog.showLanguage)
	showControl(dialog.english, dialog.showLanguage)
	description := uiThemeDescription
	if dialog.showLanguage {
		description = uiLanguageChoice
	}
	setText(dialog.description, a.uiText(description))
	user32.NewProc("RedrawWindow").Call(dialog.window, 0, 0, 0x0085)
}

func (a *app) drawInterfaceButton(dialog *interfaceDialog, item *buttonDraw) {
	id, bounds := int(item.controlID), item.bounds
	selected := false
	symbol := ""
	switch id {
	case settingsAppearance:
		selected, symbol = !dialog.showLanguage, "i-settings_interface"
	case settingsLanguage:
		selected, symbol = dialog.showLanguage, "i-language"
	case settingsCzech:
		selected = a.uiLocale == uiCzech
	case settingsEnglish:
		selected = a.uiLocale == uiEnglish
	case settingsLight:
		selected = a.uiTheme.Mode == "light"
	case settingsDark:
		selected = a.uiTheme.Mode == "dark"
	default:
		index := id - settingsPaletteFirst
		selected = index >= 0 && index < len(uiPaletteOrder) && a.uiTheme.Palette == uiPaletteOrder[index]
	}
	hovered := a.navigationHot(item)
	if symbol != "" {
		style := a.paintNavigation(item.dc, bounds, panelColor, false, selected, hovered)
		size := a.scale(18)
		drawUIIcon(item.dc, symbol, int(bounds.left)+a.scale(10), int(bounds.top+bounds.bottom)/2-size/2, size, style.ink, style.fill)
		caption := bounds
		caption.left += int32(a.scale(36))
		drawText(item.dc, getText(item.handle), caption, a.labelFont, style.ink, 0x8000)
	} else {
		fill, edge := uiColors["control"], divider
		if selected {
			fill, edge = uiColors["selected"], selectionColor
		} else if id == settingsApply {
			fill = uiColors["toolbarActive"]
		}
		if hovered {
			fill = uiColors["hover"]
		}
		fillColor(item.dc, bounds, panelColor)
		rounded(item.dc, bounds, a.scale(6), fill, edge)
		drawText(item.dc, getText(item.handle), bounds, a.labelFont, textColor, 1|0x8000)
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
	}
}
