# PDF Blacksmith – HTML v2.9: Globální Nastavení a šířka stránky s rezervou pro ouška

**Datum:** 22. listopadu 2026. **Stav:** pracovní interaktivní demonstrace. Zdrojový kód Go/Win32 zůstává nezměněn.

## Oprava Přizpůsobit šířce

- Dosavadní režim v2.8 počítal volnou šířku `editor.clientWidth - 22px`, čímž mohl výsledek **překrývat svislá ouška** pomocných panelů, která se vykreslují nad plochou dokumentu.
- V2.9 měří **skutečnou polohu** hlavních i pomocných oušek vůči hranám editoru pomocí `getBoundingClientRect()`, vypočítá nezbytné bezpečné odsazení každé strany plus mezeru 9 px a aktualizuje CSS proměnné `--editor-safe-left/right`.
- `fit('width')` používá **aktuální vnitřní šířku editoru po odečtení obou bezpečných odsazení** a drobné rezervy pro vykreslování. Režim se přepočítá při sbalení/rozbalení panelů a při změně velikosti okna.
- Činnosti oušek, sdílená jediná pravá pomocná pozice, ikony, Historie a přepínání nástrojů zůstávají zachovány.

## Nově viditelné schválené globální Nastavení

Dřívější specifikace `specifikace_Codex_v5.0.md` obsahuje závazná rozhodnutí Q10, Q12, Q13 a Q14. V předchozích HTML návrzích **nebyla Nastavení zobrazena**, přestože byla již schválena. V2.9 tuto mezeru pouze v prototypu napravuje.

- **Umístění:** ikona ozubeného kolečka `nav_settings` v horní liště vedle Nápovědy.
- **Modalita:** samostatné **modální** okno se seznamem kategorií vlevo a obsahem vpravo; nelze současně editovat dokument.
- **Osm schválených kategorií:** Obecné, Rozhraní, Zobrazení PDF, Klávesové zkratky, Soubory, Text a OCR, Barvy a tisk, O aplikaci.
- **Potvrzování:** **OK** = použít a zavřít; **Použít** = potvrdit a nechat otevřené; **Zrušit / Esc / ×** = zrušit pouze nepotvrzené změny od posledního Použít a zavřít.
- Nastavení ve vizuálním HTML je **pouze in-memory ukázka**, nezapisuje konfiguraci ani nijak nezakládá automatické obnovovací soubory. V produkční verzi mají být viditelné pouze skutečně funkční volby (Q14).
- Kategorie Klávesové zkratky zobrazuje návrh přehledu zkratek; produktová možnost změny mapování a řešení konfliktů musí být implementována zvlášť. Kategorie O aplikaci zřetelně uvádí verzi **HTML prototypu** a nevymýšlí číslo sestavení desktopu.

## AutoRecover vs. Historie

Podle schváleného Q19 je AutoRecover **ochranou rozpracované práce při pádu**, nikoli běžným `Uložit`, které zapisuje cílový soubor.

- Výchozí AutoRecover **zapnut**.
- Po **60 s nečinnosti**.
- Při souvislé práci **nejpozději po 2 min**.
- **Dvě** poslední validní obnovovací verze.
- `%LOCALAPPDATA%\PDFBlacksmith\Recovery` odděleně od `%TEMP%`.
- Po pádu nabídnout obnovu; po vědomém „Neuložit“ obnovovací data daného dokumentu smazat. Nově uložená kopie má nahradit originál ve stejné kartě a přebrat obnovovací relaci.

**Doporučené rozhraní:** hlavní volby a časování AutoRecover v **Nastavení → Soubory**; v pomocném panelu **Historie** jen status (zapnuto, interval) a přímá cesta **Nastavení obnovy…**. `Úpravy → Historie` neobsahuje duplicitní nastavení. Obnovovací snímky **nejsou příkazy Ctrl+Z/Ctrl+Y** a nemají se bez rozlišení míchat s historií editací.

## Interakční ověření v Chromium

- Režim Přizpůsobit šířce s otevřenými panely při viewportu 1600×900: stránka 901,6 px, volná editorová šířka 977 px; vůči ouškům je vlevo mezera **14,7 px** a vpravo **16,7 px**.
- S oběma sbalenými skupinami šířka stránky **1506,4 px**; po obnově skupin se vrátí **901,6 px**. Režim dynamicky sleduje okolní panely.
- Globální Nastavení obsahuje **8 kategorií**; otestováno otevření z ozubeného kolečka, **Použít** bez zavření, **Zrušit/Esc** s návratem k poslednímu potvrzenému stavu a **OK** s potvrzením/zavřením.
- Stav AutoRecover z Nastavení → Soubory se zobrazuje v Historii. Prototyp nic skutečně neukládá.
- Zaznamenáno **0 chyb JavaScriptu**.

**Při budoucí implementaci:** použít reálný systém preferencí, atomic write konfigurace, verzi build metadata a globální klávesové mapování; AutoRecover navázat na existující infrastrukturu bezpečného ukládání a crash recovery. Bez schválení nepřidávat prázdné funkční kategorie a nespouštět souběžnou editaci dokumentu za modálním Nastavením.
