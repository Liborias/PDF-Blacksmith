# Plán vývoje PDF Blacksmith

Export má nejvyšší prioritu. Každý další krok začíná po odsouhlasení.

1. **Výběr stran při exportu.** Nabídnout všechny strany, aktuální stranu a jednotlivé strany či rozsahy zadané například jako `1, 3–5, 8`. Výběr se řídí aktuálním pořadím stran v aplikaci a platí i pro převod na desku K nebo do profilu CMYK. [Hotovo ve verzi 0.4.8](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.8)
2. **Předvolby PDF podle účelu.** Přidat PDF pro tisk, PDF připravené pro tiskárnu, digitální PDF s menším souborem, digitální PDF s vysokou kvalitou, PDF pro další zpracování a zploštěné PDF. V nabídce uvést, co daná volba mění u barev, obrázků a průhlednosti. [Hotovo ve verzi 0.4.9](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.9)
3. **Normy PDF/X.** Přidat předvolby PDF/X-1a:2003, PDF/X-3:2003 a PDF/X-4 s vestavěnou kontrolou výstupního záměru ICC, písem, tiskových rámečků a průhlednosti. Úplné ověření shody podle ISO vyžaduje nezávislý preflight. [Export a vestavěná kontrola ve verzi 0.4.11](https://github.com/Liborias/PDF-Blacksmith/releases/tag/v0.4.11)
4. **Rozdělení PDF po stránkách.** Uložit každou vybranou stránku do vlastního PDF se jménem podle čísla strany.
5. **Přesun stránek.** Přemístit jednu nebo více vybraných stran přímo na určenou pozici v seznamu stránek a umožnit změnu vrátit.
6. **Sloučení dokumentů.** Sestavit jedno PDF z více souborů v zadaném pořadí a umožnit výběr stran z každého souboru.
7. **Záložky v PDF.** Zobrazit existující záložky v postranním panelu, přejít na jejich cíl a přidávat, přejmenovávat nebo upravovat vlastní záložky.
8. **Fulltextové vyhledávání.** Prohledat textovou vrstvu dokumentu, zobrazit výsledky podle stran a zvýraznit shody v náhledu.
9. **OCR.** Rozpoznat text na vybraných naskenovaných stranách, včetně češtiny, a přidat vyhledávatelnou textovou vrstvu při zachování vzhledu skenu.
10. **Více otevřených dokumentů.** Otevřít PDF ve vlastních kartách, kopírovat mezi nimi, při zavírání nabídnout uložení změn a umožnit zobrazení dvou dokumentů vedle sebe.
11. **Úpravy obsahu.** Vybrat, přesouvat, kopírovat a mazat objekty. U textu měnit písmo, velikost a řez i podtržení.
12. **Poznámky a kreslení.** Přidat poznámky, elipsy, obdélníky, volné čáry a šipky s nastavitelnou barvou a tloušťkou čáry.

## Levý panel nástrojů

Postupně sem přibudou **Nástroje výběru**, **Úprava textu**, **Poznámky**, **OCR**, **Sloučit PDF**, **Rozdělit PDF** a **Pořadí stránek**. Každá položka dostane piktogram a otevře své nastavení v pravém panelu.
