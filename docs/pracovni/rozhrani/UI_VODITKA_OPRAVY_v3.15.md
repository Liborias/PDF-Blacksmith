# PDF Blacksmith — opravy vodítek a kontextové nabídky, UI v3.15

**Datum:** 24. listopadu 2026  
**Stav:** schválený návrh ovládání, interaktivní HTML prototyp. **Žádná implementace do Go/Win32/PDFium zatím nebyla provedena.**  
**Aktuální HTML:** [cely_vizual_v3.15.html](cely_vizual_v3.15.html)  
**Navazuje:** [v3.14](UI_VODITKA_INTERAKCE_v3.14.md), [v3.13](UI_POMUCKY_ZKRATKY_v3.13.md), [hlavní specifikace Codex v5.0](../specifikace_Codex_v5.0.md).

## 1. Závazné opravy oproti v3.14

1. **Z vodorovného (horního) pravítka vytáhnout vodorovné vodítko** — souřadnice Y, manipulace převážně ve směru nahoru/dolů. **Ze svislého (levého) pravítka vytáhnout svislé vodítko** — souřadnice X, manipulace převážně vlevo/vpravo. Toto uživatelské pravidlo **nahrazuje opačnou orientaci ve v3.14**.
2. Pouhé kliknutí do pravítka **vodítko nevytváří**. Musí proběhnout skutečné tažení nad práh 7 obrazových pixelů a přechod ze stupnice do stránky; při uvolnění na pravítku nevznikne nic. `Esc` či `pointercancel` vrací stav bez nového vodítka. V prototypu má explicitní klávesový příkaz `Enter/mezerník` při fokusu na pravítku nadále vlastní přístupnou alternativu, **nikoliv prostý klik myší**.
3. **Všechny viditelné pevné linky jsou ze stejného datového modelu vodítek.** Staré dekorativní/ukázkové vrstvy nesmějí kreslit další nevybratelná vodítka. Výchozí dvě ukázkové linky mají regulérní ID, souřadnice, uzamčení a vlastní barvu, takže je lze přesouvat, upravovat i mazat.
4. **Správa vodítka na stránce:** při najetí se zobrazí malá kontextová nabídka **Vybrat / Barva / Zamknout–Odemknout / Smazat**. V pravém panelu zachovat číselné X/Y, jednotlivá tlačítka Zamknout/Smazat i hromadné vymazání neuzamčených vodítek. Zamčené vodítko nelze přesouvat ani mazat. Pravítka mají přednost v hit-testu tam, kde k nim vodítka sahají.
5. **Barvy odděleně:**
   - `viewAids.guideColor` je **výchozí barva nových vodítek**, nikoli barva všech již existujících.
   - Každé vodítko ukládá vlastní `guide.color`. Barvu lze změnit pro jednotlivé vodítko přes příslušné barevné políčko v seznamu či nad čarou.
   - Vícenásobný výběr (checkboxy v seznamu, `Ctrl+klik` na vodítka) umožňuje **Použít na označená**, bez přebarvení ostatních. Změna nové výchozí barvy již existující linky nepřebarví.
   - Krytí může zůstat společnou výchozí vlastností, budoucí individuální krytí je samostatná rozšiřující vlastnost.
6. **Vodítka se sama přichytávají**, pokud je magnet zapnutý. Kandidáty: jiná vodítka stejného směru, mřížka, hrany a středy objektů, účaří textu a okraje stránky, podle `viewAids.targets` a tolerance v obrazových pixelech. Během manipulace dočasně použít vínově červené zvýraznění snapu; uložení do pozice až po uvolnění myši.
7. **Více stránek:** zachovat per-page seznam. Při změně aktivní stránky nevymazat model ostatních a nepovažovat staré ID za aktivní výběr nové stránky. Vodítka jsou pouze pomocné pracovní překryvy, netisknou se a nejsou PDF anotace.

## 2. Vysvětlení konfliktů opravených v HTML

- Přímá tvorba po kliknutí pocházela z toho, že se původně zpracovával každý `pointerup` po `pointerdown`. V3.15 používá `activated` + minimální délku tažení + přechod do oblasti stránky.
- Překrytí svislého pravítka vodorovnou linkou bylo způsobeno pořadím vrstev: pravítka nyní přebírají události i na křížení, takže lze vytahovat další vodítka.
- Dekorativní staré `.pdf-fixed-guide` je trvale skryto; vodítka s `data-guide-id` mají jednotný model a události.
- Pro mazání se používá jeden kontrolovaný `removePageGuide` podle ID a stránky, namísto několika nesouvisejících větví. Všechny editace respektují individuální a globální zámek.
- Změny barvy se nepropisují do celého CSS modelu stránky, ale ukládají se ke konkrétnímu vodítku a vykreslují přes `--guide-color` na dané lince.

## 3. Klávesy a ochrana proti kolizím

- `Delete` maže označené vodítko pouze v režimu **Zobrazení → Pomůcky rozvržení**, při neuzamčeném vodítku a mimo formulářové/textové vstupní pole; jinak zachovat existující význam.
- `Esc` zruší aktivní tažení, pokud existuje; v jiném kontextu se nechová jako nový globální příkaz.
- `Ctrl+klik` na vodítko rozšiřuje/zužuje jeho výběr, **nenahrazuje** schválené Ctrl-výběry textu či objektů.
- Navazuje na [mapu kláves v1.2](../mapa_klavesovych_zkratek_v1.2.md). Žádná nová globální zkratka zatím nebyla implementována.

## 4. Ověření a omezení

Ověřeno v Chromiu při šířkách **980, 1100, 1250, 1600 px**: kliknutí do pravítka nevytváří vodítko; tažení z horního vytváří vodorovnou a z levého svislou linku; hover nabídka se otevře; JavaScriptové chyby nebyly zjištěny. Další testy při 1350 px ověřily změnu barvy jednotlivých i více označených vodítek, výchozí barvu nových, odstranění tlačítkem i z hover nabídky, zamykání a přichycení jednoho vodítka k druhému.

**Výslovná omezení:** SnapEngine pro skutečné PDF objekty a transformace stále není implementován. Prototyp používá CSS geometrii ukázkových objektů. Kompletní automatizované testy výrobní aplikace Go/Win32 ani skutečné PDF neproběhly.

## 5. Odložená vylepšení do dalších etap

- **Symetrické zrcadlení vzhledem k vybranému bodu/ose nebo vodítku**, možnost vytvářet sdružená vodítka a provádět transformace více objektů: teprve po schválení podrobného návrhu se souřadnicemi a historií.
- **Jazyk rozhraní:** v příštím kroku dopracovat **Nastavení → Obecné → Jazyk**. Starší HTML už obsahuje ukázkovou volbu Čeština/English, ale bez skutečné lokalizace všech popisků. Doplnit jednoznačné chování přepnutí, jednotný registr překladových klíčů, výchozí jazyk, ukládání předvolby a případné restartování aplikace. V3.15 jazykový systém nemění.
- Před implementací Go nejdříve inventura PDF souřadnic, společné snap kandidáty/hysteréze, kolize výběrových gest, per-page metadata a testy undo/exportu. **Neimplementovat bez zvláštního schválení.**
