# PDF Blacksmith – rozhraní se dvěma horními lištami (v2.1)
**Stav: pracovní vizuální návrh k revizi, nikoli schválený implementační předpis.**

## Zachované principy
- Barevné schéma vychází z původního HTML prototypu Formuláře F1: tmavé neutrální panely a jemná vínová akcentace výběru.
- Dvě horní řady: 1) záhlaví s **původním SVG kovadliny** vloženým přímo v HTML (beze změny geometrie cesty; upravena jen hranice viewBox a barva), záložkami dokumentů; 2) menu a kompaktní ovládání výběru, zoomu, 1:1 a navigace mezi stránkami.
- **Hranaté panely**, tenké oddělovací linie, malé ikony z již dříve připravené SVG sady.
- **Menší ouška panelů**, mírně zaoblené výhradně směrem od panelu. Levé: Nástroje, Náhledy, sbalení levé strany; pravé: Vlastnosti, Záložky, Poznámky.
- Tři ilustrační stránky, různé stavy nástroje Formuláře, demonstrace panelů a validace.
- V tomto HTML **není** implementováno čtení/ukládání PDF, podpis, tisk ani skutečný OCR či Preflight.

## Interaktivní kontrola
Otevřít `cely_vizual_dve_listy_v2.1.html` přímo v prohlížeči:
- přepínání stran náhledy a horní/dolní navigací;
- přiblížení: +/- a 1:1 (100 %), přizpůsobení šířce/stránce;
- sbalit/rozbalit panely kliknutím na malá ouška;
- aktivní nástroj Text/Formuláře/Separace a pravý panel;
- zvýraznění polí, editace ukázkového formuláře, kontrola chyb a dialog Uložit přesto;
- ouška Záložky a Poznámky zobrazují zástupný obsah.

**Důležité:** ilustrační HTML slouží pouze k revizi vzhledu a interakcí. Codex má navázat na existující Go/Win32 aplikaci a její funkce, ne nahrazovat ji webovou aplikací. Přesnost ikon, textů, vertikálního zarovnání a implementace oušek se musí schválit zvlášť.
