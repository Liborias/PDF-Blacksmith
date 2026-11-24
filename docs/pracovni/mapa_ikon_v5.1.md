# PDF Blacksmith – mapa zapojení ikon pro Codex

Tento soubor propojuje novou SVG sadu s **existujícím rozhraním** PDF Blacksmith. Úplná mapa všech ikon, významů a statusů je v `ikonova_mapa.md` a strojově ve `manifest.json`.

## 1. Hlavní navigace

Relevantní zdroj: `workspace_windows.go`, pole `workspaceTabs` a definice ID nástrojů. SVG je pouze vizuální zdroj; žádný příkaz neodstraňovat ani nenahrazovat.

| Zobrazená položka | SVG | Identifikátor v kódu |
|---|---|---|
| Soubor | `nav_file.svg` | `idFileTab` |
| Zobrazení | `nav_view.svg` | `idBrowseTab` |
| Najít v PDF | `nav_search_pdf.svg` | `idSearchTab` |
| Objekty | `nav_objects.svg` | `idObjectsTab` |
| Text | `nav_text.svg` | `idTextTab` |
| OCR | `nav_ocr.svg` | `idOCRTab` |
| Separace | `nav_separations.svg` | `idProofTab` |
| Digitální podpis | `nav_signature.svg` | `idSignTab` |
| Rozdělit PDF | `nav_split_pdf.svg` | `idSplitTab` |
| Sloučit PDF | `nav_merge_pdf.svg` | `idMergeTab` |
| Pořadí stránek | `nav_page_order.svg` | `idPageOrderTool` |
| Poznámky | `nav_notes.svg` | Nový schválený nástroj; nemá se předstírat existující ID |
| Nastavení | `nav_settings.svg` | Schválené samostatné okno; ID doplní implementace |
| Nápověda | `nav_help.svg` | Schválená nápověda; kliknutí otevře hlavní stránku manuálu, F1 kontext |

## 2. Text a objekty

Relevantní zdroje: `text_windows.go`, `typography_windows.go`, `inline_text_windows.go`, `object_windows.go`, `object_selection_windows.go` a související soubory. Nejdříve ověřit odpovídající existující ovladač v aktuálním kódu.

- Výběr textu/objektů: `cursor.svg`, `rect_select.svg`, `lasso.svg`. Obdélník stiskem kolečka myši respektovat podle specifikace.
- Základní typografie: `regular.svg`, `bold.svg`, `italic.svg`, `underline.svg`, `strikethrough.svg`, `font_size_down.svg`, `font_size_up.svg`, `insert_text.svg`, `paragraph.svg`, `align_left.svg`, `align_center.svg`, `align_right.svg`, `align_justify.svg`.
- Již viditelné parametry textového panelu: `font_family.svg`, `font_size.svg`, `font_width.svg`, `line_spacing.svg`, `letter_spacing.svg` (proklad). Ikony nemají automaticky nahradit číselná vstupní pole ani textové štítky.
- Funkční rezerva: `kerning.svg`, `word_spacing.svg`, `paragraph_spacing.svg`, indexy, kapitálky, odsazení a pokročilá sazba; implementovat jen po schválení konkrétní funkce.
- Kopírování, vložení, duplikace a historie: `copy.svg`, `paste.svg`, `duplicate.svg`, `undo.svg`, `redo.svg`, `delete.svg`.
- Objekty do budoucna: `group_objects.svg`, `ungroup_objects.svg`, `node_edit.svg`, `bezier_curve.svg`, `eyedropper.svg`, `fill_color.svg`, `stroke_width.svg`, geometrická transformace a rozložení objektů.

## 3. Separace a tisk

Relevantní zdroje: `viewer_windows.go` (`buildProofPanel`, `syncPlateSelection`), `proof.py`, `COLOR_MANAGEMENT.md`.

- Existující desky, přepínání a kontrola: `nav_separations.svg`, `plate_all.svg`, `plate_single.svg` (schválené doplnění funkce izolace jediné desky), `print.svg`, `color_profile.svg`, `spot_color.svg`.
- Schválené rozšíření: `separation_gray.svg`, `separation_color.svg`, `tac_measure.svg`, `ink_limit.svg`.
- Existující převody: `convert_gray.svg`, `convert_k.svg`; **neodstraňovat** DeviceGray, složenou černou K100 ani explicitní pouze K.
- Vizuálně přidat malé barevné značky C/M/Y/K a přímých barev. Jsou to informační vzorky, nikoli náhrada hlavních přebarvitelných SVG ikon.
- **Rozložení panelu Separace zachovat** podle schválení, bez automatického rozdělování do dvou záložek.

## 4. Stránky, soubory, dokumenty

Relevantní zdroje: `pages_windows.go`, `page_transfer_windows.go`, `document_tabs_windows.go`, `file_menu_windows.go`, `workspace_windows.go`.

- Ukládání a export: `save.svg` (`idSaveDocument`), `save_copy.svg` (`idSaveDocumentAs`), `export.svg` (exportní ovládání), `print.svg` (`idPrintDocument`).
- Pořadí a úpravy stránek: `move_before.svg`, `move_after.svg`, `rotate_left.svg`, `rotate_right.svg`, `page_delete.svg`, `merge_pages.svg`, `split_pages.svg`.
- Funkce související s ochranou práce: `auto_recover.svg` pro schválenou obnovu po havárii.
- Výběr stránky a přiblížení: `nav_thumbnails.svg`, `fit_page.svg`, `fit_width.svg`, `continuous.svg`, `zoom_in.svg`, `zoom_out.svg`, **`zoom_1_1.svg`** (Zobrazení 1:1 / 100 %).

## 5. Poznámky

Nový schválený nástroj. Při implementaci používat standardní anotace PDF oddělené od původního obsahu.

- `nav_notes.svg`: hlavní nástroj Poznámky.
- `sticky_note.svg`, `comment.svg`, `insert_text.svg`, `freehand_pen.svg`, `arrow_line.svg`, `rectangle_shape.svg`, `ellipse_shape.svg`, `highlight.svg`, `underline.svg`, `strikethrough.svg`: schválená základní funkčnost.
- `note_show.svg`, `note_hide.svg`: viditelnost **samostatně pro každý dokument**.
- `annotations_export.svg`: konzultační export s poznámkami. Běžný tisk a export zůstávají **bez poznámek**, běžné uložení pracovní kopie je zachová.
- `stamp.svg`, `note_reply.svg` a další rezervní ikony nejsou schválením dalších funkcí.

## 6. Implementační pravidla

1. **Zachovat aktuální chování nástrojů**; nejprve provést inventuru dostupných ovladačů.
2. Náhrady ikon provést bez změny ovládacích kláves, událostí a dostupnosti funkcí.
3. Přebarvování: SVG používají `currentColor`, kořenový atribut `color` je výchozí vínová; pro Win32 připravit DPI-odpovídající rastrování či renderer SVG a centrální barevný token.
4. SVG mají jednotný `viewBox="0 0 24 24"`, centrovaný optický obsah a sjednocenou základní tloušťku tahu. Zmenšení písma má záměrně menší A.
5. Nové/rezervní ikony nezavádějí automaticky nové funkce ani nové záložky, pokud nebyly zvlášť schváleny.
6. **Logo**: používat výhradně uživatelem dodané `kovadlina křivky.svg`. Balíček ikon žádné náhradní logo neobsahuje.

## 7. Zobrazení 1:1 – nové schválené ovládání (v4.1)

- Soubor: `icons/zoom_1_1.svg`.
- Popisek/tooltip: **Zobrazení 1:1 (100 %)**.
- Umístění: nabídka režimů přiblížení v dolní navigační liště a případně u přibližování v pravém panelu.
- Kliknutí nastaví aktivní PDF na **100% zoom**, přepne z režimu „Přizpůsobit stránce/šířce“ na vlastní zoom, zachová pokud možno střed aktuálně viditelné oblasti a znovu vykreslí z původních PDF dat. V duálním zobrazení se týká pouze aktivního PDF.
- Implementační poznámka ke stávající Go aplikaci: `documentView.zoom` přepočítává PDF body na obrazové pixely; základ při 100 % je `float64(v.app.dpi) / 72`, nikoli doslovně `1`. Nepoužívat zvětšení původního rastrového náhledu jako konečné vykreslení.
- **Pozor na přesný fyzický rozměr:** běžné zobrazení 100 % nemusí odpovídat reálné velikosti 1:1 na konkrétním monitoru. Volitelná kalibrace fyzického měřítka monitoru je samostatný budoucí návrh; nezaměňovat s příkazem 100 %.
- Zkratka: volitelně navrhnout `Ctrl+1` až po ověření kolize a sjednocení klávesových zkratek. Aktuální `Ctrl+0` pro přizpůsobení stránky zachovat.
- Ikona mění pouze měřítko zobrazení, nikoliv velikost stránky či export.


---

# Rozšíření mapy pro budoucí moduly (v4.2)

**Stav: návrh roadmapy, nikoli nový schválený seznam funkcí.** Zde uvedené názvy `nav_*` jsou pouze připravené soubory SVG. V současné větvi `workspace_windows.go` pro tyto čtyři moduly nejsou odpovídající `id*Tab` konstanty. Nezavádět nové ID ani funkce bez odsouhlasení návrhu a testů.

## Formuláře (nově zařazená plánovaná oblast)

- **Hlavní nástroj:** `nav_forms.svg`, zpočátku režimy `form_fill.svg`, `form_edit.svg`, `form_preview.svg`.
- **Pole AcroForm:** `form_text_field.svg`, `form_multiline.svg`, `form_checkbox.svg`, `form_radio.svg`, `form_dropdown.svg`, `form_listbox.svg`, `form_button.svg`, `form_signature_field.svg`.
- **Vlastnosti a ovládání:** `form_properties.svg`, `form_required.svg`, `form_readonly.svg`, `form_tab_order.svg`, `form_validate.svg`, `form_clear.svg`, `form_duplicate_field.svg`, `form_align_fields.svg`.
- **Tisk a data:** `form_flatten.svg` (pouze explicitní operace!), `form_import_data.svg`, `form_export_data.svg` (odložená etapa); `form_date.svg`, `form_calculation.svg`, `form_password.svg` (budoucí rezerva s bezpečnostním posouzením).
- **Kompatibilita:** AcroForm `/AcroForm` a widget anotace; chybějící nebo poškozené appearance streams `/AP` ošetřit. Rozlišení mezi komentáři a widgety je zásadní: **Export bez poznámek NESMÍ zahazovat formulářová pole, jejich vyplněné hodnoty ani formulářové podpisy**.
- **Ochrana dat:** bezpečné ukládání kopie, práce s přístupovými právy, bez nevyžádaného spouštění JavaScriptu; u podepsaných formulářů respektovat platnost podpisu. XFA odhalit a srozumitelně oznámit omezení, neprovádět tichý převod.

## Preflight (navržený nový nástroj)

- **Hlavní nástroj:** `nav_preflight.svg` (obecné `preflight.svg` existuje jako symbol, nikoli jako hotový kompletní panel).
- **Akce a výsledky:** `preflight_run.svg`, `preflight_error.svg`, `preflight_warning.svg`, `preflight_pass.svg`, `preflight_report.svg`, `preflight_fix.svg`.
- **Konkrétní kontroly:** `preflight_image_dpi.svg`, `preflight_fonts.svg`, `preflight_white_overprint.svg`, `preflight_pdfx.svg`, `preflight_hairlines.svg`.
- **Vazby:** znovupoužít `proof.py`, `pdfx.py`, `print_profiles.py`, export a existující barevné testy; rozlišit dílčí PDF/X kontrolu od plnohodnotného Preflight.

## Grafické porovnávání (navržené rozšíření duálního zobrazení)

- **Hlavní nástroj:** `nav_compare.svg`; doplnit k existujícímu `compare_documents.svg`.
- **Zobrazení:** `compare_overlay.svg`, `compare_difference.svg`, `compare_slider.svg`, `compare_sync.svg`, `compare_swap.svg`.
- **Vazby:** `comparison_windows.go`, `reader_render.py`, `viewer_windows.go`; využít stávající synchronizaci stránek/panelů, nevytvářet nezávislý prohlížeč.
- **Pozor:** barevná odchylka po řízeném ICC převodu nemusí být geometrická chyba; oddělit geometrické a barevné porovnání.

## Měření a vodítka (navržený modul)

- **Hlavní nástroj:** `nav_measure.svg`; použít i již existující ikonové položky `measure_distance.svg`, `ruler.svg`, `stroke_width.svg`.
- **Měření:** `measure_angle.svg`, `measure_area.svg`, `measure_stroke.svg`, `measure_calibrate.svg`.
- **Zobrazení a pomůcky:** `guides_toggle.svg`, `grid_toggle.svg`, `snap_to_guides.svg`.
- **Vazby:** stránkové souřadnice v bodech + page boxes; správně zohlednit otočení stránky a různé hodnoty UserUnit. Kalibrace pro fyzicky přesné měřítko je oddělená od ikony `zoom_1_1.svg`.

## Dávkové zpracování (navržený modul)

- **Hlavní nástroj:** `nav_batch.svg`; akce `batch_add_files.svg`, `batch_queue.svg`, `batch_run.svg`, `batch_pause.svg`, `batch_presets.svg`, `batch_output_folder.svg`, `batch_report.svg`.
- **Vazby:** na existující exportní předvolby, `engine.py` a převody; volby výstupu přebírat z existujícího nástroje, nevytvářet druhou konfliktní správu ICC.
- **Bezpečnost:** nikdy implicitně nepřepisovat zdrojová PDF, transakční výstupy, záznam dílčích selhání a možnost přeskočit soubor.

## Pravidla jednotného zapojení ikon

1. Vždy zachovat původní funkčnost a zkratky; samotné SVG není důvod vytvářet nový příkaz.
2. Aktivní nástroj zvýraznit vínově; vypnutý ovladač vizuálně odlišit; tooltip i vysvětlující pole zachovat.
3. Používat `currentColor` a jednotné plátno `24 × 24`; konkrétní velikost ve Win32 kontrolovat při DPI škálování.
4. Schválené uživatelské logo `kovadlina křivky.svg` nenahrazovat návrhem loga z ikonové sady.
5. Plánované moduly zavést v roadmapě s etapami a integračními testy; nepronikají automaticky mezi aktuální schválené implementační požadavky.

## Doplnění v4.3 — HTML návrh v2.8

- **`duplicate` — Duplikovat objekt.** Původní, nezměněné křivky ze schválené sady `PDF_Blacksmith_ikony_nastroju_SVG_v4.1.zip` jsou nově vložené přímo do HTML sprite jako `#i-duplicate`. Původní chybějící ikona se nenahrazuje textovou náhražkou.
- **`history` — Historie úprav.** Nový střídmý piktogram kruhové šipky s ručičkami hodin pro nabídku `Úpravy → Historie úprav`, její pravé pomocné ouško a hlavičku panelu.
- **`undo` / `redo` — Zpět / Znovu**, existující schválené SVG, používané také v ilustračním panelu Historie.
- **Ostatní akční ikony** v jednotlivých skupinách vlastností se zobrazují kompaktně **bez trvalého textu**; titulky zůstávají v nativním tooltipu a `aria-label`. Pod skupinou ikon je dočasný vysvětlující řádek reagující na najetí a fokus.
- Zobrazovaný nástroj / výběr tlačítka zůstává jemně vínový; příkazy typu `Vložit strany` nadále mají důležitý textový popisek.
- **Neimplementovat jinou podobu ikony duplikace**, jestliže není výslovně změněna schválená sada.


## Doplnění v4.4 — panel Text v HTML v3.0

**Zdroj:** schválená sada SVG v4.1 obsahuje téměř všechny potřebné kresby; v HTML v3.0 jsou vložené beze změn. Výjimkou je nový `text_caret.svg`, vytvořený ve stejném vizuálním stylu a uložený samostatně.

| UI / funkce | Ikona SVG | Chování / nápověda |
|---|---|---|
| Upravit existující text – I-kurzor | `text_caret` **NOVÁ** | Umístí bod vložení v textu; není totožná s Vložit nový text |
| Vytvořit textový blok | `text_box` | Založí nový textový objekt |
| Výběr textové oblasti | `rect_select`, `lasso` | Obdélníkový / volný výběr |
| Písmo a velikost | `font_family`, `font_size` | Rodina písma, velikost v bodech |
| Zvětšit / zmenšit typograficky | `font_size_up`, `font_size_down` | A+/A−, následující / předchozí běžná velikost |
| Běžný / tučný / kurzíva / podtržení / přeškrtnutí | `regular`, `bold`, `italic`, `underline`, `strikethrough` | Řezy a dekorace vybraného textu |
| Kapitálky, horní a dolní index | `small_caps`, `superscript`, `subscript` | Volby formátování znaku |
| Šířka znaků | `font_width` | % šířky znaku |
| Proklad písmen | `letter_spacing` | Rozestup sousedních znaků |
| Mezera mezi slovy | `word_spacing` | Mezi slovy, nikoliv písmeny |
| Barva písma | `text_color` | Barevný vzorník a výběr |
| Zvýraznění za znaky | `highlight` | Zvýraznění samotného textu, ne celého bloku |
| Kapátko, obrys | `eyedropper`, `text_outline` | Zatím pouze návrh chování |
| Zarovnání odstavců | `align_left`, `align_center`, `align_right`, `align_justify` | Levé/střed/pravé/blok |
| **Řádkování** | **`line_spacing`** | Správná původní SVG ikona, místo nesouvisejícího `paragraph` |
| Mezery před a za odstavcem | `paragraph_spacing` | Dvě oddělené číselné hodnoty |
| Odsazení zleva/zprava | `indent_increase`, `indent_decrease` | Polohové nastavení, jednotka mm |
| První řádek odstavce | `paragraph` | Samostatná hodnota, může být i záporná |
| Zvětšit/zmenšit odsazení | `indent_increase`, `indent_decrease` | Kroky po 5 mm |
| Posun účaří | `baseline_up` | Svislý posun vybraných znaků |

**Nápověda:** každá ikonová volba musí mít popis v nativním tooltipu a `aria-label`; pod skupinou ikon se při najetí a fokusu dočasně zobrazí **konkrétní účel**, ne obecná věta. Při opuštění ikony se vrátí výchozí nápověda. A+/A− mají stejné průběžné vysvětlení.

**Interakční demo:** v HTML se mění pouze **ukázková sazba v pravém panelu**; PDF text se nepřepisuje. Pro skutečný editor musí Codex po analýze implementace chránit původní kresbu, vložená písma, tvarování znaků, směry textu a stávající práci s textovými objekty. Nepovažovat vizuální panel za hotový PDF typografický engine.

**Textový kurzor:** samostatná kresba uložena v `rozhrani/ikony/text_caret.svg` na plátně 24×24, `stroke=currentColor`, šířka 1.8, umožňuje škálování DPI.


---

## Dodatek v4.5 – aktuální pracovní UI v3.9 (24. 11. 2026)

**Závazná aktuální specifikace návrhu:** [rozhrani/AKTUALNI_UI_CODEX_v3.9.md](rozhrani/AKTUALNI_UI_CODEX_v3.9.md).  
**Aktuální HTML:** [rozhrani/cely_vizual_v3.9.html](rozhrani/cely_vizual_v3.9.html).  
**SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.9.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.9.svg).

Původní mapování produktových ikon a interních příkazů výše zůstává platné, pokud není níže výslovně uvedena nová návrhová vazba. Nové názvy jsou identifikátory *HTML prototypu*, nikoli potvrzená existující Win32 command ID. Codex je nesmí zaměňovat za již implementované ID v produkční aplikaci.

| Ikona v HTML (symbol) | Prvek / účel | Stav / vlastnictví |
|---|---|---|
| `i-cursor`, `i-text_caret` | `#selectPointer`: jedno dvoustavové tlačítko, přepíná šipka / textový kurzor | Jeden výběrový dispatcher, ikona ukazuje **příští** akci; bez rozbalovacího split tlačítka |
| `i-rect_select` | `#selectRect`, Text/OCR/Objekty: univerzální obdélník | Výsledný typ určuje hit-test, ne tři samostatné handlery |
| `i-lasso` | `#selectLasso`, Text/OCR/Objekty: univerzální laso | Výsledný typ určuje hit-test, sdílené dokončení Esc/Enter |
| `i-view_pages`, `i-view_continuous` | `#viewModeCycle`: po stránkách / souvislé zobrazení | Jediné dvoustavové tlačítko; ikona ukazuje příští akci; synchronizovat Zobrazení a stavový řádek |
| `i-snap_magnet` | `#magnetToggle`, Zobrazení → Přichytit | Přepíná aktivitu přichytávání; zatím UI stav, vlastní SnapEngine v plánu |
| `i-toggle_grid` | `#gridToggle`, Zobrazení: mřížka | Zapíná viditelnou netiskovou mřížku; přichytit k mřížce má vlastní volbu |
| `i-toggle_rulers` | `#rulerToggle`, Zobrazení: pravítka | Zobrazuje netisková pravítka; vodítka z nich jsou plán |
| `i-image_edit` | Objekty → Úpravy vybraného rastrového obrázku | Jen pro označený raster: PPI, resampling, režimy, ICC |
| `i-image_crop` | Objekty → Oříznout | Ořez vybraného rastru; zatím demo náhled |
| `i-text_box`, `i-text_caret` | Text → Výběr | Textový rámec není univerzální výběr, bodový text je úloha kurzoru |

**Důležité:** V HTML jsou nové ikony v inline `<symbol>`, v SVG atlasu jako výkresový zdroj; nikoli jako nové garantované funkce PDF. Každý přepínač má `aria-label` / `title`; hover tooltip nad ikonou, textová nápověda na stabilní vyhrazené výšce pod ní. Přidání ikon nemá odstranit stávající ikony z manifestu ani jiná menu.


---

## Dodatek v4.6 – nové kresby a nabídka přichytávání (24. 11. 2026)

**Aktuální HTML:** [rozhrani/cely_vizual_v3.10.html](rozhrani/cely_vizual_v3.10.html).  
**Aktuální SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.10.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.10.svg).  
**Podrobnosti pro Codex:** [rozhrani/UI_PRICHYTAVANI_v3.10.md](rozhrani/UI_PRICHYTAVANI_v3.10.md).

Následující vazby **nahrazují pouze vzhled a položky nabídky přichytávání ve v3.9**, nezakládají další nezávislé výběrové režimy ani nové garantované produkční příkazy.

| Interní symbol HTML | Funkce | Podoba / místo |
| --- | --- | --- |
| \`i-snap_magnet\` | Přichytávání zapnuto/vypnuto | Výrazná silná podkova ve tvaru **U**, dva opticky oddělené póly; druhá horní lišta + Zobrazení |
| \`i-toggle_rulers\` | Zobrazit pravítka | Zachovaná ikonka pravoúhlého měřítka **L**; nedávat k vodítkům |
| \`i-toggle_guides\` | Zobrazit vodítka | **Nová samostatná ikona:** dvě vodorovné úsečky nad sebou nahoře, dvě svislé vedle sebe vlevo; Zobrazení |
| \`i-toggle_grid\` | Zobrazit mřížku | Beze změny; vizuální mřížka se nezaměňuje za volbu „Přichytávat k: Mřížka“ |

**Zobrazení → Přichytávání zapnuto/vypnuto** zůstává hlavním přepínačem \`viewAids.snap\`. Bezprostředně pod ním je **nový vodorovný oddělovač**, poté rozbalovací nabídka **„Přichytávat k:“** s následujícími stávajícími klíči, v pořadí:
\`baselines\` Účaří textu; \`edges\` Okraje objektů; \`guides\` Vodicí linky; \`centers\` Středy objektů; \`sideMidpoints\` Středy stran objektů; \`nodes\` Uzly křivek; \`grid\` Mřížka; \`page\` Okraje a střed stránky; \`spacing\` Rovnoměrné rozestupy.

Jednotlivé hodnoty \`viewAids.targets\` se při vypnutí hlavního magnetu **neztrácejí**. Výběr cíle přichytávání je nezávislý na přepínači jeho viditelnosti (mřížka, vodítka). Checkboxy a rozbalovací nabídka jsou v HTML interaktivní, **samotná geometrie SnapEngine zůstává plánovaná**. Vysoká nabídka má vlastní posuvník, aby se nic neořízlo při menším okně. Všechny ostatní závazné mapy ikon v4.5 beze změny.


---

## Dodatek v4.7 — Separace a magnet v UI v3.11 (24. 11. 2026)

**Aktuální HTML:** [rozhrani/cely_vizual_v3.11.html](rozhrani/cely_vizual_v3.11.html).  
**Aktuální změnové zadání:** [rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md).  
**Atlas SVG:** [rozhrani/ikony/pdf_blacksmith_nove_v3.11.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.11.svg).  
**Samostatná mapa kláves:** [mapa_klavesovych_zkratek_v1.0.md](mapa_klavesovych_zkratek_v1.0.md).

- \`i-snap_magnet\`: **plná vínově červená podkova**, ale dva horní koncové póly jsou výhradně **obrysové**, s prázdným vnitřkem a kontrastnějším tahem. Je to stejná ikona v tlačítku druhé horní lišty i v nabídce Zobrazení.
- \`i-toggle_guides\`: zachovat jako samostatnou ikonu vodítek odlišnou od pravítek \`i-toggle_rulers\`. Obě ikony se využívají v nabídce Zobrazení podle účelu; nepřemapovat mezi sebou.
- \`i-toggle_grid\`: hlavní/vedlejší čáry mřížky definují vizuální hustotu; změna rozteče/podrozdělení je UI parametr, ne nová ikona.
- Sekce Separace → Výstup a konverze: \`i-convert_gray\`, \`i-convert_k\`, \`i-color_profile\` v **svislých třech řádcích**, každý má vedle sebe příslušný dropdown; žádná duplicita ikon, žádné nové předstírané ID Go. V HTML odpovídají \`separationConversionRow\`.
- Dočasná přichytávací linka a pevné vodítko **nejsou ikonami** a nepatří do manifestu; jejich rozdílný vzhled je uveden v UI v3.11.

**Žádný zásah do dosavadních ID akcí**, odstranění historických ikon ani změna existujícího významu SVG. Jde o dokumentaci a prototyp.


---

## Doplněk v4.8 – uživatelská ikona magnetu (24. 11. 2026)

**Referenční návrh HTML:** [rozhrani/cely_vizual_v3.12.html](rozhrani/cely_vizual_v3.12.html).  
**SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.12.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.12.svg).

- Ikona `i-snap_magnet` je od nynějška převzata z **uživatelem dodaného** `magnet(2).svg` (CorelDRAW).
- Geometrie **jedné vyplněné podkovy a dvou obrysových koncovek** je zachována přesně podle předlohy; koncovky nesmějí přesahovat šířku U.
- Nepoužívat pevné odstíny vínové, ani jiné odlišné barvy: `fill="currentColor"` a `stroke="currentColor"` sdílejí barvu s ostatními ikonami daného tlačítka a respektují aktivní/hover stav.
- Zachovat `viewBox="0 0 6348.58 6600.29"`, původní obrysy i polygony, opticky neopravovat proporce ručním přepočtem.
- Všechny ostatní prvky a ikony z v4.7 zůstávají beze změny. Tato změna není samostatným novým příkazem a nijak nemění význam přichytávání.


---

## Dodatek v4.9 – jazyk a kontextové ovládání vodítek (24. 11. 2026)

**Aktuální HTML:** [rozhrani/cely_vizual_v3.16.html](rozhrani/cely_vizual_v3.16.html).  
**Pokyny pro Codex:** [rozhrani/UI_JAZYK_VODITKA_v3.16.md](rozhrani/UI_JAZYK_VODITKA_v3.16.md).  
**SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.16.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.16.svg).

| Grafický prvek | Význam a umístění | Pravidlo |
|---|---|---|
| \`i-language\` (zeměkoule) | **Nastavení → Jazyk**, samostatná sekce | Nezasahuje do starších ikon, jen UI navigace; linka/šířka v souladu s ostatními světlými ikonami, \`currentColor\` |
| Pravé kliknutí na vodítko → **barva** | Malý \`<input type=color>\` v kontextové nabídce | Upravuje jen konkrétní \`guide.color\`, nikoliv \`viewAids.guideColor\` pro budoucí vodítka |
| Pravé kliknutí → **zámeček** | Ikonové tlačítko zamknout/odemknout; v HTML malé inline SVG | Text pouze v \`title\`/\`aria-label\`, nevypisovat zbytečný text vedle ikony |
| Pravé kliknutí → **koš** | Ikonové tlačítko odstranit; v HTML malé inline SVG | Při individuálním nebo globálním zámku deaktivovat; neodstraňovat jiná vodítka |
| Kurzor při najetí na vodítko | Vodorovná linka: \`ns-resize\`; svislá: \`ew-resize\` | Nejde o samostatný soubor SVG, jde o stav kurzoru. Menu se **nezobrazuje na hover**, pouze na pravý klik |

**Zrušená UI varianta:** starší plovoucí menu s tlačítkem **Vybrat** při najetí na linku se nepoužívá. Vodítko se vybere levým kliknutím. Přímá změna aktivního panelu má dočasný charakter a po odznačení se vrací předchozí nástroj.

**Zůstává:** původní autorské \`i-snap_magnet\` (\`currentColor\`, přesná podkova v4.8), samostatné pravítko, vodítka a mřížka, univerzální výběr, konverzní a OCR ikony i všechny existující identifikátory akcí. Výše uvedené ikonky pravého menu jsou v HTML inline cestami a nemají automaticky vytvořit nové Go/Win32 command ID.


---

## Aktualizace mapy v5.0 — skupiny a viditelnost vodítek (24. 11. 2026)

**HTML:** [rozhrani/cely_vizual_v3.17.html](rozhrani/cely_vizual_v3.17.html).  
**Zadání pro Codex:** [rozhrani/UI_PREKLADY_SKUPINY_VODITEK_v3.17.md](rozhrani/UI_PREKLADY_SKUPINY_VODITEK_v3.17.md).  
**SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.17.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.17.svg).

| Symbol | Využití | Požadované chování |
| --- | --- | --- |
| \`i-guide_eye\` | Oko při zobrazeném vodítku/skupině | Klik skryje konkrétní vodítko nebo celou skupinu, nesmí odstranit data |
| \`i-guide_eye_off\` | Přeškrtnuté oko při skrytém vodítku/skupině | Klik zobrazí prvky zpět, bez změny PDF |
| \`i-guide_lock\` | Zamčeno | Označuje zámek, tooltip vyjasňuje odemčení |
| \`i-guide_unlock\` | Odemčeno | Tooltip a klik nabídnou zamknutí |
| \`i-guide_trash\` | Koš | Smaže vodítko nebo celou skupinu i se členy, pouze po ověření zámků |
| \`i-guide_group\` | Seskupení vybraných vodítek | Jediná skupina nad existujícími ID, žádná duplicita vodítek |
| \`i-guide_ungroup\` | Rozdělení skupiny | Zruší příslušnost skupiny bez odstranění členů |

**Stávající ikony** magnetu z uživatelského SVG, mřížky, pravítek, OCR, textu a výběru zůstávají beze změny. Všechny nové symboly používají \`currentColor\` a odpovídající \`aria-label\`/\`title\`, většina ikonových tlačítek neukazuje dlouhý text trvale.

**Upozornění:** V tomto dodatku jde o mapu *HTML demonstrace*, nikoli deklaraci existujících produkčních Win32 příkazových ID. Před implementací všechny ID ověřit v kódu a zabránit kolizím event handlerů. Mapa klávesových zkratek zůstává v1.2.


---

## Aktualizace mapy v5.1 – vzhled, ořez a potvrzení barev (24. 11. 2026)

**HTML:** [rozhrani/cely_vizual_v3.18.html](rozhrani/cely_vizual_v3.18.html).  
**Codex:** [rozhrani/UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md](rozhrani/UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md).  
**SVG atlas:** [rozhrani/ikony/pdf_blacksmith_nove_v3.18.svg](rozhrani/ikony/pdf_blacksmith_nove_v3.18.svg).

- \`i-nav_settings\`: nahradit původní mírně deformované ozubené kolečko **symetrickým větším** obrysem při zachování stejných \`id\` i navazujících příkazů. SVG používá \`currentColor\` a zobrazení 21×21 px v horní liště.
- \`i-guide_trash\`: jediná ikona pro všechny operace **smazání vodítka nebo skupiny** v seznamu, kontextovém menu i v hromadných akcích. Křížek neznamená smazání vodítka, ale pouze zavření/zrušení nebo zavření záložky.
- \`i-guide_group\` a \`i-guide_ungroup\`: viz v5.0, **rozbalovací šipka skupiny** je samostatný stav UI a nemění obsah.
- \`i-guide_eye\`, \`i-guide_eye_off\`, \`i-guide_lock\`, \`i-guide_unlock\`: zachovat stávající mapování a tooltipy.
- Tlačítko **✓** vedle políčka barvy znamená **Použít barvu**; není mazací ani výběrovou ikonou a bez aktivace nesmí hodnotu změnit. U jednotlivého vodítka, skupiny i předvolby nových linek jednotně.
- V **Objekty → Ořez** jsou ve v3.18 základní tvary zastoupené jednoduchými znakovo-vektorovými symboly \`▭ □ ⬭ ○\`; pokud se v produkční sadě nahradí SVG, přidělit jim explicitní ID \`image-crop-rect/square/ellipse/circle\` až po auditu manifestu. Neslučovat tyto tvary s \`i-image_crop\` (příkaz zahájit ořez).
- Barvy akcentů a světlý/tmavý režim **nejsou ikonami**, ale tokeny vzhledu. Neměnit značky CMYK, PDF obsah ani ikony dokumentů při změně palety.
