package main

const (
	uiMergeSourcePageCount          uiMessage = "win32.properties.MergeSourcePageCount"
	uiReadPageCountFirst            uiMessage = "win32.properties.ReadPageCountFirst"
	uiTargetPositionError           uiMessage = "win32.properties.TargetPositionError"
	uiSearchStopped                 uiMessage = "win32.properties.SearchStopped"
	uiSearchingDocument             uiMessage = "win32.properties.SearchingDocument"
	uiSearchProgress                uiMessage = "win32.properties.SearchProgress"
	uiSearchCount                   uiMessage = "win32.properties.SearchCount"
	uiSearchLimit                   uiMessage = "win32.properties.SearchLimit"
	uiNoSearchMatch                 uiMessage = "win32.properties.NoSearchMatch"
	uiOutputIntentUnspecified       uiMessage = "win32.properties.OutputIntentUnspecified"
	uiDocumentInfo                  uiMessage = "win32.properties.DocumentInfo"
	uiProofInfo                     uiMessage = "win32.properties.ProofInfo"
	uiPageSample                    uiMessage = "win32.properties.PageSample"
	uiPageTotal                     uiMessage = "win32.properties.PageTotal"
	uiNoEditableObject              uiMessage = "win32.properties.NoEditableObject"
	uiVectorShape                   uiMessage = "win32.properties.VectorShape"
	uiSelectedImage                 uiMessage = "win32.properties.SelectedImage"
	uiSelectedWidget                uiMessage = "win32.properties.SelectedWidget"
	uiPDFGroup                      uiMessage = "win32.properties.PDFGroup"
	uiSelectedObjects               uiMessage = "win32.properties.SelectedObjects"
	uiMixedCanvasSelection          uiMessage = "win32.properties.MixedCanvasSelection"
	uiCanvasRegionUnavailable       uiMessage = "win32.properties.CanvasRegionUnavailable"
	uiObjectsUnavailable            uiMessage = "win32.properties.ObjectsUnavailable"
	uiPresetManualName              uiMessage = "win32.export.preset.manual.name"
	uiPresetManualDescription       uiMessage = "win32.export.preset.manual.description"
	uiPresetPrintName               uiMessage = "win32.export.preset.print.name"
	uiPresetPrintDescription        uiMessage = "win32.export.preset.print.description"
	uiPresetPressName               uiMessage = "win32.export.preset.press.name"
	uiPresetPressDescription        uiMessage = "win32.export.preset.press.description"
	uiPresetDigitalSmallName        uiMessage = "win32.export.preset.digital-small.name"
	uiPresetDigitalSmallDescription uiMessage = "win32.export.preset.digital-small.description"
	uiPresetDigitalHighName         uiMessage = "win32.export.preset.digital-high.name"
	uiPresetDigitalHighDescription  uiMessage = "win32.export.preset.digital-high.description"
	uiPresetExportName              uiMessage = "win32.export.preset.export.name"
	uiPresetExportDescription       uiMessage = "win32.export.preset.export.description"
	uiPresetFlattenName             uiMessage = "win32.export.preset.flatten.name"
	uiPresetFlattenDescription      uiMessage = "win32.export.preset.flatten.description"
	uiPresetPdfx1AName              uiMessage = "win32.export.preset.pdfx-1a.name"
	uiPresetPdfx1ADescription       uiMessage = "win32.export.preset.pdfx-1a.description"
	uiPresetPdfx3Name               uiMessage = "win32.export.preset.pdfx-3.name"
	uiPresetPdfx3Description        uiMessage = "win32.export.preset.pdfx-3.description"
	uiPresetPdfx4Name               uiMessage = "win32.export.preset.pdfx-4.name"
	uiPresetPdfx4Description        uiMessage = "win32.export.preset.pdfx-4.description"
	uiCollapseSection               uiMessage = "win32.properties.CollapseSection"
	uiExpandSection                 uiMessage = "win32.properties.ExpandSection"
	uiViewMode                      uiMessage = "win32.properties.ViewMode"
	uiViewModeHelp                  uiMessage = "win32.properties.ViewModeHelp"
	uiLayoutAids                    uiMessage = "ui.pomucky.rozvrzeni.6ad91743f5"
	uiLayoutGrid                    uiMessage = "ui.mrizka.862e4fe064"
	uiShowLayoutGrid                uiMessage = "ui.zobrazit.mrizku.8cfde93724"
	uiLayoutGridHelp                uiMessage = "win32.properties.LayoutGridHelp"
	uiDocumentPassword              uiMessage = "win32.properties.DocumentPassword"
	uiDocumentPasswordHelp          uiMessage = "win32.properties.DocumentPasswordHelp"
	uiViewFullscreen                uiMessage = "win32.properties.ViewFullscreen"
	uiFlowContinuous                uiMessage = "win32.properties.FlowContinuous"
	uiScreenBox                     uiMessage = "win32.properties.ScreenBox"
	uiCropBox                       uiMessage = "win32.properties.CropBox"
	uiMediaBox                      uiMessage = "win32.properties.MediaBox"
	uiTrimBox                       uiMessage = "win32.properties.TrimBox"
	uiBleedBox                      uiMessage = "win32.properties.BleedBox"
	uiReloadPassword                uiMessage = "win32.properties.ReloadPassword"
	uiSaveExport                    uiMessage = "win32.properties.SaveExport"
	uiExportPreset                  uiMessage = "win32.properties.ExportPreset"
	uiExportPresetHelp              uiMessage = "win32.properties.ExportPresetHelp"
	uiExportPages                   uiMessage = "win32.properties.ExportPages"
	uiExportPagesHelp               uiMessage = "win32.properties.ExportPagesHelp"
	uiAllPages                      uiMessage = "ui.vsechny.strany.8e2330fc34"
	uiCurrentPage                   uiMessage = "ui.aktualni.strana.e5bc8052eb"
	uiCustomRange                   uiMessage = "win32.properties.CustomRange"
	uiSelectedThumbnails            uiMessage = "ui.vybrane.nahledy.e8b71a8bf7"
	uiRangeExample                  uiMessage = "win32.properties.RangeExample"
	uiOutputColor                   uiMessage = "win32.properties.OutputColor"
	uiOutputColorHelp               uiMessage = "win32.properties.OutputColorHelp"
	uiKeepColors                    uiMessage = "win32.properties.KeepColors"
	uiGrayDevice                    uiMessage = "win32.properties.GrayDevice"
	uiRichBlackK                    uiMessage = "win32.properties.RichBlackK"
	uiExplicitCMYK                  uiMessage = "win32.properties.ExplicitCMYK"
	uiCustomOutputICC               uiMessage = "win32.properties.CustomOutputICC"
	uiBrowseICC                     uiMessage = "win32.properties.BrowseICC"
	uiGrayConversionICC             uiMessage = "win32.properties.GrayConversionICC"
	uiGrayProofICC                  uiMessage = "win32.properties.GrayProofICC"
	uiAutomaticFromDocument         uiMessage = "win32.properties.AutomaticFromDocument"
	uiGrayGamma22                   uiMessage = "win32.properties.GrayGamma22"
	uiGrayGamma18                   uiMessage = "win32.properties.GrayGamma18"
	uiCustomGrayICC                 uiMessage = "win32.properties.CustomGrayICC"
	uiKOutputProfile                uiMessage = "win32.properties.KOutputProfile"
	uiConvertTextOutlines           uiMessage = "win32.properties.ConvertTextOutlines"
	uiRemovePasswordRestrictions    uiMessage = "win32.properties.RemovePasswordRestrictions"
	uiAllowKRaster                  uiMessage = "win32.properties.AllowKRaster"
	uiExportParameters              uiMessage = "win32.properties.ExportParameters"
	uiExportParametersHelp          uiMessage = "win32.properties.ExportParametersHelp"
	uiRenderingIntent               uiMessage = "win32.properties.RenderingIntent"
	uiBlackPoint                    uiMessage = "win32.properties.BlackPoint"
	uiImages                        uiMessage = "win32.properties.Images"
	uiCompression                   uiMessage = "win32.properties.Compression"
	uiPlateCheck                    uiMessage = "win32.properties.PlateCheck"
	uiKRepair                       uiMessage = "win32.properties.KRepair"
	uiFromPDF                       uiMessage = "win32.properties.FromPDF"
	uiPerceptual                    uiMessage = "win32.properties.Perceptual"
	uiRelative                      uiMessage = "win32.properties.Relative"
	uiSaturation                    uiMessage = "win32.properties.Saturation"
	uiAbsolute                      uiMessage = "win32.properties.Absolute"
	uiAutomatic                     uiMessage = "win32.properties.Automatic"
	uiTurnOn                        uiMessage = "win32.properties.TurnOn"
	uiTurnOff                       uiMessage = "win32.properties.TurnOff"
	uiUnchanged                     uiMessage = "win32.properties.Unchanged"
	uiResolution75                  uiMessage = "win32.properties.Resolution75"
	uiResolution150                 uiMessage = "win32.properties.Resolution150"
	uiResolution300                 uiMessage = "win32.properties.Resolution300"
	uiResolution600                 uiMessage = "win32.properties.Resolution600"
	uiAutomaticCompression          uiMessage = "win32.properties.AutomaticCompression"
	uiLossless                      uiMessage = "win32.properties.Lossless"
	uiResetExport                   uiMessage = "win32.properties.ResetExport"
	uiSignDocument                  uiMessage = "ui.podepsat.dokument.bb07a5d6ed"
	uiSignDocumentHelp              uiMessage = "win32.properties.SignDocumentHelp"
	uiPKCSFile                      uiMessage = "win32.properties.PKCSFile"
	uiPropertyNoShortcut            uiMessage = "win32.properties.shortcut.NoShortcut"
	uiPropertyTab                   uiMessage = "win32.properties.shortcut.Tab"
	uiPropertyCloseTab              uiMessage = "win32.properties.shortcut.CloseTab"
	uiPropertySwitchTab             uiMessage = "win32.properties.shortcut.SwitchTab"
	uiPropertySearchResults         uiMessage = "win32.properties.shortcut.SearchResults"
	uiPropertyZoom                  uiMessage = "win32.properties.shortcut.Zoom"
	uiPropertyFullscreen            uiMessage = "win32.properties.shortcut.Fullscreen"
	uiPropertyPageOrder             uiMessage = "win32.properties.shortcut.PageOrder"
	uiPropertyScroll                uiMessage = "win32.properties.shortcut.Scroll"
	uiPropertyPageEnds              uiMessage = "win32.properties.shortcut.PageEnds"
	uiPropertySearchEnter           uiMessage = "win32.properties.shortcut.SearchEnter"
	uiPropertyApplyText             uiMessage = "win32.properties.shortcut.ApplyText"
	uiPropertyPointSelect           uiMessage = "win32.properties.shortcut.PointSelect"
	uiPropertyRectSelect            uiMessage = "win32.properties.shortcut.RectSelect"
	uiPropertyLasso                 uiMessage = "win32.properties.shortcut.Lasso"
	uiPropertyFinishLasso           uiMessage = "win32.properties.shortcut.FinishLasso"
	uiPropertySelectModifiers       uiMessage = "win32.properties.shortcut.SelectModifiers"
	uiPropertyAltDrag               uiMessage = "win32.properties.shortcut.AltDrag"
	uiPropertyDuplicate             uiMessage = "win32.properties.shortcut.Duplicate"
	uiPropertyCropWheel             uiMessage = "win32.properties.shortcut.CropWheel"
	uiPropertySelectPages           uiMessage = "win32.properties.shortcut.SelectPages"
	uiPropertyDeletePages           uiMessage = "win32.properties.shortcut.DeletePages"
	uiChoose                        uiMessage = "win32.properties.Choose"
	uiCertificatePassword           uiMessage = "win32.properties.CertificatePassword"
	uiVisibleSignature              uiMessage = "win32.properties.VisibleSignature"
	uiSaveSignedPDF                 uiMessage = "win32.properties.SaveSignedPDF"
	uiVerifySignatures              uiMessage = "ui.overit.podpisy.23bad6f728"
	uiVerifySignatureHelp           uiMessage = "win32.properties.VerifySignatureHelp"
	uiExtraRootCertificate          uiMessage = "win32.properties.ExtraRootCertificate"
	uiProofProfile                  uiMessage = "win32.properties.ProofProfile"
	uiProofProfileHelp              uiMessage = "win32.properties.ProofProfileHelp"
	uiDefaultCMYK                   uiMessage = "win32.properties.DefaultCMYK"
	uiPrintArea                     uiMessage = "win32.properties.PrintArea"
	uiPagePlates                    uiMessage = "win32.properties.PagePlates"
	uiPagePlatesHelp                uiMessage = "win32.properties.PagePlatesHelp"
	uiAllPlates                     uiMessage = "ui.vsechny.desky.2e69c02aa7"
	uiDisablePlates                 uiMessage = "win32.properties.DisablePlates"
	uiCheckDocument                 uiMessage = "win32.properties.CheckDocument"
	uiCheckSelected                 uiMessage = "win32.properties.CheckSelected"
	uiConvertDocumentColors         uiMessage = "win32.properties.ConvertDocumentColors"
	uiConvertDocumentColorsHelp     uiMessage = "win32.properties.ConvertDocumentColorsHelp"
	uiBlackOnlyCMYK                 uiMessage = "win32.properties.BlackOnlyCMYK"
	uiTargetCMYK                    uiMessage = "win32.properties.TargetCMYK"
	uiConvertAndOpen                uiMessage = "win32.properties.ConvertAndOpen"
	uiOCRPages                      uiMessage = "win32.properties.OCRPages"
	uiOCRPagesHelp                  uiMessage = "win32.properties.OCRPagesHelp"
	uiRecognitionLanguage           uiMessage = "ui.jazyk.rozpoznavani.bd28cb10ed"
	uiRecognitionLanguageHelp       uiMessage = "win32.properties.RecognitionLanguageHelp"
	uiCzechEnglish                  uiMessage = "ui.cestina.anglictina.047e63b32a"
	uiCzechLanguage                 uiMessage = "ui.cestina.f09dc57494"
	uiEnglishLanguage               uiMessage = "ui.anglictina.e1b2b02e38"
	uiRecognizeText                 uiMessage = "ui.rozpoznat.text.48f1a7abf7"
	uiUndoOCR                       uiMessage = "win32.properties.UndoOCR"
	uiMergeSources                  uiMessage = "win32.properties.MergeSources"
	uiMergeSourcesHelp              uiMessage = "win32.properties.MergeSourcesHelp"
	uiAddFiles                      uiMessage = "win32.properties.AddFiles"
	uiAddOpenPDF                    uiMessage = "win32.properties.AddOpenPDF"
	uiAddSelectedPages              uiMessage = "win32.properties.AddSelectedPages"
	uiFilesToMerge                  uiMessage = "win32.properties.FilesToMerge"
	uiMoveUp                        uiMessage = "win32.properties.MoveUp"
	uiMoveDown                      uiMessage = "win32.properties.MoveDown"
	uiRemove                        uiMessage = "win32.properties.Remove"
	uiPagesToAdd                    uiMessage = "win32.properties.PagesToAdd"
	uiMergeRangeExample             uiMessage = "win32.properties.MergeRangeExample"
	uiMergeInputPassword            uiMessage = "win32.properties.MergeInputPassword"
	uiInspectPageCount              uiMessage = "win32.properties.InspectPageCount"
	uiMergeOutput                   uiMessage = "win32.properties.MergeOutput"
	uiMergeOutputHelp               uiMessage = "win32.properties.MergeOutputHelp"
	uiSaveMerged                    uiMessage = "win32.properties.SaveMerged"
	uiMovePages                     uiMessage = "win32.properties.MovePages"
	uiMovePagesHelp                 uiMessage = "win32.properties.MovePagesHelp"
	uiTargetPosition                uiMessage = "win32.properties.TargetPosition"
	uiTargetPositionRange           uiMessage = "win32.properties.TargetPositionRange"
	uiMove                          uiMessage = "win32.properties.Move"
	uiMoveBefore                    uiMessage = "win32.properties.MoveBefore"
	uiMoveAfter                     uiMessage = "win32.properties.MoveAfter"
	uiRotateAndDelete               uiMessage = "win32.properties.RotateAndDelete"
	uiRotateAndDeleteHelp           uiMessage = "win32.properties.RotateAndDeleteHelp"
	uiRotateLeft                    uiMessage = "ui.otocit.vlevo.0ee9474c45"
	uiRotateRight                   uiMessage = "ui.otocit.vpravo.2a1784a847"
	uiDeleteSelection               uiMessage = "win32.properties.DeleteSelection"
	uiUndoChange                    uiMessage = "win32.properties.UndoChange"
	uiSaveEditedCopy                uiMessage = "win32.properties.SaveEditedCopy"
	uiDocumentText                  uiMessage = "win32.properties.DocumentText"
	uiDocumentTextHelp              uiMessage = "win32.properties.DocumentTextHelp"
	uiSearch                        uiMessage = "win32.properties.Search"
	uiClear                         uiMessage = "win32.properties.Clear"
	uiPreviousMatch                 uiMessage = "win32.properties.PreviousMatch"
	uiNextMatch                     uiMessage = "win32.properties.NextMatch"
	uiSearchResults                 uiMessage = "win32.properties.SearchResults"
	uiObjectSelection               uiMessage = "ui.vyber.objektu.bde41a0443"
	uiObjectSelectionHelp           uiMessage = "win32.properties.ObjectSelectionHelp"
	uiSelectByClick                 uiMessage = "win32.properties.SelectByClick"
	uiRectangleSelection            uiMessage = "win32.properties.RectangleSelection"
	uiPolygonLassoKey               uiMessage = "win32.properties.PolygonLassoKey"
	uiSelectPageObject              uiMessage = "win32.properties.SelectPageObject"
	uiObjectActions                 uiMessage = "win32.properties.ObjectActions"
	uiObjectActionsHelp             uiMessage = "win32.properties.ObjectActionsHelp"
	uiCropImage                     uiMessage = "win32.properties.CropImage"
	uiConfirmImageCrop              uiMessage = "win32.properties.ConfirmImageCrop"
	uiCancelImageCrop               uiMessage = "win32.properties.CancelImageCrop"
	uiCropImageHint                 uiMessage = "win32.properties.CropImageHint"
	uiEditImageCropHint             uiMessage = "win32.properties.EditImageCropHint"
	uiCropMovePair                  uiMessage = "win32.properties.CropMovePair"
	uiCropMoveMask                  uiMessage = "win32.properties.CropMoveMask"
	uiCropMoveImage                 uiMessage = "win32.properties.CropMoveImage"
	uiCropMoveHelp                  uiMessage = "win32.properties.CropMoveHelp"
	uiCropImageScale                uiMessage = "win32.properties.CropImageScale"
	uiCropZoomOut                   uiMessage = "win32.properties.CropZoomOut"
	uiCropZoomIn                    uiMessage = "win32.properties.CropZoomIn"
	uiCropMoveImageHint             uiMessage = "win32.properties.CropMoveImageHint"
	uiResetImageCrop                uiMessage = "win32.properties.ResetImageCrop"
	uiChooseImageCrop               uiMessage = "win32.properties.ChooseImageCrop"
	uiCropAspectRatio               uiMessage = "win32.properties.CropAspectRatio"
	uiCropFreeRatio                 uiMessage = "win32.properties.CropFreeRatio"
	uiCropSquareRatio               uiMessage = "win32.properties.CropSquareRatio"
	uiCropRectangle                 uiMessage = "ui.obdelnik.e80dea7331"
	uiCropSquare                    uiMessage = "ui.ctverec.410ecb62ac"
	uiCropOval                      uiMessage = "ui.oval.f6f90ef7db"
	uiCropCircle                    uiMessage = "ui.kruh.74b14ff22d"
	uiObjectCopyKey                 uiMessage = "win32.properties.ObjectCopyKey"
	uiObjectPasteKey                uiMessage = "win32.properties.ObjectPasteKey"
	uiObjectDuplicateKey            uiMessage = "win32.properties.ObjectDuplicateKey"
	uiObjectRemoveKey               uiMessage = "win32.properties.ObjectRemoveKey"
	uiObjectUndoKey                 uiMessage = "win32.properties.ObjectUndoKey"
	uiRecentFiles                   uiMessage = "win32.properties.RecentFiles"
	uiShowMore                      uiMessage = "win32.properties.ShowMore"
	uiShowFewer                     uiMessage = "win32.properties.ShowFewer"
	uiNoRecentPDFs                  uiMessage = "win32.properties.NoRecentPDFs"
	uiRecentPDFList                 uiMessage = "win32.properties.RecentPDFList"
	uiSplitPages                    uiMessage = "win32.properties.SplitPages"
	uiSplitPagesHelp                uiMessage = "win32.properties.SplitPagesHelp"
	uiSplitToFolder                 uiMessage = "win32.properties.SplitToFolder"
	uiTextSelection                 uiMessage = "win32.properties.TextSelection"
	uiTextCursorSelection           uiMessage = "win32.properties.TextCursorSelection"
	uiTextRectangleSelection        uiMessage = "win32.properties.TextRectangleSelection"
	uiTextLassoSelection            uiMessage = "win32.properties.TextLassoSelection"
	uiTextParagraphFrame            uiMessage = "win32.properties.TextParagraphFrame"
	uiNewTextFrameHint              uiMessage = "win32.properties.NewTextFrameHint"
	uiNewPointTextHint              uiMessage = "win32.properties.NewPointTextHint"
	uiTextSelectionInstructions     uiMessage = "win32.properties.TextSelectionInstructions"
	uiHideSelectionHelp             uiMessage = "win32.properties.HideSelectionHelp"
	uiTextSelectionHelp             uiMessage = "win32.properties.TextSelectionHelp"
	uiMixedTypography               uiMessage = "win32.properties.MixedTypography"
	uiTextFormat                    uiMessage = "win32.properties.TextFormat"
	uiTextFormatHelp                uiMessage = "win32.properties.TextFormatHelp"
	uiFont                          uiMessage = "win32.properties.Font"
	uiFontSize                      uiMessage = "win32.properties.FontSize"
	uiSmallerType                   uiMessage = "win32.properties.SmallerType"
	uiLargerType                    uiMessage = "win32.properties.LargerType"
	uiTypeSizes                     uiMessage = "win32.properties.TypeSizes"
	uiRegular                       uiMessage = "win32.properties.Regular"
	uiBold                          uiMessage = "win32.properties.Bold"
	uiItalic                        uiMessage = "win32.properties.Italic"
	uiUnderline                     uiMessage = "win32.properties.Underline"
	uiAlignLeft                     uiMessage = "ui.zarovnat.vlevo.49f20a229f"
	uiAlignCenter                   uiMessage = "win32.properties.AlignCenter"
	uiAlignRight                    uiMessage = "ui.zarovnat.vpravo.564917d6d5"
	uiJustify                       uiMessage = "win32.properties.Justify"
	uiTextWidth                     uiMessage = "win32.properties.TextWidth"
	uiLineSpacing                   uiMessage = "win32.properties.LineSpacing"
	uiTracking                      uiMessage = "win32.properties.Tracking"
	uiTextWidthHelp                 uiMessage = "win32.properties.TextWidthHelp"
	uiLineSpacingHelp               uiMessage = "win32.properties.LineSpacingHelp"
	uiTrackingHelp                  uiMessage = "win32.properties.TrackingHelp"
	uiCopyText                      uiMessage = "win32.properties.CopyText"
	uiUndoTextEdit                  uiMessage = "win32.properties.UndoTextEdit"
	uiOCRSelection                  uiMessage = "win32.properties.OCRSelection"
	uiWarnFlattenedText             uiMessage = "win32.properties.WarnFlattenedText"
	uiWarnFlattenedTransparency     uiMessage = "win32.properties.WarnFlattenedTransparency"
	uiEmptyPageRange                uiMessage = "win32.properties.EmptyPageRange"
	uiInvalidPageRange              uiMessage = "win32.properties.InvalidPageRange"
	uiSelectCMYKFile                uiMessage = "win32.properties.SelectCMYKFile"
	uiSelectExportSettings          uiMessage = "win32.properties.SelectExportSettings"
	uiSelectCertificate             uiMessage = "win32.properties.SelectCertificate"
	uiInvalidTextMetrics            uiMessage = "win32.properties.InvalidTextMetrics"
	uiSelectValidFile               uiMessage = "win32.properties.SelectValidFile"
	uiSelectExistingFolder          uiMessage = "win32.properties.SelectExistingFolder"
)

// Native property labels and reports that the shared interface catalog does not cover.
var propertyTranslations = map[uiMessage]uiTranslation{
	uiMergeSourcePageCount:          {Czech: "Strany k přidání · ve zdroji %d", English: "Pages to add · source has %d"},
	uiReadPageCountFirst:            {Czech: "Nejprve načtěte počet stran.", English: "Read the page count first."},
	uiTargetPositionError:           {Czech: "Zadejte cílovou pozici od 1 do %d.", English: "Enter a target position from 1 to %d."},
	uiSearchStopped:                 {Czech: "Hledání zastaveno.", English: "Search stopped."},
	uiSearchingDocument:             {Czech: "Prohledávám dokument…", English: "Searching the document…"},
	uiSearchProgress:                {Czech: "Prohledáno %d z %d stran. Shod zatím %d.", English: "Searched %d of %d pages. Matches so far: %d."},
	uiSearchCount:                   {Czech: "Nalezené shody: %d", English: "Matches found: %d"},
	uiSearchLimit:                   {Czech: "Nalezené shody: %d\nZobrazeno prvních %d shod.", English: "Matches found: %d\nShowing the first %d matches."},
	uiNoSearchMatch:                 {Czech: "Žádná shoda v textové vrstvě PDF.", English: "No matches in the PDF text layer."},
	uiOutputIntentUnspecified:       {Czech: "Neuveden", English: "Unspecified"},
	uiDocumentInfo:                  {Czech: "%d stran\nVýstupní záměr PDF\n%s\nV paměti %d z %d stran", English: "%d pages\nPDF output intent\n%s\n%d of %d pages in memory"},
	uiProofInfo:                     {Czech: "Výstupní záměr\n%s\nSimulace  %s", English: "Output intent\n%s\nSimulation  %s"},
	uiPageSample:                    {Czech: "Str. %d · plocha / maximální tón · %d dpi", English: "Page %d · coverage / maximum tint · %d dpi"},
	uiPageTotal:                     {Czech: "z %d", English: "of %d"},
	uiNoEditableObject:              {Czech: "Na tomto místě není upravitelný objekt.", English: "There is no editable object here."},
	uiVectorShape:                   {Czech: "Vektorový tvar", English: "Vector shape"},
	uiSelectedImage:                 {Czech: "Obrázek", English: "Image"},
	uiSelectedWidget:                {Czech: "Formulářové pole", English: "Form field"},
	uiPDFGroup:                      {Czech: "Skupina PDF", English: "PDF group"},
	uiSelectedObjects:               {Czech: "Vybrané objekty: %d", English: "Selected objects: %d"},
	uiMixedCanvasSelection:          {Czech: "Výběr · znaky %d, objekty %d", English: "Selection · %d characters, %d objects"},
	uiCanvasRegionUnavailable:       {Czech: "Obsah vyznačené oblasti se nepodařilo načíst.", English: "Could not load the contents of the selected area."},
	uiObjectsUnavailable:            {Czech: "Objekty na této straně nelze vybrat.", English: "Objects on this page cannot be selected."},
	uiPresetManualName:              {Czech: "Vlastní nastavení", English: "Custom settings"},
	uiPresetManualDescription:       {Czech: "Ruční volba barev, desky K, převodu textu a komprese.", English: "Set colors, K-only output, text conversion, and compression manually."},
	uiPresetPrintName:               {Czech: "PDF pro tisk", English: "PDF for print"},
	uiPresetPrintDescription:        {Czech: "Původní barvy a průhlednost. Obrázky do 300 dpi, bezeztrátová komprese.", English: "Original colors and transparency. Images up to 300 dpi with lossless compression."},
	uiPresetPressName:               {Czech: "PDF připravené pro tiskárnu", English: "Print-ready PDF"},
	uiPresetPressDescription:        {Czech: "Procesní barvy do CMYK se zvoleným ICC profilem. Obrázky do 300 dpi.", English: "Convert process colors to CMYK with the selected ICC profile. Images up to 300 dpi."},
	uiPresetDigitalSmallName:        {Czech: "Digitální PDF · malá velikost", English: "Digital PDF · small size"},
	uiPresetDigitalSmallDescription: {Czech: "sRGB pro obrazovku. Obrázky do 150 dpi a komprese JPEG.", English: "sRGB for screens. Images up to 150 dpi with JPEG compression."},
	uiPresetDigitalHighName:         {Czech: "Digitální PDF · vysoká kvalita", English: "Digital PDF · high quality"},
	uiPresetDigitalHighDescription:  {Czech: "sRGB pro obrazovku. Původní rozlišení obrázků a bezeztrátová komprese.", English: "sRGB for screens. Original image resolution with lossless compression."},
	uiPresetExportName:              {Czech: "PDF pro další zpracování", English: "PDF for further processing"},
	uiPresetExportDescription:       {Czech: "Kopie se zachováním původních barev, křivek, textu i průhlednosti.", English: "Copy with original colors, outlines, text, and transparency."},
	uiPresetFlattenName:             {Czech: "Zploštěné PDF", English: "Flattened PDF"},
	uiPresetFlattenDescription:      {Czech: "Každá strana jako jeden obraz v sRGB při 300 dpi. Text už nelze označovat.", English: "Each page becomes a 300 dpi sRGB image. Text cannot be selected."},
	uiPresetPdfx1AName:              {Czech: "PDF/X-1a:2003", English: "PDF/X-1a:2003"},
	uiPresetPdfx1ADescription:       {Czech: "Procesní CMYK s výstupním profilem ICC. Průhlednost se zploští při 300 dpi.", English: "Process CMYK with an ICC output profile. Transparency is flattened at 300 dpi."},
	uiPresetPdfx3Name:               {Czech: "PDF/X-3:2003", English: "PDF/X-3:2003"},
	uiPresetPdfx3Description:        {Czech: "Tisková varianta PDF/X-3 v procesním CMYK. Průhlednost se zploští při 300 dpi.", English: "PDF/X-3 print variant in process CMYK. Transparency is flattened at 300 dpi."},
	uiPresetPdfx4Name:               {Czech: "PDF/X-4", English: "PDF/X-4"},
	uiPresetPdfx4Description:        {Czech: "Výstupní profil CMYK pro tisk. Živá průhlednost a vektory se zachovají.", English: "CMYK output profile for print. Live transparency and vectors are preserved."},
	uiCollapseSection:               {Czech: "sbalit", English: "collapse"},
	uiExpandSection:                 {Czech: "rozbalit", English: "expand"},
	uiViewMode:                      {Czech: "Režim zobrazení", English: "View mode"},
	uiViewModeHelp:                  {Czech: "Ctrl + kolečko přiblíží PDF pod kurzorem. Tiskový rámeček určuje oblast zobrazenou v náhledu.", English: "Ctrl + mouse wheel zooms at the pointer. The page box determines the area shown in the preview."},
	uiLayoutGridHelp:                {Czech: "Hlavní čáry jsou po 5 mm a jemné po 1 mm. Mřížka se zobrazuje jen v náhledu.", English: "Main lines are spaced 5 mm apart and minor lines 1 mm apart. The grid appears only in the preview."},
	uiDocumentPassword:              {Czech: "Heslo k otevření PDF", English: "PDF open password"},
	uiDocumentPasswordHelp:          {Czech: "Zadejte heslo a načtěte otevřený dokument znovu.", English: "Enter the password and reload the open document."},
	uiViewFullscreen:                {Czech: "Zobrazit na celou obrazovku · F11", English: "Full screen · F11"},
	uiFlowContinuous:                {Czech: "Souvisle", English: "Continuous"},
	uiScreenBox:                     {Czech: "Rámeček PDF v náhledu", English: "Preview page box"},
	uiCropBox:                       {Czech: "Ořez zobrazení", English: "Crop box"},
	uiMediaBox:                      {Czech: "Celý arch", English: "Full sheet"},
	uiTrimBox:                       {Czech: "Čistý formát", English: "Trim box"},
	uiBleedBox:                      {Czech: "Spadávka", English: "Bleed box"},
	uiReloadPassword:                {Czech: "Načíst znovu s heslem", English: "Reload with password"},
	uiSaveExport:                    {Czech: "Uložit exportované PDF…", English: "Save exported PDF…"},
	uiExportPreset:                  {Czech: "Předvolba PDF", English: "PDF preset"},
	uiExportPresetHelp:              {Czech: "Předvolba určuje účel a výchozí nastavení exportu. Otevřením nápovědy zobrazíte popis vybrané předvolby.", English: "A preset sets the purpose and default export options. Open help for a description of the selected preset."},
	uiExportPages:                   {Czech: "Strany k exportu", English: "Pages to export"},
	uiExportPagesHelp:               {Czech: "Vyberte rozsah stran nebo použijte strany označené v panelu Náhledy.", English: "Enter a page range or use pages selected in the Thumbnails panel."},
	uiCustomRange:                   {Czech: "Vlastní výběr", English: "Custom range"},
	uiRangeExample:                  {Czech: "Např. 1, 3–5, 8", English: "e.g. 1, 3–5, 8"},
	uiOutputColor:                   {Czech: "Barevný výstup", English: "Output color"},
	uiOutputColorHelp:               {Czech: "Výstupní režim a profil ICC určují, jak se při exportu převedou barvy PDF.", English: "The output mode and ICC profile determine how PDF colors are converted during export."},
	uiKeepColors:                    {Czech: "Zachovat barvy", English: "Preserve colors"},
	uiGrayDevice:                    {Czech: "Stupně šedi · DeviceGray", English: "Grayscale · DeviceGray"},
	uiRichBlackK:                    {Czech: "Složená černá → K100", English: "Rich black → K100"},
	uiExplicitCMYK:                  {Czech: "Pouze K · explicitní CMYK", English: "K only · explicit CMYK"},
	uiCustomOutputICC:               {Czech: "Vlastní výstupní profil ICC", English: "Custom output ICC profile"},
	uiBrowseICC:                     {Czech: "Vybrat ICC…", English: "Choose ICC…"},
	uiGrayConversionICC:             {Czech: "ICC profil převodu na stupně šedi", English: "Grayscale conversion ICC profile"},
	uiGrayProofICC:                  {Czech: "ICC profil převodu do šedi", English: "Grayscale conversion ICC profile"},
	uiAutomaticFromDocument:         {Czech: "Automaticky podle dokumentu", English: "Automatic from document"},
	uiGrayGamma22:                   {Czech: "Gray Gamma 2.2", English: "Gray Gamma 2.2"},
	uiGrayGamma18:                   {Czech: "Gray Gamma 1.8", English: "Gray Gamma 1.8"},
	uiCustomGrayICC:                 {Czech: "Vlastní ICC Gray…", English: "Custom Gray ICC…"},
	uiKOutputProfile:                {Czech: "Výstupní CMYK profil pro K", English: "CMYK output profile for K"},
	uiConvertTextOutlines:           {Czech: "Převést text na křivky", English: "Convert text to outlines"},
	uiRemovePasswordRestrictions:    {Czech: "Pouze odstranit omezení heslem", English: "Only remove password restrictions"},
	uiAllowKRaster:                  {Czech: "Povolit rastrovou opravu K (%d dpi)", English: "Allow K raster repair (%d dpi)"},
	uiExportParameters:              {Czech: "Parametry exportu", English: "Export parameters"},
	uiExportParametersHelp:          {Czech: "U ručního exportu lze nastavit převod barev, převzorkování obrázků, kompresi a rozlišení kontroly desek.", English: "Manual export lets you set color conversion, image resampling, compression, and plate-check resolution."},
	uiRenderingIntent:               {Czech: "Záměr reprodukce", English: "Rendering intent"},
	uiBlackPoint:                    {Czech: "Černý bod", English: "Black point"},
	uiImages:                        {Czech: "Obrázky", English: "Images"},
	uiCompression:                   {Czech: "Komprese", English: "Compression"},
	uiPlateCheck:                    {Czech: "Kontrola desek", English: "Plate check"},
	uiKRepair:                       {Czech: "Oprava K", English: "K repair"},
	uiFromPDF:                       {Czech: "Podle PDF", English: "From PDF"},
	uiPerceptual:                    {Czech: "Perceptuální", English: "Perceptual"},
	uiRelative:                      {Czech: "Relativní", English: "Relative"},
	uiSaturation:                    {Czech: "Saturace", English: "Saturation"},
	uiAbsolute:                      {Czech: "Absolutní", English: "Absolute"},
	uiAutomatic:                     {Czech: "Automaticky", English: "Automatic"},
	uiTurnOn:                        {Czech: "Zapnout", English: "On"},
	uiTurnOff:                       {Czech: "Vypnout", English: "Off"},
	uiUnchanged:                     {Czech: "Beze změny", English: "Unchanged"},
	uiResolution75:                  {Czech: "75 dpi", English: "75 dpi"},
	uiResolution150:                 {Czech: "150 dpi", English: "150 dpi"},
	uiResolution300:                 {Czech: "300 dpi", English: "300 dpi"},
	uiResolution600:                 {Czech: "600 dpi", English: "600 dpi"},
	uiAutomaticCompression:          {Czech: "Automatická", English: "Automatic"},
	uiLossless:                      {Czech: "Bezeztrátová", English: "Lossless"},
	uiResetExport:                   {Czech: "Obnovit výchozí parametry", English: "Restore defaults"},
	uiSignDocumentHelp:              {Czech: "K vytvoření podpisu vyberte osobní certifikát PKCS#12, zadejte jeho heslo a uložte podepsanou kopii.", English: "Select a personal PKCS#12 certificate, enter its password, and save a signed copy."},
	uiPKCSFile:                      {Czech: "Certifikát PKCS#12 (.p12 nebo .pfx)", English: "PKCS#12 certificate (.p12 or .pfx)"},
	uiPropertyNoShortcut:            {Czech: "Tato sekce nemá vlastní zkratky.", English: "This section has no dedicated shortcuts."},
	uiPropertyTab:                   {Czech: "Další nebo předchozí ovládací prvek", English: "Next or previous control"},
	uiPropertyCloseTab:              {Czech: "Zavřít PDF", English: "Close PDF"},
	uiPropertySwitchTab:             {Czech: "Další nebo předchozí dokumentová karta", English: "Next or previous document tab"},
	uiPropertySearchResults:         {Czech: "Další nebo předchozí výsledek hledání", English: "Next or previous search result"},
	uiPropertyZoom:                  {Czech: "Přiblížit nebo oddálit stránku", English: "Zoom page in or out"},
	uiPropertyFullscreen:            {Czech: "Zapnout nebo opustit celou obrazovku", English: "Enter or exit full screen"},
	uiPropertyPageOrder:             {Czech: "Předchozí nebo následující stránka", English: "Previous or next page"},
	uiPropertyScroll:                {Czech: "Posun zobrazení", English: "Scroll the view"},
	uiPropertyPageEnds:              {Czech: "První nebo poslední stránka", English: "First or last page"},
	uiPropertySearchEnter:           {Czech: "Spustit hledání v poli nebo přejít na označený výsledek", English: "Search from the input or jump to the selected result"},
	uiPropertyApplyText:             {Czech: "Použít úpravu při psaní na stránce", English: "Apply changes while editing text on the page"},
	uiPropertyPointSelect:           {Czech: "Bodový výběr objektu", English: "Point selection"},
	uiPropertyRectSelect:            {Czech: "Obdélníkový výběr", English: "Rectangle selection"},
	uiPropertyLasso:                 {Czech: "Laso a druhým stiskem křížový výběr", English: "Lasso, then press again for Crossing Selection"},
	uiPropertyFinishLasso:           {Czech: "Dokončit laso nebo křížový výběr", English: "Finish the lasso or Crossing Selection"},
	uiPropertySelectModifiers:       {Czech: "Shift přidává, Ctrl přepíná označení a Shift+Ctrl odebírá.", English: "Shift adds to the selection, Ctrl toggles items and Shift+Ctrl removes items."},
	uiPropertyAltDrag:               {Czech: "Alt + tažení  ·  Přesunout výběr", English: "Alt + drag  ·  Move selection"},
	uiPropertyDuplicate:             {Czech: "Duplikovat výběr objektů", English: "Duplicate selected objects"},
	uiPropertyCropWheel:             {Czech: "Kolečko nad maskou  ·  Měnit měřítko v režimu Posun obsahu v masce", English: "Wheel over the mask  ·  Scale the image in Move image inside mask mode"},
	uiPropertySelectPages:           {Czech: "Shift+klik na miniaturu vybere rozsah. Ctrl+klik přidá nebo odebere jednotlivou stránku.", English: "Shift+click a thumbnail for a range. Ctrl+click to add or remove one page."},
	uiPropertyDeletePages:           {Czech: "Smazat stránky vybrané v panelu miniatur", English: "Delete pages selected in the thumbnail panel"},
	uiChoose:                        {Czech: "Vybrat…", English: "Choose…"},
	uiCertificatePassword:           {Czech: "Heslo k certifikátu", English: "Certificate password"},
	uiVisibleSignature:              {Czech: "Viditelný podpis v novém poli", English: "Visible signature in a new field"},
	uiSaveSignedPDF:                 {Czech: "Uložit podepsané PDF…", English: "Save signed PDF…"},
	uiVerifySignatureHelp:           {Czech: "Ověření zkontroluje podpisy v otevřeném PDF. Volitelně můžete přidat důvěryhodný kořenový certifikát.", English: "Verification checks the signatures in the open PDF. You can add a trusted root certificate."},
	uiExtraRootCertificate:          {Czech: "Doplňkový kořenový certifikát", English: "Optional root certificate"},
	uiProofProfile:                  {Czech: "Náhledový profil tisku", English: "Print proof profile"},
	uiProofProfileHelp:              {Czech: "Profil a tisková oblast určují simulaci separací na obrazovce.", English: "The profile and print area control the on-screen separation preview."},
	uiDefaultCMYK:                   {Czech: "Podle PDF / výchozí CMYK", English: "From PDF / default CMYK"},
	uiPrintArea:                     {Czech: "Tisková oblast", English: "Print area"},
	uiPagePlates:                    {Czech: "Tiskové desky na stránce", English: "Printing plates on this page"},
	uiPagePlatesHelp:                {Czech: "Přepínejte desky v náhledu. Kontrola změří skutečné pokrytí na vybraných stranách nebo v celém dokumentu.", English: "Toggle plates in the preview. The check measures actual ink coverage for selected pages or the entire document."},
	uiDisablePlates:                 {Czech: "Vypnout všechny", English: "Turn all off"},
	uiCheckDocument:                 {Czech: "Zkontrolovat celý dokument", English: "Check entire document"},
	uiCheckSelected:                 {Czech: "Zkontrolovat vybrané náhledy", English: "Check selected thumbnails"},
	uiConvertDocumentColors:         {Czech: "Převod barev aktuálního dokumentu", English: "Convert document colors"},
	uiConvertDocumentColorsHelp:     {Czech: "Zvolte barevný režim a potřebný profil ICC. Převod otevře novou kopii.", English: "Choose a color mode and ICC profile. Conversion opens a new copy."},
	uiBlackOnlyCMYK:                 {Czech: "Pouze K · CMYK C0 M0 Y0", English: "K only · CMYK C0 M0 Y0"},
	uiTargetCMYK:                    {Czech: "Cílový tiskový profil CMYK", English: "Target CMYK print profile"},
	uiConvertAndOpen:                {Czech: "Použít převod a otevřít kopii…", English: "Convert and open a copy…"},
	uiOCRPages:                      {Czech: "Strany pro OCR", English: "Pages for OCR"},
	uiOCRPagesHelp:                  {Czech: "OCR vyhledává text i v obrázcích a křivkách. Zvolte stránky nebo označte jejich náhledy.", English: "OCR recognizes text in images and outlines. Choose pages or select their thumbnails."},
	uiRecognitionLanguageHelp:       {Czech: "Zvolte jazyk podle textu na vybraných stránkách.", English: "Choose the language of the text on the selected pages."},
	uiUndoOCR:                       {Czech: "Vrátit OCR", English: "Undo OCR"},
	uiMergeSources:                  {Czech: "Zdroje ke sloučení", English: "Files to merge"},
	uiMergeSourcesHelp:              {Czech: "Přidejte soubory nebo vybrané stránky a upravte pořadí jejich vložení.", English: "Add files or selected pages and arrange them in the order to merge them."},
	uiAddFiles:                      {Czech: "Přidat PDF soubory…", English: "Add PDF files…"},
	uiAddOpenPDF:                    {Czech: "Přidat otevřené PDF", English: "Add open PDF"},
	uiAddSelectedPages:              {Czech: "Přidat výběr stran", English: "Add selection"},
	uiFilesToMerge:                  {Czech: "Soubory ke sloučení", English: "Files to merge"},
	uiMoveUp:                        {Czech: "Nahoru", English: "Up"},
	uiMoveDown:                      {Czech: "Dolů", English: "Down"},
	uiRemove:                        {Czech: "Odebrat", English: "Remove"},
	uiPagesToAdd:                    {Czech: "Strany k přidání", English: "Pages to add"},
	uiMergeRangeExample:             {Czech: "Všechny strany · např. 1, 3–5", English: "All pages · e.g. 1, 3–5"},
	uiMergeInputPassword:            {Czech: "Heslo k otevření vybraného souboru", English: "Password for the selected file"},
	uiInspectPageCount:              {Czech: "Načíst počet stran", English: "Read page count"},
	uiMergeOutput:                   {Czech: "Výstup", English: "Output"},
	uiMergeOutputHelp:               {Czech: "Sloučený dokument se uloží jako nový soubor.", English: "The merged document is saved as a new file."},
	uiSaveMerged:                    {Czech: "Uložit sloučené PDF…", English: "Save merged PDF…"},
	uiMovePages:                     {Czech: "Přesun stránek", English: "Move pages"},
	uiMovePagesHelp:                 {Czech: "Strany přesuňte na cílovou pozici nebo tažením v panelu Náhledy.", English: "Move pages to a target position or drag them in the Thumbnails panel."},
	uiTargetPosition:                {Czech: "Cílová pozice", English: "Target position"},
	uiTargetPositionRange:           {Czech: "Cílová pozice (1–%d)", English: "Target position (1–%d)"},
	uiMove:                          {Czech: "Přesunout", English: "Move"},
	uiMoveBefore:                    {Czech: "Posunout před", English: "Move before"},
	uiMoveAfter:                     {Czech: "Posunout za", English: "Move after"},
	uiRotateAndDelete:               {Czech: "Otočení a odstranění", English: "Rotate and delete"},
	uiRotateAndDeleteHelp:           {Czech: "Otočení a odstranění se použijí na vybrané náhledy. Změny lze vrátit.", English: "Rotation and deletion apply to selected thumbnails. You can undo the changes."},
	uiDeleteSelection:               {Czech: "Smazat vybrané", English: "Delete selected"},
	uiUndoChange:                    {Czech: "Vrátit změnu", English: "Undo change"},
	uiSaveEditedCopy:                {Czech: "Uložit upravenou kopii…", English: "Save edited copy…"},
	uiDocumentText:                  {Czech: "Text v dokumentu", English: "Document text"},
	uiDocumentTextHelp:              {Czech: "Hledání prochází textovou vrstvu PDF včetně textu rozpoznaného nástrojem OCR.", English: "Search includes the PDF text layer and text recognized by OCR."},
	uiSearch:                        {Czech: "Hledat", English: "Search"},
	uiClear:                         {Czech: "Vymazat", English: "Clear"},
	uiPreviousMatch:                 {Czech: "Předchozí shoda", English: "Previous match"},
	uiNextMatch:                     {Czech: "Další shoda", English: "Next match"},
	uiSearchResults:                 {Czech: "Výsledky hledání", English: "Search results"},
	uiObjectSelectionHelp:           {Czech: "Vybraný objekt přesunete šipkou nebo tažením s Alt. Laso uzavřete Enterem nebo kliknutím na první bod.", English: "Drag the selected object with the arrow tool or hold Alt while dragging. Close the lasso with Enter or by clicking its first point."},
	uiSelectByClick:                 {Czech: "Výběr kliknutím (V)", English: "Click selection (V)"},
	uiRectangleSelection:            {Czech: "Obdélníkový výběr (M)", English: "Rectangle selection (M)"},
	uiPolygonLassoKey:               {Czech: "Mnohoúhelníkové laso (L)", English: "Polygon lasso (L)"},
	uiSelectPageObject:              {Czech: "Vyberte objekt na stránce.", English: "Select an object on the page."},
	uiObjectActions:                 {Czech: "Akce s objekty", English: "Object actions"},
	uiCropImage:                     {Czech: "Ořez", English: "Crop image"},
	uiConfirmImageCrop:              {Czech: "Potvrdit ořez", English: "Confirm crop"},
	uiCancelImageCrop:               {Czech: "Zrušit ořez", English: "Cancel crop"},
	uiCropImageHint:                 {Czech: "Výřez vyznačte tažením. Potvrďte Enterem nebo tlačítkem.", English: "Drag to mark the area to keep, then confirm or press Enter."},
	uiEditImageCropHint:             {Czech: "Výřez vyznačte v rámci celého obrázku.", English: "Draw a new crop within the full image."},
	uiCropMovePair:                  {Czech: "Přesunout masku i obrázek", English: "Move mask and image"},
	uiCropMoveMask:                  {Czech: "Přesunout jen ořezovou masku", English: "Move crop mask only"},
	uiCropMoveImage:                 {Czech: "Posun obsahu v masce", English: "Move image inside mask"},
	uiCropMoveHelp:                  {Czech: "Volba určuje, co se pohne při tažení uvnitř ořezové masky.", English: "Controls what moves when you drag inside the crop mask."},
	uiCropImageScale:                {Czech: "Měřítko obsahu %.0f %%", English: "Image scale %.0f %%"},
	uiCropZoomOut:                   {Czech: "Oddálit obsah masky od jejího středu", English: "Zoom out from the center of the crop mask"},
	uiCropZoomIn:                    {Czech: "Přiblížit obsah masky ke středu výřezu", English: "Zoom in on the center of the crop mask"},
	uiCropMoveImageHint:             {Czech: "Tažením posunete obrázek. Kolečkem nad maskou změníte jeho měřítko.", English: "Drag to move the image. Scroll over the mask to scale it."},
	uiResetImageCrop:                {Czech: "Obnovit ořez", English: "Reset crop"},
	uiChooseImageCrop:               {Czech: "Zvolte tvar, poměr stran nebo spusťte nástroj ořezu.", English: "Choose a shape or aspect ratio, or click the crop icon to start."},
	uiCropAspectRatio:               {Czech: "Poměr stran", English: "Aspect ratio"},
	uiCropFreeRatio:                 {Czech: "Volný", English: "Free"},
	uiCropSquareRatio:               {Czech: "1 : 1 · Čtverec", English: "1 : 1 · Square"},
	uiCropRectangle:                 {Czech: "Obdélník", English: "Rectangle"},
	uiCropSquare:                    {Czech: "Čtverec", English: "Square"},
	uiCropOval:                      {Czech: "Ovál", English: "Oval"},
	uiCropCircle:                    {Czech: "Kruh", English: "Circle"},
	uiObjectActionsHelp:             {Czech: "Akce pracují s aktuálním výběrem. Změny lze vrátit klávesami Ctrl + Z.", English: "Actions apply to the current selection. Press Ctrl + Z to undo changes."},
	uiObjectCopyKey:                 {Czech: "Kopírovat · Ctrl+C", English: "Copy · Ctrl+C"},
	uiObjectPasteKey:                {Czech: "Vložit · Ctrl+V", English: "Paste · Ctrl+V"},
	uiObjectDuplicateKey:            {Czech: "Duplikovat · Ctrl+D", English: "Duplicate · Ctrl+D"},
	uiObjectRemoveKey:               {Czech: "Smazat · Delete", English: "Delete · Delete"},
	uiObjectUndoKey:                 {Czech: "Vrátit změnu · Ctrl+Z", English: "Undo change · Ctrl+Z"},
	uiRecentFiles:                   {Czech: "Nedávné soubory", English: "Recent files"},
	uiShowMore:                      {Czech: "Zobrazit více", English: "Show more"},
	uiShowFewer:                     {Czech: "Zobrazit méně", English: "Show fewer"},
	uiNoRecentPDFs:                  {Czech: "Zatím tu nejsou žádná nedávná PDF.", English: "There are no recent PDFs yet."},
	uiRecentPDFList:                 {Czech: "Nedávno otevřené soubory PDF", English: "Recently opened PDFs"},
	uiSplitPages:                    {Czech: "Strany k rozdělení", English: "Pages to split"},
	uiSplitPagesHelp:                {Czech: "Každá vybraná strana se uloží do samostatného PDF ve zvolené složce.", English: "Each selected page is saved as its own PDF in the chosen folder."},
	uiSplitToFolder:                 {Czech: "Vybrat složku a rozdělit PDF…", English: "Choose folder and split PDF…"},
	uiTextSelection:                 {Czech: "Výběr textu", English: "Text selection"},
	uiTextSelectionHelp:             {Czech: "Základní a rozšířený výběr", English: "Basic and extended selection"},
	uiMixedTypography:               {Czech: "Smíšené", English: "Mixed"},
	uiTextCursorSelection:           {Czech: "Textový kurzor (T)", English: "Text cursor (T)"},
	uiTextRectangleSelection:        {Czech: "Výběr obdélníkem (M)", English: "Rectangle selection (M)"},
	uiTextLassoSelection:            {Czech: "Výběr lasem (L)", English: "Lasso selection (L)"},
	uiTextParagraphFrame:            {Czech: "Odstavcový textový rámec · vytvořit tažením", English: "Paragraph text frame · drag to create"},
	uiNewTextFrameHint:              {Czech: "Pište do rámce. Text potvrdíte kliknutím mimo rámec nebo stiskem Ctrl+Enter.", English: "Type in the frame. Click outside the frame or press Ctrl+Enter to apply the text."},
	uiNewPointTextHint:              {Czech: "Pište od zvoleného místa. Text potvrdíte kliknutím mimo něj nebo stiskem Ctrl+Enter.", English: "Type at the chosen position. Click outside the text or press Ctrl+Enter to apply it."},
	uiHideSelectionHelp:             {Czech: "Skrýt nápovědu k výběru", English: "Hide selection help"},
	uiTextFormat:                    {Czech: "Typografie", English: "Typography"},
	uiTextFormatHelp:                {Czech: "Změny se projeví ve výběru přímo v PDF. Proklad je mezera mezi znaky, řádkování určuje vzdálenost mezi řádky.", English: "Changes appear in the selection in the PDF. Tracking is the space between characters. Line spacing sets the space between lines."},
	uiFont:                          {Czech: "Písmo", English: "Font"},
	uiFontSize:                      {Czech: "Velikost (pt)", English: "Size (pt)"},
	uiSmallerType:                   {Czech: "Zmenšit na předchozí typografickou velikost", English: "Decrease to the previous type size"},
	uiLargerType:                    {Czech: "Zvětšit na následující typografickou velikost", English: "Increase to the next type size"},
	uiTypeSizes:                     {Czech: "Typografické velikosti v bodech", English: "Type sizes in points"},
	uiRegular:                       {Czech: "Běžný řez (Regular)", English: "Regular"},
	uiBold:                          {Czech: "Tučné", English: "Bold"},
	uiItalic:                        {Czech: "Kurzíva", English: "Italic"},
	uiUnderline:                     {Czech: "Podtržené", English: "Underline"},
	uiAlignCenter:                   {Czech: "Zarovnat na střed", English: "Center"},
	uiJustify:                       {Czech: "Zarovnat do bloku", English: "Justify"},
	uiTextWidth:                     {Czech: "Šířka (pt)", English: "Width (pt)"},
	uiLineSpacing:                   {Czech: "Řádkování (×)", English: "Leading (×)"},
	uiTracking:                      {Czech: "Proklad (pt)", English: "Tracking (pt)"},
	uiTextWidthHelp:                 {Czech: "Šířka textového bloku v bodech. Automatická šířka zachová původní řádky.", English: "Text block width in points. Automatic width keeps the original lines."},
	uiLineSpacingHelp:               {Czech: "Vzdálenost mezi účařími řádků jako násobek velikosti písma.", English: "Distance between baselines as a multiple of the font size."},
	uiTrackingHelp:                  {Czech: "Přídavná mezera mezi písmeny v bodech. Záporná hodnota písmena přiblíží.", English: "Extra spacing between letters in points. Negative values bring letters closer together."},
	uiCopyText:                      {Czech: "Kopírovat text", English: "Copy text"},
	uiUndoTextEdit:                  {Czech: "Vrátit úpravu", English: "Undo edit"},
	uiOCRSelection:                  {Czech: "OCR výběru", English: "OCR selection"},
	uiWarnFlattenedText:             {Czech: "Pozor: text ve zploštěném PDF nepůjde označit.", English: "Text in the flattened PDF cannot be selected."},
	uiWarnFlattenedTransparency:     {Czech: "Pozor: průhlednost se při exportu zploští.", English: "Transparency will be flattened during export."},
	uiEmptyPageRange:                {Czech: "Zadejte výběr stran, například 1, 3–5, 8.", English: "Enter a page range, for example 1, 3–5, 8."},
	uiInvalidPageRange:              {Czech: "Neplatná strana nebo rozsah %q. Zadejte strany od 1 do %d.", English: "Invalid page or range %q. Enter pages from 1 to %d."},
	uiSelectCMYKFile:                {Czech: "Vyberte soubor s výstupním profilem CMYK ICC nebo ICM.", English: "Choose a CMYK output profile file in ICC or ICM format."},
	uiSelectExportSettings:          {Czech: "Vyberte strany k exportu, úpravu nebo výstupní profil.", English: "Choose pages to export, a conversion option, or an output profile."},
	uiSelectCertificate:             {Czech: "Vyberte osobní certifikát.", English: "Choose a personal certificate."},
	uiInvalidTextMetrics:            {Czech: "Zkontrolujte velikost písma, šířku rámečku a řádkování.", English: "Check the font size, frame width, and line spacing."},
	uiSelectValidFile:               {Czech: "Vyberte soubor s platnou cestou.", English: "Choose a file with a valid path."},
	uiSelectExistingFolder:          {Czech: "Vyberte existující složku v souborovém systému.", English: "Choose an existing folder."},
	uiTextSelectionInstructions: {
		Czech:   "Základní výběr\nT · Kliknutím do textu otevřete editaci. Tažením v textu označíte znaky. Kliknutím na volné místo stránky vložíte bodový text.\nM · Tažením vymezíte obdélník.\nL · Laso vybírá prvky celé uvnitř. Druhým stiskem L zapnete křížový výběr otevřenou čárou. V těchto režimech kliknutí kolečkem kreslí druhý typ výběru. Enter nebo dvojklik výběr dokončí.\nV · Kliknutím vyberete stávající odstavec.\nRámec · Tažením na stránce vytvoříte nový odstavec.\n\nRozšířený výběr\nShift přidává, Ctrl přepíná označení a Shift+Ctrl odebírá.\nAlt + tažení přesune vybraný text po stránce.\nObdélník tažený doleva vybírá i protnuté prvky. Tažený doprava vybírá jen prvky celé uvnitř.",
		English: "Basic selection\nT · Click existing text to edit it. Drag across text to select characters. Click a blank area on the page to insert point text.\nM · Drag to draw a rectangle.\nL · Lasso selects fully enclosed items. Press L again for an open Crossing Selection. Middle clicks draw the alternate selection. Enter or double-click finishes the selection.\nV · Click to select an existing paragraph.\nFrame · Drag on the page to create a new paragraph.\n\nExtended selection\nShift adds to the selection, Ctrl toggles items, and Shift+Ctrl removes items.\nAlt-drag moves the selected text on the page.\nDrag a rectangle left to select intersecting items. Drag right to select fully enclosed items.",
	},
}

var exportPresetMessages = map[string]struct{ name, description uiMessage }{
	"":              {name: uiPresetManualName, description: uiPresetManualDescription},
	"print":         {name: uiPresetPrintName, description: uiPresetPrintDescription},
	"press":         {name: uiPresetPressName, description: uiPresetPressDescription},
	"digital-small": {name: uiPresetDigitalSmallName, description: uiPresetDigitalSmallDescription},
	"digital-high":  {name: uiPresetDigitalHighName, description: uiPresetDigitalHighDescription},
	"export":        {name: uiPresetExportName, description: uiPresetExportDescription},
	"flatten":       {name: uiPresetFlattenName, description: uiPresetFlattenDescription},
	"pdfx-1a":       {name: uiPresetPdfx1AName, description: uiPresetPdfx1ADescription},
	"pdfx-3":        {name: uiPresetPdfx3Name, description: uiPresetPdfx3Description},
	"pdfx-4":        {name: uiPresetPdfx4Name, description: uiPresetPdfx4Description},
}
