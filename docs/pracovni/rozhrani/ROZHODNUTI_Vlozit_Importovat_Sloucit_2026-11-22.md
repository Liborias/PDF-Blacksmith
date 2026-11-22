# PDF Blacksmith — rozlišení Vložit / Importovat / Sloučit a otevřít

**Datum rozhodnutí:** 22. listopadu 2026  
**Stav:** schválený směr a význam příkazů; detaily převodního modulu, podporovaných formátů a volby stránkování dořešit před implementací. **Zatím neměnit HTML ani zdrojový kód.**

## 1. Dva oddělené nástroje: Vložit objekt a Vložit strany — upřesněno uživatelem 22. 11. 2026

**Cíl:** oba příkazy upravují **aktuálně otevřené PDF**, ale každý pracuje s jiným typem obsahu a má samostatnou obsluhu.

### 1A. Objekty → Vložit objekt

- Je **součástí nástroje Objekty** / grafických úprav; není totožný s vložením celé stránky.
- Vkládá obrázek, vektorovou grafiku nebo jiný technicky podporovaný **grafický objekt do existující stránky**.
- Uživatel volí umístění, rozměry, případně otočení, průhlednost, vrstvení a další vlastnosti podle možností stávajícího editoru Objekty.
- Preferovat zachování vektorů a barevných profilů. Při převodu složitých grafických formátů (PSD, AI, XCF, KRA atd.) může dojít ke ztrátě editovatelnosti vrstev nebo efektů; jasně upozornit.
- Vkládání se musí účastnit historie Ctrl+Z/Y a AutoRecover a nesmí měnit okolní obsah PDF.

### 1B. Soubor → Vložit strany

**Samostatný příkaz v nabídce Soubor**. Jeho výsledkem je přidání **jedné nebo více celých stránek** do právě otevřeného PDF.

Má dvě volby zdroje:

**(1) Z jiného PDF**
- Vybrat existující PDF.
- Vybrat rozsah **obdobným ovládáním jako při Tisku**: všechny stránky, jednotlivé stránky, rozsah nebo kombinace např. `1, 3–5, 9`; kde je relevantní, volby liché/sudé.
- Ukázat počet a případně náhled vybíraných stránek.
- Určit **místo vložení** (před aktuální stránku / za aktuální stránku / na zvolenou pozici / na konec) — konkrétní výchozí volba a případné doplňky budou ještě předmětem revize UX.

**(2) Z jiného dokumentu**
- Umožnit běžné kancelářské, prezentační a grafické dokumenty: Word, Excel, PowerPoint a další podporované komerční i open-source formáty dle katalogu níže.
- Zdroj se nejdříve převede na **dočasné PDF stejným konverzním modulem jako `Soubor → Importovat`**. Nesmí vzniknout druhá nezávislá implementace převodu.
- Po převodu nabídnout kontrolu výsledku a **výběr stránek / rozsahu obdobně jako při Tisku**; poté vložit zvolené stránky do **aktuálního PDF**, nikoli otevřít nový samostatný dokument.
- Zohlednit vstupní volby příslušného formátu (např. listy a tiskové oblasti u tabulek, snímky u prezentací); případné chyby převodu nezamlčovat.

**Společná pravidla obou větví:** zachovat zdrojový soubor, neprovádět nechtěné přepsání aktuálního PDF, začlenit vložení do Ctrl+Z/Y a AutoRecover; při editaci podepsaného PDF varovat před dopadem na platnost podpisů. Při problémech s rozměry, formuláři, odkazy a anotacemi zachovat pokud možno jejich strukturu a nespouštět nevratný převod bez jasného upozornění.

**Rozhraní:** levé podmenu `Soubor → Vložit strany` a pravý panel `Vlastnosti` s výběrem jedné z obou větví, rozsahu a místa vložení. `Objekty → Vložit objekt` má vlastní nastavení v panelu Objekty; tyto dvě funkce **nezaměňovat ani nevytvářet duplicitní ovládání**.

## 2. Importovat

**Cíl:** převést **jeden zdrojový dokument** jiného formátu do samostatného PDF, nikoli jej automaticky vkládat do otevřeného dokumentu.

- **Cílový rozsah formátů je široký:** běžné kancelářské dokumenty (Word, Excel, PowerPoint), grafika z Photoshopu a Illustratoru i dokumenty z obvyklých open-source programů. Podrobný seznam kandidátů a pravidla ověřování podpory jsou v části **Katalog zdrojových formátů** níže. Podporu jednotlivých nativních formátů nevydávat za hotovou, dokud nebude ověřen dostupný převodní engine.
- U Wordu: volby pro zachování stránkování, velikosti stránky a kvalitní převod obrazů/písem v rámci možností dostupného enginu.
- U Excelu a OpenDocument tabulek: výběr listů, tiskových oblastí, orientace, papíru a vhodného měřítka; před konverzí přehled stran.
- U PowerPointu a OpenDocument prezentací: vybrat snímky, rozložení na stránce, velikost papíru a případně zahrnutí poznámek k prezentaci.
- Výsledné PDF se otevře **v nové záložce**. Původní soubor se nikdy automaticky nepřepíše.
- Konkrétní engine (LibreOffice/Office či jiný) a jeho licence, dostupnost a kvalita jsou **technické rozhodnutí k ověření**, nikoli již hotová funkce. Makra ani nedůvěryhodný aktivní obsah nikdy automaticky nespouštět.

## 3. Sloučit a otevřít

**Cíl:** vytvořit **nové společné PDF** ze **dvou a více zdrojových souborů**, přičemž zdroje mohou mít různé formáty.

- Sestava zdrojů: **dvě nebo více položek**, libovolná kombinace PDF a podporovaných kancelářských, prezentačních, grafických i dalších dokumentových formátů (včetně Word/Excel/PowerPoint, zdrojů Photoshop/Illustrator a běžných open-source programů). Nepodporovaný formát musí být jednoznačně oznámen před zahájením spojování.
- Seznam s viditelným pořadím; možnost přetažením změnit pořadí, odstranit položku, přidat další soubor, upravit rozsah stránek (u PDF) a nastavení konverze (u ostatních).
- Před převodem ukázat finální pořadí a podle možností náhled výsledné sestavy.
- Ne-PDF zdroje se nejprve převedou na přechodné PDF; spojení proběhne ve zvoleném pořadí; výsledný soubor se **otevře v nové záložce**.
- Žádný zdroj se při převodu/sloučení nemění. Při dílčím selhání převodu nezamlčet chybu, nabídnout opravný krok a nedokončit výstup tiše se ztracenými stranami.
- Zachovat původní existující funkce slučování PDF (Q31) a teprve potom rozšířit jejich vstupy o další typy.

## 3a. Katalog zdrojových formátů – požadovaný rozsah (ověřit technicky)

| Skupina | Příklady zdrojů k prověření | Pravidlo převodu |
|---|---|---|
| Kancelářské texty | DOC, DOCX, RTF, ODT a další běžné textové dokumenty | Zachovat vzhled a stránkování, upozornit na náhrady fontů |
| Tabulky | XLS, XLSX, ODS | Respektovat listy, tiskové oblasti, orientaci a měřítko |
| Prezentace | PPT, PPTX, ODP | Převést snímky na stránky; volitelné předvolby rozložení |
| Běžné obrázky | PNG, JPEG, TIFF, BMP, WebP a další bezpečně podporované typy | Zachovat vhodné rozlišení a barevné profily |
| Vektorová grafika | SVG, EPS, případně PDF-kompatibilní část AI | Pokud možno zachovat vektory; zkontrolovat převod písem, efektů a průhlednosti |
| Nativní komerční grafika | Photoshop PSD/PSB, Illustrator AI, případně další běžné formáty | **Výslovně ověřit možnost načtení/konverze** a způsob práce s vrstvami; při nemožnosti přímého převodu nabídnout uživateli výstupní PDF nebo kompatibilní meziformát |
| Nativní open-source grafika | GIMP XCF, Krita KRA, Scribus SLA, Inkscape SVG a další obvyklé formáty | Podle dostupného exportéru nebo aplikace; nevydávat za nativní PDF editaci |
| Další otevřené formáty | OpenDocument včetně ODP, případně další běžné výstupy open-source kancelářských aplikací | Zařazovat podle ověřené kompatibility a předvídatelného vzhledu |

**Požadavek uživatele je podpora širokého spektra běžných formátů, nikoli pouze Word/Excel.** Tabulka představuje **cíl a kandidáty**, nikoli zaručený seznam již funkčních importérů. U formátů Photoshopu/Illustratoru, vrstvených nativních grafik a některých proprietárních souborů mohou být nutné dodatečné aplikace, moduly či kompatibilní exportní meziformáty. Upřednostnit kvalitní převod s upozorněním na odchylky před tichým ztrátovým výsledkem.

**Bezpečnost a předvídatelnost:** neprovádět makra, externí skripty ani odkazy z dokumentů; neimportovat skrytý aktivní obsah bez vědomí uživatele. Chybějící písma, efekty, vrstvy, průhlednosti či barevné profily uvést v diagnostice. Pro skládání 2+ souborů se má stejný převodní modul využít jednotlivě, potom sloučit výstupy podle zvoleného pořadí.

## 4. Uživatelské rozhraní – návrh k dalšímu projití

- **Soubor → Importovat** = konverze externího zdroje do PDF, s pokročilými volbami v pravém panelu.
- **Soubor → Sloučit a otevřít** = sestava 2+ vstupů, požadované pořadí, konverze, společné PDF v nové záložce.
- **Soubor → Vložit strany** = vložit do aktuálního PDF celé stránky buď z jiného PDF s výběrem rozsahu jako při Tisku, nebo z jiného dokumentu po převodu společným modulem Importovat.
- **Objekty → Vložit objekt** = vložit obrazový či vektorový objekt do existující stránky a následně jej upravovat grafickými nástroji. Oba příkazy jsou **samostatné**.
- **Soubor → Otevřít = VÝHRADNĚ PDF.** Nepřijímá dokumenty Word, Excel, PowerPoint ani nativní grafické formáty a **nespouští implicitní konverzi**. Pro jiný formát slouží Importovat nebo Sloučit a otevřít; výběr souboru má odpovídající filtr PDF a při nepodporovaném vstupu nabídnout jasné přesměrování na Importovat.
- Pravý panel **Vlastnosti** má pro každý z těchto úkonů vlastní pokročilá nastavení, ne duplicitní tlačítka celých nástrojů.
- Aktuální HTML v2.6 může ještě zobrazovat starší význam Importovat: před implementací je nutné **sjednotit s tímto novějším rozhodnutím**.

## 5. Otevřené otázky pro příští kolo

1. Podrobnosti výběru rozsahu vložených stran: shodná ergonomie jako při Tisku, validace kombinovaného rozsahu, zachování pořadí a případná volba lichých/sudých stran.
2. Výchozí pozice vložení stran (před / za aktuální, na zadané místo, na konec) a pravidla při odlišných rozměrech, rotaci, formulářích a anotacích.
3. Nastavení Excelu: tiskové oblasti, listy, skryté listy, opakované řádky, měřítko; zachování původního formátování.
4. Ověřit reálnou podporu katalogu DOC/DOCX, XLS/XLSX, PPT/PPTX, ODT/ODS/ODP, SVG, PNG/JPEG/TIFF, PSD/PSB, AI/EPS, XCF, KRA, SLA a dalších; zvlášť chování velkých, chráněných, poškozených a vrstvených dokumentů.
5. Zachování písma, barevných ICC profilů, průhledností a vektorů podle možností vstupního formátu/enginu; bezpečné chování při nepodporovaných prvcích.
6. Případné předvolby konverze a šablony sestav pro opakované operace.

**Stav:** rozdělení **Vložit objekt** (Objekty) a **Vložit strany** (Soubor) bylo uživatelem upřesněno 22. 11. 2026. Podrobnosti a technická realizace zůstávají k dopracování. Otevřít je pouze pro PDF. Podpora širokého katalogu zdrojů, včetně PPT/PPTX, PSD/AI a formátů open-source aplikací, je **cílový požadavek k technickému ověření**, nikoli již implementovaná funkcionalita.
