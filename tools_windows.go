package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type exportPreset struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Suffix      string `json:"suffix"`
	Profile     string `json:"profile"`
}

type exportOptions struct {
	Intent      int    `json:"intent"`
	BlackPoint  int    `json:"black_point"`
	ImageDPI    int    `json:"image_dpi"`
	Compression string `json:"compression"`
	ProofDPI    int    `json:"proof_dpi"`
	RepairDPI   int    `json:"repair_dpi"`
	ColorMode   string `json:"color_mode"`
	GrayProfile string `json:"gray_profile"`
	KProfile    string `json:"k_profile"`
}

var defaults = exportOptions{Intent: -1, BlackPoint: -1, ImageDPI: 0, Compression: "auto", ProofDPI: 150, RepairDPI: 600, ColorMode: "gray", GrayProfile: "automatic", KProfile: "fogra51"}
var (
	intents      = []int{-1, 0, 1, 2, 3}
	blackPoints  = []int{-1, 1, 0}
	imageDPIs    = []int{0, 150, 300, 600}
	compressions = []string{"auto", "flate"}
	proofDPIs    = []int{75, 150, 300}
	repairDPIs   = []int{300, 600}
	colorModes   = []string{"gray", "rich_black", "k_cmyk"}
	grayProfiles = []string{"automatic", "gamma22", "gamma18", "custom"}
)

func defaultIndex[T comparable](values []T, value, fallback T) int {
	fallbackIndex := 0
	for index, candidate := range values {
		if candidate == value {
			return index
		}
		if candidate == fallback {
			fallbackIndex = index
		}
	}
	return fallbackIndex
}

func localAppFile(name string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, storageName, name), nil
}

func loadExportOptions() exportOptions {
	options := defaults
	path, err := localAppFile("export-settings.json")
	if err != nil {
		return options
	}
	if data, err := os.ReadFile(path); err == nil {
		if err = json.Unmarshal(data, &options); err != nil {
			return defaults
		}
	}
	options.Intent = intents[defaultIndex(intents, options.Intent, defaults.Intent)]
	options.BlackPoint = blackPoints[defaultIndex(blackPoints, options.BlackPoint, defaults.BlackPoint)]
	options.ImageDPI = imageDPIs[defaultIndex(imageDPIs, options.ImageDPI, defaults.ImageDPI)]
	options.Compression = compressions[defaultIndex(compressions, options.Compression, defaults.Compression)]
	options.ProofDPI = proofDPIs[defaultIndex(proofDPIs, options.ProofDPI, defaults.ProofDPI)]
	options.RepairDPI = repairDPIs[defaultIndex(repairDPIs, options.RepairDPI, defaults.RepairDPI)]
	options.ColorMode = colorModes[defaultIndex(colorModes, options.ColorMode, defaults.ColorMode)]
	options.GrayProfile = grayProfiles[defaultIndex(grayProfiles, options.GrayProfile, defaults.GrayProfile)]
	if options.KProfile == "" {
		options.KProfile = defaults.KProfile
	}
	return options
}

func saveExportOptions(options exportOptions) error {
	path, err := localAppFile("export-settings.json")
	if err != nil {
		return err
	}
	return saveLocalJSON(path, options)
}

func saveLocalJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	writeErr := json.NewEncoder(file).Encode(value)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}
