# PDF Blacksmith — automatické otevření příslušného pomocného panelu

**Zaznamenáno:** 22. listopadu 2026. **Stav:** uživatelský požadavek uložen k pozdější úpravě HTML i aplikace. **Nyní nic neimplementovat ani neupravovat v HTML v3.0.**

## Závazné chování

Při **aktivaci nástroje**, k němuž existuje související pomocný panel, se musí **zároveň**:
1. Aktivovat vybraný nástroj a zobrazit jeho hlavní pravý panel **Vlastnosti/nastavení nástroje**.
2. Otevřít odpovídající **pravý pomocný panel** v jediné společné pozici vyhrazené pomocným panelům.
3. Pokud je v této pozici zobrazen jiný pomocný panel, přepnout ho na právě související panel; pomocné panely se nikdy nesmějí řadit vedle sebe.

Příklady výslovně uvedené uživatelem:
- **Najít v PDF / vyhledávání → Výsledky vyhledávání** (včetně aktivace přes horní vyhledávací pole).
- **Úpravy → Historie úprav**.

Ostatní nástroje dostanou toto propojení **jen tehdy**, je-li pro ně pomocný panel funkčně definován. Nástroj **Text** má zatím jen svůj hlavní panel **Vlastnosti**; nezavádět bez požadavku prázdný pomocný panel.

## Interakční zásady k implementačnímu upřesnění

- Respektovat stávající koncepci **jediné sdílené pozice pravých pomocných panelů**, nikoli mnoho paralelních sloupců.
- Zachovat samostatné ouško pomocného panelu umožňující ruční sbalení a otevření. Při **nové aktivaci** příslušného nástroje se panel znovu automaticky nabídne.
- Sbalení celé pravé skupiny společným horním ouškem **nezničí zapamatovaný stav panelů**; při návratu se má obnovit předchozí uspořádání. Rozlišit vědomé sbalení celé skupiny od aktivace nového nástroje.
- Hlavní panel Vlastnosti nesmí být nahrazován pomocným panelem; oba mají své oddělené místo.
- Panel Historie používá skutečnou historii editací, nikoli snímky AutoRecover; stav a nastavení AutoRecover patří primárně do **Nastavení → Soubory**, viz `POZNAMKY_NAVRHU_v2.9.md`.

## Kontext

Toto pravidlo bylo řečeno po vytvoření **HTML v3.0** s rozšířenou typografií. Uživatel požádal **zatím pouze uložit** a následně prodiskutovat chování Textu. Proto záměrně **neprovádět žádné změny HTML** a neměnit zdrojový Go/Win32 kód.
