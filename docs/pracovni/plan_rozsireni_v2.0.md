> **ARCHIV / NAHRAZENO 25. 11. 2026:** Tento dokument slouží pouze k dohledání historie rozhodnutí. **NENÍ samostatným aktuálním pokynem pro Codex.** Jediné závazné zadání: [CODEX_START.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_START.md) a [CODEX_MASTER_v6.0.md](/Liborias/PDF-Blacksmith/blob/main/docs/pracovni/CODEX_MASTER_v6.0.md). **Nová priorita: nejprve rozhraní, F1 až později po schválení.** Platí finální mapy v6.0. Původní historický obsah níže zůstává zachován.

---

# PDF Blacksmith – plán dalšího vývoje v2.0 (21. 11. 2026)

**Stav:** pracovní návrh priorit a rozsahu. Nejde o příkaz automaticky vše implementovat.  
**Repozitář:** `Liborias/PDF-Blacksmith` (`main`, dle čteného `PLAN.md`, `NAVOD.md`, souborů UI a backendu).  
**Závazný princip:** žádné stávající funkce nenahrazovat redukovanou verzí; vždy ověřit regresi.  
**Značky:** ✅ již v projektu doloženo; 🟢 už schváleno v předchozí specifikaci; 🟡 nový návrh k upřesnění; ⚪ pozdější rezerva.

## 0. Co už existuje (nebudovat podruhé)

V repozitáři jsou doloženy základní prohlížení v PDFium, OCR, fulltext, rozložení a přesun stránek, spojování a dělení, úpravy textu/objektů v pracovním stavu, dvě PDF vedle sebe, třírežimový převod Gray / K100 / explicitně K, předvolby exportu, základní PDF/X validace, náhled separací a podepisování i ověřování digitálních podpisů. Přesný stav nasazení ověřit testy a Windows buildem. Jednotlivé stávající formulářové widgety lze v PDF zachovat, **plnohodnotný editor formulářů však v současném `workspaceTabs` doložen není**.

## Priorita P0 – základy spolehlivosti a práce (🟢 dříve schválené)

1. **Bezpečné ukládání a AutoRecover.** Obnovovací snímky při nečinnosti i nepřetržité práci, obnova po pádu, přepnutí aktivního souboru po „Uložit upravenou kopii…“, původní karta se nahradí kopií; potvrzování zavírání více dokumentů. Testovat pád/zápis/zrušení; originál nesmí být neúmyslně přepsán.
2. **Historie a výběr.** Chronologické Ctrl+Z/Y i pro kroky výběru; přesný zásah na geometrii, průhledné bitmapy, Ctrl+klik cyklický výběr; laso protnutím (výchozí) nebo celé objekty.
3. **Kvalita vykreslení.** Adaptivní znovuvykreslení z původních PDF dat, dlaždice při zoomu, hladké přechody, volby vyhlazování. Přepnutí 100 % (`zoom_1_1.svg`), bez záměny s fyzickou kalibrací 1:1.
4. **Dokončit schválené Poznámky.** Textové a grafické anotace jako oddělená standardní PDF data; pracovní zobrazení viditelné samostatně pro každý dokument; tisk i export standardně bez pracovních poznámek; možnost zahrnout.
5. **Separace.** Zachovat stávající UI a převody; přidat barevné značky CMYK/spot, izolovanou desku se šedotónovým (default) či barevným zobrazením, měření pod kurzorem, TAC a limit (výchozí 300 %). Rozlišovat podíl potištěné plochy vs. tón v bodě a nepředstírat přesnost nevypočtené hodnoty.

## Priorita P1 – Formuláře (🟢 etapa F1 schválena, F2/F3 v návrhu)

**Cíl:** zachovat, vyplňovat, vytvářet, upravovat a bezpečně exportovat standardní interaktivní formuláře PDF **AcroForm**. V UI nový levý nástroj „Formuláře“, v pravém panelu režimy **Vyplnit**, **Návrh polí**, **Vlastnosti / Pořadí Tab**, **Kontrola**.

### Etapa F1 – kompatibilita, vyplňování, ukládání (SCHVÁLENO – varianta A)

- Detekovat `/AcroForm` a widget anotace, bezpečně zobrazit známé typy polí a udržet vazby v `/Fields` / `/Kids` a vzhled `/AP`.
- Vyplnit text, víceřádkový text, checkboxy, radio skupiny, seznamy a comboboxy; Tab pořadí, textová navigace, klávesnice, diakritika a české znaky.
- `Ctrl+S` / „Uložit upravenou kopii…“ uchová interaktivní pole, hodnoty a appearance streams; pracovat s historií a AutoRecover.
- Oddělit pracovní poznámky od formulářových widgetů: **Export PDF bez poznámek NESMÍ odstranit vyplněná pole, odkazy, widgety podpisu ani jiný aktivní obsah**, pokud to není výslovně součástí zvoleného profilu exportu.
- U XFA formulářů detekce a jasné upozornění na nepodporované chování; bez tichého ztrátového převodu nebo uložení.

### Etapa F2 – návrh a editace formulářových polí (ZATÍM PLÁN)

- Přidání textového pole, víceřádkového pole, checkboxu, radio skupiny, seznamu, rozbalovací nabídky, běžného tlačítka a podpisového pole.
- Geometrie a vlastnosti: název/identifikátor, viditelný popisek a tooltip, poloha/rozměr, povinné, pouze ke čtení, výchozí hodnota, řez a velikost fontu, barva, rámeček, viditelnost/tisknutelnost; čitelné názvy pro asistivní technologie.
- Ochrana proti duplicitnímu názvu pole při kopírování; zarovnání polí, pořadí Tab, náhled chování.
- Digitální podpisové pole napojit na **existující** podpisový nástroj; úpravy již podepsaného PDF nepovolovat bez kontroly dopadu na podpisy.
- Výchozí implementace bez vykonávání libovolných vložených skriptů/JavaScriptu. Při výskytu složitých akcí/validací zobrazit omezení; nepředstírat podporu.

### Etapa F3 – výstupy pro běžnou praxi a předtisk (ZATÍM PLÁN)

- Přepínač pro zachování editovatelného formuláře při běžném ukládání a explicitní **Zploštit formulář pro tisk** (zvlášť od exportu „bez poznámek“).
- Zploštění až po ověření obsahu/vzhledu, s varováním o ztrátě editovatelnosti a dopadu na elektronické podpisy, vždy nová kopie nebo potvrzená změna.
- Testy běžného tiskového výstupu s textem, checkboxy a podpůrnými PDF pro ověření vzhledu v Acrobat/Edge.
- **Pokročilé funkce odložit:** FDF/XFDF import/export, výpočtová pole, datové formáty, složitý JavaScript a XFA konverze – po samostatném schválení.

**Akceptační testy:** generování a vyplnění vícestránkového formuláře s češtinou; uložení/kopie/nové otevření; kompatibilní vzhled v Acrobat/Edge; ověření zachování widgetů při exportu bez poznámek; žádný zásah do podpisů; bezpečná reakce na XFA; Undo/Redo a AutoRecover.

## Priorita P2 – Preflight (🟡 návrh)

- **Nepřidávat duplicitní kontrolu** současného PDF/X, ICC a separací. Rozšířit nad ně profilované pravidlo kontroly: fonty, DPI, RGB ve výstupu, přetisk bílé, hairlines, geometrie/bleed, TAC, transparentnost a specifikace tiskárny.
- Výsledky po stránkách a objektech, úroveň závažnosti, možnost označit výjimky; bezpečná automatická oprava pouze u jednoznačných scénářů.
- Export protokolu a kontrola před tiskovým exportem; zavést jednotný profil pravidel.

## Priorita P3 – Grafické porovnání PDF (🟡 návrh)

- Rozšířit existující duální pohled o synchronizované posouvání/zoom, částečně průhledné překrytí, před/po posuvník, geometrické i vizuální rozdíly.
- Rozlišit změnu barev způsobenou cíleným ICC převodem od skutečné chyby obsahu. Neničit již fungující přetažení stránek mezi PDF.

## Priorita P4 – Měření, pravítka a vodítka (🟡 návrh)

- Měření vzdálenosti, úhlu, plochy a tloušťky vektorové čáry; milimetry, body a případně palce, čtení skutečných boxů PDF.
- Pravidla pro otočené stránky, UserUnit, spadávku a kalibraci referenční délky; nesměšovat zobrazení 100 % se zaručenou fyzickou velikostí na monitoru.
- Vodítka, mřížka, přichytávání až v další fázi podle potřeby.

## Priorita P5 – Dávkové zpracování (🟡 návrh)

- Znovupoužít existující exportní presety, ICC profily, validace a bezpečnou konverzi; neposkytovat druhé nezávislé nastavení stejných barev.
- Fronta souborů, zdroj/výstupní složka, předvolby kroků, průběh, protokol a rozumné obnovení po chybě.
- Bez automatického přepisu zdrojových PDF; konečný výstup do nového souboru, nezávislé testy dávky.

## Závislosti a navržené pořadí

1. Stabilizovat ukládání/undo/vykreslování (P0).
2. Rozpracovat **Formuláře F1**, protože jsou samostatně testovatelné a mohou být užitečné i mimo předtisk.
3. Paralelně připravit specifikace Preflight a nástroje Porovnání, ale vlastní implementaci řadit podle stability backendu.
4. Formuláře F2/F3, měření a dávky přidávat po integračních testech.

## Co jde dokončit přímo v tomto chatu (bez placeného Codexu)

- **Teď hotovo:** tento plán, návrh formulářů F1–F3, úplná mapa ikon a SVG pro budoucí nástroje, obrázkové náhledy.
- **Další možné výstupy:** uživatelské scénáře, podrobné akceptační testy, návrhy rozložení pravých panelů, stavové diagramy, vzorová testovací PDF včetně jednoduchých AcroForm, testovací fixture s chybnými separacemi, návrh README/NAVOD, migrace ikon do kódu jako konkrétní patch bez přímého spuštění Windows aplikace.
- **Ověření bez sestavení:** kontrola dokumentace, JSON struktur, SVG validace, statické čtení Go/Python kódu, generování testovacích dat a vývojových zadání po malých dávkách.
- **Omezení:** bez spuštění programu ve Windows nelze potvrdit skutečné chování Win32 UI, kvalitu vykreslování nebo integraci PDFium v distribuované aplikaci. Bez účetních přístupů a skutečných uživatelských PDF nelze simulovat všechny externí formy podpisů, tiskárenské RIPy a XFA.
- **Příští krok pro Codex po obnovení limitu:** předat etapové zadání P0 a F1 společně s tímto plánem a mapou; ne jeden obří příkaz k implementaci celé roadmapy.

## Zdrojové podklady

- Původní plán projektu: `PLAN.md` v repozitáři; popis stávajícího UI a operací: `NAVOD.md`, `workspace_windows.go`, `viewer_windows.go`, `proof.py`, `COLOR_MANAGEMENT.md`.
- Rozšířená SVG sada: `PDF_Blacksmith_ikony_nastroju_SVG_v4.2.zip` – není instrukcí automaticky všechny rezervy zapojit.
- Kompletní specifikace dříve schválených funkcí: `PDF_Blacksmith_specifikace_Codex_v4.0.md`.


## Detailní návrh a testovací podklady pro Formuláře (v1.1)

- **F-Q1 schváleno, varianta A:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.1.md` – čtyři funkční režimy panelu, F1/F2/F3, UI, regresní testy, technická integrace.
- **Skutečná testovací fixture:** `PDF_Blacksmith_Formulare_testovaci_AcroForm_v1.pdf` – dvě stránky a 17 interaktivních AcroForm polí.
- **Technický soupis polí:** `PDF_Blacksmith_Formulare_testovaci_AcroForm_fields.json` – kontrolní názvy, typy, hodnoty.
- **Návaznost na ikony:** beze změny sady 218 SVG v4.2; ikony plánovaných formulářových funkcí nepoužívat v aktivním UI bez funkční implementace.
- **Výsledek:** soubor a základní struktura otestována na fixture, fungování plánovaného editoru v aplikaci PDF Blacksmith dosud netestováno a není implementováno.


## Rozhodnutí 21. 11. 2026 – Formuláře F-Q1

**Výslovně schváleno:** varianta A, tedy **F1 jako první samostatná implementační etapa**. Čtení, vyplňování a bezpečné ukládání existujících formulářů AcroForm je prioritou. **F2 (tvorba a úpravy polí) a F3 (flatten a specializované tiskové/exportní funkce) zatím nezahajovat bez dalšího odsouhlasení.**

Etapu F1 rozdělit na: A inventura a datový model; B zobrazení a fokus; C editace hodnot; D bezpečné uložení a test nezávislými PDF aplikacemi; E Undo/Redo, AutoRecover a ochrana podepsaných/XFA PDF; F regresní a integrační testy. Viz `PDF_Blacksmith_Formulare_specifikace_navrh_v1.1.md`.

**Schválení etapy neznamená, že je funkce hotová.** Před implementací ověřit aktuální zdrojový kód, vytvořit regresní baseline a přidávat nové UI prvky jen tehdy, když mají funkční backend. Pokročilé úpravy jiných nástrojů se tímto rozhodnutím nemění.


## Formuláře F-Q2 – schváleno 21. 11. 2026

**Varianta A:** při načtení PDF s vyplnitelným AcroForm automaticky nenutit aktivaci nástroje Formuláře; jen zobrazit neblokující odkaz „Dokument obsahuje formulář – Otevřít Formuláře“. Interakce nesmí přerušit dosavadní nástroj, nastavit neuložený stav nebo vytvořit záznam v historii. Aktivace pouze po kliknutí uživatele, vždy ve správné kartě a panelu. F1 je první implementační etapa, nikoliv hotová funkce.

**Zbývající UI rozhodnutí:** F-Q3 – přímá editace polí mimo aktivní režim Formuláře; následně detaily panelu Vyplnit a chování při chybách validace.


---

## SCHVÁLENO: Formuláře F-Q3 – 21. 11. 2026

**Varianta A s podmínkou:** interaktivní pole se vyplňují **pouze v nástroji `Formuláře → Vyplnit`**, bez konfliktů s Text/Objekty. **Při pozdější tvorbě v F2** nabídnout typografické možnosti podobné stávajícímu nástroji Text (font a řez, velikost, barva, zarovnání, víceřádkové vlastnosti a rozměry pole). **Při vyplňování ve F1** se vlastnosti textu přebírají z formuláře; uživatel je běžně nemění. Implementace F2 zůstává na samostatné pozdější schválení. Viz specifikace Formulářů v1.3.

## Priorita P1b – Vylepšení stávajícího vyhledávání (po Formulářích; nový požadavek, detaily k upřesnění)

Po rozpracování Formulářů se vrátit k nástroji **Najít v PDF**. Zachovat stávající fulltext, seznam a navigaci. Požadovaná rozšíření: **hledání ihned při psaní**, akce **„Označit všechny výskyty šipkou“**, samostatný **křížek `×`** pro odstranění jednotlivé šipky, hromadné **„Odstranit šipky“**, **uložení výskytů včetně šipek** do pracovního PDF a **textová poznámka ke každé šipce**. Trvalé šipky vhodně navázat na standardní PDF anotace z modulu Poznámky. Běžný export/tisk bez poznámek zůstává výchozí; vyznačené revizní PDF se exportuje s výslovně zahrnutými poznámkami.

**Stav:** uživatelsky požadováno a zařazeno do plánu, podrobné UX/technická rozhodnutí až **po Formulářích**. **Samostatný rozpracovaný návrh**: `PDF_Blacksmith_Vyhledavani_navrh_k_upresneni_v0.1.md`. Nejde o příkaz spustit implementaci bez podrobného odsouhlasení. Podle aktuální kapacity pořadí P1b vs. P2 Preflight případně znovu potvrdit.


## Formuláře – schváleno F-Q4 (21. 11. 2026)

Varianta A: v **Formuláře → Vyplnit** ve výchozím stavu jemně zvýraznit dostupná formulářová pole s přepínačem **„Zvýraznit vyplnitelná pole“**. Při přepnutí na jiný nástroj zvýraznění zmizí. Zvýraznění je pouze dočasná obrazovková pomůcka; nijak nezasahuje do uloženého PDF, tisku ani exportu a nezakládá změnu v historii. Netýká se nedostupných/ReadOnly/skrytých widgetů. Implementace F2/F3 nadále neschválena; viz F-Q4 a T-21 až T-25 v dokumentu `PDF_Blacksmith_Formulare_specifikace_navrh_v1.4.md`.


## Formuláře F-Q5 – schváleno 21. 11. 2026

Do **Formuláře → Vyplnit** v etapě **F1** přidat **sbalitelný přehled formulářových polí v pravém panelu**. U polí zobrazit stav vyplnění, povinnost a ReadOnly, klikem přejít na konkrétní pole a synchronizovat výběr na stránce se seznamem. Přehled funguje pro aktivní PDF i při duálním zobrazení; jeho sbalení není editace PDF. V případě formulářů s opakovanými widgety respektovat vazbu na jedno logické pole. Viz T-26–T-30 a **F-Q5 v `PDF_Blacksmith_Formulare_specifikace_navrh_v1.5.md`**. Navazující **F-Q6 je již schváleno**: seznam dle stránek a vizuální polohy, Tab podle PDF.


## Formuláře F-Q6 – SCHVÁLENO, 21. 11. 2026

**Varianta A:** seznam formulářových polí na pravé straně řadit podle **aktuálního pořadí stránek a jejich umístění na stránce** (shora dolů, v řádku zleva doprava). **Klávesa Tab / Shift+Tab** prochází vyplnitelná pole samostatně podle pravidel definovaných formulářem, nikoli podle zobrazeného seznamu. Sbalení panelu, zoom a obnova náhledu seznam nemění; přeuspořádání stran aktualizuje řazení pouze v UI. Opakované widgety nepřidávají duplicitní hodnoty. Změna řazení neovlivňuje soubor, podpisy, historii ani AutoRecover. Rozšířená specifikace `PDF_Blacksmith_Formulare_specifikace_navrh_v1.6.md`, T-31 až T-35.

**F-Q7 schváleno – varianta A:** varování při uložení neúplného formuláře, ale možnost „Uložit přesto“; podrobnosti a testy T-36 až T-41 ve formulářové specifikaci v1.7.


## Formuláře F-Q7 – SCHVÁLENO, 21. 11. 2026

**Varianta A:** aplikace při **ručním uložení** upozorní na skutečně **nevyplněná povinná pole** a nabídne přechod k jejich doplnění, ale také **„Uložit přesto“**, aby mohl být rozpracovaný formulář bezpečně uchován. AutoRecover neblokovat, nedoplňovat neznámé hodnoty a nepřepisovat vlastnosti polí. Zachovat souvislost s F-Q5/F-Q6 (seznam a navigace), obecnými pravidly Ctrl+S / Uložit kopii, s ochranou originálu a podpisů. F2 a F3 zatím nevytvářet.

**Specifikace a testy T-36 až T-41:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.7.md`. Příští případné dopracování v rámci F1: rozsah a načasování průběžné kontroly hodnot polí – dosud neschváleno.


## Formuláře F-Q8 – SCHVÁLENO, varianta A (21. 11. 2026)

**Průběžná aktualizace stavu polí bez průběžného obtěžování chybovými hlášeními.** Formátové chyby hlásit **po opuštění pole**; při **ručním ukládání** provést souhrnnou kontrolu vadných hodnot, nevyplněných povinných polí a neověřitelných pravidel. Nabídnout navigaci k poli a možnost **„Uložit přesto“**, v souladu s F-Q7. Validace **nespouští cizí PDF JavaScript**, sama o sobě není změnou obsahu PDF a neblokuje AutoRecover. Podrobnosti a testy **T-42 až T-49**: `PDF_Blacksmith_Formulare_specifikace_navrh_v1.8.md`.

**Aktuální závazné pořadí:** F1 (Q1–Q8), následně samostatně odsouhlasit F2/F3 a až pak upřesnit změny „Najít v PDF“ (okamžité hledání, šipky, poznámky). Tento plán sám o sobě není pokynem k hromadnému přepsání stávající aplikace.


## Formuláře F-Q9 – SCHVÁLENO, varianta A (21. 11. 2026)

Po opuštění **prokazatelně chybně vyplněného pole** se toto pole v nástroji **Formuláře → Vyplnit** jemně **červeně orámuje na stránce** a současně se označí v pravém sbalitelném seznamu. Po opravě a opětovném vyhodnocení chyba v obou místech zmizí. Orámování je výlučně dočasná **obrazovková vrstva UI** – neukládá se do PDF ani se netiskne/exportuje. Zůstává zachována možnost vypnout běžné zvýraznění polí (F-Q4), avšak tento přepínač neskryje ověřené chyby, dokud je aktivní Formuláře. Jiné nástroje nejsou chybovými obrysy ovlivněny. Chyby nesignalizovat jen barvou, neověřitelná pravidla nepovažovat za chyby. Zachovat F-Q7 „Uložit přesto“, F-Q8 validaci až po opuštění pole a souhrn při ukládání.

**Nové akceptační testy T-50 až T-57:** `PDF_Blacksmith_Formulare_specifikace_navrh_v1.9.md`. **Aktuálně schválená rozhodnutí F1: F-Q1 až F-Q9.** F2/F3 a vývoj vyhledávání se řídí dosavadním plánem; jejich implementace není tímto rozhodnutím zahájena.