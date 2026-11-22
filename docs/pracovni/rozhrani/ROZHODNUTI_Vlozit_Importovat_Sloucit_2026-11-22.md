# PDF Blacksmith — rozlišení Vložit / Importovat / Sloučit a otevřít

**Datum rozhodnutí:** 22. listopadu 2026  
**Stav:** schválený směr a význam příkazů; detaily převodního modulu, podporovaných formátů a volby stránkování dořešit před implementací. **Zatím neměnit HTML ani zdrojový kód.**

## 1. Vložit

**Cíl:** přidat obsah do *již otevřeného* PDF.

- **Kancelářské i jiné vícestránkové dokumenty**, včetně Word/Excel/PowerPoint, OpenDocument formátů a dalších běžných formátů, lze vložit **jako stránky**. Zdroj se bezpečně převede do PDF, teprve pak se stránky vloží do otevřeného dokumentu. Při převodu PowerPointu rozlišit snímky a poznámky/tiskové rozložení; u tabulek listy a tiskové oblasti.
- **Grafický podklad** (běžný obrázek, SVG, případně podporovaný výstup z Photoshopu, Illustratoru, GIMPu, Inkscapu, Krity, Scribusu a podobných programů) lze podle jeho povahy vložit **do aktuální stránky jako grafický objekt**, nebo převést/vložit jako samostatnou stránku. Zajistit polohu, velikost a další podporované vlastnosti bez svévolné rasterizace ostatního obsahu PDF. **Editovatelnost vrstev/křivek závisí na konkrétním formátu a převodní cestě**; neslibovat zachování nativních objektů proprietárních editorů bez ověření.
- Rozlišit v dialogu `Vložit stránky ze souboru` a `Vložit obrázek do stránky`; přesnou nabídku a její umístění v menu navrhnout následně.
- Zásahy do dokumentu jsou vratné, respektují historii Ctrl+Z/Y a AutoRecover, nesmějí nechtěně poškodit formulářová pole, poznámky, odkazy, podpisy nebo další obsah. U podepsaného PDF předem vysvětlit důsledky změny.

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
- **Vložit** = kontextový příkaz pro aktivní PDF. O přesném umístění v nabídce **Úpravy**, práci se stránkami a objekty, klávesové zkratce a zdroji rozhodnout při dalším ladění UI.
- **Soubor → Otevřít = VÝHRADNĚ PDF.** Nepřijímá dokumenty Word, Excel, PowerPoint ani nativní grafické formáty a **nespouští implicitní konverzi**. Pro jiný formát slouží Importovat nebo Sloučit a otevřít; výběr souboru má odpovídající filtr PDF a při nepodporovaném vstupu nabídnout jasné přesměrování na Importovat.
- Pravý panel **Vlastnosti** má pro každý z těchto úkonů vlastní pokročilá nastavení, ne duplicitní tlačítka celých nástrojů.
- Aktuální HTML v2.6 může ještě zobrazovat starší význam Importovat: před implementací je nutné **sjednotit s tímto novějším rozhodnutím**.

## 5. Otevřené otázky pro příští kolo

1. Vložený Word/Excel: zda vždy převádět na stránky PDF, nebo nabídnout i samostatnou přílohu v PDF (odlišná funkce).
2. Při vložení více stránek před/za aktuální stránku nebo na zadanou pozici; chování při konfliktních rozměrech stránky.
3. Nastavení Excelu: tiskové oblasti, listy, skryté listy, opakované řádky, měřítko; zachování původního formátování.
4. Ověřit reálnou podporu katalogu DOC/DOCX, XLS/XLSX, PPT/PPTX, ODT/ODS/ODP, SVG, PNG/JPEG/TIFF, PSD/PSB, AI/EPS, XCF, KRA, SLA a dalších; zvlášť chování velkých, chráněných, poškozených a vrstvených dokumentů.
5. Zachování písma, barevných ICC profilů, průhledností a vektorů podle možností vstupního formátu/enginu; bezpečné chování při nepodporovaných prvcích.
6. Případné předvolby konverze a šablony sestav pro opakované operace.

**Stav:** jde o požadavky a jejich rozdělení. Otevřít je pouze pro PDF. Podpora širokého katalogu zdrojů, včetně PPT/PPTX, PSD/AI a formátů open-source aplikací, je **cílový požadavek k technickému ověření**, nikoli již implementovaná funkcionalita.
