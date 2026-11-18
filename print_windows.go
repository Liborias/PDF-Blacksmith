package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// PRINTDLGW on Windows x64.
type printerDialog struct {
	size                                             uint32
	owner, devMode, devNames, dc                     uintptr
	flags                                            uint32
	from, to, minimum, maximum, copies               uint16
	instance, data, printHook, setupHook             uintptr
	printTemplate, setupTemplate, printMem, setupMem uintptr
}

type printDocInfo struct {
	size                   int32
	name, output, dataType *uint16
	flags                  uint32
}

type printPage struct {
	number, rotation int
	size             pageSize
}

func (a *app) printDocument() {
	v := a.view
	if a.busy || !v.ready() {
		return
	}
	const (
		pdSelection       = 0x00000001
		pdPageNums        = 0x00000002
		pdNoSelection     = 0x00000004
		pdReturnDC        = 0x00000100
		pdDriverCopies    = 0x00040000
		pdHidePrintToFile = 0x00100000
	)
	maximum := uint16(min(len(v.order), 65535))
	dialog := printerDialog{
		owner: a.window, flags: pdReturnDC | pdDriverCopies | pdHidePrintToFile,
		minimum: 1, maximum: maximum, from: 1, to: maximum, copies: 1,
	}
	if len(v.pageSelection) == 0 {
		dialog.flags |= pdNoSelection
	}
	dialog.size = uint32(unsafe.Sizeof(dialog))
	ok, _, _ := comdlg32.NewProc("PrintDlgW").Call(uintptr(unsafe.Pointer(&dialog)))
	if dialog.devMode != 0 {
		kernel32.NewProc("GlobalFree").Call(dialog.devMode)
	}
	if dialog.devNames != 0 {
		kernel32.NewProc("GlobalFree").Call(dialog.devNames)
	}
	if ok == 0 {
		if code, _, _ := dialogError.Call(); code != 0 {
			message := fmt.Sprintf("Tiskový dialog se nepodařilo otevřít (%#x).", code)
			if code == 0x1007 || code == 0x1008 {
				message = "Windows nemá k dispozici tiskárnu. Přidejte ji v nastavení tiskáren systému."
			}
			notice(a.window, message, 0x30)
		}
		return
	}
	if dialog.dc == 0 {
		notice(a.window, "Vybranou tiskárnu se nepodařilo otevřít.", 0x10)
		return
	}
	first, last := 1, len(v.order)
	if dialog.flags&pdPageNums != 0 {
		first, last = int(dialog.from), int(dialog.to)
	}
	if first < 1 || last < first || last > len(v.order) {
		gdi32.NewProc("DeleteDC").Call(dialog.dc)
		notice(a.window, "Rozsah stran k tisku je neplatný.", 0x30)
		return
	}
	numbers := v.order[first-1 : last]
	if dialog.flags&pdSelection != 0 && len(v.pageSelection) > 0 {
		numbers = v.selectedSourcePages()
	}
	pages := make([]printPage, 0, len(numbers))
	for _, number := range numbers {
		pages = append(pages, printPage{number, v.rotations[number-1], v.pageDimensions(number, "crop")})
	}
	req := v.request("print_page")
	req.View.Box = "crop"
	original := v.original
	if a.beginJob(req, "Připravuji tisk…") {
		a.launchJob(func() response { return spoolPDF(dialog.dc, req, pages, original) })
	} else {
		gdi32.NewProc("DeleteDC").Call(dialog.dc)
	}
}

func spoolPDF(dc uintptr, req request, pages []printPage, original os.FileInfo) response {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer gdi32.NewProc("DeleteDC").Call(dc)
	getCap := gdi32.NewProc("GetDeviceCaps")
	capability := func(index uintptr) float64 {
		n, _, _ := getCap.Call(dc, index)
		return float64(int32(n))
	}
	areaW, areaH := capability(8), capability(10) // HORZRES, VERTRES
	dpiX, dpiY := capability(88), capability(90)  // LOGPIXELSX, LOGPIXELSY
	if min(areaW, areaH, dpiX, dpiY) <= 0 {
		return operationError("Tiskárna nevrátila platnou velikost tisknutelné plochy.")
	}
	info := printDocInfo{name: wideString(filepath.Base(req.Input))}
	info.size = int32(unsafe.Sizeof(info))
	if ok, _, _ := gdi32.NewProc("StartDocW").Call(dc, uintptr(unsafe.Pointer(&info))); int32(ok) <= 0 {
		return operationError("Tiskovou úlohu se nepodařilo zahájit.")
	}
	complete := false
	defer func() {
		if !complete {
			gdi32.NewProc("AbortDoc").Call(dc)
		}
	}()
	startPage, endPage := gdi32.NewProc("StartPage"), gdi32.NewProc("EndPage")
	engine := &readerEngine{}
	defer engine.close()
	for _, page := range pages {
		if !unchanged(req.Input, original) {
			return operationError("Vstupní PDF se při tisku změnilo. Otevřete je znovu.")
		}
		fit := min(areaW/dpiX/page.size.width, areaH/dpiY/page.size.height)
		w := max(1, int(math.Round(page.size.width*fit*dpiX)))
		h := max(1, int(math.Round(page.size.height*fit*dpiY)))
		target := max(page.size.width, page.size.height) * fit * min(300, dpiX, dpiY)
		req.View.Page, req.View.Rotation, req.View.Pixels = page.number, page.rotation, min(4096, max(400, int(math.Ceil(target))))
		result := engine.run(req)
		if result.Status != "ok" {
			return result
		}
		if result.View == nil {
			return operationError("Tiskový obraz stránky chybí.")
		}
		bitmap, err := decodePrintBitmap(result.View.Image)
		if err != nil {
			return operationError("Tiskový obraz stránky se nepodařilo načíst.")
		}
		if ok, _, _ := startPage.Call(dc); int32(ok) <= 0 {
			return operationError("Tiskárna nepřijala další stránku.")
		}
		gdi32.NewProc("SetStretchBltMode").Call(dc, 4) // HALFTONE
		gdi32.NewProc("SetBrushOrgEx").Call(dc, 0, 0, 0)
		if drawBitmap(dc, bitmap, (int(areaW)-w)/2, (int(areaH)-h)/2, w, h) <= 0 {
			return operationError("Stránku se nepodařilo vykreslit na tiskárnu.")
		}
		if ok, _, _ := endPage.Call(dc); int32(ok) <= 0 {
			return operationError("Tiskárna nedokončila aktuální stránku.")
		}
	}
	if ok, _, _ := gdi32.NewProc("EndDoc").Call(dc); int32(ok) <= 0 {
		return operationError("Tiskovou úlohu se nepodařilo dokončit.")
	}
	complete = true
	return response{Status: "ok", Message: fmt.Sprintf("Odesláno do tiskové fronty. Počet stran %d.", len(pages))}
}
