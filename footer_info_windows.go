package main

type footerInfoControls struct {
	mode, dirty, tool uintptr
	unsaved bool
}

func (a *app) buildFooterInfo() {
	for _, handle := range []*uintptr{&a.footerInfo.mode, &a.footerInfo.dirty, &a.footerInfo.tool} {
		*handle = a.add("STATIC", "", 0, 0, 0, 1, 22, 0x4000, 0, a.normalFont)
	}
}

func (a *app) footerToolCaption() string {
	if a.layoutAids.temporary || a.mode == idLayoutAids {
		return a.uiText(uiLayoutAids)
	}
	for _, tab := range workspaceTabs {
		if tab.id == a.tool {
			name := a.uiText(tab.name)
			switch {
			case a.tool == idFileTab && a.mode == idExportTab:
				return name + " → " + a.uiText(uiExport)
			case a.tool == idFileTab && a.mode == idSignTab:
				return name + " → " + a.uiText(uiSignature)
			case a.tool == idBrowseTab && a.mode == idPagesTab:
				return name + " → " + a.uiText(uiPages)
			}
			return name
		}
	}
	return ""
}

func (a *app) refreshFooterInfo() {
	p := &a.footerInfo
	if p.mode == 0 || a.view == nil {
		return
	}
	v := a.view
	mode, dirty := "", ""
	p.unsaved = v.ready() && (v.changed() || a.inlineDirty())
	if v.ready() {
		key := uiPageByPage
		if v.overview {
			key = uiPageOverview
		} else if v.continuous {
			key = uiContinuous
		}
		mode = a.uiText(key)
		key = uiFooterNoChanges
		if p.unsaved {
			key = uiFooterUnsaved
		}
		dirty = a.uiText(key)
	}
	for _, entry := range []struct{ handle uintptr; label string }{
		{p.mode, mode}, {p.dirty, dirty}, {p.tool, a.footerToolCaption()},
	} {
		if getText(entry.handle) != entry.label {
			setText(entry.handle, entry.label)
			invalidate(entry.handle)
		}
	}
}

func (a *app) layoutFooterInfo() {
	p := &a.footerInfo
	if a.full {
		for _, handle := range []uintptr{p.mode,p.dirty,p.tool,a.focusHelp,a.status} {
			showControl(handle,false)
		}
		return
	}
	a.refreshFooterInfo()
	d := a.scale
	x, y := d(628), a.height-d(33)
	available := max(0,a.width-x-d(13))
	measure := func(keys ...uiMessage) int {
		width := 0
		for _, key := range keys {
			width = max(width,a.captionWidth(a.uiText(key),a.normalFont))
		}
		return width+d(10)
	}
	mode := measure(uiPageByPage,uiPageOverview,uiContinuous)
	dirty := measure(uiFooterNoChanges,uiFooterUnsaved)
	// The active tool can change without a resize; give it a stable slot.
	tool := d(172)
	widths := fitFooterSlots(available,d(8),mode,dirty,tool,d(70),d(140),d(218))
	for _, entry := range []struct{ handle uintptr; width int }{
		{p.mode,widths.mode},{a.focusHelp,widths.help},{p.dirty,widths.dirty},
		{p.tool,widths.tool},{a.status,widths.status},
	} {
		show := entry.width > 0
		if show {
			place(entry.handle,x,y,entry.width,d(22))
			x += entry.width+d(8)
		}
		showControl(entry.handle,show)
	}
}
