# PDF Blacksmith — JEDINÁ MAPA BAREV v6.0

**Referenční HTML:** [v3.26](rozhrani/cely_vizual_v3.26.html). **Přesná úplná data:** [JSON v6.0](mapa_barev_FINAL_v6.0.json) (kopie původních schválených hodnot v1.4, bez jejich přepočítání; aktualizovaná pouze metadata a normativní pravidla).

**Závaznost:** tato mapa nahrazuje odkazy na v1.0–v1.4. Rozhraní má 6 nezávislých schémat × 2 režimy. Mapa se týká **pouze aplikačního UI**. Obsah PDF, autentické barvy obrázků, CMYK přímé barvy, barevné exporty, uživatelsky určené odstíny vodítek/štítků a skutečné hodnoty swatchů jsou **nedotknutelné**.

## Palety – přesné základní hodnoty

| Paleta | Režim | Base | Panel | Text | Vybraný podklad | Akcent |
|---|---|---|---|---|---|---|
| Šedo-vínová (`wine`) | light | `#ECEAED` | `#FFFCFE` | `#262029` | `#F3E4EB` | `#753C55` |
| Šedo-vínová (`wine`) | dark | `#19191F` | `#26242D` | `#F5F1F4` | `#45303D` | `#AD5B80` |
| Celošedá (`gray`) | light | `#E9EBED` | `#FAFBFC` | `#242930` | `#E3E7EC` | `#48535F` |
| Celošedá (`gray`) | dark | `#191B1E` | `#272B30` | `#F4F5F6` | `#454E57` | `#66717D` |
| Krémová (`cream`) | light | `#F0E8D9` | `#FFF9EF` | `#332819` | `#F4E2C6` | `#775327` |
| Krémová (`cream`) | dark | `#24211C` | `#332E26` | `#F8F0E2` | `#554637` | `#896639` |
| Mořská modrá (`blue`) | light | `#E7F0F7` | `#FBFDFF` | `#192D3E` | `#DDEFFA` | `#1A658E` |
| Mořská modrá (`blue`) | dark | `#131F2A` | `#223444` | `#F0F7FD` | `#31516B` | `#2D789F` |
| Olivově zelená (`olive`) | light | `#EDF0E4` | `#FDFEF8` | `#26321E` | `#E4EECE` | `#537033` |
| Olivově zelená (`olive`) | dark | `#1E241B` | `#2F392C` | `#F3F7EA` | `#4C6044` | `#536B34` |
| Petrolejová (`teal`) | light | `#E5F1EF` | `#FCFFFE` | `#1C3633` | `#DCF1EB` | `#12736A` |
| Petrolejová (`teal`) | dark | `#162326` | `#25383C` | `#F0FAF8` | `#385C60` | `#327F77` |

## Sémantické tokeny a jejich hranice

**Běžné povrchy:** base/top/panel/surface/subtle/input/control/hover/selected/border/text/muted/accent/accentStrong/onAccent/focus. **Panelové a scrollbarové:** panelTabIdle/Hover/Active, tabIdle/Hover/Active, scrollTrack/Thumb/ThumbHover. **Kompaktní text a předvolby:** interactiveText/helpText, thumbnailOutline/thumbnailOutlineSoft, colorWell/sectionBg/popoverBg/iconInk. **Novější role:** editor, tooltipBg/tooltipText/tooltipBorder, infoBannerBg/infoBannerText, noteStickerBg/noteStickerText/noteStickerBorder, toolbarActive/toolbarActiveHover.

1. **Nikdy natvrdo vínová** v celošedé, krémové, modré, olivové a petrolejové. Dokonce i vybraný/hover přepínač, focus, výběr miniatur, nápověda, progress, OCR a barevný picker používají přesné tokeny aktivního schématu.
2. **Světlý režim** používá světlé panely, tlačítka, color wells, vstupy, rozbalovací nabídky, posuvníky a pomocné „ouško“; žádný zapomenutý tmavý fallback. Běžný text musí mít kontrast nejméně **4,5:1**, relevantní grafická/fokusová linie nejméně **3:1**. Kontrastní testy původního JSON byly podle tokenů; skutečné ověření všech DOM/Win32 prvků je samostatný úkol.
3. **Editorová plocha** (okolí PDF) je odvozena z palety; ne každá dark paleta má černou. Vnější **dekorativní obrys PDF stránkového listu je zrušen ve všech režimech**. To nemá nic společného s PDF obrysy widgetů, aktivními objekty nebo anotacemi.
4. **Horní textové menu** a levá navigace jsou borderless. Navigační aktivní/hover používá paletový hover/selected a úzkou barevnou linku vlevo; naproti tomu aktivní **ikonové tlačítko druhé lišty** má paletové toolbarActive + důsledné rozlišení.
5. **Panelové ouška** na obou stranách a v horní oblasti v idle/hover/active zachovávají světlý/tmavý stav. Scrollbar track/thumb/hover odpovídají tématu. Styl miniatur je UI, ale obraz stránky miniatury se nikdy nepřebarví.
6. **Vlastní barva anotace nebo vodítka** je uživatelská volba, **ne barva UI**. Výchozí štítek poznámky může být podle motivu, ale individuální a globální vlastní barva mají přednost. Nastavení barev má náhled a explicitní „Použít“, Zrušit vrací předchozí uložený stav.
7. **Color input**: barva rámu a pozadí vstupu patří aktuálnímu motivu, skutečný vnitřní barevný vzorek zůstává beze změny. Testovat i Chromium `::-webkit-color-swatch-wrapper` a Firefox.
8. Půlený kruhový přepínač horní lišty a Nastavení → Rozhraní sdílí **jediný stav režimu**, ne dvě nezávislé preference.
9. Neměnit PDF obsah, profil ICC, skutečné separace ani dokumentový text při změně palety. Uživatelské ukládání preferencí je odděleno od PDF dirty/Undo.

## Akceptační test
Pro každou z **12 kombinací** na šířkách 980/1100/1250/1600 a na DPI 100/150/200 % testovat naživo: hlavní menu vs ikonová lišta, horní/boční ouška, miniatury a vybraný stav, Text/Objekty/OCR/Separace/Formuláře/Poznámky, nápovědu, stavové pruhy, měřítko, body fokus/hover, color wells a scrollbary. Bez screenshotu/computed stylů nikdy nepsat „vizuálně ověřeno“.

**Poznámka ke zdroji:** původní mapy nesou záznamy předem vypočtených kontrastů a rolí. Jejich převzetí do JSON v6.0 neznamená, že HTML v3.26 kompletně prošlo prohlížečovou regresí.