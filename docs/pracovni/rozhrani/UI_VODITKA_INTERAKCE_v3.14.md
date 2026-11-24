# PDF Blacksmith – interaktivní vodítka, magnetické zarovnání a měření (HTML v3.14)

**Datum:** 24. listopadu 2026 · **Stav:** odsouhlasená interakční demonstrace UI a návrhová specifikace. **Nejde o implementaci SnapEngine do Go/Win32/PDFium.**

**Aktuální HTML:** [cely_vizual_v3.14.html](cely_vizual_v3.14.html)  
**Základ:** [UI v3.13](UI_POMUCKY_ZKRATKY_v3.13.md), [v3.11](UI_SEPARACE_PANELY_VODITKA_v3.11.md), [konsolidace v3.9](AKTUALNI_UI_CODEX_v3.9.md), [hlavní specifikace](../specifikace_Codex_v5.0.md), [mapa ikon v4.8](../mapa_ikon_v4.8.md), [mapa kláves v1.1](../mapa_klavesovych_zkratek_v1.1.md).

## 1. Ovládání vodítek

- **Pravítka mimo PDF:** vodorovné nad stránkou a svislé vlevo, se značkami po 1 mm, výraznějšími po 5 mm a číselnými popisky po 25 mm. Jde o HTML ilustraci při 96 CSS px/palec, ne o fyzickou kalibraci monitoru.
- **Tažení z vodorovného pravítka** založí **svislé vodítko** (X). **Tažení ze svislého pravítka** založí **vodorovné vodítko** (Y). Během tažení zobrazit živou polohu s jedním desetinným místem v mm. Vyvolané vodítko nesmí současně spustit obdélníkový, textový ani objektový výběr.
- **Přesunutí existujícího vodítka:** přetažení příslušné plné čáry; po uvolnění uložit polohu do pracovního modelu pro danou stránku. **Esc** rozpracované tažení bez změny zruší.
- **Výběr a odstranění:** výběr vodítka a **Delete** jen v kontextu \`Zobrazení → Pomůcky rozvržení\` a mimo vstupní pole; rovněž tlačítko \`×\` v jeho řádku. K dispozici jsou \`+ Svislé\`, \`+ Vodorovné\` a \`Odstranit všechna\`.
- **Zamykání:** jednotlivé vodítko lze zamknout/odemknout tlačítkem v řádku; \`Uzamknout vodítka\` funguje jako globální zákaz pohybu, změn souřadnic i vytváření/mazání. Při kliknutí na uzamčenou čáru nevyvolat žádnou jinou akci plátna.
- **Numerický vstup:** seznam uvádí osu, identifikátor a přesnou pozici v mm. Dále zachovat jednoduché vstupy prvního vodorovného a svislého vodítka pro návaznost na v3.13. Přednost má skutečný model \`pageGuides\`; pole nejsou druhým nezávislým úložištěm.
- **Stav po stránkách:** každá stránka má vlastní nezávislý seznam a ID vodítek (\`guideByPage\`). V souvislém zobrazení i při přepnutí aktivní stránky se má používat správná soustava dané stránky. Demo stav je pouze v paměti běžícího HTML; neukládá se do PDF ani po reloadu.

## 2. Vzhled a přesná geometrie

- **Pevná vodítka** jsou **plné čáry** v uživatelem nastavené barvě a krytí, protahují se přes okraje listu až k pravítkům. Zvolená čára má jemné zvýraznění.
- **Při vytahování/přesouvání** se zobrazuje pomocný čárkovaný náhled pohybu a souřadnice; po ukončení zůstane jen plná pevná linka.
- **Dočasné inteligentní přichytávání** používá oddělené vínové čárkované nebo tečkované čáry a bodový marker. Kandidáty tvoří hrany, středy, účaří textu, aktivní pevná vodítka a hranice stránky podle \`viewAids.targets\`. Blízkost se vyhodnocuje podle \`snapTolerancePx\`.
- Přichycení ke **hraně** smí zobrazit linku protaženou až k pravítkům. Přichycení k **bodu** má končit v cílovém bodě. Následné dopracování režimu obsahuje hysterézi proti přeskakování kandidátů, prioritu vrstev, přesnější baseline/rotace/uzly a vizualizaci rovnoměrných rozestupů.
- **Měření mezi objekty:** při demonstraci virtuálního tažení ukázkového textu nebo obrázku se zobrazuje mezera v mm mezi jejich hranicemi. Podporovány horizontální i vertikální mezery. **Textové označení hodnoty je ilustrativní:** HTML pracuje s CSS geometrií, nikoli s reálným PDF bounding-boxem.
- Vypnutí magnetu musí schovat dočasné linky a ghost objekty, **nesmí odstranit ani vypnout pevná vodítka**. Vypnutí viditelnosti vodítek nesmí zrušit jejich souřadnice ani preference \`targets.guides\`.

## 3. Implementační pravidla pro Codex

- Jeden společný modul \`SnapEngine\` a společná transformace \`PDF page points → device pixels\`, přihlédnout k PDF \`UserUnit\`, \`MediaBox/CropBox\`, otočení, zoomu, DPI a dokumentovým spreadům.
- Oddělit **editorové pracovní pomůcky**, **hit-test**, **snap candidate selection** a **potvrzenou transformaci PDF obsahu**. Samotné vodítko, mřížka, náhledy a přepínače **nesmějí měnit obsah PDF, příznak dirty, digitální podpis, tisk ani export**. Potvrzený přesun objektu je teprve obsahová operace a jeden atomický krok historie.
- Vlastní per-document/per-page metadata pro vodítka ukládat mimo PDF nebo jen do explicitně zvoleného pracovního formátu. Volitelný export nastavení musí být oddělen od exportu výrobního PDF. V prototypu žádná persistence.
- **Žádné nové výběrové handlery:** tažení pravítek a čar má vyšší prioritu v hit-testu než výběr textu/objektu, ale nemění už schválené Q1, SEL/TXT/OBJ.
- Klávesa \`Delete\` je v této etapě **jen v kontextu vybraného vodítka při aktivních Pomůckách rozvržení** a mimo textové vstupy. \`Esc\` ruší rozpracované tažení; již existující zkratky z [mapy v1.1](../mapa_klavesovych_zkratek_v1.1.md) se nemění.
- Doplnit návrh uživatelských stavů: práce v nesouvislém zoomu, uzamčené stránky, různé orientace, více vybraných objektů, přesahující vodítka, výběr mimo stránku a obnovení preference panelů.

## 4. Akceptační testy HTML v3.14

1. 980, 1100, 1250, 1600 px: obě pravítka a obě orientace vodítek, bez JavaScriptových chyb a bez překryvu pravých panelů.
2. Vytáhnout vodítko z horního i levého pravítka; po uvolnění se objeví nový řádek se správnou osou a pozicí.
3. Přetáhnout čáru, uzamknout ji, ověřit zákaz přesunu, odemknout a odstranit ji tlačítkem nebo \`Delete\`.
4. Při globálním zámku nevytvořit nové vodítko; při \`Esc\` se rozpracované tažení nezapíše do modelu.
5. Vodítka jedné stránky nezměnit při přepnutí na jinou stránku.
6. Nastavení číselného pole aktualizuje čáru na plátně a nevyvolá PDF editaci.
7. Při demonstrativním tažení textového/obrazového vzorku ukázat odpovídající vzdálenost, snap marker a dočasné linky; po uvolnění/úniku zrušit bez změny PDF.
8. Regresně prověřit zkratky, univerzální výběr, OCR, Separace, dvoustavové nástroje, historii, pomocný panel Výsledky a všechna schválená starší pravidla.

**Testovací skutečnost:** interakce na HTML byly ověřeny v Chromiu na uvedených šířkách. Neproběhla produkční implementace ve Windows ani test geometrie skutečného PDF.

## 5. Další doporučená etapa

Nejprve definovat datové typy \`Anchor\`, \`SnapCandidate\`, \`Guide\`, \`SnapResult\` a chování priority/hysteréze. Poté zvlášť schválit implementaci bezkolizního výběrového dispatchera a skutečné transformace objektů v Go. Teprve následně testovat vybrané reálné PDF dokumenty s textem, fonty, obrázky, otočenými stránkami a smíšeným obsahem.
