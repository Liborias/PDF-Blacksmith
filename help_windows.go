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


// TTN_SHOW lets the native tooltip retain its timing, wrapping and theme
// while replacing the cursor-relative placement with anchor-relative placement.
func (a *app) positionIconTooltip(param uintptr) bool {
	if param == 0 || a.tooltip == 0 {
		return false
	}
	notice := (*struct {
		from, id uintptr
		code uint32
	})(unsafe.Pointer(param))
	if notice.from != a.tooltip || notice.code != 0xfffffdf7 { // TTN_SHOW = -521
		return false
	}
	info := nativeTooltip{}
	info.size = uint32(unsafe.Sizeof(info))
	ok, _, _ := sendMessage.Call(a.tooltip, 0x0400+59, 0, uintptr(unsafe.Pointer(&info))) // TTM_GETCURRENTTOOLW
	if ok == 0 || info.flags&1 == 0 {
		return false
	}
	handle := info.id
	side := 0
	switch handle {
	case a.leftPanelsToggle, a.toolsToggle, a.thumbnailsToggle:
		side = 1
	case a.rightPanelsToggle, a.detailsToggle, a.resultsToggle, a.bookmarksToggle, a.layers.toggle:
		side = -1
	}
	if side == 0 {
		found := false
		for _, control := range a.commandBar.controls {
			if control.handle == handle && !control.menu {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	var anchor, tip rect
	getRect := user32.NewProc("GetWindowRect")
	if ok, _, _ := getRect.Call(handle, uintptr(unsafe.Pointer(&anchor))); ok == 0 {
		return false
	}
	if ok, _, _ := getRect.Call(a.tooltip, uintptr(unsafe.Pointer(&tip))); ok == 0 {
		return false
	}
	monitor, _, _ := user32.NewProc("MonitorFromRect").Call(uintptr(unsafe.Pointer(&anchor)), 2)
	mi := monitorInfo{size:uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := user32.NewProc("GetMonitorInfoW").Call(monitor, uintptr(unsafe.Pointer(&mi))); ok == 0 {
		return false
	}
	toBox := func(r rect) tooltipBox { return tooltipBox{int(r.left),int(r.top),int(r.right),int(r.bottom)} }
	x,y,okPlacement := iconTooltipPosition(toBox(anchor),toBox(mi.work),int(tip.right-tip.left),int(tip.bottom-tip.top),a.scale(6),side)
	if !okPlacement {
		return false
	}
	ok, _, _ = user32.NewProc("SetWindowPos").Call(a.tooltip,0,uintptr(x),uintptr(y),0,0,0x15) // NOSIZE | NOZORDER | NOACTIVATE
	return ok != 0
}
