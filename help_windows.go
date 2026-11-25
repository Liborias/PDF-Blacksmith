package main

import (
	"syscall"
	"unsafe"
)

type nativeTooltip struct {
	size, flags  uint32
	owner, id    uintptr
	bounds       rect
	instance     uintptr
	text         *uint16
	param, extra uintptr
}

type tooltipCaption struct {
	label string
	text  *uint16
}

var tooltipFocusSubclass uintptr

func init() {
	tooltipFocusSubclass = syscall.NewCallback(tooltipFocusProc)
}

func tooltipFocusProc(hwnd uintptr, message uint32, wParam, lParam, id, data uintptr) uintptr {
	a := window
	switch message {
	case 0x0007: // WM_SETFOCUS
		a.setFocusedHelp(hwnd)
	case 0x0008, 0x0082: // WM_KILLFOCUS, WM_NCDESTROY
		if a.focusedHelpTarget == hwnd {
			a.setFocusedHelp(0)
		}
		if message == 0x0082 {
			delete(a.tooltipCaptions, hwnd)
			comctl32.NewProc("RemoveWindowSubclass").Call(hwnd, tooltipFocusSubclass, 1)
		}
	}
	result, _, _ := defSubclass.Call(hwnd, uintptr(message), wParam, lParam)
	return result
}

func (a *app) setFocusedHelp(handle uintptr) {
	a.focusedHelpTarget = handle
	if a.focusHelp != 0 {
		setText(a.focusHelp, a.tooltipCaptions[handle].label)
	}
}

func (a *app) setTooltip(handle uintptr, label string) {
	if handle == 0 || label == "" || a.tooltipCaptions[handle].label == label {
		return
	}
	if a.tooltip == 0 {
		a.tooltip, _, _ = createWindow.Call(8, uintptr(unsafe.Pointer(wideString("tooltips_class32"))), 0,
			0x80000003, 0, 0, 0, 0, a.window, 0, a.instance, 0)
		if a.tooltip == 0 {
			return
		}
		a.tooltipCaptions = make(map[uintptr]tooltipCaption)
		// Disable the system skin for this window so the palette applies.
		empty := wideString("")
		syscall.NewLazyDLL("uxtheme.dll").NewProc("SetWindowTheme").Call(a.tooltip, uintptr(unsafe.Pointer(empty)), uintptr(unsafe.Pointer(empty)))
		a.recolorTooltips()
		sendMessage.Call(a.tooltip, 0x0400+24, 0, uintptr(a.scale(350)))
	}
	text := wideString(label)
	parent, _, _ := user32.NewProc("GetParent").Call(handle)
	info := nativeTooltip{flags: 0x11, owner: parent, id: handle, text: text} // TTF_SUBCLASS | TTF_IDISHWND
	info.size = uint32(unsafe.Sizeof(info))
	if _, exists := a.tooltipCaptions[handle]; exists {
		sendMessage.Call(a.tooltip, 0x0400+57, 0, uintptr(unsafe.Pointer(&info))) // TTM_UPDATETIPTEXTW
	} else {
		if ok, _, _ := sendMessage.Call(a.tooltip, 0x0400+50, 0, uintptr(unsafe.Pointer(&info))); ok == 0 {
			return
		}
		comctl32.NewProc("SetWindowSubclass").Call(handle, tooltipFocusSubclass, 1, 0)
	}
	a.tooltipCaptions[handle] = tooltipCaption{label, text}
	if a.focusedHelpTarget == handle {
		a.setFocusedHelp(handle)
	}
}

func (a *app) recolorTooltips() {
	if a.tooltip == 0 {
		return
	}
	sendMessage.Call(a.tooltip, 0x0400+28, 0, 0) // TTM_POP
	sendMessage.Call(a.tooltip, 0x0400+19, uiColors["tooltipBg"].win(), 0)
	sendMessage.Call(a.tooltip, 0x0400+20, uiColors["tooltipText"].win(), 0)
}
