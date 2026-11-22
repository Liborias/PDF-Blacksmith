# PDF Blacksmith — odložené úpravy bočních panelů po HTML v2.6

**Zapsáno:** 22. listopadu 2026
**Stav:** schválené zadání pro *příští* úpravu vzhledu, **NEIMPLEMENTOVAT NYNÍ**. Uživatel výslovně požádal pokyny pouze uložit a následně se věnovat diskusi o Souboru a Zobrazení.

## 1. Umístění oušek pomocných panelů

- **Odstranit samostatnou boční lištu/kolejnici**, do níž jsou v aktuální verzi v2.6 vložena ouška pomocných panelů.
- Ouška každého pomocného panelu mají **vyčnívat přímo do pracovní plochy dokumentu** (nikoli zabírat trvalý vodorovný pruh vyhrazený pro lištu). Vnější rohy oušek mohou být jemně zaoblené; panely zůstávají hranaté.
- Zachovat již dohodnuté horní **skupinové ouško hlavního panelu** s dvojitými směrovými šipkami na obou stranách, umístěné na vlastní hraně hlavního panelu. Horní skupinové ouško přesahuje k pomocným panelům; jejich obsah musí mít vyhrazený horní odstup, aby nedocházelo ke kolizím.
- Po sbalení celé levé nebo pravé skupiny se má její původní stav přesně obnovit, včetně individuálně otevřených/zavřených pomocných panelů. Tuto funkčnost nepřepisovat.

## 2. Posuvníky

- V **panelu miniatur/Náhledy** i v dalších posuvných panelech zavést posuvníky **se zaoblenými hranami**, ve stylu dříve uživatelem zaslaných snímků původní aplikace.
- Uprostřed posuvníku má být středová „**kolejnička**“ jako **tenká čára**.
- Zvolit malé a střídmé provedení v původním tmavém barevném schématu; nepřidávat výrazné barevné prvky.
- Při provedení vycházet z dříve dodaných referenčních screenshotů aplikace, ne z náhodných stylizovaných scrollbarů.

## 3. Podklady a pořadí

- Výchozí vizuální stav: `cely_vizual_panely_ouska_v2.6.html`.
- Nejprve **probrat Soubor a Zobrazení**. **Až v následujícím kroku**, na výslovnou žádost uživatele, vytvořit novou HTML verzi s těmito úpravami.
- Nyní neměnit stávající HTML v2.6, zdrojový kód Go/Win32, schválenou strukturu nástrojů ani panelů.
