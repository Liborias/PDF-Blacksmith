package main

import (
	"archive/zip"
	"bytes"
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
	"syscall"
)

//go:embed runtime.zip
var runtimeZip []byte

type request struct {
	Action          string `json:"action"`
	Input           string `json:"input"`
	Output          string `json:"output,omitempty"`
	PDFPassword     string `json:"pdf_password,omitempty"`
	Profile         string `json:"profile,omitempty"`
	CustomICC       string `json:"custom_icc,omitempty"`
	Certificate     string `json:"certificate,omitempty"`
	CertPassword    string `json:"cert_password,omitempty"`
	Trust           string `json:"trust,omitempty"`
	Gray            bool   `json:"gray,omitempty"`
	AllowRaster     bool   `json:"allow_raster,omitempty"`
	Outline         bool   `json:"outline,omitempty"`
	Decrypt         bool   `json:"decrypt,omitempty"`
	Visible         bool   `json:"visible,omitempty"`
	Overwrite       bool   `json:"overwrite,omitempty"`
	ClearSignatures bool   `json:"clear_signatures,omitempty"`
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func appRuntime() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(runtimeZip)
	root := filepath.Join(cache, appName)
	target := filepath.Join(root, "runtime-"+hex.EncodeToString(hash[:12]))
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
	for _, name := range []string{"ready", "python.exe", "engine.py", "print_profiles.py", "profiles.json", filepath.Join("ghostscript", "bin", "gswin64c.exe"), filepath.Join("ghostscript", "bin", "gsdll64.dll")} {
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

func runOperation(req request) response {
	dir, err := appRuntime()
	if err != nil {
		return response{"error", "Součásti aplikace se nepodařilo rozbalit.\n" + err.Error()}
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return response{"error", err.Error()}
	}
	cmd := exec.Command(filepath.Join(dir, "python.exe"), "-B", "-I", "-X", "utf8", filepath.Join(dir, "engine.py"))
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(payload)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	output, err := cmd.Output()
	if err != nil {
		return response{"error", fmt.Sprintf("Zpracování PDF se nepodařilo spustit.\n%v", err)}
	}
	var result response
	if err := json.Unmarshal(output, &result); err != nil {
		return response{"error", "Komunikace se zpracováním PDF se nezdařila."}
	}
	return result
}
