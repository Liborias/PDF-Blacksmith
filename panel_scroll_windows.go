package main

import (
	"syscall"
	"unsafe"
)

var (
	comctl32         = syscall.NewLazyDLL("comctl32.dll")
	defSubclass      = comctl32.NewProc("DefSubclassProc")
	scrollTrackColor = paletteColor{103, 99, 109}
	scrollThumbColor = paletteColor{140, 136, 147}
)

const trackClassName = "PDFBlacksmithScrollTrack"

type panelScrollbar struct {
	handle                   uintptr
	total, visible, position int
	apply                    func(int)
	dragging                 bool
	dragOffset               int
}

func (a *app) registerScrollTrack(cursor uintptr) bool {
	class := windowClass{instance: a.instance, className: uintptr(unsafe.Pointer(wideString(trackClassName))),
		windowProc: syscall.NewCallback(scrollTrackProc), cursor: cursor}
	class.size = uint32(unsafe.Sizeof(class))
	ok, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class)))
	return ok != 0
}

func (a *app) addPanelScrollbar(parent uintptr, apply func(int)) *panelScrollbar {
	handle, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(wideString(trackClassName))), 0,
		wsChild, 0, 0, 1, 1, parent, 0, a.instance, 0)
	bar := &panelScrollbar{handle: handle, apply: apply}
	a.scrollbars[handle] = bar
	return bar
}

func (bar *panelScrollbar) update(total, visible, position int, shown bool) {
	bar.total, bar.visible = total, visible
	bar.position = max(0, min(position, total-visible))
	showControl(bar.handle, shown && total > visible)
	invalidate(bar.handle)
}

func (bar *panelScrollbar) setPosition(position int) bool {
	position = max(0, min(position, bar.total-bar.visible))
	if position == bar.position {
		return false
	}
	bar.apply(position)
	return true
}

func (bar *panelScrollbar) geometry(bounds rect) (rect, rect) {
	d := window.scale
	height := max(1, int(bounds.bottom-bounds.top)-2*d(2))
	thumbHeight := min(height, max(d(18), height*bar.visible/max(1, bar.total)))
	top := d(2) + (height-thumbHeight)*bar.position/max(1, bar.total-bar.visible)
	middle := (bounds.left + bounds.right) / 2
	track := rect{middle - 1, int32(d(2)), middle + 1, int32(d(2) + height)}
	thumb := rect{middle - int32(d(3)), int32(top), middle + int32(d(3)), int32(top + thumbHeight)}
	return track, thumb
}

func scrollTrackProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	bar := window.scrollbars[hwnd]
	if bar != nil {
		var bounds rect
		getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
		track, thumb := bar.geometry(bounds)
		y := int(int16(lParam >> 16))
		switch msg {
		case wmEraseBkgnd:
			return 1
		case wmPaint:
			var info paintInfo
			dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
			fillColor(dc, bounds, panelColor)
			rounded(dc, track, 2, scrollTrackColor, scrollTrackColor)
			rounded(dc, thumb, window.scale(6), scrollThumbColor, scrollThumbColor)
			user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&info)))
			return 0
		case 0x0201: // WM_LBUTTONDOWN
			switch {
			case y < int(thumb.top):
				bar.setPosition(bar.position - bar.visible)
			case y >= int(thumb.bottom):
				bar.setPosition(bar.position + bar.visible)
			default:
				bar.dragging, bar.dragOffset = true, y-int(thumb.top)
				user32.NewProc("SetCapture").Call(hwnd)
			}
			return 0
		case 0x0200: // WM_MOUSEMOVE
			if bar.dragging {
				travel := max(1, int(track.bottom-track.top-thumb.bottom+thumb.top))
				position := (y - bar.dragOffset - int(track.top)) * (bar.total - bar.visible) / travel
				bar.setPosition(position)
			}
			return 0
		case 0x0202: // WM_LBUTTONUP
			if bar.dragging {
				user32.NewProc("ReleaseCapture").Call()
			}
			return 0
		case 0x0215: // WM_CAPTURECHANGED
			bar.dragging = false
			return 0
		}
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}

type scrollPane struct {
	handle        uintptr
	bar           *panelScrollbar
	frame         rect
	shown, narrow bool
	lineHeight    int
	list          bool
}

var paneSubclass = syscall.NewCallback(func(hwnd uintptr, message uint32, wParam, lParam, id, data uintptr) uintptr {
	v := window.view
	if v != nil && hwnd == v.pageList && v.pageListMouse(message, wParam, lParam) {
		return 0
	}
	bookmarkList := v != nil && hwnd == v.bookmarkPanel.list && v.ready()
	if bookmarkList && (message == 0x0201 || message == 0x0203) { // WM_LBUTTONDOWN, WM_LBUTTONDBLCLK
		if index, branch := v.bookmarkAt(lParam); branch {
			if message == 0x0201 {
				v.toggleBookmarkBranch(index)
			}
			return 0
		}
	}
	result, _, _ := defSubclass.Call(hwnd, uintptr(message), wParam, lParam)
	if bookmarkList && message == 0x0202 { // WM_LBUTTONUP, including a click on the already selected row.
		if index, branch := v.bookmarkAt(lParam); index >= 0 && !branch && index == v.bookmarkPanel.selected {
			v.goToBookmark()
		}
	}
	if pane := window.scrollPanes[hwnd]; pane != nil {
		switch message {
		case 0x000c: // WM_SETTEXT
			pane.reflow()
		case 0x00b6, 0x00b7, 0x0113, 0x0100, 0x0102, 0x0201, 0x0202, 0x0186, 0x0197, lbSetCaretIndex:
			pane.syncScrollbar()
		case 0x0200: // Dragging a text selection can scroll the native edit control.
			if wParam&1 != 0 {
				pane.syncScrollbar()
			}
		}
	}
	return result
})

func (a *app) addScrollPane(handle uintptr, lineHeight int, list bool) {
	pane := &scrollPane{handle: handle, lineHeight: max(1, lineHeight), list: list}
	parent, _, _ := user32.NewProc("GetParent").Call(handle)
	pane.bar = a.addPanelScrollbar(parent, pane.scrollTo)
	a.scrollPanes[handle] = pane
	comctl32.NewProc("SetWindowSubclass").Call(handle, paneSubclass, 1, 0)
	showControl(handle, false)
}

func (a *app) textLineHeight(handle uintptr) int {
	dc, _, _ := user32.NewProc("GetDC").Call(handle)
	old, _, _ := gdiSelect.Call(dc, a.normalFont)
	var metrics [16]int32 // TEXTMETRICW, beginning with tmHeight.
	gdi32.NewProc("GetTextMetricsW").Call(dc, uintptr(unsafe.Pointer(&metrics[0])))
	gdiSelect.Call(dc, old)
	user32.NewProc("ReleaseDC").Call(handle, dc)
	return int(metrics[0])
}

func (pane *scrollPane) lines() (int, int, int) {
	countMessage, topMessage := uintptr(0x00ba), uintptr(0x00ce) // EM_GETLINECOUNT, EM_GETFIRSTVISIBLELINE
	if pane.list {
		countMessage, topMessage = 0x018b, 0x018e // LB_GETCOUNT, LB_GETTOPINDEX
	}
	count, _, _ := sendMessage.Call(pane.handle, countMessage, 0, 0)
	top, _, _ := sendMessage.Call(pane.handle, topMessage, 0, 0)
	height := pane.frame.bottom - pane.frame.top
	if !pane.list {
		var format rect
		sendMessage.Call(pane.handle, 0x00b2, 0, uintptr(unsafe.Pointer(&format))) // EM_GETRECT
		height = format.bottom - format.top
	}
	return int(count), max(1, int(height)/pane.lineHeight), int(top)
}

func (pane *scrollPane) syncScrollbar() {
	if !pane.shown {
		return
	}
	total, visible, position := pane.lines()
	old := pane.bar.position
	pane.bar.update(total, visible, position, true)
	if v := window.view; v != nil && pane.handle == v.pageList && old != pane.bar.position {
		postMessage.Call(window.window, wmThumbnailUpdate, 0, 0)
	}
}

func (pane *scrollPane) reflow() {
	if !pane.shown {
		return
	}
	width := int(pane.frame.right - pane.frame.left)
	pane.narrow = false
	pane.place()
	total, visible, _ := pane.lines()
	pane.narrow = total > visible && width > window.scale(16)
	if pane.narrow {
		pane.place()
	}
	pane.syncScrollbar()
}

func (pane *scrollPane) place() {
	bounds := pane.frame
	left, top := int(bounds.left), int(bounds.top)
	width, height := int(bounds.right-bounds.left), int(bounds.bottom-bounds.top)
	gutter := window.scale(14)
	if pane.narrow {
		width -= gutter
	}
	place(pane.handle, left, top, width, height)
	place(pane.bar.handle, int(bounds.right)-gutter, top, gutter, height)
}

func (pane *scrollPane) scrollTo(position int) {
	if pane.list {
		sendMessage.Call(pane.handle, 0x0197, uintptr(position), 0) // LB_SETTOPINDEX
	} else {
		_, _, current := pane.lines()
		sendMessage.Call(pane.handle, 0x00b6, 0, uintptr(position-current)) // EM_LINESCROLL
	}
}

func (a *app) layoutScrollPane(pane *scrollPane, item panelItem, offset int, shown bool) {
	d := a.scale
	left, top := int32(d(item.x)), int32(d(item.y-offset))
	pane.arrange(rect{left, top, left + int32(d(item.width)), top + int32(d(item.height))}, shown)
}

func (pane *scrollPane) arrange(frame rect, shown bool) {
	if !shown {
		if pane.shown {
			pane.shown = false
			showControl(pane.handle, false)
			showControl(pane.bar.handle, false)
		}
		return
	}
	old := pane.frame
	pane.frame = frame
	if resized := old.right-old.left != frame.right-frame.left || old.bottom-old.top != frame.bottom-frame.top; !pane.shown || resized {
		pane.shown = true
		pane.reflow()
		showControl(pane.handle, true)
	} else if old != pane.frame {
		pane.place()
	}
}

func (a *app) wheelSettings(location uintptr, delta int, ctrl bool) {
	screen := point{int32(int16(location)), int32(int16(location >> 16))}
	hover, _, _ := user32.NewProc("WindowFromPoint").Call(uintptr(uint32(screen.x)) | uintptr(uint32(screen.y))<<32)
	rows := delta / 40
	if delta > 0 {
		rows = max(1, rows)
	} else if delta < 0 {
		rows = min(-1, rows)
	}
	thumbs := a.scrollPanes[a.view.pageList]
	parent, _, _ := user32.NewProc("GetParent").Call(hover)
	if hood := a.thumbnailSidebar; hover == hood || parent == hood {
		if ctrl {
			a.view.zoomThumbnails(screen, delta)
		} else {
			thumbs.bar.setPosition(thumbs.bar.position - thumbnailWheelSteps(delta))
		}
		return
	}
	if hood := a.bookmarkSidebar; hover == hood || parent == hood {
		bar := a.scrollPanes[a.view.bookmarkPanel.list].bar
		bar.setPosition(bar.position - rows)
		return
	}
	if a.detailsCollapsed {
		return
	}
	bar := a.scrollbars[hover]
	if pane := a.scrollPanes[hover]; pane != nil {
		bar = pane.bar
	}
	if bar != nil && bar != a.sideScroll && delta != 0 && bar.setPosition(bar.position-rows) {
		return
	}
	a.sideScroll.setPosition(a.sideScroll.position - delta*62/120)
}
