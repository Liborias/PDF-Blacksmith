package main

import (
	"context"
	"time"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	idOpen = 100 + iota
	idShowOutput
	idToggleTools
	idFullscreen
	idBrowseTab
	idPagesTab
	idProofTab
	idExportTab
	idSignTab
	idReload
	idICCProfile
	idICCBrowse
	idOutline
	idDecrypt
	idRaster
	idRun
	idExportOption
	idResetExportOptions
	idCertBrowse
	idTrustBrowse
	idSign
	idVerify
	idVisible
	idClosePDF
	idFileTab
	idSaveDocument
	idSaveDocumentAs
	idPrintDocument
	idExportScope
	idExportPreset
	idPressICC
	idToggleDetails
	idFileSign
	_
	idFileDetails
	idSplitTab
	idSplitScope
	idSplitRun
	idPageOrderTool
	idGrayICC
	idGrayProfile
	idKOutputProfile
	idProofApply
	idProofColorMode
	idProofGrayProfile
	idProofGrayICC
	idProofKProfile
	idRecentFiles
	idRecentList
	idRecentMore
)

const (
	exportAllPages = iota
	exportCurrentPage
	exportPageRanges
	exportThumbnails
)

const (
	panelWidth   = 350
	headerHeight = 50
	footerHeight = 48
)

var workspaceTabs = []struct {
	id       int
	name     uiMessage
	needsPDF bool
}{
	{idFileTab, uiFile, false},
	{idBrowseTab, uiView, false},
	{idSearchTab, uiFind, true},
	{idObjectsTab, uiObjects, true},
	{idTextTab, uiText, true},
	{idOCRTab, uiOCR, true},
	{idProofTab, uiSeparations, true},
	{idSignTab, uiSignature, true},
	{idSplitTab, uiSplitPDF, true},
	{idMergeTab, uiMergePDF, false},
	{idPageOrderTool, uiPageOrder, true},
}

var pageScopeNames = []uiMessage{uiAllPages, uiCurrentPage, uiCustomRange, uiSelectedThumbnails}

type point struct{ x, y int32 }

type windowPlacement struct {
	length, flags, show uint32
	minimum, maximum    point
	normal              rect
}

type minMaxInfo struct {
	reserved, maxSize, maxPosition, minimumTrack, maximumTrack point
}

type monitorInfo struct {
	size          uint32
	monitor, work rect
	flags         uint32
}

func (a *app) setupWorkspace() bool {
	controls := [2]uint32{8, 4} // ICC_BAR_CLASSES includes native tooltips.
	comctl32.NewProc("InitCommonControlsEx").Call(uintptr(unsafe.Pointer(&controls)))
	className := wideString("PDFBlaksmithSidebar")
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	class := windowClass{instance: a.instance, className: uintptr(unsafe.Pointer(className)),
		windowProc: syscall.NewCallback(sidebarProc), cursor: cursor}
	class.size = uint32(unsafe.Sizeof(class))
	if ok, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&class))); ok == 0 || !a.registerScrollTrack(cursor) {
		return false
	}
	a.sidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	a.thumbnailSidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000|0x04000000, 0, 0, 1, 1, a.window, 0, a.instance, 0) // Clip the File menu above it.
	a.bookmarkSidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	a.searchSidebar, _, _ = createWindow.Call(0x10000, uintptr(unsafe.Pointer(className)), 0,
		wsChild|0x02000000, 0, 0, 1, 1, a.window, 0, a.instance, 0)
	if a.sidebar == 0 || a.thumbnailSidebar == 0 || a.bookmarkSidebar == 0 || a.searchSidebar == 0 {
		return false
	}
	a.sideScroll = a.addPanelScrollbar(a.sidebar, func(offset int) {
		a.panelScroll[a.detailsTab()] = offset
		a.layoutPanel()
	})

	var catalog []profileChoice
	if json.Unmarshal(printCatalog, &catalog) != nil {
		return false
	}
	a.profileChoices = append([]profileChoice{
        {"", "Zachovat barvy"},
        {"k", "Stupně šedi · DeviceGray"},
        {"rich_black", "Složená černá → K100"},
        {"k_cmyk", "Pouze K · explicitní CMYK"},
    }, catalog...)
	a.profileChoices = append(a.profileChoices, profileChoice{"custom", "Vlastní výstupní profil ICC"})
	if json.Unmarshal(exportCatalog, &a.exportPresets) != nil || len(a.exportPresets) == 0 {
		return false
	}

	a.railFont = a.orientedFont(14, 600, 900)
	a.toolsToggle = a.shellButton(uiTools, idToggleTools, sidebarRailWidth, 0)
	a.detailsToggle = a.shellButton(uiProperties, idToggleDetails, sidebarRailWidth, 0)
	a.thumbnailsToggle = a.shellButton(uiThumbnails, idToggleThumbnails, sidebarRailWidth, 0)
	a.bookmarksToggle = a.shellButton(uiBookmarks, idToggleBookmarks, sidebarRailWidth, 0)
	a.resultsToggle = a.shellButton(uiResults, idToggleSearchResults, sidebarRailWidth, 0)
	a.leftPanelsToggle = a.button("", idToggleLeftPanels, 0, 0, sidebarRailWidth, 0)
	a.rightPanelsToggle = a.button("", idToggleRightPanels, 0, 0, sidebarRailWidth, 0)
	a.windowButtons[0] = a.shellButton(uiMinimize, idMinimize, 36, 0)
	a.windowButtons[1] = a.shellButton(uiMaximize, idMaximize, 36, 0)
	a.windowButtons[2] = a.shellButton(uiClose, idCloseWindow, 36, 0)
	a.tabList = a.shellButton(uiDocumentList, idListDocumentTabs, 26, 0)
	a.status = a.add("STATIC", "", 0, 0, 0, 160, 22, 0x4000, 0, a.normalFont)
	a.focusHelp = a.add("STATIC", "", 0, 0, 0, 1, 22, 0x400c, 0, a.normalFont)
	for _, tab := range workspaceTabs {
		handle := a.shellButton(tab.name, tab.id, 1, 0x04000000) // WS_CLIPSIBLINGS
		watchNavigationHover(handle)
		a.toolButtons[tab.id] = handle
		a.panelButtons[handle] = true
	}
	a.measureToolNames()
	if !a.buildFileMenu(className) {
		return false
	}
	a.sidebarTitle = a.label("", 0, 0, 298, true)
	a.panelLabels[a.sidebarTitle] = true
	a.addDetailButton(idBrowseTab, uiPreview)
	a.addDetailButton(idPagesTab, uiPages)

	a.view = newDocumentView(a)
	a.view.thumbnailSidebar = a.thumbnailSidebar
	a.view.buildToolbar()
	a.beginPanel(idBrowseTab)
	a.uiButton(uiViewFullscreen, idFullscreen, 16, 16, 298, 0)
	a.comparison.button = a.button(a.uiText(uiStartSideBySide), idCompareOpen, 16, 56, 298, 0)
	a.beginPropertySection(uiViewMode, 106, uiViewModeHelp)
	a.view.flow = a.uiCombo(idReaderFlow, 16, 128, 298, []uiMessage{uiFlowContinuous, uiPageByPage})
	chooseCombo(a.view.flow, 1)
	a.uiLabel(uiScreenBox, 16, 171, 298, false)
	a.view.screenBox = a.uiCombo(idReaderScreenBox, 16, 193, 298, []uiMessage{uiCropBox, uiMediaBox, uiTrimBox, uiBleedBox})
	a.view.documentInfo = a.readout(16, 238, 298, 82)
	a.beginPropertySection(uiDocumentPassword, 342, uiDocumentPasswordHelp)
	a.password = a.edit(16, 365, 298, true)
	a.uiButton(uiReloadPassword, idReload, 16, 403, 298, 0)

	a.beginPanel(idLayoutAids)
	a.buildLayoutAidsPanel()
	a.beginPanel(idPagesTab)
	a.view.buildPagePanel()
	a.beginPanel(idProofTab)
	a.view.buildProofPanel(catalog)
	a.beginPanel(idExportTab)
	a.buildExportPanel()
	a.beginPanel(idSignTab)
	a.buildSignPanel()
	a.beginPanel(idFileDetails)
	a.buildRecentPanel()
	a.beginPanel(idSearchTab)
	a.view.buildSearchPanel()
	a.beginPanel(idTextTab)
	a.buildTextPanel()
	a.addingTab, a.controlParent = 0, 0
	a.buildCommandBar()
	a.beginPanel(idObjectsTab)
	a.buildObjectPanel()
	a.beginPanel(idOCRTab)
	a.buildOCRPanel()
	a.beginPanel(idSplitTab)
	a.buildSplitPanel()
	a.beginPanel(idMergeTab)
	a.buildMergePanel()
	a.addingTab, a.controlParent = 0, a.bookmarkSidebar
	a.view.buildBookmarkPanel()
	a.controlParent = a.searchSidebar
	a.view.buildSearchResults()
	a.controlParent = a.thumbnailSidebar
	a.view.buildThumbnails()
	a.controlParent = 0
	if !a.buildComparison(className) {
		return false
	}
	a.layout()
	a.updateActions()
	return true
}

func (a *app) beginPanel(tab int) {
	a.addingTab, a.controlParent = tab, a.sidebar
	a.addingSection = 0
}

func (a *app) buildExportPanel() {
	a.runButton = a.uiButton(uiSaveExport, idRun, 16, 17, 298, 0)
	a.primaryButtons[a.runButton] = true
	a.beginPropertySection(uiExportPreset, 64, uiExportPresetHelp).detail = func() string {
		preset := a.selectedExportPreset()
		if key := exportPresetMessages[preset.Key].description; key != "" {
			return a.uiText(key)
		}
		return preset.Description
	}
	formats := make([]string, 0, len(a.exportPresets))
	formatMessages := make([]uiMessage, 0, len(a.exportPresets))
	for _, preset := range a.exportPresets {
		formats = append(formats, preset.Name)
		formatMessages = append(formatMessages, exportPresetMessages[preset.Key].name)
	}
	a.exportFormat = a.comboWithMessages(idExportPreset, 16, 86, 298, formats, formatMessages)
	a.exportWarning = a.readout(16, 121, 298, 44)
	a.beginPropertySection(uiExportPages, 169, uiExportPagesHelp)
	a.exportScope = a.uiCombo(idExportScope, 16, 191, 145, pageScopeNames)
	a.exportRange = a.edit(169, 191, 145, false)
	a.bindUICue(a.exportRange, uiRangeExample) // EM_SETCUEBANNER
	color := a.beginPropertySection(uiOutputColor, 229, uiOutputColorHelp)
	a.exportColorLabel, a.exportColorHelp = color.header, color.help
	names := make([]string, 0, len(a.profileChoices))
	for _, p := range a.profileChoices {
		names = append(names, p.Name)
	}
	profileMessages := make([]uiMessage, len(names))
	copy(profileMessages, []uiMessage{uiKeepColors, uiGrayDevice, uiRichBlackK, uiExplicitCMYK})
	profileMessages[len(profileMessages)-1] = uiCustomOutputICC
	a.profiles = a.comboWithMessages(idICCProfile, 16, 251, 298, names, profileMessages)
	chooseCombo(a.profiles, 1+defaultIndex(colorModes, a.export.ColorMode, defaults.ColorMode))
    a.pressProfiles = a.comboWithMessages(idPressICC, 16, 251, 298, names[4:], profileMessages[4:])
	for _, preset := range a.exportPresets {
		if preset.Profile != "" {
			index := slices.IndexFunc(a.profileChoices[4:], func(choice profileChoice) bool { return choice.Key == preset.Profile })
			chooseCombo(a.pressProfiles, index)
			break
		}
	}
	a.customICC = a.edit(16, 285, 187, false)
	a.customBrowse = a.uiButton(uiBrowseICC, idICCBrowse, 212, 282, 102, 0)
	a.exportManualStart = len(a.panelItems[idExportTab])
    a.grayProfileLabel = a.uiLabel(uiGrayConversionICC, 16, 285, 298, false)
    a.grayProfileCombo = a.uiCombo(idGrayProfile, 16, 307, 298, []uiMessage{
        uiAutomaticFromDocument, uiGrayGamma22, uiGrayGamma18, uiCustomGrayICC,
    })
    chooseCombo(a.grayProfileCombo, defaultIndex(grayProfiles, a.export.GrayProfile, defaults.GrayProfile))
    a.grayCustomICC = a.edit(16, 344, 187, false)
    a.grayBrowse = a.uiButton(uiBrowseICC, idGrayICC, 212, 341, 102, 0)
    a.kOutputLabel = a.uiLabel(uiKOutputProfile, 16, 376, 298, false)
    a.kOutputProfile = a.combo(idKOutputProfile, 16, 398, 298, names[4:len(names)-1])
    kIndex := slices.IndexFunc(a.profileChoices[4:], func(choice profileChoice) bool {return choice.Key == a.export.KProfile})
    chooseCombo(a.kOutputProfile, max(0, kIndex))
    a.outline = a.uiButton(uiConvertTextOutlines, idOutline, 16, 438, 298, bsAutoCheck)
	a.decrypt = a.uiButton(uiRemovePasswordRestrictions, idDecrypt, 16, 470, 298, bsAutoCheck)
    a.raster = a.button("", idRaster, 16, 502, 298, bsAutoCheck)
	a.setUIMessage(a.raster, uiAllowKRaster, a.export.RepairDPI)
	a.exportReport = a.readout(16, 542, 298, 105)
	a.beginPropertySection(uiExportParameters, 661, uiExportParametersHelp)
	labels := []uiMessage{uiRenderingIntent, uiBlackPoint, uiImages, uiCompression, uiPlateCheck, uiKRepair}
	choices := [][]uiMessage{
		{uiFromPDF, uiPerceptual, uiRelative, uiSaturation, uiAbsolute},
		{uiAutomatic, uiTurnOn, uiTurnOff},
		{uiUnchanged, uiResolution150, uiResolution300, uiResolution600},
		{uiAutomaticCompression, uiLossless},
		{uiResolution75, uiResolution150, uiResolution300},
		{uiResolution300, uiResolution600},
	}
	for index, label := range labels {
		x := 16 + (index%2)*153
		y := 696 + (index/2)*63
		a.uiLabel(label, x, y, 145, false)
		a.exportOptionControls[index] = a.uiCombo(idExportOption, x, y+21, 145, choices[index])
	}
	a.uiButton(uiResetExport, idResetExportOptions, 16, 892, 298, 0)
	a.applyExportOptions(a.export)
}

func (a *app) buildSignPanel() {
	a.beginPropertySection(uiSignDocument, 18, uiSignDocumentHelp)
	a.uiLabel(uiPKCSFile, 16, 53, 298, false)
	a.cert = a.edit(16, 76, 194, false)
	a.uiButton(uiChoose, idCertBrowse, 219, 74, 95, 0)
	a.uiLabel(uiCertificatePassword, 16, 122, 298, false)
	a.certPass = a.edit(16, 144, 298, true)
	a.visible = a.uiButton(uiVisibleSignature, idVisible, 16, 183, 298, bsAutoCheck)
	a.signButton = a.uiButton(uiSaveSignedPDF, idSign, 16, 224, 298, 0)
	a.beginPropertySection(uiVerifySignatures, 284, uiVerifySignatureHelp)
	a.uiLabel(uiExtraRootCertificate, 16, 318, 298, false)
	a.trust = a.edit(16, 341, 194, false)
	a.uiButton(uiChoose, idTrustBrowse, 219, 339, 95, 0)
	a.verifyButton = a.uiButton(uiVerifySignatures, idVerify, 16, 384, 298, 0)
	a.signReport = a.readout(16, 431, 298, 179)
}

func (a *app) applyExportOptions(value exportOptions) {
    a.export = value
    chooseCombo(a.grayProfileCombo, defaultIndex(grayProfiles, value.GrayProfile, defaults.GrayProfile))
    chooseCombo(a.view.convertGrayProfile, defaultIndex(grayProfiles, value.GrayProfile, defaults.GrayProfile))
    mode := defaultIndex(colorModes, value.ColorMode, defaults.ColorMode)
    chooseCombo(a.profiles, 1+mode)
    chooseCombo(a.view.convertMode, mode)
    kidx := slices.IndexFunc(a.profileChoices[4:], func(p profileChoice) bool { return p.Key == value.KProfile })
    chooseCombo(a.kOutputProfile, max(0, kidx))
    chooseCombo(a.view.convertKProfile, max(0, kidx))
	indexes := []int{
		defaultIndex(intents, value.Intent, defaults.Intent),
		defaultIndex(blackPoints, value.BlackPoint, defaults.BlackPoint),
		defaultIndex(imageDPIs, value.ImageDPI, defaults.ImageDPI),
		defaultIndex(compressions, value.Compression, defaults.Compression),
		defaultIndex(proofDPIs, value.ProofDPI, defaults.ProofDPI),
		defaultIndex(repairDPIs, value.RepairDPI, defaults.RepairDPI),
	}
	for index, handle := range a.exportOptionControls {
		chooseCombo(handle, indexes[index])
	}
}

func (a *app) updateExportOptions() {
	c := a.exportOptionControls
	a.export.Intent = intents[max(0, comboIndex(c[0]))]
    a.export.BlackPoint = blackPoints[max(0, comboIndex(c[1]))]
    a.export.ImageDPI = imageDPIs[max(0, comboIndex(c[2]))]
    a.export.Compression = compressions[max(0, comboIndex(c[3]))]
    a.export.ProofDPI = proofDPIs[max(0, comboIndex(c[4]))]
    a.export.RepairDPI = repairDPIs[max(0, comboIndex(c[5]))]
    a.export.GrayProfile = grayProfiles[max(0, comboIndex(a.grayProfileCombo))]
	if err := saveExportOptions(a.export); err != nil {
		a.report("Předvolby nelze uložit. " + err.Error())
	}
	a.setUIMessage(a.raster, uiAllowKRaster, a.export.RepairDPI)
}

func (a *app) selectedExportPreset() exportPreset {
	index := comboIndex(a.exportFormat)
	if index < 0 || index >= len(a.exportPresets) {
		return exportPreset{}
	}
	return a.exportPresets[index]
}

func (a *app) profileKey() string {
	control, choices := a.profiles, a.profileChoices
	if preset := a.selectedExportPreset(); preset.Key != "" {
		if preset.Profile == "" {
			return ""
		}
		control, choices = a.pressProfiles, choices[4:]
	}
	index := comboIndex(control)
	if index < 0 || index >= len(choices) {
		return ""
	}
	return choices[index].Key
}

func (a *app) activate(id int) {
	if a.mode == idTextTab && id != idTextTab && a.saveInlineBefore(func() { a.activate(id) }) {
		return
	}
	old := a.mode
	switch id {
	case idFileTab:
		if a.tool != idFileTab {
			a.mode = idFileDetails
		}
		a.tool = idFileTab
	case idExportTab, idFileSign:
		if !a.fileOptionEnabled(id) {
			return
		}
		a.tool, a.mode = idFileTab, id
		if id == idFileSign {
			a.mode = idSignTab
		}
	case idBrowseTab, idPagesTab, idLayoutAids:
		if id != idBrowseTab && !a.documentTools {
			return
		}
		a.tool, a.mode = idBrowseTab, id
	case idPageOrderTool:
		if !a.documentTools {
			return
		}
		a.tool, a.mode = idPageOrderTool, idPagesTab
	case idMergeTab:
		a.tool, a.mode = idMergeTab, idMergeTab
	case idProofTab, idSignTab, idSplitTab, idSearchTab, idOCRTab, idTextTab, idObjectsTab:
		if !a.documentTools {
			return
		}
		a.tool, a.mode = id, id
	}
	if old != a.mode {
		a.cancelCanvasSelections()
		a.view.clearCanvasContent()
		a.comparison.view.clearCanvasContent()
	}
	a.layoutAids.temporary = false
	a.syncSearchAuxiliary()
	if id == idExportTab || id == idFileSign || id == idLayoutAids {
		a.rightDock.revealPrimary()
	}
	if (old == idProofTab) != (id == idProofTab) {
		a.view.reframe()
	}
	if old == idTextTab && a.mode != idTextTab {
		a.clearTextSelection()
	}
	if old == idObjectsTab && a.mode != idObjectsTab {
		a.clearObjectSelection()
		a.view.invalidateCanvas()
	}
	a.updateObjectActions()
	a.updateHistoryButtons()
	if (old == idSearchTab) != (a.mode == idSearchTab) {
		a.view.invalidateCanvas()
	}
	a.layout()
	a.updateCommandBar()
	if scope, _ := a.scopeControls(); old != a.mode && (a.thumbnailScopes[scope] || comboIndex(scope) == exportAllPages) {
		a.applyThumbnailScope()
	}
	if old == idPagesTab && a.mode != idPagesTab {
		a.view.endPageDrag()
	}
	a.view.schedule()
}

func (a *app) layout() {
	a.hideFileMenu(false)
	var size rect
	if ok, _, _ := getClientRect.Call(a.window, uintptr(unsafe.Pointer(&size))); ok == 0 {
		return
	}
	a.width, a.height = int(size.right), int(size.bottom)
	if a.width == 0 || a.height == 0 {
		return
	}
	d := a.scale
	a.thumbnailArea, a.rightAuxArea = rect{}, rect{}
	if a.full {
		a.viewport = size
	} else {
		a.layoutDocumentPanels()
		a.layoutHeader()
		place(a.status, a.width-d(231), a.height-d(33), d(218), d(22))
		place(a.focusHelp, d(628), a.height-d(33), a.width-d(874), d(22))
		a.view.layoutToolbar()
	}
	chrome := uintptr(5)
	if a.full {
		chrome = 0
	}
	showWindow.Call(a.status, chrome)
	showWindow.Call(a.focusHelp, chrome)
	for _, handle := range a.windowButtons {
		showWindow.Call(handle, chrome)
	}
	a.layoutCommandBar()
	a.layoutDocumentTabs()
	a.view.showToolbar(chrome)
	active := a.layoutComparison()
	a.layoutSidebars()
	a.view.resize(active)
	invalidate(a.window)
}

func (a *app) panelItemLayout(tab, index int, item panelItem) (panelItem, bool) {
	if tab == idFileDetails {
		return a.recentItemLayout(item)
	}
	if item.handle == a.splitRange {
		return item, comboIndex(a.splitScope) == exportPageRanges
	}
	if item.handle == a.ocr.pageRange {
		return item, comboIndex(a.ocr.scope) == exportPageRanges
	}
	if tab == idProofTab {
        mode := max(0, comboIndex(a.view.convertMode))
        switch item.handle {
        case a.view.convertGrayLabel, a.view.convertGrayProfile:
            return item, mode != 1
        case a.view.convertGrayICC, a.view.convertGrayBrowse:
            return item, mode != 1 && comboIndex(a.view.convertGrayProfile) == 3
        case a.view.convertKLabel, a.view.convertKProfile:
            return item, mode == 2
        }
        return item, true
    }
    if tab != idExportTab {
        return item, true
    }
	preset := a.selectedExportPreset()
	manual, press := preset.Key == "", preset.Profile != ""
	key := a.profileKey()
	switch item.handle {
	case a.exportRange:
		return item, comboIndex(a.exportScope) == exportPageRanges
	case a.exportColorLabel, a.exportColorHelp:
		return item, manual || press
	case a.profiles:
		return item, manual
	case a.pressProfiles:
		return item, press
	case a.customICC, a.customBrowse:
		return item, key == "custom"
	case a.grayProfileLabel, a.grayProfileCombo:
        return item, manual && (key == "k" || key == "k_cmyk")
    case a.grayCustomICC, a.grayBrowse:
        return item, manual && (key == "k" || key == "k_cmyk") && a.export.GrayProfile == "custom"
    case a.kOutputLabel, a.kOutputProfile:
        return item, manual && key == "k_cmyk"
    case a.raster:
        return item, manual && key == "k" && !a.checked(a.outline)
	case a.exportReport:
		switch {
		case press && key == "custom":
			item.y = 332
		case press:
			item.y = 296
		case !manual:
			item.y = 239
		}
		return item, true
	}
	return item, manual || index < a.exportManualStart
}

func (a *app) layoutPanel() {
	if a.full || a.settingsWidth == 0 {
		return
	}
	active := a.detailsTab()
	items := a.visiblePropertyItems(active)
	placed := make(map[uintptr]panelItem, len(items))
	height := 1
	for _, item := range items {
		height = max(height, item.y+item.height+18)
		placed[item.handle] = a.scaleSettingsItem(item)
	}
	area := a.detailsArea
	total := max(2, int(area.bottom-area.top))
	left, top, width := int(area.left), int(area.top), int(area.right-area.left)
	place(a.sidebar, left, top, width, total)
	available := max(1, total*96/a.dpi)
	offset := min(max(0, a.panelScroll[active]), max(0, height-available))
	a.panelScroll[active] = offset
	var inner rect
	getClientRect.Call(a.sidebar, uintptr(unsafe.Pointer(&inner)))
	place(a.sideScroll.handle, int(inner.right)-a.scale(18), 0, a.scale(14), int(inner.bottom))
	a.sideScroll.update(height, available, offset, true)
	for _, controls := range a.panelItems {
		for _, control := range controls {
			item, show := placed[control.handle]
			if pane := a.scrollPanes[control.handle]; pane != nil {
				a.layoutScrollPane(pane, item, offset, show)
				continue
			}
			if show {
				place(item.handle, item.x, a.scale(item.y-offset), item.width, a.scale(item.height))
			}
			showControl(control.handle, show)
		}
	}
	invalidate(a.sidebar)
}

func (a *app) revealFocusedControl() {
	if a.settingsWidth == 0 || a.full {
		return
	}
	focus, _, _ := user32.NewProc("GetFocus").Call()
	var bounds rect
	getClientRect.Call(a.sidebar, uintptr(unsafe.Pointer(&bounds)))
	height := max(1, int(bounds.bottom)*96/a.dpi)
	tab := a.detailsTab()
	offset := a.panelScroll[tab]
	for _, item := range a.visiblePropertyItems(tab) {
		if item.handle != focus {
			continue
		}
		if item.y < offset || item.y+item.height > offset+height {
			a.panelScroll[tab] = max(0, min(item.y, item.y+item.height-height))
			a.layoutPanel()
		}
		return
	}
}

func sidebarProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	a := window
	switch msg {
	case wmPaint:
		var paint paintInfo
		var bounds rect
		getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
		dc, _, _ := user32.NewProc("BeginPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		if hwnd == a.fileMenu.handle {
			outline := bounds
			outline.top -= int32(a.scale(9))
			rounded(dc, outline, a.scale(9), uiColors["popoverBg"], divider)
			fillColor(dc, rect{right: bounds.right, bottom: 1}, divider)
		} else {
			fillColor(dc, bounds, panelColor)
		}
		if hwnd == a.thumbnailSidebar || hwnd == a.bookmarkSidebar || hwnd == a.searchSidebar ||
			(a.comparison.view != nil && hwnd == a.comparison.view.thumbnailSidebar) {
			fillColor(dc, rect{right: 1, bottom: bounds.bottom}, divider)
		}
		user32.NewProc("EndPaint").Call(hwnd, uintptr(unsafe.Pointer(&paint)))
		return 0
	case wmEraseBkgnd:
		return 1
	case wmDrawItem:
		if a.drawControl(lParam) {
			return 1
		}
	case wmCtlColorEdit, wmCtlColorList, wmCtlColorText:
		return a.colorControl(msg, wParam, lParam)
	case wmCommand:
		if wParam>>16&0xffff == 0x0300 {
			a.inlineControlChange(lParam)
		}
		if v := a.comparison.view; v != nil && hwnd == v.thumbnailSidebar && lParam == v.pageList {
			if wParam>>16&0xffff == 1 { // LBN_SELCHANGE
				v.readPageSelection()
			}
			return 0
		}
		a.command(int(wParam&0xffff), int(wParam>>16&0xffff))
		return 0
	case 0x020a: // WM_MOUSEWHEEL
		if a.wheelSettings(lParam, int(int16(wParam>>16)), wParam&8 != 0) {
			return 0
		}
	}
	result, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}

func (a *app) minimumSize(pointer uintptr) {
	var info minMaxInfo
	copyDrawItem.Call(uintptr(unsafe.Pointer(&info)), pointer, unsafe.Sizeof(info))
	info.minimumTrack = point{int32(a.scale(minimumWorkspaceWidth)), int32(a.scale(545))}
	copyDrawItem.Call(pointer, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))
}

func (a *app) toggleFullscreen() {
	if a.saveInlineBefore(a.toggleFullscreen) {
		return
	}
	a.cancelCanvasSelections()
	a.cancelCanvasRegionProbes()
	getStyle := user32.NewProc("GetWindowLongW")
	setStyle := user32.NewProc("SetWindowLongW")
	position := user32.NewProc("SetWindowPos")
	const styleIndex = ^uintptr(15)
	if a.full {
		a.full = false
		setStyle.Call(a.window, styleIndex, a.savedStyle)
		user32.NewProc("SetWindowPlacement").Call(a.window, uintptr(unsafe.Pointer(&a.placement)))
		position.Call(a.window, 0, 0, 0, 0, 0, 0x37)
	} else {
		monitor, _, _ := user32.NewProc("MonitorFromWindow").Call(a.window, 2)
		info := monitorInfo{size: uint32(unsafe.Sizeof(monitorInfo{}))}
		if ok, _, _ := user32.NewProc("GetMonitorInfoW").Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
			return
		}
		a.placement = windowPlacement{length: uint32(unsafe.Sizeof(windowPlacement{}))}
		user32.NewProc("GetWindowPlacement").Call(a.window, uintptr(unsafe.Pointer(&a.placement)))
		a.savedStyle, _, _ = getStyle.Call(a.window, styleIndex)
		a.full = true
		setStyle.Call(a.window, styleIndex, (a.savedStyle&^0x00c40000)|0x80000000)
		r := info.monitor
		position.Call(a.window, 0, uintptr(r.left), uintptr(r.top), uintptr(r.right-r.left), uintptr(r.bottom-r.top), 0x24)
	}
	a.layout()
	setFocus.Call(a.window)
}

func (a *app) openDrop(handle uintptr) {
	proc := shell32.NewProc("DragQueryFileW")
	defer shell32.NewProc("DragFinish").Call(handle)
	if a.busy {
		return
	}
	count, _, _ := proc.Call(handle, 0xffffffff, 0, 0)
	var paths []string
	for index := range count {
		if length, _, _ := proc.Call(handle, index, 0, 0); length > 0 {
			buf := make([]uint16, length+1)
			proc.Call(handle, index, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
			paths = append(paths, syscall.UTF16ToString(buf))
		}
	}
	if a.mode == idMergeTab {
		a.addMergeFiles(paths)
	} else {
		for _, path := range paths {
			a.openDocument(path, "")
		}
	}
}

func (a *app) openDialog() {
	if a.saveInlineBefore(a.openDialog) {
		return
	}
	if a.busy {
		return
	}
	for _, path := range a.fileDialogPaths(pdfFilter, "", "", false, true) {
		a.openDocument(path, "")
	}
}

func (a *app) mayReplaceDocument() bool {
	return !a.view.changed() || notice(a.window, "Zahodit neuložené úpravy PDF?", 0x134) == 6
}

func (a *app) openDocument(path, password string) {
	if a.busy {
		return
	}
	path, err := filepath.Abs(path)
	if err != nil {
		a.report("Cestu k dokumentu nelze otevřít.")
		return
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		a.report("Vyberte existující soubor PDF.")
		return
	}
	a.addDocumentTab(path, password, info)
}

func (a *app) clearDocumentControls() {
	setText(a.password, "")
	setText(a.signReport, "")
	setText(a.exportReport, "")
	chooseCombo(a.exportScope, exportAllPages)
	setText(a.exportRange, "")
	chooseCombo(a.splitScope, exportAllPages)
	setText(a.splitRange, "")
	setText(a.splitReport, "")
	chooseCombo(a.ocr.scope, exportAllPages)
	setText(a.ocr.pageRange, "")
	setText(a.ocr.report, "")
}

func (a *app) loadDocument(path, password string, info os.FileInfo) {
	a.dropRedo(a.activeTab)
	a.retireDocumentDrafts(a.view.documentContent)
	a.clearDocumentControls()
	setText(a.password, password)
	setText(a.window, filepath.Base(path)+" · "+appName)
	a.view.open(path, password, info)
	a.layout()
}

func (a *app) command(id, notification int) {
	if index := id - idFileMenuFirst; index >= 0 && index < len(fileOptions) && notification == 0 {
		a.hideFileMenu(true)
		a.runFileOption(fileOptions[index].id)
		return
	}
	if choice := a.choiceCommands[id]; choice != nil && notification == 0 {
		a.openChoice(choice)
		return
	}
	if a.commandBarCommand(id, notification) {
		return
	}
	if notification == 0 && a.propertySectionCommand(id) {
		return
	}
	if notification == 0 && (id == idHeaderUndo || id == idHeaderRedo) {
		command := docUndo
		if id == idHeaderRedo {
			command = docRedo
		}
		setFocus.Call(a.window)
		a.doDocumentCommand(a.view, command, false)
		return
	}
	if a.comparisonCommand(id, notification) || a.textCommand(id, notification) ||
		a.objectCommand(id, notification) || a.view.command(id, notification) {
		return
	}
	if a.mergeCommand(id, notification) {
		return
	}
	if id == idRecentList {
		switch notification {
		case 2: // LBN_DBLCLK
			a.openRecentFile()
		}
		return
	}
	if notification == 1 {
		switch id {
		case idExportOption:
            a.updateExportOptions()
        case idGrayProfile:
            a.updateExportOptions()
            chooseCombo(a.view.convertGrayProfile, max(0, comboIndex(a.grayProfileCombo)))
            a.layoutPanel()
        case idProofColorMode:
            a.export.ColorMode = colorModes[max(0, comboIndex(a.view.convertMode))]
            chooseCombo(a.profiles, 1+max(0, comboIndex(a.view.convertMode)))
            _ = saveExportOptions(a.export)
            a.layoutPanel()
        case idProofGrayProfile:
            a.export.GrayProfile = grayProfiles[max(0, comboIndex(a.view.convertGrayProfile))]
            chooseCombo(a.grayProfileCombo, max(0, comboIndex(a.view.convertGrayProfile)))
            _ = saveExportOptions(a.export)
            a.layoutPanel()
        case idProofKProfile:
            idx := max(0, comboIndex(a.view.convertKProfile))
            if idx < len(a.profileChoices)-5 {
                a.export.KProfile = a.profileChoices[4+idx].Key
                chooseCombo(a.kOutputProfile, idx)
                _ = saveExportOptions(a.export)
            }
        case idKOutputProfile:
            key := a.profileChoices[4+max(0, comboIndex(a.kOutputProfile))].Key
            a.export.KProfile = key
            chooseCombo(a.view.convertKProfile, max(0, comboIndex(a.kOutputProfile)))
            if err := saveExportOptions(a.export); err != nil { a.report(err.Error()) }
            a.layoutPanel()
		case idExportPreset:
			var warning uiMessage
			switch a.selectedExportPreset().Key {
			case "flatten":
				warning = uiWarnFlattenedText
			case "pdfx-1a", "pdfx-3":
				warning = uiWarnFlattenedTransparency
			}
			a.setUIMessage(a.exportWarning, warning)
			a.layoutPanel()
		case idExportScope:
			delete(a.thumbnailScopes, a.exportScope)
			a.layoutPanel()
			if comboIndex(a.exportScope) == exportPageRanges {
				setFocus.Call(a.exportRange)
			}
		case idSplitScope:
			delete(a.thumbnailScopes, a.splitScope)
			a.layoutPanel()
			if comboIndex(a.splitScope) == exportPageRanges {
				setFocus.Call(a.splitRange)
			}
		case idOCRScope:
			delete(a.thumbnailScopes, a.ocr.scope)
			a.layoutPanel()
			if comboIndex(a.ocr.scope) == exportPageRanges {
				setFocus.Call(a.ocr.pageRange)
			}
		case idICCProfile:
            key := a.profileKey()
            if key == "k" || key == "rich_black" || key == "k_cmyk" {
                a.export.ColorMode = map[string]string{"k":"gray","rich_black":"rich_black","k_cmyk":"k_cmyk"}[key]
                chooseCombo(a.view.convertMode, max(0, comboIndex(a.profiles)-1))
                if err := saveExportOptions(a.export); err != nil { a.report(err.Error()) }
            }
            if key != "" {
				a.setChecked(a.decrypt, false)
			}
			a.layoutPanel()
		case idPressICC:
			a.layoutPanel()
		}
		return
	}
	if notification != 0 {
		return
	}
	switch id {
	case idFileTab:
		a.showFileMenu()
	case idSearchTab:
		a.showSearch()
	case idBrowseTab, idPagesTab, idLayoutAids, idPageOrderTool, idProofTab, idSignTab, idSplitTab, idMergeTab, idOCRTab, idTextTab, idObjectsTab:
		if id == idTextTab || id == idObjectsTab {
			a.canvasTextCursor = id == idTextTab
		}
		a.activate(id)
	case idToggleLayoutGrid:
		a.toggleLayoutGrid()
	case idFitOverviewPage:
		a.view.fitOverviewSelection()
	case idOpen, idRecentFiles, idShowOutput, idSaveDocument, idSaveDocumentAs, idExportTab, idFileSign, idPrintDocument:
		a.runFileOption(id)
	case idRecentMore:
		a.toggleRecentFiles()
	case idToggleTools:
		a.leftDock.primaryCollapsed = !a.leftDock.primaryCollapsed
		a.layout()
	case idToggleDetails:
		hidden := !a.effectiveRightDock().primaryCollapsed
		a.layoutAids.temporary = false
		a.rightAux.setPropertiesCollapsed(&a.rightDock, hidden)
		a.layout()
		a.view.schedule()
	case idToggleThumbnails:
		a.leftDock.auxiliaryCollapsed = !a.leftDock.auxiliaryCollapsed
		a.view.endPageDrag()
		a.layout()
		a.view.schedule()
	case idToggleLeftPanels, idToggleRightPanels:
		a.togglePanelSide(id == idToggleLeftPanels)
	case idToggleBookmarks:
		a.toggleRightAuxiliary(auxiliaryBookmarks)
	case idToggleSearchResults, idCloseSearchResults:
		a.toggleRightAuxiliary(auxiliaryResults)
		if id == idCloseSearchResults {
			setFocus.Call(a.resultsToggle)
		}
	case idFullscreen:
		a.toggleFullscreen()
	case idMinimize:
		showWindow.Call(a.window, 6)
	case idMaximize:
		a.toggleMaximized()
	case idCloseWindow:
		postMessage.Call(a.window, wmClose, 0, 0)
	case idSplitRun:
		a.splitDocument()
	case idOCRRun:
		a.startOCR()
	case idOCRUndo:
		if !a.busy && a.view.lastDraftIs("ocr") {
			a.view.undo()
		}
	case idReload:
		if !a.busy && a.view.path != "" {
			if info, err := os.Stat(a.view.path); err == nil && a.mayReplaceDocument() {
				a.loadDocument(a.view.path, getText(a.password), info)
			} else if err != nil {
				a.report("Původní PDF nelze otevřít.")
			}
		}
	case idICCBrowse:
        a.chooseInput(a.customICC, "Barevné profily ICC nebo ICM\x00*.icc;*.icm")
    case idGrayICC:
        a.chooseInput(a.grayCustomICC, "Šedé profily ICC nebo ICM\x00*.icc;*.icm")
        if path := getText(a.grayCustomICC); path != "" { setText(a.view.convertGrayICC, path) }
    case idProofGrayICC:
        a.chooseInput(a.view.convertGrayICC, "Šedé profily ICC nebo ICM\x00*.icc;*.icm")
        if path := getText(a.view.convertGrayICC); path != "" { setText(a.grayCustomICC, path) }
	case idCertBrowse:
		a.chooseInput(a.cert, "Osobní certifikát PKCS#12\x00*.p12;*.pfx")
	case idTrustBrowse:
		a.chooseInput(a.trust, "Kořenové certifikáty\x00*.pem;*.cer;*.crt;*.der")
	case idOutline:
		a.setChecked(a.outline, !a.checked(a.outline))
		if a.checked(a.outline) {
			a.setChecked(a.decrypt, false)
		}
		a.layoutPanel()
	case idDecrypt:
		a.setChecked(a.decrypt, !a.checked(a.decrypt))
		if a.checked(a.decrypt) {
			chooseCombo(a.profiles, 0)
			a.setChecked(a.outline, false)
		}
		a.layoutPanel()
	case idVisible, idRaster:
		h := a.visible
		if id == idRaster {
			h = a.raster
		}
		a.setChecked(h, !a.checked(h))
	case idResetExportOptions:
		a.applyExportOptions(defaults)
		a.updateExportOptions()
	case idRun, idSign, idVerify, idProofApply:
		a.startAction(id)
	}
}

func (a *app) selectPages(req *request, scope, pageRange uintptr) bool {
	switch comboIndex(scope) {
	case exportCurrentPage:
		req.View.Pages = []int{a.view.page + 1}
	case exportThumbnails:
		selected := a.view.selectedPagePositions()
		if len(selected) == 0 {
			a.reportUI(uiOverviewEmpty)
			return false
		}
		for _, position := range selected {
			req.View.Pages = append(req.View.Pages, position+1)
		}
	case exportPageRanges:
		pages, err := parseExportPages(getText(pageRange), len(a.view.order))
		if err != nil {
			if pageError, ok := err.(exportPagesError); ok {
				a.reportUI(pageError.key, pageError.args...)
			} else {
				a.report(err.Error())
			}
			setFocus.Call(pageRange)
			return false
		}
		req.View.Pages = pages
	}
	return true
}

func (a *app) startAction(id int) {
	if a.busy || !a.view.ready() {
		return
	}
	req := a.view.request("")
	var suffix string
	switch id {
    case idProofApply:
        req.Action, req.OpenResult = "transform", true
        req.ColorMode = colorModes[max(0, comboIndex(a.view.convertMode))]
        req.Gray = req.ColorMode == "gray"
        req.GrayProfile = grayProfiles[max(0, comboIndex(a.view.convertGrayProfile))]
        req.GrayCustomICC = strings.Trim(strings.TrimSpace(getText(a.view.convertGrayICC)), "\"")
        req.KProfile = a.export.KProfile
        suffix = map[string]string{"gray":"-Gray", "rich_black":"-cista-cerna", "k_cmyk":"-K-CMYK"}[req.ColorMode]
    case idRun:
        req.Action = "transform"
		preset := a.selectedExportPreset()
		req.Preset = preset.Key
		if !a.selectPages(&req, a.exportScope, a.exportRange) {
			return
		}
		key := a.profileKey()
        if req.Preset == "" {
            req.Outline, req.Decrypt = a.checked(a.outline), a.checked(a.decrypt)
            req.Gray = key == "k"
            req.ColorMode = map[string]string{"k":"gray", "rich_black":"rich_black", "k_cmyk":"k_cmyk"}[key]
            req.GrayProfile = a.export.GrayProfile
            req.GrayCustomICC = strings.Trim(strings.TrimSpace(getText(a.grayCustomICC)), "\"")
            req.KProfile = a.export.KProfile
            req.AllowRaster = req.Gray && !req.Outline && a.checked(a.raster)
            if key != "k" && key != "rich_black" && key != "k_cmyk" {
                req.Profile = key
            }
        } else {
            req.Profile = key
        }
		if req.Profile == "custom" {
			req.CustomICC = strings.Trim(strings.TrimSpace(getText(a.customICC)), "\"")
			if info, err := os.Stat(req.CustomICC); err != nil || !info.Mode().IsRegular() {
				a.reportUI(uiSelectCMYKFile)
				return
			}
		}
		if req.Preset == "" && key == "" && !req.Outline && !req.Decrypt && len(req.View.Pages) == 0 {
			a.reportUI(uiSelectExportSettings)
			return
		}
		switch {
		case req.Preset != "":
			suffix = preset.Suffix
		case req.ColorMode == "gray":
            suffix = "-Gray"
        case req.ColorMode == "rich_black":
            suffix = "-cista-cerna"
        case req.ColorMode == "k_cmyk":
            suffix = "-K-CMYK"
		case key != "":
			suffix = "-" + key
		case req.Decrypt:
			suffix = "-bez-hesla"
		}
		if req.Outline {
			suffix += "-krivky"
		}
		if len(req.View.Pages) > 0 {
			suffix += "-vyber"
		}
	case idSign:
		req.Action, suffix = "sign", "-podepsano"
		req.Certificate = getText(a.cert)
		req.CertPassword, req.Visible = getText(a.certPass), a.checked(a.visible)
		if info, err := os.Stat(req.Certificate); err != nil || !info.Mode().IsRegular() {
			a.reportUI(uiSelectCertificate)
			return
		}
	case idVerify:
		req.Action, req.Trust = "verify", getText(a.trust)
	}
	if req.Action != "verify" {
		if !a.chooseOutput(&req, suffix) {
			return
		}
		if a.view.changed() {
			a.view.includeEdits(&req)
		}
	}
	a.run(req)
}

func (a *app) updateActions() {
	documentReady := a.view.ready()
	if a.documentTools != documentReady {
		a.documentTools = documentReady
		if !documentReady {
			a.layoutAids.temporary = false
			switch a.mode {
			case idProofTab, idSignTab, idExportTab, idSplitTab, idPagesTab, idLayoutAids, idSearchTab, idOCRTab, idTextTab, idObjectsTab:
				if a.tool == idFileTab {
					a.mode = idFileDetails
				} else {
					a.tool, a.mode = idBrowseTab, idBrowseTab
				}
			}
		}
		a.syncSearchAuxiliary()
		a.layout()
	}
	setEnabled(a.detailButtons[idPagesTab], documentReady)
	a.updateRecentActions()
	ready := !a.busy && documentReady
	for _, tab := range a.tabs {
		setEnabled(tab.button, !a.busy)
		setEnabled(tab.close, !a.busy)
		invalidate(tab.button)
	}
	setEnabled(a.tabList, !a.busy)
	setEnabled(a.ocr.undo, ready && a.view.lastDraftIs("ocr"))
	a.updateTextActions()
	a.updateObjectActions()
	a.updateHistoryButtons()
	for _, button := range []uintptr{a.runButton, a.signButton, a.verifyButton, a.scanButton, a.splitButton, a.proofApply, a.ocr.start} {
		setEnabled(button, ready)
	}
	a.updateMergeActions()
	a.view.updateActions()
	a.updateComparisonNavigation()
}

func (a *app) reportStatus(message string) {
	setText(a.status, strings.SplitN(message, "\n", 2)[0])
}

func (a *app) report(message string) {
	a.reportStatus(message)
	if target := a.activeReportField(); target != 0 {
		setText(target, multiline(message))
	}
}

func (a *app) activeReportField() uintptr {
	switch a.mode {
	case idMergeTab:
		return a.merge.report
	case idExportTab:
		return a.exportReport
	case idSignTab:
		return a.signReport
	case idSplitTab:
		return a.splitReport
	case idTextTab:
		return a.textTool.report
    case idOCRTab:
		return a.ocr.report
    case idProofTab:
		return a.view.convertReport
	}
	return 0
}

func (a *app) run(req request) bool {
	if req.ReplaceSource {
		if a.beginJob(req, "Ukládám dokument…") {
			a.waitingOverwrite = true
			a.view.cancelSearch()
			a.startOverwrite()
			return true
		}
		return false
	}
	if a.beginJob(req, "Zpracovávám dokument…") {
		if req.Action == "text_edit" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			a.cancelTextJob = cancel
			a.launchJob(func() response { return runOperationContext(ctx, req) })
		} else {
			a.launchJob(func() response { return runOperation(req) })
		}
		return true
	}
	return false
}

func (a *app) beginJob(req request, status string) bool {
	if a.busy {
		return false
	}
	if req.Action != "merge" && req.Action != "merge_probe" && !a.view.sourceUnchanged() {
		a.report("Vstupní PDF se změnilo. Otevřete je znovu.")
		return false
	}
	a.pending, a.busy = req, true
	a.updateActions()
	a.reportStatus(status)
	return true
}

func (a *app) startOverwrite() {
	// Ghostscript and PDFium may still have the original open. Wait for all readers to
	// finish before replacing the file on Windows.
	if a.waitingOverwrite && !a.view.hasRenderJobs() && !a.comparison.view.hasRenderJobs() && a.view.search.task == nil {
		a.waitingOverwrite = false
		req := a.pending
		a.launchJob(func() response { return runOperation(req) })
	}
}

func (a *app) launchJob(operation func() response) {
	go func() {
		a.results <- operation()
		postMessage.Call(a.window, wmJobFinished, 0, 0)
	}()
}

func (a *app) finish(result response) {
	req := a.pending
	if a.cancelTextJob != nil {
		a.cancelTextJob()
		a.cancelTextJob = nil
	}
	a.busy = false
	a.pending = request{}
	if req.Action == "merge_probe" {
		a.finishMergeProbe(result)
		a.updateActions()
		a.view.schedule()
		return
	}
	if req.Action == "insert_probe" {
		a.finishInsertProbe(req, result)
		a.updateActions()
		a.view.schedule()
		return
	}

	draft := req.Action == "ocr" || req.Action == "text_edit" || req.Action == "insert_pages" || req.Action == "object_edit"
	if result.Status == "merge_encrypted" {
		question := "Uložit sloučené PDF bez šifrování?"
		if req.InsertPages != nil {
			question = "Vložit vybrané strany do nešifrovaného PDF?"
		}
		if notice(a.window, result.Message+"\n\n"+question, 0x134) == 6 {
			req.AllowUnencryptedMerge = true
			if a.run(req) {
				return
			}
		}
		result.Message = "Sloučení zrušeno."
		if req.InsertPages != nil {
			result.Message = "Vložení stránek zrušeno."
		}
	}
	if result.Status == "signed_pdf" {
		message := result.Message
		question := "Uložit kopii bez původních podpisů?"
		if req.Action == "split" {
			question = "Uložit vybrané strany bez původních podpisů?"
		}
		if req.Action == "merge" {
			question = "Sloučit dokumenty do jednoho PDF bez původních podpisů?"
		}
		if req.Action == "ocr" {
			question = "Použít OCR v pracovní kartě bez původních podpisů?"
		}
		if req.Action == "text_edit" {
			question = "Použít úpravu textu v pracovní kartě bez původních podpisů?"
		}
		if req.Action == "object_edit" {
			question = "Použít úpravu objektu v pracovní kartě bez původních podpisů?"
		}
		if req.InsertPages != nil {
			question = "Vložit strany do pracovní karty bez původních podpisů cílového PDF?"
		}
		if req.ReplaceSource {
			message = "Původní PDF obsahuje digitální podpis."
			question = "Přepsat původní soubor a odstranit jeho stávající podpisy?"
		}
		if notice(a.window, message+"\n\n"+question, 0x134) == 6 {
			req.ClearSignatures = true
			if a.run(req) {
				return
			}
		}
		result.Message = "Uložení zrušeno."
	}
	if draft {
		if result.Status == "ok" && !a.applyDocumentDraft(req, result) {
			result = operationError("Upravené PDF se nepodařilo načíst do pracovní karty.")
		}
		if result.Status != "ok" {
			a.retiredDrafts = append(a.retiredDrafts, req.Output)
		}
	}
	if result.Status == "ok" && req.Output != "" && !req.ReplaceSource && !draft {
		a.lastOutput, a.lastPassword = req.Output, ""
		if req.Action == "save_document" ||
			(req.Action == "transform" && (req.Preset == "export" ||
				(req.Preset == "" && !req.Gray && !req.Outline && req.Profile == "" && !req.Decrypt))) {
			a.lastPassword = req.PDFPassword
		}
	}
	if result.Status == "ok" && req.Action == "save_document" {
		if !a.reopenSaved(req) {
			result = operationError("PDF bylo uloženo, ale nepodařilo se znovu otevřít. Karta zůstala otevřená.")
		}
	} else if result.Status == "ok" && req.OpenResult {
		a.openDocument(req.Output, a.lastPassword)
	}
	if result.Status == "ok" && req.Action == "scan" && result.View != nil {
		a.view.showScan(result.View)
		a.reportStatus("Kontrola tiskových desek dokončena.")
	} else {
		message := result.Message
		if result.Status == "ok" && req.Output != "" && !draft {
			message += "\n" + req.Output
		}
		a.reportStatus(message)
		if req.Action == "sign" || req.Action == "verify" {
			setText(a.signReport, multiline(message))
		}
		if req.Action == "transform" {
            target := a.exportReport
            if req.OpenResult { target = a.view.convertReport }
            setText(target, multiline(message))
        }
		if req.Action == "split" {
			if result.Status == "ok" {
				message += "\n" + req.OutputDir
			}
			setText(a.splitReport, multiline(message))
		}
		if req.Action == "merge" {
			setText(a.merge.report, multiline(message))
		}
		if req.Action == "ocr" {
			setText(a.ocr.report, multiline(message))
		}
		if req.Action == "text_edit" || len(req.OCRRegion) > 0 {
			setText(a.textTool.report, multiline(message))
		}
	}
	if req.Action == "sign" {
		setText(a.certPass, "")
	}
	if result.Status == "error" && (req.Action == "save_document" || req.Action == "print_page" || req.Action == "insert_pages" || req.Action == "object_edit") {
		notice(a.window, result.Message, 0x10)
	}
	a.updateActions()
	if active, _, _ := user32.NewProc("GetForegroundWindow").Call(); active == a.window {
		focus, _, _ := user32.NewProc("GetFocus").Call()
		if enabled, _, _ := user32.NewProc("IsWindowEnabled").Call(focus); enabled == 0 {
			setFocus.Call(a.window)
		}
	}
	a.view.schedule()
	a.pruneDrafts()
	if req.Action == "text_edit" {
		next := a.textTool.afterEdit
		a.textTool.afterEdit = nil
		if a.closeAfterTextJob {
			a.closeAfterTextJob = false
			a.closeWindow()
			return
		}
		if result.Status == "ok" && next != nil {
			next()
			return
		}
	}
	if a.closing != nil {
		if result.Status == "ok" && req.Action == "save_document" {
			a.continueCloseDocuments()
		} else {
			a.closing = nil
		}
	}
}

func (a *app) handleInput(msg *winMessage) bool {
	if a.fileMenuInput(msg) || a.recentContextInput(msg) || a.documentMouseContext(msg) {
		return true
	}
	if msg.message == 0x020a { // WM_MOUSEWHEEL
		pos := point{int32(int16(msg.lParam)), int32(int16(msg.lParam >> 16))}
		user32.NewProc("ScreenToClient").Call(a.window, uintptr(unsafe.Pointer(&pos)))
		delta, control := int(int16(msg.wParam>>16)), msg.wParam&8 != 0
		var reader *documentView
		switch {
		case a.comparisonVisible() && contains(a.comparison.view.viewport, pos.x, pos.y):
			reader = a.comparison.view
		case contains(a.view.viewport, pos.x, pos.y):
			reader = a.view
		}
		if reader != nil {
			if !control && a.wheelTargetsEdit(msg.window) {
				reader.pageWheel = pageWheelGesture{}
				return false
			}
			reader.wheel(delta, control, pos, msg.time)
			return true
		}
		if contains(a.thumbnailArea, pos.x, pos.y) || contains(a.comparison.thumbnailArea, pos.x, pos.y) {
			return a.wheelSettings(msg.lParam, delta, control)
		}
		if !control && (contains(a.rightAuxArea, pos.x, pos.y) ||
			(!a.full && a.settingsWidth > 0 && pos.x >= a.detailsArea.left && pos.x < int32(a.width) &&
				pos.y >= a.viewport.top && pos.y < a.viewport.bottom)) {
			return a.wheelSettings(msg.lParam, delta, false)
		}
	}
	if msg.message != 0x0100 {
		return false
	}
	if msg.wParam == 0x1b && (msg.window == a.window && a.cancelCanvasSelections() || a.cancelPageDrag()) {
		return true
	}
	if msg.wParam == 0x1b && msg.window == a.window && !a.full && a.focusedDocumentView().clearCanvasContent() {
		return true
	}
	if msg.wParam == 0x0d && msg.window == a.window && !keyDown(0x11) && !keyDown(0x12) {
		v := a.focusedDocumentView()
		if v.selection.active() && v.selection.pick.mode == textLasso {
			v.finishCanvasSelection()
			return true
		}
		if v.overview {
			v.fitOverviewSelection()
			return true
		}
	}
	if msg.wParam == 0x1b && a.closeFocusedPropertyHelp(msg.window) {
		return true
	}
	if a.objectKey(msg) || a.textKey(msg) || a.documentKey(msg) {
		return true
	}
	if msg.wParam == 0x0d && a.mode == idTextTab && a.textTool.inline.visible && !a.busy {
		p := &a.textTool
		switch msg.window {
		case p.size, p.width, p.leading, p.tracking:
			a.focusInlineEditor()
			return true
		}
	}
	if msg.wParam == 0x0d && msg.window == a.view.moveTarget {
		a.view.moveToPosition()
		return true
	}
	if msg.wParam == 0x0d && msg.window == a.recent.list {
		a.openRecentFile()
		return true
	}
	if msg.wParam == 0x0d && msg.window == a.view.bookmarkPanel.list {
		a.view.goToBookmark()
		return true
	}
	if msg.wParam == 0x0d {
		p := &a.view.search
		switch msg.window {
		case p.input:
			a.view.queueSearch(getText(p.input))
			return true
		case p.list:
			a.view.jumpToSearch(p.selected())
			return true
		}
	}
	if msg.wParam == 0x2e && msg.window == a.view.pageList && a.mode == idPagesTab {
		a.view.deleteSelectedPages()
		return true
	}
	ctrl, _, _ := user32.NewProc("GetKeyState").Call(0x11)
	if ctrl&0x8000 != 0 {
		switch msg.wParam {
		case 'O':
			a.openDialog()
			return true
		case 'F':
			a.showSearch()
			return true
		case 'W':
			a.closeDocument()
			return true
		case 9:
			shift, _, _ := user32.NewProc("GetKeyState").Call(0x10)
			step := 1
			if shift&0x8000 != 0 {
				step = -1
			}
			a.stepDocumentTab(step)
			return true
		case 'S':
			shift, _, _ := user32.NewProc("GetKeyState").Call(0x10)
			a.saveDocument(shift&0x8000 != 0)
			return true
		case 'P':
			a.printDocument()
			return true
		}
	}
	switch msg.wParam {
	case 0x72: // F3
		if a.mode == idSearchTab || msg.window == a.view.search.list {
			shift, _, _ := user32.NewProc("GetKeyState").Call(0x10)
			step := 1
			if shift&0x8000 != 0 {
				step = -1
			}
			a.view.stepSearch(step)
			return true
		}
	case 0x7a: // F11
		a.toggleFullscreen()
		return true
	case 0x1b: // Escape
		if a.full {
			a.toggleFullscreen()
			return true
		}
		if a.dismissLayoutAids() {
			setFocus.Call(a.window)
			return true
		}
		if a.mode == idSearchTab {
			a.activate(idBrowseTab)
			setFocus.Call(a.window)
			return true
		}
	}
	if msg.window == a.window || a.full {
		if a.comparisonVisible() && a.comparison.canvasFocused {
			return a.comparison.view.key(msg.wParam)
		}
		return a.view.key(msg.wParam)
	}
	return false
}

func (a *app) wheelTargetsEdit(receiver uintptr) bool {
	focus, _, _ := user32.NewProc("GetFocus").Call()
	return a.writableEdit(focus) || (receiver != focus && a.writableEdit(receiver))
}

func (a *app) writableEdit(handle uintptr) bool {
	if handle == 0 || handle == a.window || a.readOnlyFields[handle] {
		return false
	}
	if handle == a.textTool.inline.handle {
		return true
	}
	var name [16]uint16
	user32.NewProc("GetClassNameW").Call(handle, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
	return strings.EqualFold(syscall.UTF16ToString(name[:]), "Edit")
}
