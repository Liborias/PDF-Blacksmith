# PDF Blacksmith — JEDINÁ MAPA IKON v6.0

**Platí pro:** [jednotnou specifikaci v6.0](CODEX_MASTER_v6.0.md), [HTML v3.26](rozhrani/cely_vizual_v3.26.html) a [JSON mapu v6.0](mapa_ikon_FINAL_v6.0.json).

Tato mapa konsoliduje **v5.1** (všechny aktualizace magnetu, vodítek, skupin, ořezu) a pozdější samostatný doplněk **v4.7** (nová ikona Nastavení → Rozhraní), plus všechny SVG symboly aktuálního HTML. Historické hlavičky „aktuální v3.9/v3.18“ jsou překonané. Přítomnost ikony **není důkaz existujícího Win32 ID ani oprávnění realizovat novou funkci**.

## Závazná pravidla
- Logo kovadliny je pouze originální autorský vektor `kovadlina křivky.svg`. Ikona `i-snap_magnet` vychází z uživatelem dodaného `magnet(2).svg` a zachovává plnou podkovu/dvě obrysové koncovky; `currentColor`.
- `i-cursor` ↔ `i-text_caret`: jedno dvoustavové tlačítko **ikona ukazuje aktuální stav, nápověda popisuje příští akci**; `i-rect_select` má vlastní tlačítko. Tlačítko lasa přepíná mezi `i-lasso` a `i-crossing_selection`. `i-text_box` vytváří pouze odstavcový textový rámec.
- `i-view_pages` ↔ `i-view_continuous`: druhé dvoustavové tlačítko s ikonou aktuálního režimu. Samostatný třetí Přehled stran používá `i-nav_thumbnails` bez změny tohoto dvoustavu.
- `i-settings_interface`: samostatná ikona v Nastavení → Rozhraní, zdroj `rozhrani/ikony/settings_interface.svg`. Nepoužívat neexistující `i-panel_left` jako ikonu. Jazyk používá `i-language`. Horní Nápověda má prostý znak ? bez kruhového obrysu uvnitř tlačítka.
- `i-guide_eye`, `i-guide_eye_off`, `i-guide_lock`, `i-guide_unlock`, `i-guide_trash`, `i-guide_group`, `i-guide_ungroup`: oko/lock/trash a skupiny jsou nezaměnitelné; koš = smazat, křížek = zavřít. Volba barvy a potvrzení ✓ nepatří do hit-testu ikon.
- `i-toggle_rulers`, `i-toggle_guides`, `i-toggle_grid`, `i-snap_magnet` jsou čtyři různé akce. Form F2, Preflight, Batch a Compare ikony nesmějí založit produktový command ID bez schválení.
- UI SVG mají centrální paletovou barvu `currentColor`, typicky plátno 24×24; pro Go/Win32 převod odpovídající DPI, zachování kontrastu a optických velikostí. CMYK značky, originální swatche a obsah PDF se nepřebarvují.
- **Stav níže:** definice inline `<symbol>` v HTML v3.26. Neznamená fungující Go/PDFium příkaz; ten nutno auditovat v příslušném zdrojovém souboru.

## Soupis všech 147 symbolů (bez duplicit)

### Formuláře (10)

| Symbol | Význam | Status |
|---|---|---|
| `i-form_checkbox` | Checkbox | HTML symbol; Win32 audit |
| `i-form_dropdown` | Rozbalovací pole | HTML symbol; Win32 audit |
| `i-form_edit` | Návrh polí F2 — demo | HTML symbol; Win32 audit |
| `i-form_fill` | Vyplnit F1 | HTML symbol; Win32 audit |
| `i-form_properties` | Vlastnosti pole | HTML symbol; Win32 audit |
| `i-form_radio` | Radio | HTML symbol; Win32 audit |
| `i-form_required` | form required | HTML symbol; Win32 audit |
| `i-form_text_field` | Textové pole | HTML symbol; Win32 audit |
| `i-form_validate` | form validate | HTML symbol; Win32 audit |
| `i-nav_forms` | nav forms | HTML symbol; Win32 audit |

### Navigace (20)

| Symbol | Význam | Status |
|---|---|---|
| `i-nav_batch` | nav batch | HTML symbol; Win32 audit |
| `i-nav_bookmarks` | nav bookmarks | HTML symbol; Win32 audit |
| `i-nav_compare` | nav compare | HTML symbol; Win32 audit |
| `i-nav_file` | nav file | HTML symbol; Win32 audit |
| `i-nav_help` | nav help | HTML symbol; Win32 audit |
| `i-nav_measure` | nav measure | HTML symbol; Win32 audit |
| `i-nav_merge_pdf` | nav merge pdf | HTML symbol; Win32 audit |
| `i-nav_notes` | nav notes | HTML symbol; Win32 audit |
| `i-nav_objects` | nav objects | HTML symbol; Win32 audit |
| `i-nav_ocr` | nav ocr | HTML symbol; Win32 audit |
| `i-nav_page_order` | nav page order | HTML symbol; Win32 audit |
| `i-nav_preflight` | nav preflight | HTML symbol; Win32 audit |
| `i-nav_search_pdf` | nav search pdf | HTML symbol; Win32 audit |
| `i-nav_separations` | nav separations | HTML symbol; Win32 audit |
| `i-nav_settings` | nav settings | HTML symbol; Win32 audit |
| `i-nav_signature` | nav signature | HTML symbol; Win32 audit |
| `i-nav_split_pdf` | nav split pdf | HTML symbol; Win32 audit |
| `i-nav_text` | nav text | HTML symbol; Win32 audit |
| `i-nav_thumbnails` | Miniatury / pružný přehled | HTML symbol; Win32 audit |
| `i-nav_view` | nav view | HTML symbol; Win32 audit |

### Objekty (10)

| Symbol | Význam | Status |
|---|---|---|
| `i-align_objects_center` | align objects center | HTML symbol; Win32 audit |
| `i-group_objects` | group objects | HTML symbol; Win32 audit |
| `i-image_crop` | Ořez | Používá nástroj Objekty |
| `i-image_edit` | image edit | HTML symbol; Win32 audit |
| `i-ungroup_objects` | ungroup objects | HTML symbol; Win32 audit |
| `i-crop_mask` | Ořezová maska / Maska objektu | Definováno ve SVG a atlasu Win32 |
| `i-background_mask` | Maska pozadí | Definováno ve SVG a atlasu Win32 |
| `i-mask_move_content` | Posun obsahu v masce | Definováno ve SVG a atlasu Win32 |
| `i-mask_aspect_lock` | Poměr stran / Zámek poměru stran | Definováno ve SVG a atlasu Win32 |
| `i-mask_invert` | Inverze masky | Definováno ve SVG a atlasu Win32 |

### Ostatní (47)

| Symbol | Význam | Status |
|---|---|---|
| `i-batch_queue` | batch queue | HTML symbol; Win32 audit |
| `i-bold` | bold | HTML symbol; Win32 audit |
| `i-compare_overlay` | compare overlay | HTML symbol; Win32 audit |
| `i-copy` | copy | HTML symbol; Win32 audit |
| `i-delete` | delete | HTML symbol; Win32 audit |
| `i-document_properties` | document properties | HTML symbol; Win32 audit |
| `i-duplicate` | duplicate | HTML symbol; Win32 audit |
| `i-export` | export | HTML symbol; Win32 audit |
| `i-eyedropper` | eyedropper | HTML symbol; Win32 audit |
| `i-font_family` | font family | HTML symbol; Win32 audit |
| `i-font_size` | font size | HTML symbol; Win32 audit |
| `i-font_size_down` | font size down | HTML symbol; Win32 audit |
| `i-font_size_up` | font size up | HTML symbol; Win32 audit |
| `i-font_width` | font width | HTML symbol; Win32 audit |
| `i-highlight` | highlight | HTML symbol; Win32 audit |
| `i-history` | history | HTML symbol; Win32 audit |
| `i-indent_decrease` | indent decrease | HTML symbol; Win32 audit |
| `i-indent_increase` | indent increase | HTML symbol; Win32 audit |
| `i-italic` | italic | HTML symbol; Win32 audit |
| `i-kerning` | kerning | HTML symbol; Win32 audit |
| `i-language` | Nastavení → Jazyk | HTML symbol; Win32 audit |
| `i-letter_spacing` | letter spacing | HTML symbol; Win32 audit |
| `i-line_spacing` | line spacing | HTML symbol; Win32 audit |
| `i-measure_distance` | measure distance | HTML symbol; Win32 audit |
| `i-move_down` | move down | HTML symbol; Win32 audit |
| `i-move_up` | move up | HTML symbol; Win32 audit |
| `i-new_document` | new document | HTML symbol; Win32 audit |
| `i-ocr_region` | ocr region | HTML symbol; Win32 audit |
| `i-open_document` | open document | HTML symbol; Win32 audit |
| `i-panel_right` | panel right | HTML symbol; Win32 audit |
| `i-paragraph_spacing` | paragraph spacing | HTML symbol; Win32 audit |
| `i-paste` | paste | HTML symbol; Win32 audit |
| `i-preflight_run` | preflight run | HTML symbol; Win32 audit |
| `i-redo` | redo | HTML symbol; Win32 audit |
| `i-regular` | regular | HTML symbol; Win32 audit |
| `i-rotate_left` | rotate left | HTML symbol; Win32 audit |
| `i-rotate_right` | rotate right | HTML symbol; Win32 audit |
| `i-save` | save | HTML symbol; Win32 audit |
| `i-save_copy` | save copy | HTML symbol; Win32 audit |
| `i-search` | search | HTML symbol; Win32 audit |
| `i-settings_interface` | Nastavení → Rozhraní | HTML symbol; Win32 audit |
| `i-side_by_side` | side by side | HTML symbol; Win32 audit |
| `i-signature_verify` | signature verify | HTML symbol; Win32 audit |
| `i-strikethrough` | strikethrough | HTML symbol; Win32 audit |
| `i-stroke_width` | stroke width | HTML symbol; Win32 audit |
| `i-underline` | underline | HTML symbol; Win32 audit |
| `i-undo` | undo | HTML symbol; Win32 audit |

### Poznámky (7)

| Symbol | Význam | Status |
|---|---|---|
| `i-comment` | comment | HTML symbol; Win32 audit |
| `i-note_edit` | note edit | HTML symbol; Win32 audit |
| `i-note_filter` | note filter | HTML symbol; Win32 audit |
| `i-note_hide` | note hide | HTML symbol; Win32 audit |
| `i-note_reply` | note reply | HTML symbol; Win32 audit |
| `i-note_show` | note show | HTML symbol; Win32 audit |
| `i-sticky_note` | sticky note | HTML symbol; Win32 audit |

### Separace a barvy (9)

| Symbol | Význam | Status |
|---|---|---|
| `i-color_profile` | color profile | HTML symbol; Win32 audit |
| `i-convert_gray` | Stupně šedi | HTML symbol; Win32 audit |
| `i-convert_k` | Převod do K | HTML symbol; Win32 audit |
| `i-fill_color` | fill color | HTML symbol; Win32 audit |
| `i-overprint_preview` | overprint preview | HTML symbol; Win32 audit |
| `i-print` | print | HTML symbol; Win32 audit |
| `i-print_preview` | print preview | HTML symbol; Win32 audit |
| `i-spot_color` | spot color | HTML symbol; Win32 audit |
| `i-tac_measure` | tac measure | HTML symbol; Win32 audit |

### Stránky a zobrazení (14)

| Symbol | Význam | Status |
|---|---|---|
| `i-fit_page` | fit page | HTML symbol; Win32 audit |
| `i-fit_width` | fit width | HTML symbol; Win32 audit |
| `i-merge_pages` | merge pages | HTML symbol; Win32 audit |
| `i-page_delete` | page delete | HTML symbol; Win32 audit |
| `i-page_extract` | page extract | HTML symbol; Win32 audit |
| `i-page_insert` | page insert | HTML symbol; Win32 audit |
| `i-page_range` | page range | HTML symbol; Win32 audit |
| `i-page_replace` | page replace | HTML symbol; Win32 audit |
| `i-split_pages` | split pages | HTML symbol; Win32 audit |
| `i-view_continuous` | Souvislé | HTML symbol; Win32 audit |
| `i-view_pages` | Po stránkách | HTML symbol; Win32 audit |
| `i-zoom_1_1` | zoom 1 1 | HTML symbol; Win32 audit |
| `i-zoom_in` | zoom in | HTML symbol; Win32 audit |
| `i-zoom_out` | zoom out | HTML symbol; Win32 audit |

### Text a výběr (19)

| Symbol | Význam | Status |
|---|---|---|
| `i-align_center` | align center | HTML symbol; Win32 audit |
| `i-align_justify` | align justify | HTML symbol; Win32 audit |
| `i-align_left` | align left | HTML symbol; Win32 audit |
| `i-align_right` | align right | HTML symbol; Win32 audit |
| `i-baseline_down` | baseline down | HTML symbol; Win32 audit |
| `i-baseline_up` | baseline up | HTML symbol; Win32 audit |
| `i-cursor` | Univerzální šipka | HTML symbol; Win32 audit |
| `i-crossing_selection` | Křížový výběr | Definováno ve SVG a atlasu Win32 |
| `i-lasso` | Univerzální laso | HTML symbol; Win32 audit |
| `i-paragraph` | paragraph | HTML symbol; Win32 audit |
| `i-rect_select` | Univerzální obdélník | HTML symbol; Win32 audit |
| `i-small_caps` | small caps | HTML symbol; Win32 audit |
| `i-subscript` | subscript | HTML symbol; Win32 audit |
| `i-superscript` | superscript | HTML symbol; Win32 audit |
| `i-text_box` | Odstavcový textový rámec | HTML symbol; Win32 audit |
| `i-text_caret` | Textový kurzor | HTML symbol; Win32 audit |
| `i-text_color` | text color | HTML symbol; Win32 audit |
| `i-text_outline` | text outline | HTML symbol; Win32 audit |
| `i-word_spacing` | word spacing | HTML symbol; Win32 audit |

### Vodítka a snap (11)

| Symbol | Význam | Status |
|---|---|---|
| `i-guide_eye` | Zobrazit vodítko | HTML symbol; Win32 audit |
| `i-guide_eye_off` | Skrýt vodítko | HTML symbol; Win32 audit |
| `i-guide_group` | Seskupit vodítka | HTML symbol; Win32 audit |
| `i-guide_lock` | Zamknout | HTML symbol; Win32 audit |
| `i-guide_trash` | Odstranit vodítko | HTML symbol; Win32 audit |
| `i-guide_ungroup` | Rozdělit skupinu | HTML symbol; Win32 audit |
| `i-guide_unlock` | Odemknout | HTML symbol; Win32 audit |
| `i-snap_magnet` | Přichytávání (autorský magnet) | HTML symbol; Win32 audit |
| `i-toggle_grid` | Mřížka | HTML symbol; Win32 audit |
| `i-toggle_guides` | Vodítka | HTML symbol; Win32 audit |
| `i-toggle_rulers` | Pravítka | HTML symbol; Win32 audit |

## Akceptační kontrola
Pro každý symbol otestovat `<use href="#i-...">`, text alternativy, hover/focus/disabled/active, bez ztráty reálného příkazu. V HTML v3.26 nalezeno 141 unikátních symbolů, 0 duplicit. Pokud se zjistí rozdíl mezi touto mapou a produkčním kódem, zapsat ho jako ověřovací úkol, neimprovizovat nové ID.

## Otevřené ověřovací chyby HTML (nikoli zrušená schválená ikona)

Statické prověření volání `miniIcon`, `block` a seznamů odhalilo nejméně tyto **použité, ale v HTML v3.26 chybějící definice symbolů**: `i-panel_left`, `i-arrow_line`, `i-transform_object`, `i-plate_all`, `i-plate_single`. Jde například o přepínání panelu Nástroje, šipkové poznámky, transformace objektů a separační desky. **Nezakládat automaticky nová Win32 ID ani nevymýšlet náhradní grafiku**: ověřit, zda je přesný symbol v původním SVG atlasu/manifestu, případně bezpečně doplnit po schválení UI opravy. Strojová mapa eviduje chybějící odkazy ve `known_missing_html_symbols`; tyto symboly **nejsou** mezi 147 skutečně definovanými.
