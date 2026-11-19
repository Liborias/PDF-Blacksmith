package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

//go:embed runtime.zip
var runtimeZip []byte

var runtimeVersion = sync.OnceValue(func() string {
	hash := sha256.Sum256(runtimeZip)
	return "runtime-" + hex.EncodeToString(hash[:12])
})

type request struct {
	Action          string         `json:"action"`
	Input           string         `json:"input"`
	Output          string         `json:"output,omitempty"`
	OutputDir       string         `json:"output_dir,omitempty"`
	PDFPassword     string         `json:"pdf_password,omitempty"`
	Preset          string         `json:"preset,omitempty"`
	Profile         string         `json:"profile,omitempty"`
	CustomICC       string         `json:"custom_icc,omitempty"`
	Certificate     string         `json:"certificate,omitempty"`
	CertPassword    string         `json:"cert_password,omitempty"`
	Trust           string         `json:"trust,omitempty"`
	Gray            bool           `json:"gray,omitempty"`
	OpenResult      bool           `json:"open_result,omitempty"`
	ColorMode       string         `json:"color_mode,omitempty"`
	GrayProfile     string         `json:"gray_profile,omitempty"`
	GrayCustomICC   string         `json:"gray_custom_icc,omitempty"`
	KProfile        string         `json:"k_profile,omitempty"`
	AllowRaster     bool           `json:"allow_raster,omitempty"`
	Outline         bool           `json:"outline,omitempty"`
	Decrypt         bool           `json:"decrypt,omitempty"`
	Visible         bool           `json:"visible,omitempty"`
	Overwrite       bool           `json:"overwrite,omitempty"`
	ReplaceSource   bool           `json:"replace_source,omitempty"`
	ReplaceDocument string         `json:"replace_document,omitempty"`
	SourceSize      int64          `json:"source_size,omitempty"`
	SourceMtime     int64          `json:"source_mtime_ns,omitempty"`
	ClearSignatures bool           `json:"clear_signatures,omitempty"`
	Export          *exportOptions `json:"export,omitempty"`
	View            *viewRequest   `json:"view,omitempty"`

	MergeInputs           []mergeInput `json:"merge_inputs,omitempty"`
	InsertPages           *pageInsertion `json:"insert_pages,omitempty"`
	AllowUnencryptedMerge bool         `json:"allow_unencrypted_merge,omitempty"`
	OCRLanguage           string       `json:"ocr_language,omitempty"`
	OCRRegion             []textPoint  `json:"ocr_region,omitempty"`
	TextEdit              *textEditRequest `json:"text_edit,omitempty"`
}

type viewRequest struct {
	Pixels       int              `json:"pixels,omitempty"`
	Page         int              `json:"page,omitempty"`
	Rotation     int              `json:"rotation,omitempty"`
	Box          string           `json:"box,omitempty"`
	Region       *screenRegion    `json:"region,omitempty"`
	Pages        []int            `json:"pages,omitempty"`
	ProofProfile string           `json:"proof_profile,omitempty"`
	HidePlates   []string         `json:"hide_plates,omitempty"`
	Order        []int            `json:"order,omitempty"`
	Rotations    []int            `json:"rotations,omitempty"`
	Bookmarks    []bookmarkChange `json:"bookmarks,omitempty"`
}

type plateInfo struct {
	Name     string  `json:"name"`
	Active   bool    `json:"active"`
	Coverage float64 `json:"coverage"`
	Maximum  float64 `json:"maximum"`
	Pages    []int   `json:"pages"`
}

type viewResponse struct {
	TextPage     *textPage              `json:"text_page,omitempty"`
	Sizes        []map[string][2]float64 `json:"sizes"`
	Pages        int                     `json:"pages"`
	DPI          int                     `json:"dpi"`
	OutputIntent string                  `json:"output_intent"`
	Simulation   string                  `json:"simulation"`
	Image        []byte                  `json:"image"`
	RasterWidth  int                     `json:"raster_width"`
	RasterHeight int                     `json:"raster_height"`
	Plates       []plateInfo             `json:"plates"`
	UsedPlates   []plateInfo             `json:"used_plates"`
	MergeInfo    []mergeSourceInfo       `json:"merge_info"`
	Bookmarks    []pdfBookmark           `json:"bookmarks"`
}

type response struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	View    *viewResponse `json:"view"`
}

func operationError(message string) response {
	return response{Status: "error", Message: message}
}

func appRuntime() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(cache, storageName)
	target := filepath.Join(root, runtimeVersion())
	if runtimeReady(target) {
		return target, nil
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(root, "rozbalovani-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	if err := unpackRuntime(stage); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "ready"), nil, 0600); err != nil {
		return "", err
	}
	if err := os.Rename(stage, target); err != nil && !runtimeReady(target) {
		return "", err
	}
	return target, nil
}

func runtimeReady(path string) bool {
	for _, name := range []string{"ready", "python.exe", "engine.py", "bookmarks.py", "reader_search.py", "ocr.py", "text_edit.py", "text_block.py", "block_fonts.py", "system_fonts.py", "color_conversion.py", "export_settings.py", "export_presets.json", "pdfx.py",
		filepath.Join("text-fonts", "LiberationSans-Regular.ttf"),
		filepath.Join("packages", "fontTools", "ttLib", "__init__.py"),
		filepath.Join("tessdata", "ces.traineddata"), filepath.Join("tessdata", "eng.traineddata"),
		"proof.py", "print_profiles.py", "profiles.json", filepath.Join("ghostscript", "bin", "gswin64c.exe"),
		filepath.Join("ghostscript", "bin", "gsdll64.dll")} {
		info, err := os.Stat(filepath.Join(path, name))
		if err != nil || !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}

func unpackRuntime(stage string) error {
	files, err := zip.NewReader(bytes.NewReader(runtimeZip), int64(len(runtimeZip)))
	if err != nil {
		return err
	}
	for _, entry := range files.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		name := filepath.FromSlash(entry.Name)
		if !filepath.IsLocal(name) {
			return errors.New("neplatná cesta ve vestavěném archivu")
		}
		destination := filepath.Join(stage, name)
		if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
			return err
		}
		if err := extractEntry(entry, destination); err != nil {
			return err
		}
	}
	return nil
}

func extractEntry(entry *zip.File, destination string) error {
	src, err := entry.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := io.Copy(dst, src)
	return errors.Join(writeErr, dst.Close())
}

func engineCommand(args ...string) (*exec.Cmd, error) {
	return engineCommandContext(context.Background(), args...)
}

func engineCommandContext(ctx context.Context, args ...string) (*exec.Cmd, error) {
	dir, err := appRuntime()
	if err != nil {
		return nil, fmt.Errorf("Součásti aplikace se nepodařilo rozbalit.\n%w", err)
	}
	cmd := exec.CommandContext(ctx, filepath.Join(dir, "python.exe"),
		append([]string{"-B", "-I", "-X", "utf8", filepath.Join(dir, "engine.py")}, args...)...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd, nil
}

func runOperation(req request) response {
	cmd, err := engineCommand()
	if err != nil {
		return operationError(err.Error())
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return operationError(err.Error())
	}
	cmd.Stdin = bytes.NewReader(payload)
	output, err := cmd.Output()
	if err != nil {
		return operationError(fmt.Sprintf("Zpracování PDF se nepodařilo spustit.\n%v", err))
	}
	var result response
	if err := json.Unmarshal(output, &result); err != nil {
		return operationError("Komunikace se zpracováním PDF se nezdařila.")
	}
	return result
}
