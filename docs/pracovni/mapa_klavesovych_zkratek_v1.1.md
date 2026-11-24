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
