# PDF Blacksmith — rozlišení Vložit / Importovat / Sloučit a otevřít

**Datum rozhodnutí:** 22. listopadu 2026  
**Stav:** schválený směr a význam příkazů; detaily převodního modulu, podporovaných formátů a volby stránkování dořešit před implementací. **Zatím neměnit HTML ani zdrojový kód.**

## 1. Vložit

**Cíl:** přidat obsah do *již otevřeného* PDF.

- Word/Excel nebo jiný vícestránkový dokument lze vložit **jako stránky**. Dokument se nejprve bezpečně převede na PDF a pak se vloží na zvolené místo v aktivním PDF.
- Obrázek lze vložit **přímo do aktuální stránky jako grafický objekt**, s možností nastavit polohu, velikost a další podporované vlastnosti, bez svévolného rasterizování zbytku stránky.
- Rozlišit v dialogu `Vložit stránky ze souboru` a `Vložit obrázek do stránky`; přesnou nabídku a její umístění v menu navrhnout následně.
- Zásahy do dokumentu jsou vratné, respektují historii Ctrl+Z/Y a AutoRecover, nesmějí nechtěně poškodit formulářová pole, poznámky, odkazy, podpisy nebo další obsah. U podepsaného PDF předem vysvětlit důsledky změny.

## 2. Importovat

**Cíl:** převést **jeden zdrojový dokument** jiného formátu do samostatného PDF, nikoli jej automaticky vkládat do otevřeného dokumentu.

- Kandidátní zdroje: DOCX, XLSX, obrázky (PNG/JPEG/TIFF) a další formáty podle reálně ověřené podpory převodní knihovny.
- U Wordu: volby pro zachování stránkování, velikosti stránky a kvalitní převod obrazů/písem v rámci možností dostupného enginu.
- U Excelu: výběr listů, tiskových oblastí, orientace, papíru a vhodného měřítka; před konverzí přehled stran.
- Výsledné PDF se otevře **v nové záložce**. Původní soubor se nikdy automaticky nepřepíše.
- Konkrétní engine (LibreOffice/Office či jiný) a jeho licence, dostupnost a kvalita jsou **technické rozhodnutí k ověření**, nikoli již hotová funkce. Makra ani nedůvěryhodný aktivní obsah nikdy automaticky nespouštět.

## 3. Sloučit a otevřít

**Cíl:** vytvořit **nové společné PDF** ze **dvou a více zdrojových souborů**, přičemž zdroje mohou mít různé formáty.

- Sestava zdrojů: PDF, Word/Excel a obrázky či další podporované typy, podle ověřené převodní podpory.
- Seznam s viditelným pořadím; možnost přetažením změnit pořadí, odstranit položku, přidat další soubor, upravit rozsah stránek (u PDF) a nastavení konverze (u ostatních).
- Před převodem ukázat finální pořadí a podle možností náhled výsledné sestavy.
- Ne-PDF zdroje se nejprve převedou na přechodné PDF; spojení proběhne ve zvoleném pořadí; výsledný soubor se **otevře v nové záložce**.
- Žádný zdroj se při převodu/sloučení nemění. Při dílčím selhání převodu nezamlčet chybu, nabídnout opravný krok a nedokončit výstup tiše se ztracenými stranami.
- Zachovat původní existující funkce slučování PDF (Q31) a teprve potom rozšířit jejich vstupy o další typy.

## 4. Uživatelské rozhraní – návrh k dalšímu projití

- **Soubor → Importovat** = konverze externího zdroje do PDF, s pokročilými volbami v pravém panelu.
- **Soubor → Sloučit a otevřít** = sestava 2+ vstupů, požadované pořadí, konverze, společné PDF v nové záložce.
- **Vložit** = kontextový příkaz pro aktivní PDF. O přesném umístění v nabídce **Úpravy**, práci se stránkami a objekty, klávesové zkratce a zdroji rozhodnout při dalším ladění UI.
- **Soubor → Otevřít** = otevřít existující PDF jako samostatný dokument (bez automatické konverze různých formátů).
- Pravý panel **Vlastnosti** má pro každý z těchto úkonů vlastní pokročilá nastavení, ne duplicitní tlačítka celých nástrojů.
- Aktuální HTML v2.6 může ještě zobrazovat starší význam Importovat: před implementací je nutné **sjednotit s tímto novějším rozhodnutím**.

## 5. Otevřené otázky pro příští kolo

1. Vložený Word/Excel: zda vždy převádět na stránky PDF, nebo nabídnout i samostatnou přílohu v PDF (odlišná funkce).
2. Při vložení více stránek před/za aktuální stránku nebo na zadanou pozici; chování při konfliktních rozměrech stránky.
3. Nastavení Excelu: tiskové oblasti, listy, skryté listy, opakované řádky, měřítko; zachování původního formátování.
4. Zpracování podpory DOC/DOCX, XLS/XLSX, ODT/ODS, SVG, PNG/JPEG/TIFF a výjimek, zvláště velkých a chráněných dokumentů.
5. Zachování písma, barevných ICC profilů, průhledností a vektorů podle možností vstupního formátu/enginu; bezpečné chování při nepodporovaných prvcích.
6. Případné předvolby konverze a šablony sestav pro opakované operace.

**Stav:** jde o požadavky a jejich rozdělení. Samotná konverze Word/Excel, vložení obrázků ani vícerozměrné slučování zatím nejsou tímto dokumentem prohlášeny za implementované.
