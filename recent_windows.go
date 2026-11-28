package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	recentFileLimit   = 100
	recentFilePreview = 10
	recentRowHeight   = 44
)

type recentPanel struct {
	files                      []string
	expanded                   bool
	empty, heading, more, list uintptr
}

func (p *recentPanel) visibleCount() int {
	if p.expanded {
		return len(p.files)
	}
	return min(recentFilePreview, len(p.files))
}

func recentFilePaths(paths []string) []string {
	files := make([]string, 0, recentFileLimit)
	for _, path := range paths {
		if !filepath.IsAbs(path) || strings.ContainsRune(path, 0) {
			continue
		}
		path = filepath.Clean(path)
		if slices.ContainsFunc(files, func(file string) bool { return strings.EqualFold(file, path) }) {
			continue
		}
		files = append(files, path)
		if len(files) == recentFileLimit {
			break
		}
	}
	return files
}

func readRecentFiles(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var files []string
	if json.Unmarshal(data, &files) != nil {
		return nil
	}
	return recentFilePaths(files)
}

func storeRecentFile(history, document string) ([]string, error) {
	if err := os.MkdirAll(filepath.Dir(history), 0700); err != nil {
		return nil, err
	}
	// Lock a separate file so that the JSON can be replaced while other instances wait.
	lock, err := os.OpenFile(history+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	var offset syscall.Overlapped
	const lockExclusive = 2
	if result, _, err := kernel32.NewProc("LockFileEx").Call(lock.Fd(), lockExclusive, 0, 1, 0, uintptr(unsafe.Pointer(&offset))); result == 0 {
		return nil, err
	}
	defer kernel32.NewProc("UnlockFileEx").Call(lock.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(&offset)))
	files := readRecentFiles(history)
	updated := recentFilePaths(append([]string{document}, files...))
	if !slices.Equal(files, updated) {
		if err := saveLocalJSON(history, updated); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

func (a *app) buildRecentPanel() {
	p := &a.recent
	if history, err := localAppFile("recent-files.json"); err == nil {
		p.files = readRecentFiles(history)
	}
	p.heading = a.uiLabel(uiRecentFiles, 16, 16, 145, true)
	p.more = a.button("", idRecentMore, 169, 12, 145, 0)
	a.panelButtons[p.more] = true
	p.empty = a.uiLabel(uiNoRecentPDFs, 16, 49, 298, false)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100 // Notify, owner-drawn rows with strings.
	p.list = a.add("LISTBOX", "", idRecentList, 16, 49, 298, 120, wsTabstop|listStyle, 0, a.normalFont)
	a.bindUILabel(p.list, uiRecentPDFList)
	sendMessage.Call(p.list, 0x01a0, 0, uintptr(a.scale(recentRowHeight))) // LB_SETITEMHEIGHT
	a.addScrollPane(p.list, a.scale(recentRowHeight), true)
	a.refreshRecentPanel()
}

func (a *app) rememberRecentFile(path string) {
	p := &a.recent
	p.files = recentFilePaths(append([]string{path}, p.files...))
	if history, err := localAppFile("recent-files.json"); err == nil {
		if files, err := storeRecentFile(history, path); err == nil {
			p.files = files
		}
	}
	a.refreshRecentPanel()
	if a.detailsTab() == idFileDetails {
		a.layoutPanel()
	}
}

func (a *app) refreshRecentPanel() {
	p := &a.recent
	sendMessage.Call(p.list, 0x000b, 0, 0) // WM_SETREDRAW
	sendMessage.Call(p.list, 0x0184, 0, 0) // LB_RESETCONTENT
	for _, path := range p.files[:p.visibleCount()] {
		sendMessage.Call(p.list, 0x0180, 0, uintptr(unsafe.Pointer(wideString(filepath.Base(path))))) // LB_ADDSTRING
	}
	if len(p.files) > 0 {
		sendMessage.Call(p.list, 0x0186, 0, 0) // LB_SETCURSEL
	}
	sendMessage.Call(p.list, 0x000b, 1, 0)
	pane := a.scrollPanes[p.list]
	if !pane.shown {
		showControl(p.list, false)
	}
	pane.reflow()
	invalidate(p.list)
	a.updateRecentActions()
}

func (a *app) toggleRecentFiles() {
	p := &a.recent
	if a.busy || a.detailsTab() != idFileDetails || len(p.files) <= recentFilePreview {
		return
	}
	selected, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0) // LB_GETCURSEL
	p.expanded = !p.expanded
	a.refreshRecentPanel()
	a.layoutPanel()
	if selected < uintptr(p.visibleCount()) {
		sendMessage.Call(p.list, 0x0186, selected, 0) // LB_SETCURSEL
		a.scrollPanes[p.list].syncScrollbar()
		a.updateRecentActions()
	}
}

func (a *app) recentItemLayout(item panelItem) (panelItem, bool) {
	p := &a.recent
	switch item.handle {
	case p.heading:
		return item, true
	case p.more:
		return item, len(p.files) > recentFilePreview
	case p.empty:
		return item, len(p.files) == 0
	case p.list:
		available := int(a.detailsArea.bottom-a.detailsArea.top)*96/a.dpi - item.y - 18
		rows := min(p.visibleCount(), max(1, available/recentRowHeight))
		item.height = rows * recentRowHeight
		return item, rows > 0
	}
	return item, true
}

func (p *recentPanel) selectedPath() string {
	index, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0) // LB_GETCURSEL
	if index >= uintptr(p.visibleCount()) {
		return ""
	}
	return p.files[index]
}

func (a *app) updateRecentActions() {
	p := &a.recent
	setEnabled(p.list, !a.busy)
	setEnabled(p.more, !a.busy)
	label := uiShowMore
	if p.expanded {
		label = uiShowFewer
	}
	a.setUIMessage(p.more, label)
}

func (a *app) openRecentFile() {
	if a.detailsTab() == idFileDetails && a.fileOptionEnabled(idRecentFiles) {
		if path := a.recent.selectedPath(); path != "" {
			a.openDocument(path, "")
		}
	}
}

func (a *app) drawRecentRow(item *buttonDraw) bool {
	p := &a.recent
	if item.controlType != 2 || item.handle != p.list {
		return false
	}
	bounds := item.bounds
	fillColor(item.dc, bounds, panelColor)
	if item.itemID >= uint32(len(p.files)) {
		return true
	}
	color := textColor
	if item.state&1 != 0 {
		color = selectionColor
	}
	bounds.left += int32(a.scale(7))
	bounds.right -= int32(a.scale(6))
	path := p.files[item.itemID]
	nameTop, nameBottom, pathTop, pathBottom := recentRowLines(int(bounds.bottom-bounds.top), a.scale(3), a.scale(2))
	nameBounds := rect{bounds.left,bounds.top+int32(nameTop),bounds.right,bounds.top+int32(nameBottom)}
	pathBounds := rect{bounds.left,bounds.top+int32(pathTop),bounds.right,bounds.top+int32(pathBottom)}
	drawText(item.dc, filepath.Base(path), nameBounds, a.normalFont, color, 0x8000) // DT_END_ELLIPSIS
	drawText(item.dc, filepath.Dir(path), pathBounds, a.normalFont, mutedColor, 0x4000) // DT_PATH_ELLIPSIS
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

func (a *app) recentContextInput(msg *winMessage) bool {
	p := &a.recent
	if msg.window != p.list || a.detailsTab() != idFileDetails || a.busy {
		return false
	}
	var pos point
	switch {
	case msg.message == 0x0205: // WM_RBUTTONUP
		pos = point{int32(int16(msg.lParam)), int32(int16(msg.lParam >> 16))}
		row, _, _ := sendMessage.Call(p.list, 0x01a9, 0, msg.lParam) // LB_ITEMFROMPOINT
		if row >= uintptr(p.visibleCount()) {
			return true
		}
		var bounds rect
		sendMessage.Call(p.list, 0x0198, row, uintptr(unsafe.Pointer(&bounds)))
		if !contains(bounds, pos.x, pos.y) {
			return true
		}
		setFocus.Call(p.list)
		sendMessage.Call(p.list, 0x0186, row, 0) // LB_SETCURSEL
	case msg.message == 0x0100 && (msg.wParam == 0x5d || msg.wParam == 0x79 && keyDown(0x10)):
		row, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0)
		if row >= uintptr(p.visibleCount()) {
			return true
		}
		var bounds rect
		sendMessage.Call(p.list, 0x0198, row, uintptr(unsafe.Pointer(&bounds)))
		pos = point{bounds.left + int32(a.scale(10)), bounds.bottom}
	default:
		return false
	}
	path := p.selectedPath()
	a.clearRecentPathTooltip()
	user32.NewProc("ClientToScreen").Call(p.list, uintptr(unsafe.Pointer(&pos)))
	options := []popupItem{{label: a.uiText(uiRecentOpen)}, {label: a.uiText(uiRecentOpenLocation)}, {label: a.uiText(uiRecentCopyPath)}}
	switch a.trackPopupAt(pos, max(238, (a.captionWidth(a.uiText(uiRecentOpenLocation), a.normalFont)*96+a.dpi-1)/a.dpi+40), options) {
	case 0:
		a.openDocument(path, "")
	case 1:
		directory := filepath.Dir(path)
		result, _, _ := shell32.NewProc("ShellExecuteW").Call(a.window, uintptr(unsafe.Pointer(wideString("open"))),
			uintptr(unsafe.Pointer(wideString(directory))), 0, 0, 1)
		if result <= 32 {
			a.reportStatus(a.uiText(uiRecentLocationFailed))
		}
	case 2:
		a.copyClipboardText(path)
	}
	return true
}


// A single native tooltip follows the hovered row, not the selected row.
// The original path is never shortened or modified for opening or copying.
func (a *app) recentPathHover(message uint32, wParam,lParam uintptr) {
	p:=&a.recent
	switch message {
	case 0x0200: // WM_MOUSEMOVE
		if a.busy || a.detailsTab()!=idFileDetails || wParam&3!=0 {
			a.clearRecentPathTooltip()
			return
		}
		pos:=point{int32(int16(lParam)),int32(int16(lParam>>16))}
		var client,bounds rect
		getClientRect.Call(p.list,uintptr(unsafe.Pointer(&client)))
		row,_,_:=sendMessage.Call(p.list,0x01a9,0,lParam) // LB_ITEMFROMPOINT
		if !contains(client,pos.x,pos.y) || row>>16!=0 || row>=uintptr(p.visibleCount()) {
			a.clearRecentPathTooltip()
			return
		}
		ok,_,_:=sendMessage.Call(p.list,0x0198,row,uintptr(unsafe.Pointer(&bounds))) // LB_GETITEMRECT
		if ok==^uintptr(0) || !contains(bounds,pos.x,pos.y) {
			a.clearRecentPathTooltip()
			return
		}
		if !trackMouseLeave(p.list) { return }
		path:=p.files[row]
		if a.tooltipCaptions[p.list].label!=path {
			if a.tooltip!=0 { sendMessage.Call(a.tooltip,0x0400+28,0,0) } // TTM_POP
			a.setTooltip(p.list,path)
		}
	case 0x0018,0x000a: // hidden or disabled
		if wParam==0 { a.clearRecentPathTooltip() }
	case 0x02a3,0x0201,0x0203,0x0204,0x020a,0x0115,0x0100,0x0197,0x0184,0x0047,0x0082:
		// Mouse leaves, clicks, scrolling, keyboard navigation, list changes,
		// repositioning or destruction invalidate the hovered row.
		a.clearRecentPathTooltip()
	}
}

func (a *app) clearRecentPathTooltip() {
	handle:=a.recent.list
	if handle==0 || a.tooltip==0 { return }
	if _,exists:=a.tooltipCaptions[handle]; !exists { return }
	sendMessage.Call(a.tooltip,0x0400+28,0,0) // TTM_POP
	parent,_,_:=user32.NewProc("GetParent").Call(handle)
	info:=nativeTooltip{flags:0x11,owner:parent,id:handle}
	info.size=uint32(unsafe.Sizeof(info))
	sendMessage.Call(a.tooltip,0x0400+51,0,uintptr(unsafe.Pointer(&info))) // TTM_DELTOOLW
	comctl32.NewProc("RemoveWindowSubclass").Call(handle,tooltipFocusSubclass,1)
	delete(a.tooltipCaptions,handle)
	if a.focusedHelpTarget==handle { a.setFocusedHelp(0) }
}
