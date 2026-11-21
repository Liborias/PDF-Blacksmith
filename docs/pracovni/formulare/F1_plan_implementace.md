# Formuláře F1 – etapový implementační plán

**Pracovní návrh pro Codex.** Implementace zatím neproběhla. Hlavní dokumentace: [F1 specifikace](F1_specifikace_v1.9.md), [schválená rozhodnutí](F1_rozhodnuti.md), [HTML návrh](F1_mockup_v1.0.html).

## F1-0 Inventura stávající aplikace
Analyzovat `workspace_windows.go`, `viewer_windows.go`, správu dokumentů, ukládání, tisk, export a backend. Zjistit již existující práci s PDF anotacemi a formuláři, konkrétní PDFium API a podporu fontů. Před úpravou sepsat regresní testy.

## F1-1 Detekce a datový model
Přečíst AcroForm /Fields /Kids, typy /Tx /Btn /Ch /Sig a widget /Rect. Rozlišit skutečné pole, sdílené widgety, opakované výskyty, akce, XFA a podpisy. Nabídnout přechod do Formulářů bez automatického přepínání a bez příznaku neuložené změny.

## F1-2 Bezpečné vyplňování
V aktivním nástroji Formuláře → Vyplnit umožnit text, víceřádkový text, checkbox, radio a výběrová pole; ostatní nástroje nesmějí změnit formulář. Zachovat předepsanou typografii a Tab pořadí. Změny propojit s Ctrl+Z/Y a AutoRecover.

## F1-3 UI a pracovní vrstvy
Přidat nástroj Formuláře do stávající levé navigace, pravý panel s přepínačem zvýraznění, sbalitelným přehledem a navigací. Seznam řadit podle stránky a polohy, Tab podle PDF. Zobrazení vínových/červených obrysů řešit čistě jako UI; funkční i při zoomu, rotaci a duálním pohledu.

## F1-4 Validace a bezpečné ukládání
Průběžný stav bez rušivého upozornění, chybné formáty označit po opuštění pole, před ukládáním zobrazit souhrn a možnost Uložit přesto. Výsledné hodnoty /V a vzhled /AP /AS musí odpovídat, být znovu čitelné a tisknutelné. **Nemažte ostatní PDF anotace/widgety/odkazy/podpisy.**

## F1-5 Akceptace
Ověřit typy polí v PDFium a nezávislých čtečkách, češtinu, opakované widgety, podpisy, XFA, rozpracované formuláře, bezpečné kopie, AutoRecover, tisk bez pracovních poznámek a regresi starých nástrojů. Win32 UI otestovat přímo ve Windows.

**Odloženo:** F2 tvorba/editor polí, F3 zploštění, nový modul Preflight a vyhledávací šipky.
