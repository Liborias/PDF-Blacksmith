package main

import (
	"syscall"
	"unsafe"
)

var navigationSubclass = syscall.NewCallback(navigationButtonProc)

func watchNavigationHover(handle uintptr) {
	comctl32.NewProc("SetWindowSubclass").Call(handle, navigationSubclass, 1, 0)
}

func trackMouseLeave(hwnd uintptr) bool {
	track := struct {
		size, flags uint32
		window      uintptr
		hover       uint32
	}{flags: 2, window: hwnd} // TME_LEAVE
	track.size = uint32(unsafe.Sizeof(track))
	ok, _, _ := user32.NewProc("TrackMouseEvent").Call(uintptr(unsafe.Pointer(&track)))
	return ok != 0
}

func navigationButtonProc(hwnd uintptr, message uint32, wParam, lParam, id, data uintptr) uintptr {
	a := window
	switch message {
	case 0x0200: // WM_MOUSEMOVE
		if a.hoveredNavigation != hwnd && trackMouseLeave(hwnd) {
			a.setNavigationHover(hwnd)
		}
	case 0x0018, 0x000a: // WM_SHOWWINDOW, WM_ENABLE
		if wParam == 0 && a.hoveredNavigation == hwnd {
			a.setNavigationHover(0)
		}
	case 0x02a3, 0x0082: // WM_MOUSELEAVE, WM_NCDESTROY
		if a.hoveredNavigation == hwnd {
			a.setNavigationHover(0)
		}
	}
	result, _, _ := defSubclass.Call(hwnd, uintptr(message), wParam, lParam)
	return result
}

func (a *app) setNavigationHover(handle uintptr) {
	previous := a.hoveredNavigation
	a.hoveredNavigation = handle
	if previous != 0 {
		invalidate(previous)
	}
	if handle != 0 {
		invalidate(handle)
	}
}

func (a *app) navigationHot(item *buttonDraw) bool {
	return item.state&(1|64) != 0 || a.hoveredNavigation == item.handle
}

func (a *app) paintNavigation(dc uintptr, bounds rect, base paletteColor, disabled, selected, hovered bool) navigationStyle {
	style := uiColors.navigation(base, disabled, selected, hovered)
	fillColor(dc, bounds, style.fill)
	if style.accent {
		edge := bounds
		edge.right = min(edge.right, edge.left+int32(a.scale(2)))
		fillColor(dc, edge, selectionColor)
	}
	return style
}

func (a *app) drawToolNavigation(item *buttonDraw) {
	disabled := item.state&4 != 0
	style := a.paintNavigation(item.dc, item.bounds, panelColor, disabled, int(item.controlID) == a.tool, a.navigationHot(item))
	if !a.drawToolIconLabel(item, style.ink, style.fill) {
		drawText(item.dc, getText(item.handle), item.bounds, a.labelFont, style.ink, 1)
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
}
