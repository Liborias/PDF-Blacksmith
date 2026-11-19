package main

import (
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	idTextSelectObject = 470 + iota
	idTextSelectParagraph
	idTextSelectRectangle
	idTextSelectLasso
	idTextRegular
	idTextAlignLeft
	idTextAlignCenter
	idTextAlignRight
	idTextAlignJustify
)

type nativeTooltip struct {
	size, flags  uint32
	owner, id    uintptr
	bounds       rect
	instance     uintptr
	text         *uint16
	param, extra uintptr
}

var textIcons = map[int][][]point{
	idTextSelectObject:    {{{4, 2}, {14, 2}, {14, 5}}, {{4, 2}, {4, 5}}, {{9, 2}, {9, 16}}, {{6, 16}, {12, 16}}},
	idTextSelectParagraph: {{{13, 2}, {7, 2}, {4, 4}, {4, 7}, {7, 9}, {9, 9}}, {{9, 2}, {9, 16}}, {{13, 2}, {13, 16}}},
	idTextSelectRectangle: {{{2, 4}, {2, 2}, {5, 2}}, {{8, 2}, {10, 2}}, {{13, 2}, {16, 2}, {16, 4}}, {{16, 7}, {16, 10}}, {{16, 13}, {16, 16}, {13, 16}}, {{10, 16}, {8, 16}}, {{5, 16}, {2, 16}, {2, 13}}, {{2, 10}, {2, 7}}},
	idTextSelectLasso:     {{{2, 9}, {3, 6}, {7, 3}, {12, 2}, {16, 4}, {17, 7}, {15, 10}, {10, 12}, {5, 12}, {2, 9}}, {{3, 10}, {4, 13}, {3, 16}, {6, 17}}},
	idTextRegular:         {{{5, 16}, {5, 2}, {10, 2}, {13, 4}, {13, 7}, {10, 9}, {5, 9}}, {{9, 9}, {14, 16}}},
	idTextBold:            {{{5, 16}, {5, 2}, {11, 2}, {14, 4}, {14, 6}, {11, 9}, {5, 9}}, {{11, 9}, {15, 11}, {15, 14}, {12, 16}, {5, 16}}},
	idTextItalic:          {{{6, 2}, {15, 2}}, {{10, 2}, {7, 16}}, {{3, 16}, {12, 16}}},
	idTextUnderline:       {{{4, 2}, {4, 9}, {6, 12}, {11, 12}, {14, 9}, {14, 2}}, {{3, 16}, {15, 16}}},
	idTextAlignLeft:       {{{2, 3}, {16, 3}}, {{2, 7}, {11, 7}}, {{2, 11}, {16, 11}}, {{2, 15}, {10, 15}}},
	idTextAlignCenter:     {{{2, 3}, {16, 3}}, {{5, 7}, {13, 7}}, {{2, 11}, {16, 11}}, {{5, 15}, {13, 15}}},
	idTextAlignRight:      {{{2, 3}, {16, 3}}, {{7, 7}, {16, 7}}, {{2, 11}, {16, 11}}, {{8, 15}, {16, 15}}},
	idTextAlignJustify:    {{{2, 3}, {16, 3}}, {{2, 7}, {16, 7}}, {{2, 11}, {16, 11}}, {{2, 15}, {16, 15}}},
	idTextApply:           {{{2, 9}, {7, 14}, {16, 4}}},
	idTextCopy:            {{{6, 6}, {16, 6}, {16, 16}, {6, 16}, {6, 6}}, {{2, 12}, {2, 2}, {12, 2}}},
	idTextUndo:            {{{5, 2}, {2, 6}, {5, 10}}, {{2, 6}, {12, 6}, {16, 9}, {16, 13}, {13, 16}, {8, 16}}},
	idTextOCR: {
		{{1, 5}, {1, 1}, {5, 1}}, {{13, 1}, {17, 1}, {17, 5}}, {{17, 13}, {17, 17}, {13, 17}}, {{5, 17}, {1, 17}, {1, 13}},
		{{5, 13}, {9, 5}, {13, 13}}, {{7, 10}, {11, 10}},
	},
}

func (a *app) textIconButton(label string, id, x, y, width int) uintptr {
	h := a.button(label, id, x, y, width, 0)
	p := &a.textTool
	if p.tooltip == 0 {
		p.tooltip, _, _ = createWindow.Call(8, uintptr(unsafe.Pointer(wideString("tooltips_class32"))), 0,
			0x80000003, 0, 0, 0, 0, a.window, 0, a.instance, 0)
	}
	if p.tooltip != 0 {
		name := wideString(label)
		p.tooltipNames = append(p.tooltipNames, name)
		parent, _, _ := user32.NewProc("GetParent").Call(h)
		info := nativeTooltip{flags: 0x11, owner: parent, id: h, text: name} // TTF_SUBCLASS | TTF_IDISHWND
		info.size = uint32(unsafe.Sizeof(info))
		sendMessage.Call(p.tooltip, 0x0400+50, 0, uintptr(unsafe.Pointer(&info)))
	}
	return h
}

func (a *app) drawTextIcon(item *buttonDraw) bool {
	id := int(item.controlID)
	paths := textIcons[id]
	if len(paths) == 0 {
		return false
	}
	fillColor(item.dc, item.bounds, panelColor)
	color, background := textColor, inputColor
	switch {
	case item.state&4 != 0:
		color, background = disabledColor, panelColor
	case a.checked(item.handle):
		color = selectionColor
	}
	if item.state&1 != 0 {
		background = divider
	}
	rounded(item.dc, item.bounds, a.scale(5), background, divider)
	size := int32(a.scale(18))
	bounds := item.bounds
	x := (bounds.left + bounds.right - size) / 2
	y := (bounds.top + bounds.bottom - size) / 2
	pen, _, _ := gdiPen.Call(0, uintptr(max(1, a.scale(2))), color.win())
	old, _, _ := gdiSelect.Call(item.dc, pen)
	for _, path := range paths {
		vertices := make([]point, len(path))
		for i, p := range path {
			vertices[i] = point{x + p.x*size/18, y + p.y*size/18}
		}
		gdi32.NewProc("Polyline").Call(item.dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
	}
	gdiSelect.Call(item.dc, old)
	gdiDelete.Call(pen)
	if item.state&16 != 0 {
		a.drawFocus(item.dc, item.bounds)
	}
	return true
}

var bundledTextFonts = []string{"Liberation Sans", "Liberation Serif", "Liberation Mono"}

func (a *app) installedTextFonts() []string {
	fonts := slices.Clone(bundledTextFonts)
	dc, _, _ := user32.NewProc("GetDC").Call(a.window)
	if dc == 0 {
		return fonts
	}
	defer user32.NewProc("ReleaseDC").Call(a.window, dc)
	// LOGFONTW ends with the 32 UTF-16 code units of the Windows family.
	type fontFilter struct {
		metrics         [5]int32
		characteristics [8]byte
		family          [32]uint16
	}
	filter := fontFilter{}
	filter.characteristics[3] = 1 // DEFAULT_CHARSET
	found := map[string]string{}
	callback := syscall.NewCallback(func(face *fontFilter, _ uintptr, kind uint32, _ uintptr) uintptr {
		if kind&4 != 0 {
			name := strings.TrimSpace(syscall.UTF16ToString(face.family[:]))
			if name != "" && !strings.HasPrefix(name, "@") {
				found[strings.ToLower(name)] = name
			}
		}
		return 1
	})
	gdi32.NewProc("EnumFontFamiliesExW").Call(dc, uintptr(unsafe.Pointer(&filter)), callback, 0, 0)
	for _, name := range bundledTextFonts {
		delete(found, strings.ToLower(name))
	}
	var system []string
	for _, name := range found {
		system = append(system, name)
	}
	slices.SortFunc(system, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	return append(fonts, system...)
}
