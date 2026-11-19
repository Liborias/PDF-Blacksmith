package main

import "unsafe"

var (
	dragMonitorFromPoint = user32.NewProc("MonitorFromPoint")
	dragGetMonitorInfo   = user32.NewProc("GetMonitorInfoW")
	dragSetWindowPos     = user32.NewProc("SetWindowPos")
)

// A drag preview owns its window and accepts artwork from the drag source.
// The capture and placement of the item itself belong to the caller.
type dragPreview struct {
	handle    uintptr
	size, pos point
	offset    int32
}

func (p *dragPreview) show(owner uintptr, cursor, size point, offset, corner int32, paint func(dc uintptr)) {
	dc, _, _ := createSurfaceDC.Call(0)
	if dc == 0 {
		return
	}
	defer deleteSurfaceDC.Call(dc)
	header := bitmapHeader{size: 40, width: size.x, height: -size.y, planes: 1, bits: 32}
	var data *byte
	bitmap, _, _ := gdi32.NewProc("CreateDIBSection").Call(dc, uintptr(unsafe.Pointer(&header)), 0,
		uintptr(unsafe.Pointer(&data)), 0, 0)
	if bitmap == 0 {
		return
	}
	defer gdiDelete.Call(bitmap)
	previous, _, _ := gdiSelect.Call(dc, bitmap)
	if previous == 0 || previous == ^uintptr(0) {
		return
	}
	defer gdiSelect.Call(dc, previous)
	paint(dc)
	gdi32.NewProc("GdiFlush").Call()
	pixels := unsafe.Slice(data, int(size.x*size.y*4))
	radius := min(corner, min(size.x, size.y)/2)
	for y := range size.y {
		dy := max(0, radius-y-1, y-(size.y-radius))
		for x := range size.x {
			dx := max(0, radius-x-1, x-(size.x-radius))
			at := int(4 * (y*size.x + x))
			if dx*dx+dy*dy > radius*radius {
				clear(pixels[at : at+4])
			} else {
				pixels[at+3] = 255
			}
		}
	}

	if p.handle == 0 {
		// The owned popup follows the pointer without taking focus or hit tests.
		const exStyle = 0x00000020 | 0x00000080 | 0x00080000 | 0x08000000
		class := wideString("STATIC")
		p.handle, _, _ = createWindow.Call(exStyle, uintptr(unsafe.Pointer(class)), 0,
			0x80000000, 0, 0, 0, 0, owner, 0, 0, 0)
		if p.handle == 0 {
			return
		}
	}
	pos := dragPreviewPosition(cursor, size, offset)
	source := point{}
	blend := [4]byte{0, 0, 232, 1} // AC_SRC_OVER with source alpha.
	ok, _, _ := user32.NewProc("UpdateLayeredWindow").Call(p.handle, 0, uintptr(unsafe.Pointer(&pos)),
		uintptr(unsafe.Pointer(&size)), dc, uintptr(unsafe.Pointer(&source)), 0,
		uintptr(unsafe.Pointer(&blend)), 2) // ULW_ALPHA
	if ok == 0 {
		p.close()
		return
	}
	p.size, p.pos, p.offset = size, pos, offset
	showWindow.Call(p.handle, 4) // SW_SHOWNOACTIVATE
}

func dragPreviewPosition(cursor, size point, offset int32) point {
	pos := point{cursor.x + offset, cursor.y + offset}
	packed := uintptr(uint32(cursor.x)) | uintptr(uint32(cursor.y))<<32
	monitor, _, _ := dragMonitorFromPoint.Call(packed, 2) // MONITOR_DEFAULTTONEAREST
	info := struct {
		size          uint32
		monitor, work rect
		flags         uint32
	}{}
	info.size = uint32(unsafe.Sizeof(info))
	if ok, _, _ := dragGetMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok != 0 {
		bounds := info.monitor
		if pos.x+size.x > bounds.right {
			pos.x = cursor.x - offset - size.x
		}
		if pos.y+size.y > bounds.bottom {
			pos.y = cursor.y - offset - size.y
		}
		pos.x = max(bounds.left, min(pos.x, bounds.right-size.x))
		pos.y = max(bounds.top, min(pos.y, bounds.bottom-size.y))
	}
	return pos
}

func (p *dragPreview) move(cursor point) {
	if p.handle == 0 {
		return
	}
	pos := dragPreviewPosition(cursor, p.size, p.offset)
	if pos != p.pos {
		dragSetWindowPos.Call(p.handle, 0, uintptr(pos.x), uintptr(pos.y), 0, 0, 0x0015)
		p.pos = pos
	}
}

func (p *dragPreview) close() {
	if p.handle != 0 {
		user32.NewProc("DestroyWindow").Call(p.handle)
		*p = dragPreview{}
	}
}
