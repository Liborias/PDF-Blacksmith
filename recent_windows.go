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

const recentFileLimit = 10

type recentPanel struct {
	files                   []string
	empty, list, path, open uintptr
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
	p.empty = a.label("Zatím tu nejsou žádná nedávná PDF.", 16, 16, 298, false)
	const listStyle = 0x0001 | 0x0010 | 0x0040 | 0x0100 // Notify, owner-drawn rows with strings.
	p.list = a.add("LISTBOX", "Nedávno otevřené soubory PDF", idRecentList, 16, 16, 298, 120, wsTabstop|listStyle, 0, a.normalFont)
	sendMessage.Call(p.list, 0x01a0, 0, uintptr(a.scale(24))) // LB_SETITEMHEIGHT
	a.addScrollPane(p.list, a.scale(24), true)
	p.path = a.readout(16, 144, 298, 48)
	p.open = a.button("Otevřít vybrané PDF", idRecentOpen, 16, 202, 298, 0)
}

func (a *app) rememberRecentFile(path string) {
	p := &a.recent
	p.files = recentFilePaths(append([]string{path}, p.files...))
	if history, err := localAppFile("recent-files.json"); err == nil {
		if files, err := storeRecentFile(history, path); err == nil {
			p.files = files
		}
	}
	if a.detailsTab() == idRecentFiles {
		a.refreshRecentPanel()
		a.layoutPanel()
	}
}

func (a *app) refreshRecentPanel() {
	p := &a.recent
	sendMessage.Call(p.list, 0x000b, 0, 0) // WM_SETREDRAW
	sendMessage.Call(p.list, 0x0184, 0, 0) // LB_RESETCONTENT
	for _, path := range p.files {
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

func (p *recentPanel) selectedPath() string {
	index, _, _ := sendMessage.Call(p.list, 0x0188, 0, 0) // LB_GETCURSEL
	if index >= uintptr(len(p.files)) {
		return ""
	}
	return p.files[index]
}

func (a *app) updateRecentActions() {
	p := &a.recent
	path := p.selectedPath()
	setText(p.path, path)
	setEnabled(p.list, !a.busy)
	setEnabled(p.open, !a.busy && path != "")
}

func (a *app) openRecentFile() {
	if a.detailsTab() == idRecentFiles && a.fileOptionEnabled(idRecentFiles) {
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
	drawText(item.dc, filepath.Base(p.files[item.itemID]), bounds, a.normalFont, color, 0x8000)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}
