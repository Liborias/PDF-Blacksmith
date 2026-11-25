# PDF Blacksmith — závazný vstup pro Codex (v6.0)

Toto je existující **desktopová aplikace Go/Win32 s PDFium**, nikoli projekt k nahrazení webovým PDF editorem.

**JEDINÉ ZADÁNÍ:** [docs/pracovni/CODEX_START.md](docs/pracovni/CODEX_START.md), [docs/pracovni/CODEX_MASTER_v6.0.md](docs/pracovni/CODEX_MASTER_v6.0.md), [docs/pracovni/CODEX_IMPLEMENTACNI_PLAN_v6.0.md](docs/pracovni/CODEX_IMPLEMENTACNI_PLAN_v6.0.md) a finální mapy vydání v6.0. Toto je autoritativní instrukce pro práci. **Staré dokumenty v3.x, v4.x, v5.x a staré AGENT/ChatGPT iniciační prompty jsou archiv, nikoliv souběžné pokyny.**

**NOVÁ PRIORITA:** dokončit a sjednotit **rozhraní aplikace nejprve**. Formulářové F1 mělo dříve vyšší prioritu, ale nyní se **pozastavuje jeho další implementace** do doby schválení po UI. Rozpracované změny F1 pouze bezpečně zmapovat, **neodstraňovat**. F2, hromadné importy a další plánované funkce **neimplementovat automaticky**.

**První úkol:** přečíst aktuální v6.0 dokumenty, zkontrolovat skutečný zdrojový kód a pracovní strom, napsat audit a návrh **nejmenší UI etapy**. Bez povolení nic neupravovat, necommitovat, neprovádět push. Kontrola Windows, PDF integrity a regresní testy jsou povinné před tvrzením „hotovo“. Nemažte historii, nepoužívejte force-push ani destruktivní reset.

UI reference: docs/pracovni/rozhrani/cely_vizual_v3.26.html; barevná mapa 6×2, finální ikonová a klávesová mapa v6.0. Uživatelský obsah PDF, form widgety, podpisy, ICC a skutečné barvy se nemění při úpravě UI.
