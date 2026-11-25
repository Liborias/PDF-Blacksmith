# PDF Blacksmith – UI v3.26: lokální zoom, pružné rozložení stran, tvorba formulářů

**Datum:** 25. listopadu 2026 · **Status:** interaktivní HTML demonstrace, bez schválení implementace produkčních Go/Win32/PDFium funkcí F2 nebo batch importu.
**Aktuální prototyp:** [cely_vizual_v3.26.html](cely_vizual_v3.26.html) · **Předchozí:** [cely_vizual_v3.25.html](cely_vizual_v3.25.html). Platí také dodatky ze všech předchozích verzí a [opravy z 25. 11.](UI_ZOBRAZENI_OCR_PANELY_2026-11-25.md).
**Data:** [DATA_IMPORT_FORMULARE_v1.0.md](DATA_IMPORT_FORMULARE_v1.0.md).

## 1. Dvě zóny Ctrl + kolečko
- V `#editor`: Ctrl + kolečko mění `state.zoom` dokumentu, využívá schválené limity 30–900 %, kladný/negativní krok podle směru; přibližně zachovává bod pod kurzorem. Neovlivňuje ostatní části aplikace.
- V `#thumbPane`: Ctrl + kolečko mění jen vizuální velikost miniatur (v HTML rozsah 0,65–2,25), nemění dokumentové měřítko ani výběr stránek.
- `wheel` je registrován nepasivně ve fázi capture a v příslušných zónách vykonává `preventDefault` pro Ctrl+wheel. Mimo zóny nadále funguje standardní zoom prohlížeče. Požadované chování v Chrome/Edge ověřit skutečnými testy; není univerzálně zaručeno, že každý prohlížeč nebo operační systém umožní potlačení všech zkratek.
- Standardní kolečko bez Ctrl zachovává existující listování po stránkách a souvislé rolování. Zachovat dostupnost přes klávesnici i globální zoom prohlížeče mimo pracovní zóny.

## 2. Pružný přehled stran
- Nová třetí volba `Zobrazení → Přehled stran · pružné rozložení` a tlačítko ve druhé horní liště. Zobrazení po stránkách a souvislé zůstává nezměněno.
- Při prvním vstupu se demonstrační měřítko sníží na cca 45 %, aby se vešlo více stran. Stránky se ve flex-wrap rozložení skládají po řádcích; další oddalování umožní vidět více stran najednou (až všechny).
- Klik označí stránku. Ctrl + levý klik přidává/odebírá, Shift + klik vybírá souvislý rozsah podle pořadí. Výběr je sdílen se stávajícím `selectionUX.pageSelected`, nikoliv s výběrem znaků/objektů.
- Enter otevře vybranou stránku v režimu po stránkách a nastaví její přizpůsobení výšce pracovní plochy. Výběr ve flex přehledu se zobrazuje malým označením, nikoli starým výrazným vínovým obrysem stránky.
- Přetahováním více vybraných stran lze v HTML přeuspořádat jejich **demonstrační pořadí**. Nezmění se skutečný PDF soubor ani data formulářů. Produkční operace musí používat stejný datový model a Undo/Redo jako miniatury, nikoli zvláštní duplicitní logiku.
- Ochrana: nevykonávat operace stran během editace textu, objektu či anotace; zpracovat prioritu výběru podle aktivního režimu. Klávesa Enter se nesmí přebít ve formulářovém inputu, textovém poli ani na aktivním tlačítku.

## 3. Formuláře – nová záložka Návrh polí F2 (v HTML náhledu)
Dosud nefunkční záložka `Návrh polí · F2` je v prototypu interaktivní. Původní `Vyplnit` je oddělená a zachována. Nová `Z dat` zobrazuje průvodce importem, aniž by měnila PDF.

### Návrh polí: doporučené logické skupiny
1. **Vložit pole:** Text, víceřádkový text, zaškrtávací, radio, kombinované pole, seznam, tlačítko, datum a podpis. Použít SVG sadu v dosavadní mapě ikon. Nástroj se vybírá v panelu a místo se volí na stránce; pole lze přesouvat jako běžný objekt.
2. **Seznam polí:** seskupení podle stránky a názvu, zobrazení vybrané položky. V budoucnu podpora kopírování/duplikace a vícenásobného výběru.
3. **Vlastnosti pole:** plně kvalifikovaný název `/T` (vč. hierarchie), tooltip `/TU`, typ, výchozí hodnota `/DV`, aktuální hodnota `/V`, povinné/jen ke čtení, písmo/velikost, rámeček/výplň, volby a exportní hodnota checkboxu/radia, pořadí Tab, zarovnání a formátování. U podpisu bezpečně odlišovat definici pole od samotného digitálního podepisování.
4. **Automatická detekce pole z existujícího dokumentu** pouze jako budoucí schvalovaná akce; nikdy tiše nepřevádět OCR nebo dokument na AcroForm.
5. **Náhled vyplnění:** přepínat režim Návrh/Vyplnit, zvýraznění polí, kontrola povinných hodnot a kompatibility. V HTML jsou vytvořená pole pouze vykreslená nad stránkou v paměti – nezapisují `/AcroForm`, `/Fields`, `/Widget`, `/AP` ani do PDF.

### Zdroje inspirace
Adobe Acrobat Prepare a form (identifikace/vkládání polí), Foxit Form Editor (vlastnosti, validace, import/export) a PDF-XChange Editor Plus (nástroje polí, hromadná tvorba a Populate Forms). Převzít jen funkční principy, **nikoliv** jejich proprietární grafiku, layout nebo ikony.
- https://helpx.adobe.com/acrobat/using/pdf-forms.html
- https://www.foxit.com/resource-hub/user-guide/foxit-pdf-editor-for-windows/forms/
- https://help.pdf-xchange.com/pdfxp11/creating-and-editing-fillable-forms.html

## 4. Data – třetí záložka `Z dat`
- Ukázka načítá lokální UTF-8 CSV a zobrazí tabulkový náhled řádků, přířazení jednotlivých sloupců k existujícím či nově vytvořeným formulářovým polím a automatický pokus o mapování podle názvů.
- XLSX a Google Tabulky mají v HTML viditelný vstupní návrh, **nikoli hotový parser nebo přístup k účtu**. Bez implementace OAuth 2.0, Sheets API nebo XLSX backendu nemají tvrdit úspěšný import.
- Kontrola mapování v HTML vypíše počet záznamů a duplicity, ale skutečnou tvorbu PDF, export, zápis do souborů, validaci dat podle formátu a transakční bezpečnost musí navrhnout a implementovat Codex až po schválení samostatné etapy.
- Specifikace formátů, bezpečnosti, validace a batch výstupů je v [DATA_IMPORT_FORMULARE_v1.0.md](DATA_IMPORT_FORMULARE_v1.0.md).

## 5. Kritéria a regresní ochrana
Otestovat: Ctrl+wheel jen ve dvou zónách / normální scroll mimo ně; 3×3 stránky, změny šířky panelů, přiblížení a ztráta focusu; Ctrl/Shift/Enter v přehledu; obnovu stavu v původních režimech; aktivní režimy SEL/TXT/OBJ; formuláře Vyplnit/Návrh/Z dat; přepínání nástrojů při otevřeném návrhu; žádná změna dokumentových barev nebo uloženého PDF. V HTML je demonstrace pouze v paměti.

**Produkční povolení:** F1 zůstává schváleno po jednotlivých etapách dle F1 dokumentace. F2 a datový import **nejsou tímto HTML schváleny pro implementaci do Go/Win32**. Uvedené úpravy jsou nové návrhy do pracovního plánu Codexu. Před zahájením se vyžaduje samostatné schválení jednotlivé etapy.
