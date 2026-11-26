package main

func (a *app) uiText(key uiMessage) string {
	return a.uiLocale.text(key)
}

func (a *app) setUIMessage(handle uintptr, key uiMessage, args ...any) {
	if handle == 0 {
		return
	}
	if key == "" {
		setText(handle, "")
		return
	}
	if a.uiMessages == nil {
		a.uiMessages = make(map[uintptr]localizedMessage)
	}
	message := localizedMessage{key: key, args: args}
	a.uiMessages[handle] = message
	writeControlText(handle, multiline(a.uiLocale.formatMessage(message)))
}

func (a *app) reportUI(key uiMessage, args ...any) {
	a.setUIMessage(a.status, key, args...)
	a.setUIMessage(a.activeReportField(), key, args...)
}

func (a *app) bindUILabel(handle uintptr, key uiMessage) {
	if handle == 0 {
		return
	}
	if a.uiLabels == nil {
		a.uiLabels = make(map[uintptr]uiMessage)
	}
	a.uiLabels[handle] = key
	setText(handle, a.uiText(key))
}

func (a *app) shellButton(key uiMessage, id, width int, style uintptr) uintptr {
	handle := a.uiButton(key, id, 0, 0, width, style)
	a.setTooltip(handle, a.uiText(key))
	return handle
}

func (a *app) measureToolNames() {
	a.preferredToolsWidth = a.scale(168)
	for _, tab := range workspaceTabs {
		a.preferredToolsWidth = max(a.preferredToolsWidth, a.captionWidth(a.uiText(tab.name), a.labelFont)+a.scale(64))
	}
}

func (a *app) setUILanguage(locale uiLocale) {
	if a.uiLocale == locale {
		return
	}
	a.hidePropertyHelp(false)
	a.uiLocale = locale
	for handle, key := range a.uiLabels {
		caption := a.uiText(key)
		setText(handle, caption)
		if _, hasTooltip := a.tooltipCaptions[handle]; hasTooltip {
			a.setTooltip(handle, caption)
		}
	}
	if a.tooltip != 0 {
		sendMessage.Call(a.tooltip, 0x0400+28, 0, 0) // TTM_POP
	}
	a.localizeCommandBar()
	a.localizeReaderHelp()
	a.localizeProperties()
	a.updateObjectActions()
	a.updateComparisonNavigation()
	a.measureToolNames()
	a.layout()
}
