# Plán vývoje PDF Blacksmith

Export má nejvyšší prioritu. Každý další krok začíná po odsouhlasení.

1. **Výběr stran při exportu.** Nabídnout všechny strany, aktuální stranu a jednotlivé strany či rozsahy zadané například jako `1, 3–5, 8`. Výběr se řídí aktuálním pořadím stran v aplikaci a platí i pro převod na desku K nebo do profilu CMYK. [Hotovo ve verzi 0.4.8](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.8)
2. **Předvolby PDF podle účelu.** Přidat PDF pro tisk, PDF připravené pro tiskárnu, digitální PDF s menším souborem, digitální PDF s vysokou kvalitou, PDF pro další zpracování a zploštěné PDF. V nabídce uvést, co daná volba mění u barev, obrázků a průhlednosti. [Hotovo ve verzi 0.4.9](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.9)
3. **Normy PDF/X.** Přidat předvolby PDF/X-1a:2003, PDF/X-3:2003 a PDF/X-4 s vestavěnou kontrolou výstupního záměru ICC, písem, tiskových rámečků a průhlednosti. Úplné ověření shody podle ISO vyžaduje nezávislý preflight. [Export a vestavěná kontrola ve verzi 0.4.11](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.11)
4. **Rozdělení PDF po stránkách.** Uložit každou vybranou stránku do vlastního PDF se jménem podle čísla strany. [Hotovo ve verzi 0.4.12](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.12)
5. **Pořadí a náhledy stránek.** Přesunout jednu či více stran na zadanou pozici nebo přetažením náhledů. Vybrané strany lze také smazat. Přesun i smazání lze vrátit. [Hotovo ve verzi 0.4.13](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.13)
6. **Sloučení dokumentů.** Sestavit jedno PDF z více souborů v zadaném pořadí a umožnit výběr stran z každého souboru. [Hotovo ve verzi 0.4.15](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.15)
7. **Záložky v PDF.** Zobrazit strom záložek v postranním panelu, přejít na cílovou stranu, přidat záložku a upravit její název nebo cíl. [Hotovo ve verzi 0.4.16](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.16)
8. **Fulltextové vyhledávání.** Prohledat textovou vrstvu dokumentu, zobrazit výsledky podle stran a zvýraznit shody v náhledu. [Hotovo ve verzi 0.4.19](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.19)
9. **OCR.** Rozpoznat skeny i text převedený na křivky na vybraných stranách či rozsazích. Přidat českou nebo anglickou textovou vrstvu při zachování původní grafiky. [Hotovo ve verzi 0.4.19](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.19)
10. **Více otevřených dokumentů.** Otevřít PDF ve vlastních kartách a při zavírání nabídnout uložení změn. OCR se používá přímo v otevřené kartě a ukládá se nástrojem Soubor. [Hotovo ve verzi 0.4.20](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.20)
11. **Úprava textu.** Upravit vybrané znaky jako blok, rozpoznat řádky a odstavce a měnit písmo, řez, zarovnání, řádkování a proklad. Nabídnout výběr obdélníkem a lasem i OCR oblasti bez textové vrstvy. Úpravy v rámečcích přímo na stránce, automatické rozměry, kontextová nabídka a zkratky. [Pracovní verze 0.4.24](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.24)
12. **Výběr stran z náhledů.** Synchronizovat výběr s nástroji, které pracují s rozsahem nebo jednotlivými stranami. [Hotovo ve verzi 0.4.21](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.21)
13. **Dva dokumenty vedle sebe.** Každý dokument má vlastní sbalitelný panel náhledů, výběr stránek a zoom. Bez otevřeného PDF je výchozím nástrojem Soubor. [Hotovo ve verzi 0.4.22](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.22)
14. **Přetahování stránek mezi PDF.** Přetáhnout jeden náhled nebo více označených náhledů do druhého panelu a vložit kopie stránek na vybrané místo v jejich pořadí. [Pracovní verze 0.4.25](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.25)
15. **Barvy při přenosu stránek.** Při vložení tažením se zeptat, zda mají vložené stránky zachovat svůj barevný prostor, nebo se mají převést do barevného prostoru cílového PDF. Stránky ve stupních šedi zůstanou šedé. Když převod může vést k nevratné změně nebo ztrátě vlastností, upozornit na konkrétní dopad před potvrzením.
16. **Úpravy objektů.** Vybrat, přesouvat, kopírovat a mazat objekty v dokumentu.
17. **Laso pro vektorové objekty.** Vybrat společně více vektorových objektů. Při přesunu nebo kopírování nabídnout zachování jednotlivých objektů nebo jejich seskupení do jednoho celku.
18. **Přetahování objektů a textu.** Umožnit přenos mezi panely otevřených PDF i přetažení do jiných aplikací a složek. Pro externí přetažení využít export na pozadí ve stávajícím formátu a barevném profilu objektu. U vektorizovaného textu po OCR předat mimo aplikaci při kopírování nebo tažení neformátovaný text. Mezi PDF v aplikaci při kopírování nebo tažení přenést vektorovou skupinu společně s rozpoznaným textem.
19. **Poznámky a kreslení.** Přidat poznámky, elipsy, obdélníky, volné čáry a šipky s nastavitelnou barvou a tloušťkou čáry.

## Levý panel nástrojů

**Sloučit PDF** je v levém panelu nástrojů dostupné i bez otevřeného dokumentu. Po načtení dokumentu jsou dostupné také **Text**, **OCR**, **Rozdělit PDF** a **Pořadí stránek**. Společné náhledy mají vlastní sbalitelný panel mezi nástroji a dokumentem. Postupně přibudou **Nástroje výběru** a **Poznámky**.
