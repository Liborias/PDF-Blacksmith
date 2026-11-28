package main

import "unsafe"

const (
	idPropertyFirst         = 600
	wmPropertyReadoutLayout = 0x801b
)

type propertySection struct {
	header, help       uintptr
	title, description uiMessage
	instructions       uiMessage
	detail             func() string
	collapsed          bool
}

func (a *app) beginPropertySection(title uiMessage, y int, description uiMessage) *propertySection {
	a.addingSection = 0
	id := idPropertyFirst + 2*len(a.propertySections)
	header := a.add("BUTTON", "", id, 16, y, 265, 22, wsTabstop|bsOwnerDraw, 0, a.labelFont)
	help := a.add("BUTTON", "", id+1, 290, y, 24, 22, wsTabstop|bsOwnerDraw, 0, a.normalFont)
	watchNavigationHover(header)
	watchNavigationHover(help)
	a.propertySections = append(a.propertySections, propertySection{header: header, help: help, title: title, description: description})
	a.addingSection = len(a.propertySections)
	section := &a.propertySections[a.addingSection-1]
	a.localizePropertySection(section)
	return section
}

func (a *app) updatePropertySectionCaption(section *propertySection) {
	action := uiCollapseSection
	if section.collapsed {
		action = uiExpandSection
	}
	setText(section.header, a.uiText(section.title)+" – "+a.uiText(action))
}

func (a *app) localizePropertySection(section *propertySection) {
	a.updatePropertySectionCaption(section)
	setText(section.help, a.uiText(uiHelp)+" – "+a.uiText(section.title))
	a.setTooltip(section.help, a.uiText(uiHelp)+" – "+a.uiText(section.title))
}

func (a *app) propertySectionCommand(id int) bool {
	index := (id - idPropertyFirst) / 2
	if id < idPropertyFirst || index >= len(a.propertySections) {
		return false
	}
	section := &a.propertySections[index]
	if (id-idPropertyFirst)%2 == 1 {
		a.togglePropertyHelp(section.help, section)
	} else {
		a.hidePropertyHelp(false)
		section.collapsed = !section.collapsed
		a.updatePropertySectionCaption(section)
		a.layoutPanel()
		a.revealFocusedControl()
	}
	return true
}

func (a *app) revealPropertyControl(handle uintptr) {
	for _, item := range a.panelItems[a.detailsTab()] {
		if item.handle == handle && item.section > 0 {
			section := &a.propertySections[item.section-1]
			if section.collapsed {
				section.collapsed = false
				a.updatePropertySectionCaption(section)
				a.layoutPanel()
			}
			return
		}
	}
}

func (a *app) drawPropertySectionButton(item *buttonDraw) bool {
	index := (int(item.controlID) - idPropertyFirst) / 2
	if int(item.controlID) < idPropertyFirst || index >= len(a.propertySections) {
		return false
	}
	section := &a.propertySections[index]
	if item.handle != section.header && item.handle != section.help {
		return false
	}
	bounds := item.bounds
	d := a.scale
	fillColor(item.dc, bounds, panelColor)
	if item.handle == section.header {
		if a.navigationHot(item) {
			fillColor(item.dc, bounds, uiColors["hover"])
		}
		size := int32(d(12))
		y := (bounds.top+bounds.bottom-size)/2 - int32(d(2))
		icon := rect{bounds.right - int32(d(16)), y, bounds.right - int32(d(16)) + size, y + size}
		vertices := []point{{4, 7}, {9, 12}, {14, 7}}
		if section.collapsed {
			vertices = []point{{7, 4}, {12, 9}, {7, 14}}
		}
		for i, p := range vertices {
			vertices[i] = point{icon.left + p.x*size/18, icon.top + p.y*size/18}
		}
		pen, _, _ := gdiPen.Call(0, 1, selectionColor.win())
		old, _, _ := gdiSelect.Call(item.dc, pen)
		gdi32.NewProc("Polyline").Call(item.dc, uintptr(unsafe.Pointer(&vertices[0])), uintptr(len(vertices)))
		gdiSelect.Call(item.dc, old)
		gdiDelete.Call(pen)
		caption := bounds
		caption.right = icon.left - int32(d(2))
		drawText(item.dc, a.uiText(section.title), caption, a.labelFont, textColor, 0x8000)
		fillColor(item.dc, rect{bounds.left, bounds.bottom - 1, bounds.right, bounds.bottom}, divider)
	} else {
		size := min(int32(d(18)), min(bounds.right-bounds.left, bounds.bottom-bounds.top))
		x, y := (bounds.left+bounds.right-size)/2, (bounds.top+bounds.bottom-size)/2
		circle := rect{x, y, x + size, y + size}
		fill := panelColor
		edge := divider
		if a.propertyHelp.anchor == section.help {
			edge = selectionColor
		}
		if a.navigationHot(item) {
			fill = uiColors["hover"]
		}
		rounded(item.dc, circle, int(size), fill, edge)
		drawText(item.dc, "?", circle, a.normalFont, textColor, 1)
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
	}
	return true
}

func (a *app) uiLabel(key uiMessage, x, y, width int, strong bool) uintptr {
	handle := a.label("", x, y, width, strong)
	a.bindUILabel(handle, key)
	return handle
}

func (a *app) uiButton(key uiMessage, id, x, y, width int, style uintptr) uintptr {
	handle := a.button("", id, x, y, width, style)
	a.bindUILabel(handle, key)
	return handle
}

func (a *app) uiTextIconButton(key uiMessage, id, x, y, width int) uintptr {
	handle := a.uiButton(key, id, x, y, width, 0)
	a.bindUIHelp(handle, key)
	return handle
}

func (a *app) bindUICue(handle uintptr, key uiMessage) {
	if a.uiCues == nil {
		a.uiCues = make(map[uintptr]uiMessage)
	}
	a.uiCues[handle] = key
	a.setUICue(handle, key)
}

func (a *app) setUICue(handle uintptr, key uiMessage) {
	sendMessage.Call(handle, 0x1501, 1, uintptr(unsafe.Pointer(wideString(a.uiText(key))))) // EM_SETCUEBANNER
}

func (a *app) bindUIHelp(handle uintptr, key uiMessage) {
	if a.uiHelpKeys == nil {
		a.uiHelpKeys = make(map[uintptr]uiMessage)
	}
	a.uiHelpKeys[handle] = key
	a.setTooltip(handle, a.uiText(key))
}

func (a *app) localizeProperties() {
	for index := range a.propertySections {
		a.localizePropertySection(&a.propertySections[index])
	}
	for _, choice := range a.choices {
		if len(choice.messages) != 0 || choice.placeholder != "" {
			a.uiLocale.applyChoiceMessages(choice.names, choice.messages)
			setText(choice.handle, a.choiceLabel(choice))
			invalidate(choice.handle)
		}
	}
	for handle, key := range a.uiCues {
		a.setUICue(handle, key)
	}
	for handle, key := range a.uiHelpKeys {
		a.setTooltip(handle, a.uiText(key))
	}
	for handle, message := range a.uiMessages {
		writeControlText(handle, multiline(a.uiLocale.formatMessage(message)))
	}
}

func (a *app) visiblePropertyItems(tab int) []panelItem {
	items := a.panelItems[tab]
	visible := make([]panelItem, 0, len(items))
	for index, item := range items {
		item, shown := a.panelItemLayout(tab, index, item)
		if !shown {
			continue
		}
		if a.readOnlyFields[item.handle] {
			// Reports remain outside folding.
			if getText(item.handle) == "" {
				continue
			}
		} else if item.section > 0 && a.propertySections[item.section-1].collapsed {
			continue
		}
		visible = append(visible, item)
	}
	if tab != idFileDetails {
		headers := make(map[uintptr]bool,2*len(a.propertySections))
		for _, section := range a.propertySections {
			headers[section.header], headers[section.help] = true,true
		}
		compactPropertyRowsWithHeaders(visible,headers)
	}
	return visible
}
