# PDF Blacksmith — JEDINÝ IMPLEMENTAČNÍ PLÁN v6.0

**Autorita:** [CODEX_MASTER_v6.0.md](CODEX_MASTER_v6.0.md). Novější výslovné přání autora: **nejprve stabilizovat a sjednotit rozhraní aplikace**, nikoliv rozpracovávat F1. Tento plán nahrazuje staré implementační pořadí v „specifikace_Codex_v5.0.md“, „plan_rozsireni_v2.0.md“, „F1_plan_implementace.md“ a iniciačních dodatcích. Je to plán, nikoli automatická autorizace změn.

## Řízení etap
- Každý krok má **vstupní audit → jeden malý diff → testy → výstupní kontrola → schválení dalšího kroku**. Bez výslovného souhlasu žádný commit ani push.
- Zachovat současné produkční Go/Win32 UI a PDFium, nepřepisovat funkční části podle HTML od nuly.
- Všechny nalezené změny Codexu v rozpracovaném F1 nejprve vypsat; bez resetu a automatického mazání **pozastavit další práci na formulářích**.
- Přehled hotového vs naplánovaného dokládat konkrétním zdrojovým souborem a testem. HTML v3.26 není důkaz Go implementace.

## Fáze 0 – převzetí a inventura (nyní, bez změny kódu)
| Krok | Výstup | Akceptace |
|---|---|---|
| 0.1 | Stav větve, pracovní strom, lokální změny a poslední commity | Žádná změna cizího souboru, bezpečný diff |
| 0.2 | Matice existujících UI modulů/handlerů a procesů PDF | U každého A/B/D/P, odkaz na soubor, riziko |
| 0.3 | Srovnání produkčního rozhraní s HTML v3.26 | Seznam konkrétních neshod a požadovaných ovládání |
| 0.4 | Důsledky rozpracovaných F1 kroků | Co bylo změněno, co funguje a co je nutno ponechat rozpracované |
| 0.5 | Doporučení nejmenší bezpečné UI úpravy | Jeden konkrétní další úkol, přesné soubory, testy, rollback |

**Výstup teď:** jediná zpráva uživateli bez práce v kódu. Codex se nesmí odvolávat na staré „F1 je první priorita“ pro změnu tohoto pořadí.

## Fáze 1 – design systém a nejzákladnější rozhraní
**1A. Tokeny a palety:** 6×light/dark, žádné pevné vínové relikty, tmavé inputy ve světlých profilech, sémantika swatch vs UI. Ověřit v native Win32 způsob barvení; uvést screenshoty, pokud běží na Windows. **Nevylepšovat automaticky render PDF**.

**1B. Ikonový systém:** Součástí je ověření chybějících SVG symbolů `i-panel_left`, `i-arrow_line`, `i-transform_object`, `i-plate_all` a `i-plate_single` (v aktuálním HTML jsou některé volány, ale nejsou definované); prověřit původní atlas, nekreslit náhrady bez souhlasu.  normativní mapování symbolů a statusů; první horní lišta a druhá menu/ikony, logo a magnet z původních autorských kreseb, bez změny command ID. Zkontrolovat 150/200 % DPI a přístupnost.

**1C. Základní docky:** Nástroje vlevo, Vlastnosti vpravo, miniatury a jeden pomocný panel; korektně se skládat, změna šířky a vracení stavu, práce v malém okně. Neodpojovat stávající OCR/separace.

**Vstup:** schválení 1A nebo 1B nebo 1C. **Done:** 12 motivů, 3 DPI, 4 šířky, nepoškozené PDF a všechna původní menu.

## Fáze 2 – konzistentní chování aplikace
**2A. Menu, textové vs ikonové:** horní Soubor/Zobrazení/Nástroje/Úpravy/Formuláře/Pomoc bez rámečků; levá navigace a dropdown bez obdélníkových rámů s levým akcentním pruhem; aktivní ikona druhé lišty s výrazným stavem. Žádné změny funkčnosti příkazů.

**2B. Vlastnosti aktivního nástroje:** související podskupiny s oddělovačem a sbalením/rozbalením, bez dlouhých prázdných hint slotů. Otazník vedle sekce ukazuje metodický popis; **varování/validace se nikdy neschovávají**. Vlastnosti zachovat při sbalení.

**2C. Nápověda + lokalizace:** nativní tooltipy odpovídají CZ/EN, pole popisů podle fokusu, čtení pro klávesnici; volba jazyka v Nastavení/Jazyk, téma v Nastavení/Rozhraní i přes horní půlený přepínač. Lokalizační katalog 809 CZ/EN klíčů, nezasahovat do PDF.

**2D. Kontextové pomocné panely:** automatické Výsledky při hledání a korektní návrat, dočasné Pomůcky rozvržení při interakci s vodítkem. Uživatelova manuální volba má přednost. Nereplikovat další pravý dock.

**Done:** zachování všech stávajících ikon i kláves; návrat z vyhledávání; žádné zbytkové tmavé plochy; přístupný focus.

## Fáze 3 – zobrazení stránek a interaktivní plátno
**3A. Žádné dekorativní rámy stránky:** zcela odstranit externí vínový outline v hlavní ploše ve stránkovém, souvislém i šířkovém režimu; skutečný obsah PDF nerušit.

**3B. Normální navigace:** po stránkách scroll uvnitř listu → při dosažení kraje další/předchozí, ochrana před několikanásobným přeskokem, souvislý režim bez přerušení.

**3C. Zoom:** přizpůsobit stránce, šířce, 100 %; Ctrl+kolečko v PDF mění dokument, v miniaturách pouze jejich velikost, mimo obě zóny bez zachycení. V HTML Chromium/Edge ověřit wheel capture/passive false/preventDefault; produkční Win32 mít lokální obsluhu bez globálních zásahů.

**3D. Adaptivní Přehled stran:** až po dokončení 3A–3C a schválení. Samostatný třetí režim (dvoustavové tlačítko pages/continuous zůstává jen dvoustavové), flex-wrap, výběr klik/Ctrl/Shift, Enter fit vybrané na výšku, sdílený výběr s miniaturami. **Skutečné přetahování a hromadné transformace** řešit zvlášť po zajištění Undo/Redo/page modelu. Současné HTML v3.26 mění jen ukázkové pořadí.

**Done:** testy 3 stránek + velký soubor, zoom 30–900 % podle návrhu, DPI, focus, scroll anchoring, otáčení stran a dual view.

## Fáze 4 – výběr a editační pomůcky
**4A. Dispatcher:** sjednotit šipku/I-kurzor, univerzální obdélník/laso, Text/Objekty/OCR bez paralelních handlerů; nejdříve provést audit všech aktuálních kláves T/V/M/L.

**4B. Ovládací prvky:** Text Výběr čtyři ikony na jediné řadě + nápověda ?, Typography mixed, Objekty raster editor, OCR s vlastní nápovědou, zachování převodu text ↔ křivky jen jako operace nad výběrem.

**4C. Pravítka a vodítka:** osa X/Y správně, pouze drag z pravítka, Ctrl+klik vs Ctrl+drag, per-page stav, skupiny a zámky, barvy „Použít“, kontext pravým klikem bez „Vybrat“, dočasná aktivace Pomůcek rozvržení. Nahradit pouze dekorativní paralelní prvky, ne PDF data.

**4D. SnapEngine:** zvlášť po prokázané přesné geometrii a výkonu. Nikdy nepřepisovat aktuální výběr, neukládat vodítka do tiskového PDF.

**Done:** pro každé gesto alespoň pozitivní/negativní test, overlap, PDF s rotací, rozložení CZ/EN, historií Undo/Redo a dokumentem bez rizika.

## Fáze 5 – F1 vyplňování formulářů (až na další souhlas)
**5A. F1-0:** inventura skutečné PDFium/API, anotací, podpisů a UI.
**5B. F1-1:** detekce AcroForm/Fields/Kids/Widget, Tab pořadí, XFA, návrh přechodu do Formulářů bez automatického přepnutí.
**5C. F1-2/F1-3:** vyplňování standardních polí a jejich pravý panel; zvýraznění polí pouze UI; Text/Objekty/OCR nepoškozeny.
**5D. F1-4/F1-5:** kontrola a bezpečné ukládání /V /AS /AP, Uložit přesto, diakritika, AutoRecover, podpisy a nezávislé PDF čtečky.

**Poznámka:** Dřívější dokument F1 plánoval UI uvnitř F1 až později. Nově má **základní shell přednost**, ale vytváření samotné funkcionality F1 není tímto autorizováno.

## Fáze 6 – návrhy F2, data a F3
**F2:** tvorba polí a vlastností ve Win32. HTML v3.26 je pouze ukázka, nikoli libovolný povolený write do AcroForm.
**Importy:** CSV UTF-8 vč. BOM/quoted, XLSX list/range, Google Sheets read-only OAuth, bezpečná vazba sloupců na /T a správné /AS a /Opt, 1 row → 1 result PDF, náhled a kontrola před exportem. Až po schválení parseru, autorizace a batch exportu.
**F3:** explicitní flatten do nové kopie; neslučovat s „bez poznámek“.

## Fáze 7 – volitelné další moduly
OCR rekonstrukce fontu, náročnější SnapEngine, Preflight, vizuální diff, měření, batch, převody/importy cizích souborů, šipkové nálepky, rozšířená měření pokrytí, externí drag/drop mezi aplikacemi. Schvalovat individuálně a neplést rezervní ikonu s dostupnou funkcí.

## Řízení kvality a forma reportu
Za každý dokončený krok Codex předá:
1. Přesný rozsah změny a důvod.
2. Seznam změněných souborů a regresních rizik.
3. Čím testoval (kompilace, jednotkové, integrační, UI Windows, screenshot, PDF round-trip), co **neotestoval** a proč.
4. Ověření funkce v přímém porovnání s v3.26 a s finálními mapami.
5. Potřebný další dílčí krok a otázky, které vyžadují rozhodnutí.
**Neoznačovat jako implementované** něco, co bylo pouze nakresleno v HTML nebo přidáno do specifikace.