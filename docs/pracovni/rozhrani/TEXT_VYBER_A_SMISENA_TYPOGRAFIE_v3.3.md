# PDF Blacksmith — Text v3.3: čtyři režimy výběru a smíšené formátování

**Datum:** 23. listopadu 2026  
**Stav:** schválené UX rozhodnutí a demonstrace v HTML. **Nejde o pokyn okamžitě přepsat produkční Go/Win32/PDFium kód.** Implementovat až po inventuře stávajících handlerů a zvláštním schválení práce.  
**Podklady:** `specifikace_Codex_v5.0.md` — Q1–Q3, Q15–Q16, Q31; SEL-01 až SEL-03, TXT-01 až TXT-07, OBJ-01 až OBJ-03, UI-07; `rozhrani/TEXT_VYBER_A_ZALOZKY_v3.1.md`. Prototyp `rozhrani/cely_vizual_text_vyber_typografie_v3.3.html`.

## A. Nové rozvržení záložky Text → Výběr

**Jedna vodorovná řada čtyř skutečných režimových ikon** v následujícím pořadí, vlevo:

1. **Textový kurzor** — původní schválená ikona `text_caret`. Klik v existujícím editovatelném textu umístí caret. Klik na prázdné místo stránky má zahájit nový **bodový/volný řádkový text** (bez předem vymezeného textového rámce). Výchozí aktivní režim.
2. **Obdélníkový výběr textu** — existující `rect_select`. V tomto explicitně vybraném podrežimu levé tažení vyznačí obdélníkovou oblast pro výběr **textových znaků/úseků**. Neslouží pro rozšířený výběr pomocí klávesových zkratek — to je informační část nápovědy.
3. **Laso pro výběr textu** — existující `lasso`. V tomto explicitně vybraném podrežimu levé kliky zadávají vrcholy výběrového polygonu, levý dvojklik nebo Enter jej dokončí, Esc zruší; jde o **textový** plošný výběr, nikoliv o objektové Laso. Jde o návrh konvence podle stávajícího objektového lasa; před implementací prověřit kompatibilitu s běžícím výběrem.
4. **Odstavcový textový blok** — existující `text_box`. Levé tažení vymezí hranice textového rámce, který používá **odstavcový text se zalamováním**. Jednoduché kliknutí bez smysluplného tažení žádný odstavcový rámec nevytvoří. Tento nástroj **neslouží** k novému bodovému/řádkovému textu.

Na **témže řádku zcela vpravo** je odsazená ikonka **otazníku** `nav_help`. Pouze po kliknutí otevře/uzavře místní nápovědu **Základní výběr** + **Rozšířený výběr**. Další kliknutí a `Esc` ji zavírají. Nápověda nikdy nemění aktivní režim, caret ani vybraný obsah. Základní a rozšířený výběr **nejsou další samostatné ikony či výběrové režimy**.

Dočasný vysvětlující text pod řadou může popisovat **aktuální režim** a na hover/fokus ostatních ikon dočasně přepnout podle `UI-07`. Podrobné tabulky gest však nechat jen pod `?`, aby panel nevyžadoval zbytečné rolování. Každá ikona má `title`, přístupný název, aktivní stav/fokus; zachovat původní tmavé šedé provedení a vínový akcent.

## B. Přednost gest a prevence kolizí — zvlášť důležité

**Neměnit už schválené globální chování kolečka!** Prostřední **stisk + tažení** označuje obdélníkovou oblast v každém relevantním nástroji a nemusí se kvůli tomu volit ikona obdélníku. Samotný prostřední klik nic nevybírá, otočení kolečka roluje a `Ctrl + otáčení` přibližuje. Tato gesta se neúčastní vytváření textu, objektů ani uzavírání lasa.

| Aktivní kontext / podrežim | Levé kliknutí | Levé tažení | Levý dvojklik | Stisk kolečka + tažení |
|---|---|---|---|---|
| Text → **Textový kurzor** | V existujícím textu caret; na volném místě nový bodový text | V aktivním textu označení znaků | Slovo; další dvojkliky segment → věta → odstavec | Obdélníkový výběr oblasti |
| Text → **Obdélníkový výběr** | Samotný klik nevytváří textový objekt ani rámeček | **Obdélníkový výběr textových úseků**, pouze v tomto explicitním režimu | Nesmí zároveň spouštět stupně TXT-02 | Obdélníkový výběr dle SEL-01 |
| Text → **Laso** | Vkládá vrcholy polygonu | Náhled/gesto podle implementačního rozhodnutí; nesmí vybírat znaky či přemisťovat objekty | Uzavírá textové laso, **nikoli výběr slova** | Obdélníkový výběr dle SEL-01 |
| Text → **Odstavcový blok** | Bez tažení nový rámec nevznikne | Vymezí hranice nového odstavcového rámce; nikoli výběrový obdélník | Nesmí vybírat slovo ani přepínat editaci | Obdélníkový výběr dle SEL-01 |
| Objekty / Objektové laso | Výběr dle geometrie OBJ-02, resp. bod objektového lasa | Přesun objektu, kreslení, nebo odpovídající režim | V objektovém lasu ukončí jeho vlastní polygon | Obdélníkový výběr objektů |
| Miniatury stránek | Aktivace/výběr stránek | Přesun se schváleným potvrzením | Zachovat stávající chování | Neměnit existující chování miniatur |
| AcroForm widget či zaměřené pole UI | Převzít existující obsluhu pole, **ne PDF textový editor** | Dle widgetu | Dle widgetu | Neinterferovat s editací pole |

**Kontextové modifikátory** patří výhradně k odpovídajícímu aktivnímu nástroji: `Ctrl+klik` v Textovém kurzoru = výběr věty, v Objektech = cyklování objektů, v miniaturách = vícenásobný výběr. `Ctrl+dvojklik` v Textovém kurzoru = řádek; `Shift+dvojklik` = odstavec a pak sloupec; `Alt+dvojklik` = podobné formátování. Nespouštět těchto více významů jedním vstupem. Pokud je zvoleno Textové laso a přijde dvojklik, **uzavře laso** (žádný výběr slova). `WM_LBUTTONDBLCLK` a jednotlivé kliky směrovat podle aktivního podrežimu; nevytvořit nejprve falešný bodový text a následně teprve rozhodnout, že šlo o dvojklik.

**Bodový text na prázdné stránce:** prázdné místo znamená skutečnou stránkovou pracovní plochu bez existujícího editovatelného textu, aktivního formulářového widgetu, odkazu či klikatelné anotace. Klik může vytvořit **dočasný caret a prázdný pracovní textový objekt**, ale bez napsaného textu a potvrzení nemá vznikat nechtěný prázdný PDF objekt, příznak změny ani transakce AutoRecover. Zrušení Escape zahodí prázdný návrh. Při vstupu znaků začít normální vratnou operaci s výchozím stylem. Bodový text může přirozeně růst v délce; přetečení do dalšího řádku nezaměňovat s rámcem odstavcového textu.

**Odstavcový rámec:** krátký klik bez překročení minimální prahové vzdálenosti nevytvoří prázdný rámec. Během tažení zobrazovat průběžný náhled. Esc zruší tvorbu, bez dirty flag. Vložený rámec používá zavinování řádků podle šířky, formátování podle panelu Typografie a je vratný jedním příkazem po skutečném dokončení. Při změně šířky rámce zachovat neovlivněný ostatní PDF obsah.

**Přepínání režimů:** je vždy explicitní. Záložky **Výběr/Typografie**, nápověda `?`, změna fokusu a hover **nemění podrežim plátna**. Změna režimu při rozkresleném polygonu/rámci bezpečně ruší pouze rozpracovaný náhled (neprovádí dva příkazy, nepřevádí jej na objekt jiného typu). Zaměřené vstupní pole vlastností má přednost před klávesovými gesty plátna.

## C. Smíšené formátování výběru — schválený princip

Panel Typografie se napojuje na **aktuální textový výběr** (jediný souvislý úsek, několik úseků i výběr s rozdílnými formátovacími běhy), a musí **vyhodnocovat každou vlastnost samostatně**. Důraz na neporušení původních PDF běhů a vložených písem.

- **Společná hodnota** ve všech označených znacích → zobrazit danou hodnotu (např. velikost `12 pt`, písmo Arial, modrá).
- **Různé platné hodnoty** → jednoznačný stav **„— Smíšené —“** nebo `—` ve číselném poli. U přepínačů (Tučné, Kurzíva, Podtržení) třetí **smíšený** stav (vizuálně odlišný od zapnuto i vypnuto). U barev vícebarevný vzorník / „Více barev“.
- **Nedostupné** = vlastnost nelze bezpečně načíst nebo upravit (např. font subset či text převedený do křivek); **nesmí** být prezentováno jako „Smíšené“. Vysvětlit důvod a vyřadit neproveditelnou operaci.
- **Bez aktivního výběru / pouze caret**: hodnoty se načtou z polohy kurzoru, popř. z výchozího stylu nového bodového/odstavcového textu; změna může definovat styl následujícího psaní, **nepřepíše automaticky okolní obsah**.
- Při změně jedné vlastnosti **měnit výhradně tuto vlastnost u všech vybraných úseků**, ostatní vlastnosti, příslušné fonty, barvy, velikosti, řádkování atd. zůstávají původní. Výjimkou je explicitní souhrnný příkaz (např. **Běžný řez** výslovně ruší Tučné i Kurzívu); u něj popsat obě změny v tooltipu.
- **A+ / A− jsou relativní** vůči každému označenému běhu. Např. 12+16 pt → po A+ 14+18 pt. Udržet rozlišení jednotlivých velikostí, přitom jeden příkaz = **jeden vratný krok** pro celý výběr. Použít definovanou konzistentní stupnici a maximální/minimální velikost.
- **Přímé zadání `14 pt`** ve smíšeném výběru nastaví obě velikosti na 14 pt, ale zachová všechny ostatní rozdíly. Výchozí prázdná/neurčitá hodnota není nula, **neaplikovat žádnou hodnotu, dokud uživatel neprovede explicitní změnu**.
- **Smíšené Tučné/Kurzíva/Podtržení**: první klik sjednotí příslušnou vlastnost do **zapnuto pro všechny**, další klik vypne pro všechny. Neměnit zbývající přepínače ani rodiny písem. Všechny zapnuto → klik vypnout, všechny vypnuto → klik zapnout.
- **Barva textu**: volba konkrétní barvy sjednotí pouze barvu. Pokud zůstanou různé fonty a velikosti, zobrazují se nadále jako smíšené. Totéž platí pro řádkování, zarovnání, proklad, mezerování, odsazení a efekty v jejich vlastním relevantním rozsahu (znak/odstavec).
- **Nesouvislý výběr**: společné typografické vlastnosti lze upravit. **Neumisťovat jeden smyšlený caret do několika úseků a nezahajovat nejednoznačné psaní.**

**Historie:** každé potvrzené typografické nastavení nad jedním označeným výběrem je jediný obsahový krok v centrální historii dokumentu (`SEL-03`) a přepne příznak neuložené změny; pouhý výběr, přepnutí záložek, neurčitý stav, hover, nápověda a změna demo výběru nikoliv. Undo/Redo musí obnovit **původní hodnoty každého běhu samostatně**. Respektovat AutoRecover a nepoškozovat podpisy/anotace/formulářové widgety.

## D. Změny v HTML v3.3

V `rozhrani/cely_vizual_text_vyber_typografie_v3.3.html`:

- V záložce Výběr jsou **na jediné řadě čtyři výběrové režimy** v přesném pořadí a **otazník** vpravo. Základní a rozšířený výběr jsou v místním otevřitelném vysvětlení; předchozí dlouhé tabulky a neaktivní samostatné ikonky jsou z panelu odstraněny.
- Klik na ikonku změní pouze vybraný režim UI a jeho stručný text. Popup `?` se otevírá kliknutím a zavírá opět kliknutím, mimo plochu nebo Esc.
- V Typografii je **výslovně demonstrační** přepínač: caret / jeden úsek / dva smíšeně formátované úseky. Není součástí finální aplikace; ta bude vycházet z reálného PDF výběru.
- U smíšeného výběru se zobrazuje neurčité písmo, velikost, řádkování a tri-state tlačítka. A+ posouvá oba příklady relativně; přímá hodnota sjednocuje jen velikost; Tučné sjednocuje pouze tučnost. Bez změny skutečného PDF.
- Stále jde o **HTML demonstraci**: kliky do dokumentu, skutečný výběr PDF, vznik bodového textu, kreslení polygonu nebo odstavcového rámce ani editace původních PDF fontů nejsou implementovány. Před samotnou implementací provést inventuru reálného Go/Win32 a PDFium kódu.

## E. Akceptační testy pro Codex (před produkčním nasazením)

1. **Rozvržení/UI** při úzkém i širokém okně a 100/150/200 % DPI: čtyři výběrové ikony v jednom řádku, `?` zcela vpravo, správné tooltipy a přístupnost, žádné překrytí ostatních panelů.
2. **Výběrová pravidla:** v Textovém kurzoru single-click caret, double-click slovo/segment/věta/odstavec; klik na volné místo bodový text; drag v existujícím textu znakový výběr; žádný prázdný textový objekt při kliknutí a Esc.
3. **Obdélník/Laso:** oba jsou samostatné režimy pro **text**, nesmí spouštět objektový handler ani přepisovat modifikátory TXT-02. Textové laso dvojklikem uzavře polygon, nikoli slovo.
4. **Odstavcový rámec:** vytvoří se jen drag gestem a obsah se zalamuje; samotný klik žádný bodový text nevytváří.
5. **Prostřední tlačítko:** v každém vhodném režimu prostřední tažení obdélník, samotný prostřední klik nic; otočení kolečka scroll, `Ctrl+kolečko` zoom.
6. **Konflikty:** Objekty `Ctrl+klik` cyklus překryvů, miniatury Ctrl/Shift multivýběr a přesun s potvrzením, AcroForm vstup zůstává prioritní. Jediný vstup spustí nejvýše jeden příkaz.
7. **Typografie:** smíšené 12+16 pt a Arial+Georgia / červená+modrá zobrazí smíšené hodnoty. A+ → 14+18 bez změny fontů, barev; přímé 14 pt → 14+14 se zachováním fontů, barev. Tučné ve smíšeném stavu ON pro všechny, následující klik OFF pro všechny. Nedostupné ≠ Smíšené.
8. **Historie:** po změně jedné vlastnosti Undo vrátí každý běh do původního formátu, Redo opět aplikuje; jedna operace = jeden krok. Volba UI režimu, nápověda a samotný výběr nemění obsah dokumentu (kroky výběru dle SEL-03 se však chronologicky zaznamenávají).
9. **Stabilita a návrat:** přepnutí Výběr/Typografie, nástroje i aktivního dokumentu nezmění nevýslovně obsah ani nerozbije výběr. Kód nesmí zjednodušit původní práci s PDF textem, objekty, anotacemi, formuláři, fonty a barvovými profily.

## F. Technické omezení a rozsah

**Nesmí se vydávat za hotové:** nativní editace libovolného PDF textu, zachování všech fontů, proměna křivek/bitmap na text, složité sloupce, bidi a některé transformační matice. Upozornit na rozdíly v zachování typografie a nabídnout explicitní postup (OCR nebo vhodný náhradní font) tam, kde je to nutné. Globální Nastavení aplikace a schválený panel Separace se touto změnou **nepřestavují**.

**Pravidlo precedence:** tato v3.3 specifikace upřesňuje část `TXT-06` (nový bodový text je v nástroji Textový kurzor; samostatné Textové pole dělá pouze odstavcový text) a nahrazuje pouze dočasné neschválené v3.1 omezení, že textové ikony obdélníku/lasa jsou neaktivní. **Ostatní schválené SEL/TXT/OBJ a Q31 zůstávají závazné.**