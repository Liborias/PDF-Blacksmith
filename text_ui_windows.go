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
	idTextNewFrame
)

var textIcons = map[int][][]point{
	idHeaderUndo:       {{{6, 3}, {2, 7}, {6, 11}}, {{2, 7}, {12, 7}, {15, 9}, {15, 13}, {12, 16}}},
	idHeaderRedo:       {{{12, 3}, {16, 7}, {12, 11}}, {{16, 7}, {6, 7}, {3, 9}, {3, 13}, {6, 16}}},
	idTextSmaller:      {{{1, 15}, {6, 3}, {11, 15}}, {{3, 11}, {9, 11}}, {{12, 8}, {17, 8}}},
	idTextLarger:       {{{1, 15}, {6, 3}, {11, 15}}, {{3, 11}, {9, 11}}, {{12, 8}, {17, 8}}, {{14, 5}, {14, 11}}},
	idTextSizes:        {{{5, 7}, {9, 11}, {13, 7}}},
	idTextWidths:       {{{5, 7}, {9, 11}, {13, 7}}},
	idTextLeadings:     {{{5, 7}, {9, 11}, {13, 7}}},
	idTextTracking:     {{{5, 7}, {9, 11}, {13, 7}}},
	idTextRegular:      {{{5, 16}, {5, 2}, {10, 2}, {13, 4}, {13, 7}, {10, 9}, {5, 9}}, {{9, 9}, {14, 16}}},
	idTextBold:         {{{5, 16}, {5, 2}, {11, 2}, {14, 4}, {14, 6}, {11, 9}, {5, 9}}, {{11, 9}, {15, 11}, {15, 14}, {12, 16}, {5, 16}}},
	idTextItalic:       {{{6, 2}, {15, 2}}, {{10, 2}, {7, 16}}, {{3, 16}, {12, 16}}},
	idTextUnderline:    {{{4, 2}, {4, 9}, {6, 12}, {11, 12}, {14, 9}, {14, 2}}, {{3, 16}, {15, 16}}},
	idTextAlignLeft:    {{{2, 3}, {16, 3}}, {{2, 7}, {11, 7}}, {{2, 11}, {16, 11}}, {{2, 15}, {10, 15}}},
	idTextAlignCenter:  {{{2, 3}, {16, 3}}, {{5, 7}, {13, 7}}, {{2, 11}, {16, 11}}, {{5, 15}, {13, 15}}},
	idTextAlignRight:   {{{2, 3}, {16, 3}}, {{7, 7}, {16, 7}}, {{2, 11}, {16, 11}}, {{8, 15}, {16, 15}}},
	idTextAlignJustify: {{{2, 3}, {16, 3}}, {{2, 7}, {16, 7}}, {{2, 11}, {16, 11}}, {{2, 15}, {16, 15}}},
	idTextCopy:         {{{6, 6}, {16, 6}, {16, 16}, {6, 16}, {6, 6}}, {{2, 12}, {2, 2}, {12, 2}}},
	idTextUndo:         {{{5, 2}, {2, 6}, {5, 10}}, {{2, 6}, {12, 6}, {16, 9}, {16, 13}, {13, 16}, {8, 16}}},
	idTextOCR: {
		{{1, 5}, {1, 1}, {5, 1}}, {{13, 1}, {17, 1}, {17, 5}}, {{17, 13}, {17, 17}, {13, 17}}, {{5, 17}, {1, 17}, {1, 13}},
		{{5, 13}, {9, 5}, {13, 13}}, {{7, 10}, {11, 10}},
	},
}

var selectionIcons = map[int]string{
	idTextSelectObject:    "i-text_caret",
	idTextSelectRectangle: "i-rect_select",
	idTextSelectLasso:     "i-lasso",
	idTextNewFrame:        "i-text_box",
	idObjectPoint:         "i-cursor",
	idObjectRectangle:     "i-rect_select",
	idObjectLasso:         "i-lasso",
	idObjectCrop:          "i-image_crop",
}

func (a *app) drawTextIcon(item *buttonDraw) bool {
	id := int(item.controlID)
	paths, symbol := textIcons[id], selectionIcons[id]
	if len(paths) == 0 && symbol == "" {
		return false
	}
	base := panelColor
	if id == idHeaderUndo || id == idHeaderRedo {
		base = titleColor
	}
	fillColor(item.dc, item.bounds, base)
	color, background, edge := textColor, inputColor, divider
	pressed := item.state&1 != 0
	hovered := symbol != "" && a.navigationHot(item)
	switch {
	case item.state&4 != 0:
		color, background = disabledColor, panelColor
	case a.checked(item.handle):
		background, edge = uiColors["toolbarActive"], uiColors["accentStrong"]
		if pressed || hovered {
			background = uiColors["toolbarActiveHover"]
		}
	case pressed:
		background = divider
	case hovered:
		background = uiColors["hover"]
	}
	rounded(item.dc, item.bounds, a.scale(5), background, edge)
	var mixedMask uint32
	switch id {
	case idTextRegular:
		mixedMask = inlineBold | inlineItalic
	case idTextBold:
		mixedMask = inlineBold
	case idTextItalic:
		mixedMask = inlineItalic
	case idTextUnderline:
		mixedMask = inlineLine
	}
	if item.state&4 == 0 && a.textTool.inline.mixedFormat&mixedMask != 0 {
		right, bottom := item.bounds.right-int32(a.scale(3)), item.bounds.bottom-int32(a.scale(3))
		dot := int32(max(3, a.scale(4)))
		rounded(item.dc, rect{right - dot, bottom - dot, right, bottom}, int(dot), selectionColor, selectionColor)
	}
	size := int32(a.scale(18))
	bounds := item.bounds
	x := (bounds.left + bounds.right - size) / 2
	y := (bounds.top + bounds.bottom - size) / 2
	if symbol != "" {
		drawUIIcon(item.dc, symbol, int(x), int(y), int(size), color, background)
	} else {
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
	}
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
