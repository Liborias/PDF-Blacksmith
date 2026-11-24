# PDF Blacksmith – mapa barev a pravidla motivů v1.2

**Aktualizace:** 24. listopadu 2026 · **Referenční rozhraní:** `cely_vizual_v3.21.html`  
**Stav:** schválený návrh HTML a pravidla budoucí implementace, ne automatická změna Go/PDFium.

## Barevná schémata

Všechny palety jsou sestaveny jako **kompletní sady** pro tmavý i světlý režim. Výchozí zůstává šedo-vínová. V dalších pěti paletách **není vínová základní ani doplňková barva**.

| Schéma | ID | Tmavý akcent | Světlý akcent | Světlý panel | Tmavý panel |
|---|---|---|---|---|---|
| Šedo-vínová | `wine` | `#E0A1BC` | `#703950` | `#FFFCFE` | `#26242D` |
| Celošedá | `gray` | `#E2E5E9` | `#414B56` | `#FAFBFC` | `#272B30` |
| Krémová | `cream` | `#F1C98A` | `#704C1D` | `#FFF9EF` | `#332E26` |
| Mořská modrá | `blue` | `#99D5FF` | `#165E88` | `#FBFDFF` | `#223444` |
| Olivově zelená | `olive` | `#C6DC9B` | `#4B642F` | `#FDFEF8` | `#2F392C` |
| Petrolejová | `teal` | `#A0E3D7` | `#126B63` | `#FCFFFE` | `#25383C` |

## Sémantické role

| Token | Význam |
|---|---|
| `--ui-base` / `base` | Základní plocha okna |
| `--ui-top` / `top` | Horní lišty |
| `--ui-panel` / `panel` | Vnitřní panel |
| `--ui-surface` / `surface` | Vnořená plocha / dialog |
| `--ui-subtle` / `subtle` | Jemně oddělená sekce |
| `--ui-input` / `input` | Textové vstupní pole |
| `--ui-control` / `control` | Běžné tlačítko a ouško |
| `--ui-hover` / `hover` | Najeď myší |
| `--ui-selected` / `selected` | Aktivní či vybraný prvek |
| `--ui-border` / `border` | Obrys prvku |
| `--ui-text` / `text` | Základní text |
| `--ui-muted` / `muted` | Sekundární text |
| `--ui-accent` / `accent` | Akcentní text a ikona |
| `--ui-accentStrong` / `accentStrong` | Plná akcentní plocha / hrana |
| `--ui-onAccent` / `onAccent` | Text na plné akcentní ploše |
| `--ui-focus` / `focus` | Obrys fokusu |

V HTML se ještě pro kompatibilitu udržují původní tokeny `--base`, `--top`, `--panel`, `--line`, `--text`, `--accent` a `--accent-soft`; pro nové funkce Codex použije pouze **jedinou centralizovanou mapu**, ne pevné hodnoty `#...` po komponentách.

## Povinná pravidla pro Codex

1. Volba **Světlý/Tmavý** a **šesti schémat** patří společně do `Nastavení → Rozhraní`. Nezdvojovat volbu vzhledu v jiné záložce. Uživatelské **Vlastní** barvy jsou navíc, ne sedmou oficiální paletou.
2. Přepnutí schématu vymění **všechny** související povrchy a odstíny: horní lišty, nástrojové a pravé panely, ouška, tlačítka, vstupní pole, rozevírací seznamy, vybrané řádky, kontextové nabídky, zvýraznění, hranice a stavové prvky. **Nestačí** vyměnit pouze `--accent`.
3. Ve **světlém režimu** musí mít ouška, tlačítka, seznamové řádky a textová pole **světlé pozadí**. Nepoužívat fixní `#26262f` ani fixně bílé písmo. Vlastní světlé pozadí a tmavý text tvoří pár, který musí projít kontrolou.
4. Kontrast normálního textu dle WCAG AA je **minimálně 4,5 : 1**, u velkého textu a grafických stavových ikon alespoň **3 : 1**. V souboru JSON je 96 testovaných párů ze 12 variant. **Všechny splňují 4,5 : 1**. Další komponenty musejí používat vhodné tokeny a procházet ověřením vypočtených stylů.
5. **Ovládací text** a popisky chyb ve formuláři překládá systém `i18n`, včetně dynamických hlášení a počtů. Obsah PDF, názvy existujících souborů a názvy polí převzaté ze skutečného dokumentu se nepřekládají.
6. Barvy skutečné stránky PDF (včetně formátů CMYK, ICC, rastrů a původních formulářových hodnot) **nikdy nepřepisuje motiv aplikace**. Stejně tak změny UI nesmějí měnit dirty flag dokumentu, Undo ani export.
7. Barvy stavu **chyba/varování/úspěch** jsou samostatné sémantické role; červená chyba není dekorativní vínová. Výchozí přichytávací pomocné linky mohou sledovat akcent zvolené palety, ale ručně zvolenou barvu vodítka/pomůcky motiv nepřepisuje.
8. Vlastní barva se aplikuje až po tlačítku **Použít** a musí automaticky zabezpečit odpovídající kontrast textu. `Zrušit` obnoví poslední potvrzený vzhled. Dlouhé popisky a jazykové mutace nesmějí přetékat.
9. Ikony (včetně magnetu, pravítek, koše a symetrického ozubeného kolečka Nastavení) mají používat `currentColor`, s jedinou konzistentní velikostí v liště; Nastavení je stejně vysoké jako sousední disketa.
10. Nezavádět duplikované styly pro každý nástroj. Zdroj pravdy: jedny návrhové tokeny, strukturované palety z **`mapa_barev_v1.2.json`**, jejich testy a jednotná logika v aplikaci. Bez výslovného schválení neimplementovat do Go.

## Ověření barevných párů

| Paleta | Režim | Nejslabší testovaný kontrast | Výsledek |
|---|---|---|---|
| Šedo-vínová | Tmavý | 4.58 : 1 | Splněno |
| Šedo-vínová | Světlý | 7.89 : 1 | Splněno |
| Celošedá | Tmavý | 4.97 : 1 | Splněno |
| Celošedá | Světlý | 7.21 : 1 | Splněno |
| Krémová | Tmavý | 5.22 : 1 | Splněno |
| Krémová | Světlý | 6.88 : 1 | Splněno |
| Mořská modrá | Tmavý | 4.89 : 1 | Splněno |
| Mořská modrá | Světlý | 6.37 : 1 | Splněno |
| Olivově zelená | Tmavý | 5.97 : 1 | Splněno |
| Olivově zelená | Světlý | 5.63 : 1 | Splněno |
| Petrolejová | Tmavý | 4.73 : 1 | Splněno |
| Petrolejová | Světlý | 5.70 : 1 | Splněno |

**Navazující soubory:** `docs/pracovni/rozhrani/cely_vizual_v3.21.html`, `docs/pracovni/i18n/ui_catalog_v1.2.json`, `docs/pracovni/README.md` a `docs/pracovni/specifikace_Codex_v5.0.md`.

## Regresní testy rozhraní

- Zkontrolovat všech **6 × 2** kombinací, vždy i v Nastavení, Formulářích a Pomůckách rozvržení (ouška, inputy, volby, tlačítka, stavové štítky).
- Automaticky ověřit kontrast vypočtených barev a klávesový fokus; kontrolovat kontrast i při hover, disabled, selected, error.
- Ověřit překlad hlavního panelu formulářů: `Přehled polí` → `Field overview`, `Stránka 1 (8 polí)` → `Page 1 (8 fields)`, chybové hlášky v EN; údaje dokumentu beze změny.
- Ověřit, že **Neutrální šedá** neobsahuje dekorativní vínový akcent a barvy vodítek zadané uživatelem zůstanou zachované.
- Ověřit, že přepnutí režimu/šablony nespadne do nové editace PDF a při Zrušit obnoví předchozí styl.
## Nové role v1.2
Samostatné `panelTabIdle`, `panelTabHover`, `panelTabActive`, `tabIdle`, `tabHover`, `tabActive`, `scrollTrack`, `scrollThumb`, `scrollThumbHover`, `interactiveText`, `helpText` pro všech šest palet a oba režimy. Stejná jemná interakce oušek na levé i pravé straně a obou horních ovladačích. Ve světlém motivu žádný tmavý hover ani posuvník. Nové kontrastní páry jsou v poli `component_checks` souboru JSON. Test computed styles a interakcí v prohlížeči se vyžaduje samostatně.

## Audit v1.2 – miniatury, barvy, OCR
Nové role `thumbnailOutline`, `thumbnailOutlineSoft`, `colorWell`, `sectionBg`, `popoverBg`, `iconInk` se liší podle palety a světlého/tmavého režimu. Vybrané rámečky miniatur, povrchy barevných inputů, OCR nápovědy a historické dialogy se řídí mapou. Nezaměňovat barvy UI a skutečné barvy vybraných vzorků či PDF. Kontrolovat vypočtené CSS a kontrast všech komponent.
