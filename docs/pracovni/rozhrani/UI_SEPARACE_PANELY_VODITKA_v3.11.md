# PDF Blacksmith — UI v3.11: Separace, pomocné panely a inteligentní přichytávání

**Datum:** 24. listopadu 2026  
**Stav:** aktualizovaný pracovní HTML prototyp a implementační plán; **žádný zásah do produkční Go/Win32/PDFium aplikace**.  
**HTML:** [cely_vizual_v3.11.html](cely_vizual_v3.11.html)  
**Klávesové zkratky:** [mapa v1.0](../mapa_klavesovych_zkratek_v1.0.md)  
**Ikony:** [mapa v4.7](../mapa_ikon_v4.7.md) · [SVG atlas v3.11](ikony/pdf_blacksmith_nove_v3.11.svg)

Tento dodatek **navazuje** na v3.9 a [UI v3.10](UI_PRICHYTAVANI_v3.10.md). Neopravňuje Codex nevyžádaně měnit produktové handlery, stávající příkazy nebo formáty PDF.

## 1. Separace → Výstup a konverze

**Zachovat jednu sekci Výstup a konverze; nemažte pokrytí desek, nátisk, TAC, ICC ani separační režimy.**

- Tři převodní příkazy jsou **svisle pod sebou**. Každý řádek má vlevo **jednu ikonu-tlačítko** odpovídající příkazu a vpravo **jediný select** s rozevíracími možnostmi; stejný vizuální model jako horní řádky „Náhled tiskových desek“.
- Řádek 1: Převod do stupňů šedi. Volba způsobu/profilu \`DeviceGray\`, Gray Gamma 2.2, Dot Gain 15/20 % a aktuálního ICC záměru je **pouhý návrh**; pro realitu vyžaduje validaci profilu a konzistenci s color managementem.
- Řádek 2: Převod do K. **Nezaměnit** prostý DeviceGray, černou K 100 % a převod všech C/M/Y do K; přesná matematika převodu závisí na pravidlech a ICC profilu. Před schválením exportu zobrazit náhled a dopad na ostatní separace.
- Řádek 3: Správa ICC: **přiřadit profil** není totéž co **převést do profilu**. Nepřepisovat profil bez jednoznačné volby.
- Tlačítko provádí zvolenou operaci **až po implementaci skutečného backendu a případném potvrzení**. V HTML je volba lokální a změny PDF se nesimulují jako úspěšný export. Vybraná položka přežije přerenderování panelu.
- Popisky tlačítek, focus, tooltipy a správné \`label\` / \`aria-label\` musí být zachovány. Selektor nesmí přetékat do prostoru PDF při úzkém panelu.

## 2. Pomocné panely: dočasný Výsledky vyhledávání

Při přechodu do nástroje **Najít v PDF** (také horním vyhledávacím polem) se do jediné sdílené pomocné pozice automaticky otevře **Výsledky vyhledávání** a případně se pro tuto dobu rozbalí pravá skupina.

Při **prvním vstupu** si zapamatovat:
- předchozí pomocný panel: žádný / Záložky / Poznámky / Historie,
- původní sbalení pravé skupiny, 
- zda uživatel během vyhledávání pomocný panel nebo pravý dock ručně změnil.

Při **opuštění** nástroje Vyhledávání:
- pokud jsou stále zobrazené **automaticky otevřené Výsledky**, obnovit předchozí pomocný panel; pokud nebyl žádný, **sbalit** jej;
- pokud předtím byla pravá skupina sbalená, obnovit sbalení, **pokud ji uživatel během hledání ručně neměnil**;
- pokud uživatel v průběhu hledání vědomě aktivoval Záložky/Poznámky/Historii nebo změnil konfiguraci docku, **nepřepisovat jeho poslední rozhodnutí**;
- pokud se nástroj nemění (opětovný klik na Vyhledávání), nepřepisovat uložený původní stav;
- nesmí docházet ke vzniku dalších souběžných pomocných sloupců.

**Rozsah nové automatické obnovy: nyní jen Výsledky vyhledávání.** Pro jiné samostatné panely zachovat jejich přetrvávání. Toto upřesňuje dříve zaznamenané [pravidlo panelů](PRAVIDLO_AUTOMATICKE_POMOCNE_PANELY_2026-11-22.md), které předpokládalo automatické otevření, ale neřešilo obnovu předchozího stavu po opuštění hledání.

## 3. Vodítka, mřížka a inteligentní snap – vizuální gramatika

### Pevná vodítka

- Uživatel si může nastavit více **vodorovných a svislých** linek, tažením z pravítka, numericky i přes Zobrazení; v HTML v3.11 zatím pouze ukázková jedna horizontální a jedna vertikální linka se souřadnicemi v mm.
- **Plná** čára, jiná barva než dočasný magnetický indikátor (v návrhu mírná modrá); musí **přesahovat okraje listu až ke stupnici pravítek**, aby nešla zaměnit s obsahem PDF.
- Viditelnost lze vypnout nezávisle na přichytávání k vodítkům. Pravítka lze skrýt nezávisle na viditelnosti vodítek.
- Předmětem ukládání bude kontext dokumentu/pracovní plochy, nikoli tiskový obsah; **standardně se nevypisují do PDF a nejsou anotací**.
- V budoucí implementaci přidat zamknout/odemknout, smazat jedno/všechna, přesné souřadnice, vodítka při rotaci a transformacích, volitelný export nastavení vedle PDF.

### Dočasné inteligentní linky

- **Vínově červené**, **přerušované nebo tečkované**, automaticky se objeví pouze **při manipulaci** objektem či textem poblíž vhodného cíle snapu; při uvolnění tlačítka / Esc / vypnutí přichytávání zmizí.
- U bodu (střed, uzel, účaří) čára končí **v cílovém bodě** a lze doplnit malý koncový marker.
- U zarovnání hran objektů dočasná linie může pokračovat **až k pravítkům**; nesmí být zaměnitelná s trvalým vodítkem (proto vínová, přerušovaná).
- Při více kandidátech vybrat nejbližší validní cíl se stabilní prioritou a \`hysteresis\` (nepřeskakovat s každým pixelem). Zvolený cíl zvýraznit bez změny PDF a na vyžádání uvést typ cíle.
- Správně řešit baseline při různých fontech, středy objektů i stran, hrany, uzly křivek, vodítka, mřížku, okraje stránky a rovnoměrné rozestupy.
- **HTML v3.11:** dočasné linky jsou pouze názorná vizualizace při interakci s ukázkovým obrázkem/textem, **nejde o hotový geometrický SnapEngine** a nedochází k posunu skutečných objektů PDF.

### Mřížka a její jemnost

- V Zobrazení → Zobrazit mřížku je nastavitelná rozteč **1–50 mm** (v HTML ukázkově 5 mm) a počet dílků **1/2/4/5/10** mezi hlavními liniemi.
- Hlavní a jemná mřížka musí být vizuálně odlišeny, tenké, na plátně bez zbytečného rušení čitelnosti PDF.
- Při přiblížení/otočení dokumentu musí mřížka zůstávat **svázaná se stránkovými souřadnicemi PDF**, nikoli se screen pixely; rozlišení DPI a \`UserUnit\` mít pod kontrolou. V HTML je přepočet pro demonstraci přibližný (CSS 96 px / palec).
- **Zobrazit mřížku** a **Přichytit k mřížce** jsou **dvě oddělené volby**. Hlavní magnet je další nadřazený přepínač; vypnutí magnetu nikdy nevymaže individuální preference snap cílů.
- Cíl snapu vyhodnocovat v **PDF bodech**, toleranci pro kurzor ve **screen pixelech** (návrh 6 px u běžného kurzoru, 10 px u touch), včetně zoomu. Snap pracuje jen s editovatelnými/validními cíli a chrání uzamčené vrstvy, podpisy a formulářové widgety.

## 4. Klávesové zkratky a kolize

Související inventura: **[mapa klávesových zkratek v1.0](../mapa_klavesovych_zkratek_v1.0.md)**. Nové zkratky *nejsou v Go automaticky implementované*.

- \`T / V / M / L\` jsou v aktuálním Go již použity pro textové výběrové režimy. Nepoužívat bez migrace \`M\` pro magnet a \`V\` pro šipku.
- \`Ctrl+L / E / R / J\` již patří zarovnání textu, \`Ctrl+F\` vyhledávání, \`Ctrl+0\` fit stránce. \`Ctrl+Shift+R\` lze navrhnout pro pravítka, ale teprve po kontrole dalších stavů.
- Pro nové funkce jsou navrženy \`Ctrl+Shift+M\` (magnet), \`F9\` (mřížka), \`Ctrl+Shift+H\` (vodítka), \`Ctrl+Shift+R\` (pravítka). Dosud **pouze návrh**.
- Zkratky respektují aktivní fokus, editaci textu, nativní nabídky a české AltGr. Zobrazování přichytávání nemění dokument ani Undo/Redo; samotný potvrzený posun objektu je jeden krok historie.

## 5. Testovací scénáře a akceptace

1. **Separace:** tři svislé ikony, vedle tři přístupné selecty; při šířkách 980/1100/1250/1600 px nic nepřekrývá následující skupiny; změna selectu neprovádí PDF konverzi a uloží aktuální výběr pro HTML re-render.
2. **Pomocný panel:** vstup do hledání z žádného panelu → zobrazit Výsledky → přepnout na Text → žádný panel. Záložky → Vyhledávání → Objekty → opět Záložky. Stejně poznámky/historie; zachovat případné původní sbalení skupiny.
3. **Ruční zásah:** ve Vyhledávání klik na Poznámky → jiný nástroj → Poznámky zůstanou; ruční sbalení pravé skupiny musí zůstat zachováno.
4. **Vodítka:** plná čára přesahuje stránku ke směru pravítek; zmizí při vypnutí Zobrazit vodítka, ale předvolba \`targets.guides\` zůstane zachována.
5. **Snap:** dočasné vínové čárkované/tečkované linky se objeví při manipulaci ukázkovou oblastí a zmizí po \`pointerup\`, \`Esc\`, vypnutí magnetu; bodový a hranový cíl jsou opticky odlišeny.
6. **Mřížka:** při změně rozteče a dělení se mění jemnost zobrazení; viditelnost, snap k mřížce a globální magnet jsou nezávislé přepínače.
7. **Zkratky:** porovnat s \`documentKey\`, \`workspace\`, \`viewer.key\` a speciálními textovými dialogy; neaktivovat nové návrhy při focusu ve vstupních polích.
8. **Regrese:** režimy výběru, obě dvoustavová tlačítka, F1, psaní textu, přiblížení, export/separace a existující pomocné ouška zůstanou zachovány; nulové JS runtime chyby.

**Před implementací Go:** inventura existujících handlerů, společný \`SnapEngine\`, explicitní schválení etapy, technické i vizuální testy. Nedistribuovat pouze HTML prototyp jako hotovou PDF funkci.
