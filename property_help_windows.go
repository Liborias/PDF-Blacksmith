package main

import (
	"strings"
	"unsafe"
)

type propertyHelpPopup struct {
	handle, text, anchor uintptr
	body                 string
}

func (a *app) togglePropertyHelp(anchor uintptr, section *propertySection) {
	help := &a.propertyHelp
	if help.anchor == anchor {
		a.hidePropertyHelp(true)
		return
	}
	a.hidePropertyHelp(false)
	if help.handle == 0 {
		help.handle, _, _ = createWindow.Call(0, uintptr(unsafe.Pointer(wideString("PDFBlaksmithSidebar"))), 0,
			wsChild|0x06000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
		if help.handle == 0 {
			return
		}
		a.controlParent = help.handle
		help.text = a.readout(0, 0, 1, 1)
		a.controlParent = 0
	}
	body := a.propertyHelpText(section)
	help.anchor, help.body = anchor, body
	pane := a.scrollPanes[help.text]
	pane.arrange(rect{}, false)
	setText(help.text, multiline(body))
	a.updatePropertyHelp()
	if help.anchor == 0 {
		return
	}
	if a.tooltip != 0 {
		sendMessage.Call(a.tooltip, 0x0401, 0, 0) // TTM_ACTIVATE
	}
	setFocus.Call(help.text)
	sendMessage.Call(help.text, 0x00b1, 0, 0) // EM_SETSEL
	pane.bar.setPosition(0)
	invalidate(anchor)
}

func (a *app) hidePropertyHelp(restoreFocus bool) {
	help := &a.propertyHelp
	anchor := help.anchor
	if anchor == 0 {
		return
	}
	help.anchor = 0
	showControl(help.handle, false)
	if a.tooltip != 0 {
		sendMessage.Call(a.tooltip, 0x0401, 1, 0)
	}
	invalidate(anchor)
	if restoreFocus {
		setFocus.Call(anchor)
	}
}

func (a *app) updatePropertyHelp() {
	help := &a.propertyHelp
	if help.anchor == 0 {
		return
	}
	visible, _, _ := user32.NewProc("IsWindowVisible").Call(help.anchor)
	if visible == 0 {
		a.hidePropertyHelp(false)
		return
	}
	d := a.scale
	var anchor rect
	user32.NewProc("GetWindowRect").Call(help.anchor, uintptr(unsafe.Pointer(&anchor)))
	user32.NewProc("MapWindowPoints").Call(0, a.window, uintptr(unsafe.Pointer(&anchor)), 2)
	parent, _, _ := user32.NewProc("GetParent").Call(help.anchor)
	width, top, bottom := d(346), d(headerHeight), a.height-d(24)
	if parent == a.sidebar {
		if a.full || a.settingsWidth == 0 || anchor.bottom <= a.detailsArea.top || anchor.top >= a.detailsArea.bottom {
			a.hidePropertyHelp(false)
			return
		}
		width = a.settingsWidth - d(28)
		top, bottom = int(a.detailsArea.top), int(a.detailsArea.bottom)
	}
	maxHeight := max(int(anchor.top)-top, bottom-int(anchor.bottom)) - d(4)
	if maxHeight < d(30) {
		a.hidePropertyHelp(false)
		return
	}
	left := max(d(6), min(int(anchor.right)-width, a.width-width-d(6)))
	if parent == a.sidebar {
		left = max(int(a.detailsArea.left)+d(8), left)
	}
	pad := d(8)
	pane := a.scrollPanes[help.text]
	measure := rect{right: int32(width - 2*pad - d(5))}
	dc, _, _ := user32.NewProc("GetDC").Call(help.text)
	old, _, _ := gdiSelect.Call(dc, a.normalFont)
	winDrawText.Call(dc, uintptr(unsafe.Pointer(wideString(help.body))), ^uintptr(0),
		uintptr(unsafe.Pointer(&measure)), 0x0450) // DT_CALCRECT | DT_WORDBREAK | DT_NOPREFIX
	gdiSelect.Call(dc, old)
	user32.NewProc("ReleaseDC").Call(help.text, dc)
	height := min(maxHeight, d(272), int(measure.bottom)+2*pad+pane.lineHeight)
	y := int(anchor.bottom) + d(4)
	if y+height > bottom {
		y = int(anchor.top) - d(4) - height
	}
	user32.NewProc("SetWindowPos").Call(help.handle, 0, uintptr(left), uintptr(y), uintptr(width), uintptr(height), 0x0050) // SWP_SHOWWINDOW | SWP_NOACTIVATE
	pane.arrange(rect{int32(pad), int32(pad), int32(width - pad), int32(height - pad)}, true)
	user32.NewProc("RedrawWindow").Call(help.handle, 0, 0, 0x0185) // Invalidate and repaint the popup and its children.
}

func (a *app) propertyHelpInput(msg *winMessage) bool {
	help := &a.propertyHelp
	if help.anchor == 0 {
		return false
	}
	pane := a.scrollPanes[help.text]
	switch msg.message {
	case 0x0201, 0x0204, 0x0207, 0x00a1, 0x00a4, 0x00a7:
		switch msg.window {
		case help.anchor, help.handle, help.text, pane.bar.handle:
		default:
			a.hidePropertyHelp(false)
		}
	case 0x020a:
		screen := point{int32(int16(msg.lParam)), int32(int16(msg.lParam >> 16))}
		var bounds rect
		user32.NewProc("GetWindowRect").Call(help.handle, uintptr(unsafe.Pointer(&bounds)))
		if contains(bounds, screen.x, screen.y) {
			delta := int(int16(msg.wParam >> 16))
			rows := max(1, absTextDrag(int32(delta/40)))
			if delta < 0 {
				rows = -rows
			}
			if delta != 0 {
				pane.bar.setPosition(pane.bar.position - int(rows))
			}
			return true
		}
		a.hidePropertyHelp(false)
	case 0x0100:
		switch msg.wParam {
		case 0x1b:
			a.hidePropertyHelp(true)
			return true
		case 9:
			if msg.window == help.text {
				msg.window = help.anchor
				a.hidePropertyHelp(true)
			}
		case 0x0d:
			if msg.window == help.anchor {
				a.hidePropertyHelp(true)
				return true
			}
		}
	}
	return false
}

type propertyShortcut struct {
	key    string
	action uiMessage
}

func (a *app) propertyHelpText(section *propertySection) string {
	var body []string
	title := uiMessage("")
	if section != nil {
		title = section.title
		body = append(body, a.uiText(section.description))
		if section.detail != nil {
			if detail := section.detail(); detail != "" {
				body = append(body, detail)
			}
		}
	}
	legend := []string{a.uiText(uiShortcuts)}
	if section != nil && section.instructions != "" {
		legend = append(legend, a.uiText(section.instructions))
	}
	for _, shortcut := range propertyShortcuts(title) {
		line := a.uiText(shortcut.action)
		if shortcut.key != "" {
			line = shortcut.key + "  ·  " + line
		}
		legend = append(legend, line)
	}
	if len(legend) == 1 {
		legend = append(legend, a.uiText(uiPropertyNoShortcut))
	} else if title != "" {
		legend = append(legend, "Tab / Shift+Tab  ·  "+a.uiText(uiPropertyTab))
	}
	return strings.Join(append(body, strings.Join(legend, "\n")), "\n\n")
}

func propertyShortcuts(title uiMessage) []propertyShortcut {
	switch title {
	case "":
		return []propertyShortcut{
			{"Ctrl+O", uiOpenPDF}, {"Ctrl+S", uiSave}, {"Ctrl+Shift+S", uiSaveAs}, {"Ctrl+P", uiPrint},
			{"Ctrl+W", uiPropertyCloseTab}, {"Ctrl+Tab / Ctrl+Shift+Tab", uiPropertySwitchTab},
			{"Ctrl+F", uiFind}, {"F3 / Shift+F3", uiPropertySearchResults},
			{"Ctrl+Z", uiUndo}, {"Ctrl+Y / Ctrl+Shift+Z", uiRedo},
			{"Ctrl++ / Ctrl+−", uiPropertyZoom}, {"Ctrl+0", uiFitPage}, {"", uiWheelZoom},
			{"F11 / Esc", uiPropertyFullscreen},
		}
	case uiViewMode:
		return []propertyShortcut{
			{"← / →", uiPropertyPageOrder}, {"↑ / ↓", uiPropertyScroll}, {"PgUp / PgDn", uiPropertyScroll},
			{"Home / End", uiPropertyPageEnds}, {"Ctrl++ / Ctrl+−", uiPropertyZoom}, {"Ctrl+0", uiFitPage},
			{"", uiWheelZoom}, {"F11 / Esc", uiPropertyFullscreen},
		}
	case uiDocumentText:
		return []propertyShortcut{
			{"Ctrl+F", uiFind}, {"Enter", uiPropertySearchEnter}, {"F3 / Shift+F3", uiPropertySearchResults},
		}
	case uiTextSelection:
		return []propertyShortcut{{"Ctrl+Z", uiUndo}, {"Ctrl+Y / Ctrl+Shift+Z", uiRedo}}
	case uiTextFormat:
		return []propertyShortcut{
			{"Ctrl+B", uiBold}, {"Ctrl+I", uiItalic}, {"Ctrl+U", uiUnderline},
			{"Ctrl+L", uiAlignLeft}, {"Ctrl+E", uiAlignCenter}, {"Ctrl+R", uiAlignRight}, {"Ctrl+J", uiJustify},
			{"Ctrl+X", uiCut}, {"Ctrl+C", uiCopy}, {"Ctrl+V", uiPaste}, {"Ctrl+A", uiSelectAll},
			{"Ctrl+Z", uiUndo}, {"Ctrl+Y / Ctrl+Shift+Z", uiRedo}, {"Ctrl+Enter", uiPropertyApplyText},
		}
	case uiObjectSelection:
		return []propertyShortcut{
			{"V", uiPropertyPointSelect}, {"M", uiPropertyRectSelect}, {"L → L", uiPropertyLasso},
			{"Enter", uiPropertyFinishLasso}, {"", uiPropertySelectModifiers}, {"", uiPropertyAltDrag},
			{"Ctrl+Z", uiUndo}, {"Ctrl+Y / Ctrl+Shift+Z", uiRedo},
		}
	case uiObjectActions:
		return []propertyShortcut{
			{"Ctrl+C", uiCopy}, {"Ctrl+V", uiPaste}, {"Ctrl+D", uiPropertyDuplicate}, {"Delete", uiDelete},
			{"Enter", uiConfirmImageCrop}, {"Esc", uiCancelImageCrop}, {"", uiPropertyCropWheel},
			{"Ctrl+Z", uiUndo}, {"Ctrl+Y / Ctrl+Shift+Z", uiRedo},
		}
	case uiMovePages, uiOCRPages, uiExportPages, uiSplitPages:
		return []propertyShortcut{{"", uiPropertySelectPages}}
	case uiRotateAndDelete:
		return []propertyShortcut{{"Delete", uiPropertyDeletePages}, {"", uiPropertySelectPages}}
	default:
		return nil
	}
}
