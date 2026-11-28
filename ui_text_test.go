package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestUICatalogMessages(t *testing.T) {
	if len(uiTranslations) != 829 {
		t.Fatalf("got %d canonical entries, want 829", len(uiTranslations))
	}
	for _, sample := range []struct {
		key            uiMessage
		czech, english string
	}{
		{uiFile, "Soubor", "File"},
		{uiFitWidth, "Přizpůsobit šířce", "Fit width"},
		{uiInterfaceLanguage, "Jazyk rozhraní", "Interface language"},
		{uiOpenFile, "Otevřít soubor", "Open file"},
		{uiExpandRight, "Rozbalit pravou skupinu panelů", "Expand right panel group"},
		{uiLayersAndObjects, "Vrstvy a objekty", "Layers and objects"},
		{uiRecognitionLanguage, "Jazyk rozpoznávání", "Recognition language"},
		{uiWarnFlattenedText, "Pozor: text ve zploštěném PDF nepůjde označit.", "Text in the flattened PDF cannot be selected."},
	} {
		if cs, en := uiCzech.text(sample.key), uiEnglish.text(sample.key); cs != sample.czech || en != sample.english {
			t.Errorf("%s = (%q, %q), want (%q, %q)", sample.key, cs, en, sample.czech, sample.english)
		}
	}
}

func TestChoiceLocalizationPreservesProfileNamesAndOrder(t *testing.T) {
	labels := []string{"", "FOGRA39 · ISO Coated v2", ""}
	keys := []uiMessage{uiKeepColors, "", uiCustomOutputICC}
	uiEnglish.applyChoiceMessages(labels, keys)
	if want := []string{"Preserve colors", "FOGRA39 · ISO Coated v2", "Custom output ICC profile"}; !slices.Equal(labels, want) {
		t.Fatalf("English labels = %v, want %v", labels, want)
	}
	uiCzech.applyChoiceMessages(labels, keys)
	if want := []string{"Zachovat barvy", "FOGRA39 · ISO Coated v2", "Vlastní výstupní profil ICC"}; !slices.Equal(labels, want) {
		t.Errorf("Czech labels = %v, want %v", labels, want)
	}
}

func TestExportPresetTranslationKeys(t *testing.T) {
	data, err := os.ReadFile("export_presets.json")
	if err != nil {
		t.Fatal(err)
	}
	var presets []struct {
		Key               string `json:"key"`
		Name, Description string
	}
	if err := json.Unmarshal(data, &presets); err != nil {
		t.Fatal(err)
	}
	for _, preset := range presets {
		labels := exportPresetMessages[preset.Key]
		if uiCzech.text(labels.name) != preset.Name || uiCzech.text(labels.description) != preset.Description {
			t.Errorf("preset %q has different Czech captions from its export definition", preset.Key)
		}
		if uiEnglish.text(labels.name) == "" || uiEnglish.text(labels.description) == "" {
			t.Errorf("preset %q lacks English help", preset.Key)
		}
	}
}

func TestLocalizedMessagePreservesDocumentOutputIntent(t *testing.T) {
	const intent = "Ofsetový tisk Žluťoučký kůň"
	message := localizedMessage{key: uiDocumentInfo, args: []any{5, intent, 2, 5}}
	if got := uiEnglish.formatMessage(message); !strings.Contains(got, intent) || !strings.Contains(got, "2 of 5 pages in memory") {
		t.Errorf("unexpected English document information: %q", got)
	}
}

func TestDecodeUILanguage(t *testing.T) {
	for _, example := range []struct {
		file string
		want uiLocale
	}{
		{"", uiCzech},
		{"{", uiCzech},
		{`{"locale":"de-DE"}`, uiCzech},
		{`{"locale":"cs-CZ"}`, uiCzech},
		{`{"locale":"en-US"}`, uiEnglish},
	} {
		if got := decodeUILanguage([]byte(example.file)); got != example.want {
			t.Errorf("%q selected %s, want %s", example.file, got, example.want)
		}
	}
}


func TestRecentAndThumbnailLocalization(t *testing.T) {
	for _, sample := range []struct {
		key uiMessage
		cs,en string
	}{
		{uiRecentOpen,"Otevřít","Open"},
		{uiRecentOpenLocation,"Otevřít umístění souboru","Open file location"},
		{uiRecentCopyPath,"Kopírovat cestu","Copy path"},
		{uiRecentLocationFailed,"Umístění souboru se nepodařilo otevřít.","Could not open the file location."},
		{uiPageThumbnails,"Náhledy stránek","Page thumbnails"},
		{uiThumbnails,"Náhledy","Thumbnails"},
		{uiThumbnailWheelHint,"Ctrl + kolečko","Ctrl + mouse wheel"},
	} {
		if cs,en:=uiCzech.text(sample.key),uiEnglish.text(sample.key); cs!=sample.cs || en!=sample.en {
			t.Errorf("%s = (%q,%q), want (%q,%q)",sample.key,cs,en,sample.cs,sample.en)
		}
	}
}

func TestFileCommandsHeadingLocalization(t *testing.T) {
	if uiCzech.text(uiFileCommands)!="Příkazy souboru" || uiEnglish.text(uiFileCommands)!="File commands" { t.Fatal("missing file menu heading translation") }
}
