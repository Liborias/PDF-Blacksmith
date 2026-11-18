package main

import "unsafe"

type choiceControl struct {
	handle       uintptr
	notification int
	names        []string
	index        int
}

type menuMeasure struct {
	controlType, controlID, itemID, width, height uint32
	data                                          uintptr
}

type popupItem struct {
	label, shortcut              string
	checked, disabled, separator bool
	command                      int
}

func (a *app) combo(id, x, y, width int, names []string) uintptr {
	command := 2000 + len(a.choices)
	handle := a.add("BUTTON", names[0], command, x, y, width, 28, wsTabstop|bsOwnerDraw, 0, a.normalFont)
	choice := &choiceControl{handle: handle, notification: id, names: names}
	a.choices[handle], a.choiceCommands[command] = choice, choice
	return handle
}

func comboIndex(handle uintptr) int {
	if choice := window.choices[handle]; choice != nil {
		return choice.index
	}
	return -1
}

func chooseCombo(handle uintptr, index int) {
	choice := window.choices[handle]
	if choice == nil || index < 0 || index >= len(choice.names) || index == choice.index {
		return
	}
	choice.index = index
	setText(handle, choice.names[index])
	invalidate(handle)
}

func (a *app) drawChoice(item *buttonDraw) bool {
	choice := a.choices[item.handle]
	if choice == nil {
		return false
	}
	bounds := item.bounds
	color := inputColor
	if item.state&1 != 0 {
		color = paletteColor{38, 35, 40}
	}
	fillColor(item.dc, bounds, a.background(item.handle))
	rounded(item.dc, bounds, a.scale(5), color, divider)
	text := textColor
	if item.state&4 != 0 {
		text = mutedColor
	}
	x, y := int(bounds.right)-a.scale(13), int(bounds.top+bounds.bottom)/2
	a.arrow(item.dc, x, y, text)
	bounds.left += int32(a.scale(8))
	bounds.right -= int32(a.scale(24))
	drawText(item.dc, choice.names[choice.index], bounds, a.normalFont, text, 0x8000)
	if item.state&16 != 0 {
		bounds.left -= 1
		bounds.top += 3
		bounds.bottom -= 3
		user32.NewProc("DrawFocusRect").Call(item.dc, uintptr(unsafe.Pointer(&bounds)))
	}
	return true
}

func (a *app) background(handle uintptr) paletteColor {
	if a.panelButtons[handle] {
		return panelColor
	}
	return canvasColor
}

func (a *app) arrow(dc uintptr, x, y int, color paletteColor) {
	pen, _, _ := gdiPen.Call(0, 1, color.win())
	old, _, _ := gdiSelect.Call(dc, pen)
	size := max(3, a.scale(4))
	gdi32.NewProc("MoveToEx").Call(dc, uintptr(x-size), uintptr(y-2), 0)
	gdi32.NewProc("LineTo").Call(dc, uintptr(x), uintptr(y+2))
	gdi32.NewProc("LineTo").Call(dc, uintptr(x+size+1), uintptr(y-2))
	gdiSelect.Call(dc, old)
	gdiDelete.Call(pen)
}

func (a *app) openChoice(choice *choiceControl) {
	items := make([]popupItem, len(choice.names))
	for index, name := range choice.names {
		items[index] = popupItem{label: name, checked: index == choice.index}
	}
	if index := a.trackPopup(choice.handle, 302, items); index >= 0 && index != choice.index {
		chooseCombo(choice.handle, index)
		a.command(choice.notification, 1)
	}
}

func (a *app) trackPopup(anchorHandle uintptr, width int, items []popupItem) int {
	menu, _, _ := user32.NewProc("CreatePopupMenu").Call()
	if menu == 0 {
		return -1
	}
	defer user32.NewProc("DestroyMenu").Call(menu)
	appendItem := user32.NewProc("AppendMenuW")
	for index, entry := range items {
		item := uintptr(index + 1)
		flags := uintptr(0x100) // MF_OWNERDRAW
		if entry.disabled || entry.separator {
			flags |= 1 // MF_GRAYED
		}
		appendItem.Call(menu, flags, item, item)
	}
	anchor := rect{}
	user32.NewProc("GetWindowRect").Call(anchorHandle, uintptr(unsafe.Pointer(&anchor)))
	a.popup, a.popupWidth = items, width
	selected, _, _ := user32.NewProc("TrackPopupMenuEx").Call(menu, 0x182, uintptr(anchor.left), uintptr(anchor.bottom+2), a.window, 0)
	a.popup = nil
	if index := int(selected) - 1; index >= 0 && index < len(items) && !items[index].disabled && !items[index].separator {
		return index
	}
	return -1
}

func (a *app) measurePopup(param uintptr) bool {
	if param == 0 || a.popup == nil {
		return false
	}
	var item menuMeasure
	copyDrawItem.Call(uintptr(unsafe.Pointer(&item)), param, unsafe.Sizeof(item))
	if item.controlType != 1 || item.data == 0 || item.data > uintptr(len(a.popup)) {
		return false
	}
	item.width = uint32(a.scale(a.popupWidth))
	item.height = uint32(a.scale(31))
	if a.popup[item.data-1].separator {
		item.height = uint32(a.scale(9))
	}
	copyDrawItem.Call(param, uintptr(unsafe.Pointer(&item)), unsafe.Sizeof(item))
	return true
}

func (a *app) drawPopup(item *buttonDraw) bool {
	if item.controlType != 1 || item.itemData == 0 || item.itemData > uintptr(len(a.popup)) {
		return false
	}
	entry := a.popup[item.itemData-1]
	bounds := item.bounds
	color := panelColor
	if item.state&1 != 0 && !entry.disabled && !entry.separator {
		color = paletteColor{67, 43, 56}
	}
	fillColor(item.dc, bounds, color)
	if entry.separator {
		y := (bounds.top + bounds.bottom) / 2
		fillColor(item.dc, rect{bounds.left + int32(a.scale(10)), y, bounds.right - int32(a.scale(10)), y + 1}, divider)
		return true
	}
	if entry.checked {
		box := rect{bounds.left + 5, bounds.top + 5, bounds.left + 23, bounds.bottom - 5}
		drawCheck(item.dc, box)
	}
	bounds.left += int32(a.scale(30))
	bounds.right -= int32(a.scale(7))
	foreground := textColor
	if entry.disabled {
		foreground = paletteColor{119, 114, 122}
	}
	if entry.shortcut != "" {
		shortcutColor := mutedColor
		if entry.disabled {
			shortcutColor = foreground
		}
		drawText(item.dc, entry.shortcut, bounds, a.normalFont, shortcutColor, 2)
		bounds.right -= int32(a.scale(112))
	}
	drawText(item.dc, entry.label, bounds, a.normalFont, foreground, 0x8000)
	return true
}
