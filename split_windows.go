package main

import (
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

func (a *app) buildSplitPanel() {
	a.beginPropertySection("Strany k rozdělení", 61, "Každá vybraná strana se uloží do samostatného PDF ve zvolené složce.")
	a.splitScope = a.combo(idSplitScope, 16, 86, 145, pageScopeNames)
	a.splitRange = a.edit(169, 86, 145, false)
	sendMessage.Call(a.splitRange, 0x1501, 1, uintptr(unsafe.Pointer(wideString("Např. 1, 3–5, 8"))))
	a.splitButton = a.button("Vybrat složku a rozdělit PDF…", idSplitRun, 16, 134, 298, 0)
	a.primaryButtons[a.splitButton] = true
	a.splitReport = a.readout(16, 184, 298, 170)
}

func (a *app) splitDocument() {
	if a.busy || !a.view.ready() {
		return
	}
	req := a.view.request("split")
	if !a.selectPages(&req, a.splitScope, a.splitRange) {
		return
	}
	req.OutputDir = a.selectFolder(filepath.Dir(a.view.path))
	if req.OutputDir == "" {
		return
	}
	a.view.includeEdits(&req)
	setText(a.splitReport, "")
	a.run(req)
}

type browseInfo struct {
	owner, root, display, title uintptr
	flags                       uint32
	callback, initial           uintptr
	image                       int32
}

var folderCallback = syscall.NewCallback(func(hwnd, message, _ uintptr, initial uintptr) uintptr {
	if message == 1 { // BFFM_INITIALIZED
		sendMessage.Call(hwnd, 0x0467, 1, initial) // BFFM_SETSELECTIONW
	}
	return 0
})

func (a *app) selectFolder(initial string) string {
	com := syscall.NewLazyDLL("ole32.dll")
	result, _, _ := com.NewProc("CoInitializeEx").Call(0, 2) // COINIT_APARTMENTTHREADED
	if int32(result) < 0 {
		a.report("Výběr složky se nepodařilo otevřít.")
		return ""
	}
	defer com.NewProc("CoUninitialize").Call()
	shell := syscall.NewLazyDLL("shell32.dll")
	var display [260]uint16
	title, path := wideString("Vyberte složku pro rozdělené PDF."), wideString(initial)
	info := browseInfo{owner: a.window, display: uintptr(unsafe.Pointer(&display[0])), title: uintptr(unsafe.Pointer(title)),
		flags: 0x41, callback: folderCallback, initial: uintptr(unsafe.Pointer(path))} // BIF_NEWDIALOGSTYLE | BIF_RETURNONLYFSDIRS
	id, _, _ := shell.NewProc("SHBrowseForFolderW").Call(uintptr(unsafe.Pointer(&info)))
	runtime.KeepAlive(display)
	runtime.KeepAlive(title)
	runtime.KeepAlive(path)
	if id == 0 {
		return ""
	}
	defer com.NewProc("CoTaskMemFree").Call(id)
	var buffer [32768]uint16
	ok, _, _ := shell.NewProc("SHGetPathFromIDListEx").Call(id, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0)
	if ok == 0 {
		a.report("Vyberte existující složku v souborovém systému.")
		return ""
	}
	return syscall.UTF16ToString(buffer[:])
}
