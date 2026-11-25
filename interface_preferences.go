package main

import "encoding/json"

type interfacePreferences struct {
	Locale uiLocale `json:"locale"`
	Theme  uiTheme  `json:"theme"`
}

func decodeInterfacePreferences(data []byte, fallbackLocale uiLocale) interfacePreferences {
	settings := interfacePreferences{Locale: fallbackLocale, Theme: uiPalettes.Default}
	var stored interfacePreferences
	if json.Unmarshal(data, &stored) != nil {
		return settings
	}
	if stored.Locale == uiCzech || stored.Locale == uiEnglish {
		settings.Locale = stored.Locale
	}
	if uiPalettes.colors(stored.Theme) != nil {
		settings.Theme = stored.Theme
	}
	return settings
}
