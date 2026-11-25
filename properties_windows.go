package main

const (
	idPropertyFirst         = 600
	wmPropertyReadoutLayout = 0x801b
)

type propertySection struct {
	header, help       uintptr
	title, description string
	detail             func() string
	collapsed          bool
}

func (a *app) beginPropertySection(title string, y int, description string) *propertySection {
	a.addingSection = 0
	id := idPropertyFirst + 2*len(a.propertySections)
	header := a.add("BUTTON", title+" – sbalit", id, 16, y, 265, 22, wsTabstop|bsOwnerDraw, 0, a.labelFont)
	help := a.add("BUTTON", "Nápověda – "+title, id+1, 290, y, 24, 22, wsTabstop|bsOwnerDraw, 0, a.normalFont)
	a.addTextTooltip(help, description)
	watchNavigationHover(header)
	watchNavigationHover(help)
	a.propertySections = append(a.propertySections, propertySection{header: header, help: help, title: title, description: description})
	a.addingSection = len(a.propertySections)
	return &a.propertySections[a.addingSection-1]
}

func (a *app) updatePropertySectionCaption(section *propertySection) {
	action := "sbalit"
	if section.collapsed {
		action = "rozbalit"
	}
	setText(section.header, section.title+" – "+action)
}

func (a *app) propertySectionCommand(id int) bool {
	index := (id - idPropertyFirst) / 2
	if id < idPropertyFirst || index >= len(a.propertySections) {
		return false
	}
	section := &a.propertySections[index]
	if (id-idPropertyFirst)%2 == 1 {
		body := section.description
		if section.detail != nil {
			if detail := section.detail(); detail != "" {
				body += "\n\n" + detail
			}
		}
		notice(a.window, body, 0x40)
	} else {
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
		chevron := "⌄"
		if section.collapsed {
			chevron = "›"
		}
		icon := bounds
		icon.right = min(icon.right, icon.left+int32(d(17)))
		drawText(item.dc, chevron, icon, a.normalFont, selectionColor, 1)
		caption := bounds
		caption.left = icon.right
		drawText(item.dc, section.title, caption, a.labelFont, textColor, 0x8000)
		fillColor(item.dc, rect{bounds.left, bounds.bottom - 1, bounds.right, bounds.bottom}, divider)
	} else {
		size := min(int32(d(18)), min(bounds.right-bounds.left, bounds.bottom-bounds.top))
		x, y := (bounds.left+bounds.right-size)/2, (bounds.top+bounds.bottom-size)/2
		circle := rect{x, y, x + size, y + size}
		fill := panelColor
		if a.navigationHot(item) {
			fill = uiColors["hover"]
		}
		rounded(item.dc, circle, int(size), fill, divider)
		drawText(item.dc, "?", circle, a.normalFont, textColor, 1)
	}
	if item.state&16 != 0 {
		a.drawFocus(item.dc, bounds)
	}
	return true
}

func (a *app) visiblePropertyItems(tab int) []panelItem {
	items := a.panelItems[tab]
	visible := make([]panelItem, 0, len(items))
	for index, item := range items {
		item, shown := a.panelItemLayout(tab, index, item)
		if shown {
			if a.readOnlyFields[item.handle] {
				// Reports remain outside folding and take space only when they have content.
				shown = getText(item.handle) != ""
			} else if item.section > 0 {
				shown = !a.propertySections[item.section-1].collapsed
			}
		}
		if shown {
			visible = append(visible, item)
		}
	}
	if tab != idFileDetails {
		compactPropertyRows(visible)
	}
	return visible
}
