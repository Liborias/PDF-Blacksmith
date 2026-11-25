# PDF Blacksmith — mapa klávesových zkratek v1.0

**Datum:** 24. listopadu 2026 · **Status:** inventura skutečných zkratek + návrh pro nové funkce, **není to automatický příkaz měnit produktový kód**.

**Pro Codex:** Oddělovat (A) zkratku **skutečně obsluhovanou současným Go handlerem**, (B) **textový popisek** v UI, jehož funkčnost se ověřuje zvlášť, a (C) **nový návrh** bez implementace. Nepovažovat vše v tabulkách za již fungující. Uplatnit pouze tehdy, má-li fokus relevantní plátno nebo textový editor; při fokusu v políčku zadávání zachovat běžné vstupy.

## 1. Ověřené v repozitáři (A)

Provedena statická kontrola souborů \`document_commands_windows.go\` (\`inlineShortcuts\`, \`viewShortcuts\`, \`documentKey\`), \`workspace_windows.go\` (centrální směrování), \`viewer_windows.go\` (\`documentView.key\`), \`file_menu_windows.go\` a \`canvas_selection_windows.go\` na větvi \`main\` k 24. 11. 2026. Nejde o runtime test ve Windows.

| Zkratka | Existující význam | Rozsah/handler | Status |
|---|---|---|---|
| Ctrl+O | Otevřít soubor | workspace | A |
| Ctrl+F | Najít v PDF | workspace | A |
| Ctrl+W | Zavřít dokument | workspace | A |
| Ctrl+Tab / Ctrl+Shift+Tab | Další / předchozí karta dokumentu | workspace | A |
| Ctrl+S / Ctrl+Shift+S | Uložit / Uložit jako | workspace | A |
| Ctrl+P | Tisk | workspace | A |
| F3 / Shift+F3 | Další / předchozí výsledek (jen v režimu Vyhledávání) | workspace | A |
| F11, Esc | Celá obrazovka, Esc návrat z fullscreen | workspace | A |
| Ctrl+Z, Ctrl+Y, Ctrl+Shift+Z | Zpět, znovu | dokumentové režimy | A |
| Ctrl+X, Ctrl+C, Ctrl+V | Vyjmout, kopírovat, vložit | editor textu; Ctrl+C též plátno | A |
| Ctrl+A | Vybrat vše | text a plátno (dle fokusu) | A |
| Delete | Odstranit | kontextové příkazy; potvrdit fokus | A/B |
| Ctrl+Enter | Použít úpravu textu | inline editor / dokument dle handleru | A |
| Ctrl+B / Ctrl+I / Ctrl+U | Tučné / kurzíva / podtržené | inline text | A |
| Ctrl+L / Ctrl+E / Ctrl+R / Ctrl+J | Zarovnat vlevo / střed / vpravo / do bloku | inline text | A |
| Ctrl++ / Ctrl+− / Ctrl+0 | Přiblížit / oddálit / přizpůsobit stránce | plátno | A |
| T / V / M / L | Stávající textové výběrové režimy (objekt, odstavec, obdélník, laso) | plátno bez Ctrl/Shift, mimo inline editor | A; **POZOR NA KOLIZI** |
| Šipka vlevo / vpravo | Předchozí / následující stránka | plátno \`documentView.key\` | A |
| Šipka nahoru / dolů | Posun o krok | plátno | A |
| Page Up / Page Down, mezerník | Posun o téměř stránku / dolů | plátno | A |
| Home / End | První / poslední stránka | plátno | A |
| Ctrl+kolečko | Zoom pod kurzorem | plátno (\`wheel\`) | A |
| Prostřední tlačítko + tažení | Schválený obdélníkový výběr podle Q1 | plátno | A/specifikace — regresně ověřit |

**Zkrácený popisek neznamená existenci stejného handleru v každém panelu.** Např. \`Ctrl+1\` (100 %) je v dokumentaci jako návrh, nikoli v mapě \`viewShortcuts\` výše. \`Ctrl+D\` (duplikovat) je uvedeno v \`documentActions\`, ale není automaticky přiřazeno ve stejné \`viewShortcuts\` tabulce — ověřit trasu příkazu.

## 2. Návrhy zkratek (C) – rozvržení, vodítka a přichytávání

| Zkratka | Navržený příkaz | Kdy platí | Stav |
|---|---|---|---|
| **Ctrl+Shift+M** | Přichytávání zapnuto / vypnuto (magnet) | fokus plátno, nikoli textový input | C; bez kolize v prověřených tabulkách |
| **F9** | Zobrazit / skrýt mřížku | fokus plátno | C; výslovně ne Ctrl+Shift+G |
| **Ctrl+Shift+H** | Zobrazit / skrýt vlastní vodítka | fokus plátno | C; ověřit širší menu klávesy |
| **Ctrl+Shift+R** | Zobrazit / skrýt pravítka | fokus plátno | C; odlišit od Ctrl+R zarovnání vpravo |
| **Shift+F9** | Volitelně zapnout / vypnout přichytávání **k mřížce** | fokus plátno | C; až po schválení, není totéž co její zobrazení |
| **Ctrl+1** | Zobrazení na 100 % | fokus plátno | C; starší návrh, nepovažovat za hotové |
| **Bez výchozí zkratky** | Po stránkách ↔ souvislé zobrazení | dvoustavové tlačítko lišty | C; zkratku zvolit po kontrole dopadu na taby |
| **Bez výchozí zkratky** | Vložit vodorovné / svislé pevné vodítko | pravítka a Zobrazení | C; volit přesně nebo tažením; nemíchat se Shift+F10 |
| **Bez výchozí zkratky** | Upravit rozteč a jemnost mřížky | Zobrazení | C; nastavení číselným polem, bez potřeby zkratky |
| **Bez výchozí zkratky** | Převést vybraný text na křivky / rekonstruovat | Text → Výběr | C; nespouštět destruktivní operaci jedinou neověřenou zkratkou |
| **Bez výchozí zkratky** | OCR region / ořez obrázku / barevná konverze | podle vybraného obsahu | C; vyhnout se kolizím s již používanými zkratkami |

**Proč F9 a ne Ctrl+Shift+G:** ve vektorových editorech je \`Ctrl+Shift+G\` běžně oddělení skupiny; Blacksmith má seskupování objektů v plánu. Nezabírat zkratku snadno zaměnitelnou za budoucí Ungroup. **Proč nepoužívat Ctrl+Alt:** česká klávesnice používá AltGr, který může být hlášen jako Ctrl+Alt. Nevolit jako výchozí mezinárodně problematické kombinace s \`+\`, \`[\`, \`]\`, \`;\` pro nové příkazy.

## 3. Výběr: aktivní kolize a pravidlo migrace

- Současné **V** znamená odstavec, ne univerzální šipku. **M** znamená textový obdélník, nikoli magnet. **L** znamená textové laso. **T** znamená textový výběrový režim, ale nový pracovní UI může interpretovat textový caret jinak.
- V novém HTML v3.9+ jsou horní šipka/I-kurzor **dvoustavové** a obdélník/laso **univerzální**. **Nesmí vzniknout paralelní handlery**, které by převzaly \`V/M/L/T\` bez výslovného přemapování.
- **Navržená budoucí migrace (jen po analýze):** \`A\` na plátně pro univerzální šipku; \`T\` pro textový kurzor; \`M\` pro univerzální obdélník; \`L\` pro univerzální laso. \`V\` by se uvolnilo pro jiný příkaz nebo zůstalo dočasně jako starší volba odstavce. Vyřešit existující gesta výběru a místní fokus, nesmí se ztratit vytváření bodového a odstavcového textu.
- Tuto migraci **neimplementovat** jen kvůli novému HTML nebo ikonám. Nejprve testy \`SEL-01\` až \`SEL-03\`, \`TXT-01\` až \`TXT-07\`, \`OBJ-01\` až \`OBJ-03\`, Q1/Q3 a stávajícího \`documentKey\`.

## 4. Priorita při zpracování kláves

1. Modal/dialog, otevřená nabídka (Esc, Enter, šipky) a systémové klávesy mají přednost.
2. Editační pole, IME a klávesové kombinace při zadávání textu mají přednost před zkratkami plátna; nepohlcovat textová \`A/T/V/M/L\` při psaní.
3. Inline text editor používá \`inlineShortcuts\`, plátno \`viewShortcuts\`; mimo ně má Win32 běžné nativní chování.
4. Zkratka označená jako „globální“ musí respektovat aktivní dokument a stav nástroje; nesmí zasáhnout dialog, pokud by se měla provést na plátně.
5. Změna zapnuto/vypnuto vodítek/mřížky/magnetu je **zobrazovací preference**, ne obsahová PDF operace: nenastavit dirty, nepřidat krok do Undo.
6. Jediný centrální registr akcí a zkratek; UI popisky, tooltipy, nabídky a klávesové handlery musí čerpat ze stejné mapy. Uživatel by měl mít možnost zkratky přenastavit a systém musí odmítnout duplicity v rámci stejného fokusu.

## 5. Testy před jakoukoli implementací

- Ověřit na Windows při českém i anglickém rozložení, s AltGr, v inline editoru a v poli vyhledávání.
- \`Ctrl+Shift+R\` nesmí měnit zarovnání textu; \`Ctrl+Shift+H\` nesmí skrýt dokument; \`F9\` nesmí být zachyceno jinou komponentou.
- \`Ctrl+B/I/U\`, \`Ctrl+L/E/R/J\`, \`Ctrl+F\`, \`Ctrl+S\`, \`Ctrl+Tab\`, \`Esc\`, \`F3\` a režimy \`T/V/M/L\` musí nadále fungovat ve svém starém rozsahu, dokud nebude schválena migrace.
- Při přepnutí z Vyhledávání obnovit předchozí pomocný panel, pokud si jej uživatel během vyhledávání ručně nepřepnul.
- Automatizovaný test střetů: stejná sekvence nesmí mít dvě akce v překrývajícím se kontextu. Pro každou zkratku testovat stav „bez PDF“, „PDF otevřeno“, „vlastnosti panelu“, „inline text“, „modální okno“.

**Navazuje na:** [aktuální UI v3.11](rozhrani/UI_SEPARACE_PANELY_VODITKA_v3.11.md), [HTML v3.11](rozhrani/cely_vizual_v3.11.html), [hlavní specifikaci](specifikace_Codex_v5.0.md). Historické mapy ikon ani pravidla SEL/TXT/OBJ se nemažou.


---

## Dodatek v1.1 – uživatelské zkratky v Nastavení (24. 11. 2026)

**Referenční HTML:** [rozhrani/cely_vizual_v3.13.html](rozhrani/cely_vizual_v3.13.html).  
**UI specifikace:** [rozhrani/UI_POMUCKY_ZKRATKY_v3.13.md](rozhrani/UI_POMUCKY_ZKRATKY_v3.13.md).

Pro každý příkaz v Nastavení → Klávesové zkratky má být dostupné \`Vlastní zkratka\` + obnova výchozí hodnoty. Při volbě přejít do režimu zachycení stisku; \`Esc\` zruší, \`Backspace\` vrátí výchozí, po validaci uživatel potvrdí změnu přes OK/Použít. Pokud se stejná kombinace používá jiným příkazem v **překrývajícím se kontextu**, nebo je vyhrazena systémem, zobrazit konkrétní varování a změnu **neuložit**. Žádné tiché přepsání stávajícího příkazu.

Požadavky pro produkční Codex: jeden centrální registr přiřazení, původní zkratka a uživatelská výjimka, kontext fokusu (\`global/canvas/inline/dialog\`), lokalizované popisky v menu a tooltipu, serializace preferencí, možnost resetu všech změn, kontrola AltGr/IME pro CZ/EN rozložení a regresní testy \`T/V/M/L\`, \`Ctrl+L/E/R/J\`, \`Ctrl+F\`, \`Ctrl+Tab\` a běžných editačních kombinací.

**HTML v3.13** je pouze demonstrace tabulky, záznamu klávesy a varování. **Nepřepisuje** zkratky Go/Win32 aplikace a není konečným důkazem bezkoliznosti ve všech oknech. Před reálnou implementací ověřit kompletní Win32 dispatch a případné systémové rezervace.

**Testy prototypu:** pokus nastavit mřížce \`Ctrl+F\` musí zobrazit konflikt s Najít v PDF a ponechat původní přiřazení; nová nekolidující kombinace se lokálně zobrazí a po OK je při opětovném otevření dialogu viditelná; Zrušit ji nepoužije.


---

## Dodatek v1.2 – ovládání vodítek v UI v3.14 (24. 11. 2026)

**HTML návrh:** [rozhrani/cely_vizual_v3.14.html](rozhrani/cely_vizual_v3.14.html). **Implementační zadání:** [rozhrani/UI_VODITKA_INTERAKCE_v3.14.md](rozhrani/UI_VODITKA_INTERAKCE_v3.14.md).

| Klávesa | Akce | Povolený kontext | Status |
|---|---|---|---|
| \`Esc\` | Zrušit rozpracované vytahování/přesun vodítka bez zápisu | Pouze pokud právě běží drag vodítka; jinak předat dosavadnímu handleru | Návrhové chování HTML |
| \`Delete\` | Odstranit právě vybrané vodítko | Pouze \`Zobrazení → Pomůcky rozvržení\`, aktivní vybrané a neuzamčené vodítko, fokus mimo INPUT/TEXTAREA/SELECT | Návrhové chování HTML |
| \`Enter\` / \`mezerník\` | Vytvořit vodítko přibližně v polovině dané osy | Fokus přímo na konkrétním pravítku s \`role=button\` | Návrhové chování HTML |

Výše uvedené chování se **nesmí slepě přidat jako další globální zkratka**; \`Delete\` už patří mazání obsahu, \`Esc\` zavírání nabídek/dialogů a \`mezerník\` posuvu stránky. Rozhoduje kontext a fokus. Globální Lock vodítek blokuje jejich přesun, založení i odstranění. Vedle zkratek zůstává myš a číselné ovládání v pravém panelu. Codex musí před produkční implementací prověřit existující dispatch (\`document_commands_windows.go\`, \`workspace_windows.go\`) a nepoškodit původní SELECT/TXT/OBJ pravidla.


---

## V1.3 – sjednocení gest výběru vodítek a duplikování (24. 11. 2026)

**Referenční HTML:** [rozhrani/cely_vizual_v3.18.html](rozhrani/cely_vizual_v3.18.html).  
**Pravidla Codex:** [rozhrani/UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md](rozhrani/UI_TEMATA_PREKLADY_VODITKA_OREZ_v3.18.md).

| Gesto | Výsledek | Konfliktní stav / precedence | Status |
|---|---|---|---|
| Levý klik na vodítku | Jediné označené vodítko; odznačit ostatní | Hit-test vodítka má přednost před výběrem PDF pouze na samotné čáře | HTML v3.18; Go plán |
| **Ctrl + levý klik na vodítku** | Přidat/odebrat vodítko ze stávajícího výběru | **Pouze klik bez pohybu**. Změnu výběru provést až po \`pointerup\`, ne při \`pointerdown\` | HTML v3.18; Go plán |
| **Ctrl + tažení levým tlačítkem na vodítku** | Duplikovat originál na nové pozici, nové ID | Jen po překročení prahu 7 CSS px; nevyvolá Ctrl+klik toggle. Originál se nepohne. Nové vodítko není implicitně členem původní skupiny | HTML v3.18; Go plán |
| Levé tažení bez Ctrl | Přesun neuzamčeného vodítka | Bez duplikace, s magnetickým zarovnáním podle nastavení | HTML v3.18; Go plán |
| **Esc** při tažení | Zrušit operaci bez změny | Nesmí zavřít celý dialog ani mazat PDF | HTML v3.18; Go plán |
| Ctrl + kolečko | Přiblížení PDF | Stávající význam beze změny | Existující Go |

**Přístupnost a priority:** při zasažení skutečné linky s aktivním režimem vodítek mají gesta linek přednost; jinak \`Ctrl+klik\` nad textem a objektem používá dosavadní jejich rozšířený výběr. V textovém inputu se Ctrl kombinace nepřesměrovávají na nástroje plátna. \`Delete\` v seznamu vodítek smí mazat vybraná vodítka pouze při odpovídajícím kontextu, ne text PDF. **Žádný nový globální keybinding nevzniká.**


---

## Doplnění mapy v1.4 – 25. listopadu 2026 (HTML v3.26)

| Kontext | Gesto | Výsledek v HTML v3.26 | Produkční Go/Win32 |
|---|---|---|---|
| Hlavní pracovní plocha PDF | Ctrl + kolečko nahoru/dolů | Lokální přiblížení/oddálení dokumentu 30–900 %, přibližné zachování bodu pod kurzorem; prohlížeči se v tomto kontejneru předává `preventDefault` | Navrhnout obdobu, otestovat s Windows a přístupností |
| Levý panel miniatur | Ctrl + kolečko | Mění pouze velikost miniatur 65–225 %, ne měřítko PDF | K posouzení pro budoucí implementaci |
| Mimo PDF a miniatury | Ctrl + kolečko | Není aplikací blokováno, běžná funkce prohlížeče | Bez změny |
| Pružný přehled stran | Klik do plochy stránky | Vybere stránku podle `selectionUX.pageSelected` | Zatím HTML náhled |
| Pružný přehled stran | Ctrl + levý klik | Přidá/odebere stránku z hromadného výběru | Pozor na kolize Ctrl+klik u objektů a vodítek; musí rozhodovat aktivní kontext |
| Pružný přehled stran | Shift + klik | Vybere souvislý rozsah dle pořadí | Zatím HTML náhled |
| Pružný přehled stran s fokusem stránky | Enter | Otevře vybranou stránku na výšku pracovní plochy | Neměnit Enter v editaci textu ani ve formulářových polích |
| Pružný přehled stran | Přetáhnutí vybraných stran | Přeskupí jen demonstrační pořadí náhledu; PDF není změněno | K realizaci pouze s datovým modelem miniatur a Undo/Redo |
| Formuláře → Návrh polí | Klik na stránku po zvolení typu | Vloží ukázkový widget pouze do HTML paměti | F2 neschválené produkční implementace |
| Formuláře → Návrh polí | Tažení vytvořeného pole | Posune demonstrační pole | Sjednotit s objekty, vodítky a přichytáváním, až bude schváleno |

Při kombinaci s dalšími gesty se vždy nejprve vyhodnocuje aktivní kontext a fokus: textový editor, formulářové pole, přesouvání objektu, stránkový přehled, miniatury. Nesoučasně zobrazovat dva vlastní výběrové kurzory. Browser zoom je v některých prostředích neodchytitelný, proto potřebuje integrační test prohlížečů a případnou alternativu.

**Aktuální implementace HTML:** [v3.26](rozhrani/cely_vizual_v3.26.html). **Podrobná specifikace:** [UI_ZOOM_FORM_DESIGN_v3.26.md](rozhrani/UI_ZOOM_FORM_DESIGN_v3.26.md). Související F2 a import zdrojových dat jsou pouze návrhové – [DATA_IMPORT_FORMULARE_v1.0.md](rozhrani/DATA_IMPORT_FORMULARE_v1.0.md).
