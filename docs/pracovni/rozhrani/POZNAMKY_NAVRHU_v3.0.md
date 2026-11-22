# PDF Blacksmith — návrh panelu Text v3.0

**Datum:** 22. 11. 2026. **Typ změny:** HTML vzhled a demo interakcí; žádné změny zdrojového Go/Win32/PDFium.

## Problémy opravené proti v2.9
- **Nesprávná nápověda ikon**: nahradit generické popisy konkrétními, věcně odpovídajícími vysvětlivkami pro každou volbu. Popis viditelný jak ve standardním `title`, tak v `aria-label`, a dynamicky také pod skupinou ikon při hover/fokusu.
- **Řádkování**: použít původní `line_spacing.svg` z oficiální sady místo obecného symbolu odstavce; v pravém panelu nastavení jednoduché, 1,15, 1,25, 1,5, dvojité.
- **Textový kurzor**: nový vlastní SVG symbol `text_caret` (I-beam), který se neplete s vytvořením textového bloku (`text_box`). Doplněna též volba obdélníkového a volného výběru oblasti.
- **A+ / A−**: použít schválené ikony `font_size_up` a `font_size_down` s posunem mezi běžnými typografickými velikostmi (např. 12→14→16 pt); živé zobrazení čísla. Nápověda při najetí.
- **Typografie**: rodina/velikost, běžný řez, tučné, kurzíva, podtržení, přeškrtnutí, kapitálky, horní/dolní index, šířka znaků, proklad písmen a mezery mezi slovy.
- **Barvy**: výběr barvy písma; samostatná barva a přepínač zvýraznění za samotnými znaky; kapátko a obrys jen náznak budoucí funkce.
- **Odstavec**: zarovnání, řádkování, mezera před a za odstavcem v pt, odsazení zleva/zprava a prvního řádku v mm, zmenšení/zvětšení odsazení ikonami.
- **Ukázkový panel sazby**: poskytuje okamžitou vizuální odezvu na volby bez skutečné editace PDF.

## Rozsah ikon
Originální ikony převzaty přímo z archivu `PDF_Blacksmith_ikony_nastroju_SVG_v4.1.zip`, zejména `line_spacing`, `font_size_up/down`, `text_color`, `highlight`, `paragraph_spacing`, `indent_increase/decrease`, `text_box` a `superscript/subscript`. Nová samostatná ikona `text_caret.svg` je v `rozhrani/ikony/`. Mapa aktualizována ve `mapa_ikon_v4.4.md`.

## Kontroly
HTML otevřeno v Chromium, ověřena přítomnost všech ikon použitých v panelu Text bez chybějících SVG symbolů, správná nápověda pro A+ a Textový kurzor, velikosti 12→14→12, interakce řezů, horního/dolního indexu, řádkování 1,5, odsazení po 5 mm, zapnutí zvýraznění, různé textové režimy a zachování textových voleb při přepínání nástrojů. Zkontrolováno také Soubor → Vložit strany, Objekty → Duplikovat, Zobrazení → Přizpůsobit šířce a menší okno 1180×800. **0 chyb JavaScriptu.**

**Implementační upozornění:** toto je pouze model UI. Kreslení a převod původního textu PDF, integrace fontů a skutečná typografická měřítka vyžadují samostatné ověření. Neztratit původní ovládání Text/Objekty a zachovat bezpečné ukládání i historii.
