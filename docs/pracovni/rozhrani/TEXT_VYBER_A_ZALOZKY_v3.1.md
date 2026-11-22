# PDF Blacksmith — Text: bezkolizní výběr, editace a záložky panelu Vlastnosti (v3.1)

**Datum:** 22. 11. 2026  
**Stav:** pracovní technické zadání pro Codex a vizuální HTML prototyp; **nejde o povolení k přepsání existující výběrové logiky ani o hotovou implementaci**.  
**Výchozí podklad:** `docs/pracovni/specifikace_Codex_v5.0.md`, zejména Q1–Q3, Q15–Q16, Q31; SEL-01 až SEL-03; TXT-01 až TXT-07; OBJ-01 až OBJ-03; UI-07. Vizuální podklad `rozhrani/cely_vizual_text_zalozky_v3.1.html`.

## 1. Oprava možného konfliktu ze staršího návrhu

V debatě zazněl jako **pracovní, neschválený** návrh „první levé kliknutí označí textový blok, dvojklik zahájí editaci“. **Nepřebírat jej do kódu.** Je v přímém rozporu s již zaznamenanými pravidly `SEL-01` a `TXT-02/04`:

- **Jedno levé kliknutí v nástroji Text** do bezpečně editovatelného textu **aktivuje editor a umístí textový kurzor (caret)**.
- **Levé dvojkliknutí** označí **slovo**, nikoli „zahájí editaci“. Další dvojkliky v téže výběrové posloupnosti: **slovo → segment → věta → odstavec**. Výběr celého řádku a sloupce má své samostatné kombinace, viz níže.
- **Levé tažení uvnitř aktivního textového rámce** označuje **znakový rozsah**. Nesmí se automaticky reinterpretovat jako plošný výběr obdélníkem.
- **Stisk kolečka + tažení** je vyhrazen pro **obdélníkový výběr oblasti**, a to i při aktivním nástroji Laso. Samotný prostřední klik nic nevybere, otáčení kolečka scrolluje, `Ctrl+kolečko` přibližuje.
- **Levé tažení nad objektem** v režimu Objekty může přesouvat objekt; v režimu kreslení tvoří tvar. **Aktivní režim má přednost**, nevytvářet globální handler levého tažení, který přebije text nebo kreslení.

**Výběr celého textového bloku** lze nabídnout jako explicitní příkaz v záložce Výběr nebo v kontextové nabídce **až po samostatném rozhodnutí o jeho gestu**. Nespojovat ho s prvním kliknutím a nepřepisovat schválené zkratky.

## 2. Konkrétní mapa gest s jednoznačným kontextem

| Gesto | Text / aktivní textový rámec | Objekty / kreslení | Náhledy stran |
|---|---|---|---|
| Levé kliknutí | Umístí caret do editovatelného textu | Výběr viditelného objektu podle skutečné geometrie `OBJ-02`; ve zvláštním režimu kreslení platí jeho vlastní pravidla | Aktivuje miniaturu/stránku |
| Levé tažení | Znakový rozsah | Přesun vybraného objektu / kreslení aktivního tvaru | Přesun miniatury, případně výběr podle stávajícího ovládání |
| Levý dvojklik | Slovo; opakováním segment → věta → odstavec | V režimu Laso **ukončuje polygon**; nesmí se zachytit textovým handlery | Zachovat stávající chování miniatur |
| `Ctrl + levý klik` | Přímo označit **větu** | Procházet geometricky zasažené překrývající se objekty podle `OBJ-03` | Přidat/odebrat stránku z výběru |
| `Ctrl + levý dvojklik` | Přímo označit celý **vizuální řádek** | Nesmí nechtěně spustit výběr věty ani cyklus hlubokých objektů | Dle existující logiky |
| `Shift + levý klik` | Rozšíří znakový rozsah od kotvy | Podle již ověřené existující logiky objektů | Souvislý rozsah stránek |
| `Shift + levý dvojklik` | Odstavec, opakováním celý příslušný sloupec | Nezasahovat do objektového lasa | Bez změn |
| `Alt + levý dvojklik` | Stejně formátované úseky **v segmentu** | Bez nově přiděleného globálního významu | Bez změn |
| `Alt + Shift + levý dvojklik` | Stejné formátování v **aktuálním sloupci** | Bez nově přiděleného globálního významu | Bez změn |
| Prostřední tlačítko + tažení | Obdélníkový výběr oblasti textu | Obdélníkový výběr objektů (také když je aktivní Laso) | Nezavádět konflikt s posouváním miniatur |
| `Ctrl + Z / Ctrl + Y` | Společná chronologická historie **výběru i úprav** podle `SEL-03` | Zachovat funkční kroky úprav | Zachovat vratné změny pořadí |

**Priorita zachytávání událostí:** nejprve oblast UI (textové vstupní pole v panelu vs. samotné PDF plátno), pak aktivní dokument a nástroj, následně jeho podrežim, až potom gesto a modifikátory. Neprovádět dva příkazy jedním klikem. Rozlišovat Windows `WM_LBUTTONDBLCLK` od obyčejného kliknutí tak, aby nevznikaly vedlejší kroky, které vzápětí dvojklik ruší; konkrétní implementaci naplánovat po auditu stávajících handlerů.

**Laso:** stávající objektové Laso (`OBJ-01`, režimy Protnutí / Celý objekt) se **nesmí zaměnit** za samostatnou volbu textového výběru. Ikonky obdélníku/volného textového výběru v HTML v3.0 představovaly možné režimy; v3.1 je má pouze v sekci **Připravované režimy**, **neaktivní**. Rozhodnutí o případném samostatném textovém lasu je odložené. Univerzální obdélníkový výběr prostředním tažením musí fungovat dál.

## 3. Výběr, formátování a historie bez ztráty dat

1. Typografický panel se váže na **aktuální textový výběr v aktivním PDF**, případně na aktivní textový rámec nebo na budoucí výchozí styl nového textu. **Pouhé přepnutí záložek Výběr/Typografie nesmí měnit výběr, caret, aktivní režim plátna ani samotný PDF dokument.**
2. Pokud je vybrána jen část textu, formátování se uplatní **jen na tento rozsah**, nikoli implicitně na celý PDF textový objekt. Při více nesouvislých úsecích lze změnit jejich společné vlastnosti, ale **nelze do všech umístit jeden falešný caret** ani začít nejednoznačně psát.
3. Výběry a obsahové úpravy mají podle schváleného Q3 **jednu chronologickou historii na dokument**. Výběr sám nepřepne stav „neuložené změny“. Opakované náhledy, hover, fokus, přepnutí záložky či rozbalení sekce nevytváří historické ani obsahové kroky.
4. Při smíšené typografii (více písem, velikostí, barev) zobrazit **neurčitý/smíšený stav**; vyhodnotit konkrétní vlastnost až po výslovném zásahu. Tento bod je návrhem UX k ověření, nikoliv již implementovanou logikou.
5. Respektovat omezení PDF: neúplné vložené fonty, podmnožiny, transformace, rotace, text převedený do křivek, OCR. Bezpečné varování a případná nabídka OCR podle `TXT-04`; **nezplošťovat stránku, nezměnit nesouvisející formátovací běhy a neprovádět tiché nahrazení písem**.
6. Dvě varianty tvorby nového textu z `TXT-06`: **bodový text kliknutím** a **odstavcový rámec tažením**. Nástroj Nový text nemá přebírat gesta režimu Úprava existujícího textu; při implementaci nabídnout jasný výběr varianty.
7. Žádný nový textový ovladač nesmí přepsat stávající výběr či editaci formulářových widgetů AcroForm ani funkce nástroje Objekty. Režim a typ zasaženého obsahu rozhoduje o cíli akce.

## 4. Nové rozvržení pravého panelu **Vlastnosti → Text**

**Zásada:** jde o **vnořené záložky nástroje Text**, nikoli o nové globální okno Nastavení, ani o samostatné pomocné panely. Pravidlo jediného pravého pomocného panelu zůstává nedotčené.

### Záložka Výběr

- Rychlá volba I-kurzoru pro existující text a ikona **Nový text** pro vytvoření bodového/odstavcového textu; obě používají konkrétní tooltip a detailní nápovědu pod ikonami podle `UI-07`.
- Přehled základních gest bez zavádění konfliktních tlačítek. **Jedno levé kliknutí = caret, první dvojklik = slovo.**
- Rozbalitelná sekce **Rozšířený výběr a zkratky** obsahuje pravidla z `SEL-02/TXT-02` a případně později schválené explicitní příkazy pro slovo/segment/větu/řádek/odstavec/sloupec.
- Rozbalitelná sekce **Výběr oblastí – omezení a další režimy** obsahuje neaktivní návrhy textového obdélníku/lasa. Dokud nebude rozhodnuto jinak, **nemají přiřazena nová gesta**.

### Záložka Typografie

- **Základní část bez dlouhého rolování:** Rodina písma, velikost v bodech, A+/A−, běžný/tučný/kurzíva/podtržení/přeškrtnutí, barva písma; zarovnání a řádkování.
- **Rozbalitelné Pokročilá typografie:** kapitálky, horní/dolní index, šířka znaků, proklad, mezery mezi slovy, posun účaří.
- **Rozbalitelné Pokročilé nastavení odstavce:** mezery před/za odstavcem, odsazení zleva/zprava/prvního řádku, zvýšit/snížit odsazení.
- **Rozbalitelné Zvýraznění a barevné efekty:** barva zvýraznění za znaky, přepínač zvýraznění; kapátko a obrys pouze jako zřetelně ilustrační či neaktivní návrhy do doby potvrzení jejich funkce.
- Náhled sazby lze ponechat viditelný pod složenými sekcemi; **v HTML jde pouze o demonstrační náhled, nikoli úpravu PDF**.
- Při přepnutí obou záložek **zachovat hodnoty, aktivní podrežim i stavy rozbalených pokročilých sekcí**. Rolování má být omezeno na skutečně otevřené pokročilé sekce, bez dalšího trvale otevřeného panelu.
- Ikonová tlačítka a přístupné názvy/tooltipy zachovat; krátký tooltip ukazuje název a zkratku, trvalá kontextová nápověda pod skupinou se při najetí/fokusu dočasně změní a po opuštění vrátí na právě aktivní funkci (`UI-07`).

**Pozor:** toto rozdělení se týká **výhradně Textu**. Neprolínat jej do Separací: tam uživatel výslovně odmítl rozdělení panelu na záložky (`Q32`, `RND-03`). Globální **Nastavení aplikace** zůstává samostatným modálním oknem s osmi kategoriemi (`Q10/Q12–Q14`).

## 5. Audit kódu – povinné před implementací

Začít inventurou skutečného chování a obsluhy událostí v tomto pořadí (ověřit existenci souborů a funkcí v aktuálním commitu):

- `canvas_selection_windows.go`: `middleSelection`, `canvasSelectionChange()`, rozlišení klik/tažení kolečka.
- `text_windows.go`, `inline_text_windows.go`, `typography_windows.go`: klikací výběr, textový caret, segmenty a formátování.
- `object_windows.go`, `object_selection_windows.go`: geometrický výběr, Ctrl+klik pro cyklus, Laso.
- `history_windows.go`, `document_commands_windows.go`: jednotný chronologický Undo/Redo.
- `pages_windows.go`, `page_transfer_windows.go`: výběry a tažení miniatur.
- `workspace_windows.go`, příslušné rozvržení panelů: přepínání nástrojů a pravého panelu Vlastnosti.

**Pořadí prací:** (1) inventura skutečných handlerů a konfliktů; (2) tabulka rozhodovacího směrování gest dle aktivního nástroje; (3) návrh změny pro review; (4) úpravy UI s vazbou na stávající stav; (5) regresní testy v níže uvedené matici. **Neměnit produkční soubory jen proto, že existuje HTML prototyp.**

## 6. Regresní testy (minimální sada)

- **Text:** single-click caret; dblclick slovo; druhý/třetí/čtvrtý dblclick segment/věta/odstavec; `Ctrl+klik` věta, `Ctrl+dblclick` řádek; `Shift+dblclick` odstavec/sloupec; `Alt+dblclick` stejný styl; levé tažení znakový výběr.
- **Objekty:** obyčejný levý klik přesná geometrie a překryvy; `Ctrl+klik` cyklus; žádný textový výběr nevstoupí do objektového handleru; levé tažení přesun.
- **Laso:** levý klik vrchol, dvojklik/Enter uzavření; výběr Protnutí/Celý objekt; `Esc`; stisk kolečka+tažení stále obdélník.
- **Kolečko:** samotný prostřední klik nic; tažení označí obdélník podle nástroje; otočení scroll; `Ctrl+otočení` zoom. Kontrola zoomu/rotace/HiDPI.
- **Miniatury:** Ctrl/Shift multivýběr, přesun a povinné potvrzení při přesunu; textové zkratky nepřepisují stránkový výběr.
- **Historie:** sekvence výběr slovo/segment/věta → formátování → nový výběr → Ctrl+Z/Y přesně chronologicky; samotný výběr nedělá dirty flag ani AutoRecover snapshot.
- **Textové panelové záložky:** Výběr ↔ Typografie, zachování fontu 12→14 přes A+, řádkování 1,5, odsazení; rozbalení/sbalení pokročilých sekcí, přepnutí na Objekty a zpět, zachování fokusu a žádný zásah do PDF; testy při 100/150/200 % DPI a úzkém okně.
- **Formuláře/OCR:** kliknutí do vyplňovacího pole není převzato textovým editorem PDF; OCR pouze explicitně, bez změny originálu bez souhlasu.

## 7. Stav HTML v3.1

Interaktivní HTML je **výhradně simulace vzhledu**: záložky, rozbalování, typografické vstupy, živý náhled a tooltipy. Vkládání ani editace skutečného PDF zde není implementována. Při testu Chromium ověřit funkčnost záložek, zachování hodnot i po přepnutí nástroje, vypnuté návrhy samostatného textového lasa/obdélníku a nulové JavaScript chyby. Produkční změny až po samostatném schválení implementace.