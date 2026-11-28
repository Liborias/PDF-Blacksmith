package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type uiLocale string
type uiMessage string

type localizedMessage struct {
	key  uiMessage
	args []any
}

func (locale uiLocale) formatMessage(message localizedMessage) string {
	pattern := locale.text(message.key)
	if len(message.args) == 0 {
		return pattern
	}
	args := make([]any, len(message.args))
	for i, arg := range message.args {
		args[i] = arg
		if key, ok := arg.(uiMessage); ok {
			args[i] = locale.text(key)
		}
	}
	return fmt.Sprintf(pattern, args...)
}

func (locale uiLocale) applyChoiceMessages(names []string, messages []uiMessage) {
	for index, key := range messages {
		if key != "" {
			names[index] = locale.text(key)
		}
	}
}

const (
	uiCzech   uiLocale = "cs-CZ"
	uiEnglish uiLocale = "en-US"
)

const (
	uiFile                 uiMessage = "ui.soubor.5ee136c77a"
	uiView                 uiMessage = "ui.zobrazeni.77f1559aa4"
	uiTools                uiMessage = "ui.nastroje.a096f53723"
	uiEdit                 uiMessage = "ui.upravy.07ca5172db"
	uiForms                uiMessage = "ui.formulare.a578cb13d5"
	uiHelpMenu             uiMessage = "ui.pomoc.fc144bacfd"
	uiSettings             uiMessage = "ui.nastaveni.e01356eccc"
	uiSettingsTitle        uiMessage = "ui.nastaveni.pdf.blacksmith.4868b702d1"
	uiInterfaceLanguage    uiMessage = "ui.jazyk.rozhrani.edcf28a483"
	uiLanguage             uiMessage = "ui.jazyk.011a5d575a"
	uiInterface            uiMessage = "ui.rozhrani.d282a02b00"
	uiDarkMode             uiMessage = "ui.tmavy.38d04407f4"
	uiLightMode            uiMessage = "ui.svetly.e0d08090fa"
	uiInterfaceColors      uiMessage = "ui.barvy.rozhrani.733a07be19"
	uiApply                uiMessage = "ui.pouzit.190439bbc3"
	uiCancel               uiMessage = "ui.zrusit.6c77d8a4b0"
	uiHelp                 uiMessage = "ui.napoveda.249ae863e9"
	uiShortcuts            uiMessage = "ui.klavesove.zkratky.40cd311f0e"
	uiFind                 uiMessage = "ui.najit.v.pdf.8069740ad4"
	uiObjects              uiMessage = "ui.objekty.0c404098a0"
	uiText                 uiMessage = "ui.text.71988c4d8e"
	uiSeparations          uiMessage = "ui.separace.7468b3564b"
	uiSignature            uiMessage = "ui.digitalni.podpis.5e66a36014"
	uiSplitPDF             uiMessage = "ui.rozdelit.pdf.5dc7e235f5"
	uiMergePDF             uiMessage = "ui.sloucit.pdf.1dc5ce3c61"
	uiPageOrder            uiMessage = "ui.poradi.stranek.98753ea603"
	uiOpenPDF              uiMessage = "ui.otevrit.pdf.e53096d743"
	uiSave                 uiMessage = "ui.ulozit.78a5fe4f6b"
	uiSaveAs               uiMessage = "ui.ulozit.jako.540ff00830"
	uiProperties           uiMessage = "ui.vlastnosti.43d8ff25d4"
	uiThumbnails           uiMessage = "ui.nahledy.36a951f2d3"
	uiBookmarks            uiMessage = "ui.zalozky.d5988f16cb"
	uiPreview              uiMessage = "ui.nahled.68c447bc27"
	uiPages                uiMessage = "ui.stranky.abf779a41b"
	uiPageByPage           uiMessage = "ui.po.strankach.ae9475b722"
	uiFooterNoChanges      uiMessage = "ui.bez.zmen.249bf45ea6"
	uiFooterUnsaved        uiMessage = "win32.FooterUnsaved"
	uiFullscreen           uiMessage = "ui.cela.obrazovka.c44c09e551"
	uiFitPage              uiMessage = "ui.prizpusobit.strance.1656c87fd6"
	uiFitWidth             uiMessage = "ui.prizpusobit.sirce.8bca0aacbe"
	uiActualSize           uiMessage = "ui.skutecna.velikost.100.498ea622bc"
	uiZoom100              uiMessage = "win32.Zoom100"
	uiPageOverview         uiMessage = "win32.PageOverview"
	uiOverviewFit          uiMessage = "win32.OverviewFit"
	uiOverviewSelection    uiMessage = "win32.OverviewSelection"
	uiOverviewEmpty        uiMessage = "win32.OverviewEmpty"
	uiZoomIn               uiMessage = "ui.priblizit.d5195d8843"
	uiZoomOut              uiMessage = "ui.oddalit.2ca97714a5"
	uiContinuous           uiMessage = "ui.souvisle.zobrazeni.00efb2eec0"
	uiSwitchToContinuous   uiMessage = "ui.prepnout.na.souvisle.zobrazeni.59582bbe8d"
	uiSideBySide           uiMessage = "ui.vedle.sebe.727b60c855"
	uiUndo                 uiMessage = "ui.zpet.3e220c7830"
	uiRedo                 uiMessage = "ui.znovu.ef86d3752e"
	uiCut                  uiMessage = "ui.vyjmout.19e5fbdb1c"
	uiCopy                 uiMessage = "ui.kopirovat.89c0199fee"
	uiPaste                uiMessage = "ui.vlozit.d24832d822"
	uiDelete               uiMessage = "ui.odstranit.7904e49271"
	uiPreviousPage         uiMessage = "ui.predchozi.strana.597f2deff1"
	uiNextPage             uiMessage = "ui.nasledujici.strana.2765ab8c6f"
	uiOCR                  uiMessage = "win32.OCR"
	uiOpenFile             uiMessage = "win32.OpenFile"
	uiRecent               uiMessage = "win32.Recent"
	uiOpenOutput           uiMessage = "win32.OpenOutput"
	uiExport               uiMessage = "win32.Export"
	uiSignedCopy           uiMessage = "win32.SignedCopy"
	uiPrint                uiMessage = "win32.Print"
	uiMovingSelection      uiMessage = "win32.MovingSelection"
	uiSwitchToCanvasText   uiMessage = "win32.SwitchToCanvasText"
	uiSwitchToCanvasArrow  uiMessage = "win32.SwitchToCanvasArrow"
	uiCanvasRectangle      uiMessage = "win32.CanvasRectangle"
	uiCrossingSelection    uiMessage = "win32.CrossingSelection"
	uiCanvasLasso          uiMessage = "win32.CanvasLasso"
	uiSwitchToPages        uiMessage = "win32.SwitchToPages"
	uiStopSideBySide       uiMessage = "win32.StopSideBySide"
	uiStartSideBySide      uiMessage = "win32.StartSideBySide"
	uiSelectAll            uiMessage = "win32.SelectAll"
	uiGoToPage             uiMessage = "win32.GoToPage"
	uiWholePage            uiMessage = "win32.WholePage"
	uiPageWidth            uiMessage = "win32.PageWidth"
	uiCustomZoom           uiMessage = "win32.CustomZoom"
	uiMinimize             uiMessage = "win32.Minimize"
	uiMaximize             uiMessage = "win32.Maximize"
	uiClose                uiMessage = "win32.Close"
	uiDocumentList         uiMessage = "win32.DocumentList"
	uiCollapseBothDocks    uiMessage = "win32.CollapseBothDocks"
	uiRestoreBothDocks     uiMessage = "win32.RestoreBothDocks"
	uiCollapseLeft         uiMessage = "win32.CollapseLeft"
	uiExpandLeft           uiMessage = "win32.ExpandLeft"
	uiCollapseRight        uiMessage = "win32.CollapseRight"
	uiExpandRight          uiMessage = "win32.ExpandRight"
	uiLanguageChoice       uiMessage = "win32.LanguageChoice"
	uiLanguageSaveFailed   uiMessage = "win32.LanguageSaveFailed"
	uiLanguageDialogFailed uiMessage = "win32.LanguageDialogFailed"
	uiThemeDescription     uiMessage = "win32.ThemeDescription"
	uiSwitchToLight        uiMessage = "win32.SwitchToLight"
	uiSwitchToDark         uiMessage = "win32.SwitchToDark"
	uiWheelZoom            uiMessage = "win32.WheelZoom"
)

type uiTranslation struct {
	Czech   string `json:"cs"`
	English string `json:"en"`
}

// Strings specific to the Win32 shell that have no corresponding prototype key.
var win32Translations = map[uiMessage]uiTranslation{
	uiFooterUnsaved:        {Czech: "Neuložené změny", English: "Unsaved changes"},
	uiCollapseBothDocks:    {Czech: "Dočasně sbalit obě skupiny panelů", English: "Temporarily collapse both panel groups"},
	uiRestoreBothDocks:     {Czech: "Obnovit původní rozložení panelů", English: "Restore previous panel layout"},
	uiOCR:                  {Czech: "OCR", English: "OCR"},
	uiOpenFile:             {Czech: "Otevřít soubor", English: "Open file"},
	uiRecent:               {Czech: "Nedávné", English: "Recent"},
	uiOpenOutput:           {Czech: "Otevřít výstup", English: "Open output"},
	uiExport:               {Czech: "Export", English: "Export"},
	uiSignedCopy:           {Czech: "Podepsaná kopie", English: "Signed copy"},
	uiPrint:                {Czech: "Tisk", English: "Print"},
	uiMovingSelection:      {Czech: "Přesouvám výběr…", English: "Moving selection…"},
	uiSwitchToCanvasText:   {Czech: "Přepnout na textový kurzor", English: "Switch to text cursor"},
	uiSwitchToCanvasArrow:  {Czech: "Přepnout na šipku", English: "Switch to arrow"},
	uiCanvasRectangle:      {Czech: "Výběr obdélníkem", English: "Select with rectangle"},
	uiCrossingSelection:    {Czech: "Křížový výběr", English: "Crossing Selection"},
	uiCanvasLasso:          {Czech: "Výběr lasem", English: "Select with lasso"},
	uiSwitchToPages:        {Czech: "Přepnout na zobrazení po stránkách", English: "Switch to page-by-page view"},
	uiStopSideBySide:       {Czech: "Ukončit zobrazení vedle sebe", English: "Close side-by-side view"},
	uiStartSideBySide:      {Czech: "Zobrazit vedle sebe…", English: "Show side by side…"},
	uiSelectAll:            {Czech: "Vybrat vše", English: "Select all"},
	uiGoToPage:             {Czech: "Přejít", English: "Go"},
	uiWholePage:            {Czech: "Celá stránka", English: "Whole page"},
	uiPageWidth:            {Czech: "Šířka stránky", English: "Page width"},
	uiCustomZoom:           {Czech: "Vlastní zoom", English: "Custom zoom"},
	uiZoom100:              {Czech: "100 %", English: "100%"},
	uiPageOverview:         {Czech: "Přehled stran", English: "Page overview"},
	uiOverviewFit:          {Czech: "Na výšku (Enter)", English: "Fit page (Enter)"},
	uiOverviewSelection:    {Czech: "Přehled stran · Vybráno %d", English: "Page overview · %d selected"},
	uiOverviewEmpty:        {Czech: "Vyberte alespoň jednu stránku.", English: "Select at least one page."},
	uiMinimize:             {Czech: "Minimalizovat", English: "Minimize"},
	uiMaximize:             {Czech: "Maximalizovat nebo obnovit", English: "Maximize or restore"},
	uiClose:                {Czech: "Zavřít", English: "Close"},
	uiDocumentList:         {Czech: "Seznam otevřených PDF", English: "List of open PDFs"},
	uiCollapseLeft:         {Czech: "Sbalit levou skupinu panelů", English: "Collapse left panel group"},
	uiExpandLeft:           {Czech: "Rozbalit levou skupinu panelů", English: "Expand left panel group"},
	uiCollapseRight:        {Czech: "Sbalit pravou skupinu panelů", English: "Collapse right panel group"},
	uiExpandRight:          {Czech: "Rozbalit pravou skupinu panelů", English: "Expand right panel group"},
	uiLanguageChoice:       {Czech: "Zvolte jazyk. Rozhraní se přepne ihned. Použít volbu uloží, Zrušit vrátí původní jazyk.", English: "Choose a language to preview it. Apply saves your choice, and Cancel restores the previous language."},
	uiLanguageSaveFailed:   {Czech: "Nastavení rozhraní se nepodařilo uložit.", English: "Could not save the interface settings."},
	uiLanguageDialogFailed: {Czech: "Nastavení rozhraní se nepodařilo otevřít.", English: "Could not open the interface settings."},
	uiThemeDescription:     {Czech: "Vyberte paletu a režim. Změnu uvidíte hned. Použít ji uloží, Zrušit vrátí původní vzhled.", English: "Choose a palette and mode to preview them. Apply saves the changes. Cancel restores the previous appearance."},
	uiSwitchToLight:        {Czech: "Přepnout na světlý režim", English: "Switch to light mode"},
	uiSwitchToDark:         {Czech: "Přepnout na tmavý režim", English: "Switch to dark mode"},
	uiWheelZoom:            {Czech: "Ctrl+kolečko   Přiblížení dokumentu nebo miniatur", English: "Ctrl+mouse wheel   Zoom the document or thumbnails"},
}

//go:embed docs/pracovni/i18n/ui_catalog_FINAL_v6.0.json
var uiCatalogData []byte

var uiTranslations = loadUITranslations()

func loadUITranslations() map[uiMessage]uiTranslation {
	var catalog struct {
		Entries map[uiMessage]uiTranslation `json:"entries"`
	}
	if err := json.Unmarshal(uiCatalogData, &catalog); err != nil {
		panic(err)
	}
	return catalog.Entries
}

func (locale uiLocale) text(key uiMessage) string {
	translation, found := uiTranslations[key]
	if !found {
		translation, found = win32Translations[key]
	}
	if !found {
		translation = propertyTranslations[key]
	}
	if locale == uiEnglish && translation.English != "" {
		return translation.English
	}
	return translation.Czech
}

type uiLanguageSettings struct {
	Locale uiLocale `json:"locale"`
}

func decodeUILanguage(data []byte) uiLocale {
	var settings uiLanguageSettings
	if json.Unmarshal(data, &settings) == nil && settings.Locale == uiEnglish {
		return uiEnglish
	}
	return uiCzech
}
