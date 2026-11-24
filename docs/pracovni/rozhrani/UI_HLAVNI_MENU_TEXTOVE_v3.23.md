# PDF Blacksmith – UI v3.23: hlavní menu bez tlačítkových rámečků

**Datum:** 24. 11. 2026 · **Rozsah:** pouze podoba druhé lišty shora v HTML prototypu. Nejde o povolení změn produkční aplikace.

**HTML:** [v3.23](cely_vizual_v3.23.html). **Předchozí verze:** [v3.22](cely_vizual_v3.22.html).

Položky **Soubor, Zobrazení, Nástroje, Úpravy, Formuláře, Pomoc** jsou prosté textové položky nabídky bez trvalého podkladu a bez viditelného rámečku. Při hover, rozbalení menu a klávesovém fokusu zůstává jemné zřetelné zvýraznění podle aktivní palety. Oprava je výhradně pro `.commandbar .menus .menubtn`; ostatní tlačítka horní lišty a ikonové ovladače jsou beze změn.

Zachovat stávající click/hover/menu handlery, pořadí, zkratky, lokalizaci, nastavení motivů, horní půlený přepínač, veškeré PDF/Go funkce, stav Undo i předchozí verze HTML. Ověřit všech šest palet v obou režimech a interakce hover/open/focus.

**Poznámka ke kontrole:** CSS je cílené a struktura HTML se jinak nemění. Vykreslení v reálném prohlížeči dosud nebylo kompletně ověřeno.
