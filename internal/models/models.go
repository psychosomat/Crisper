package models

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"Crisper/internal/hardware"
)

type ModelInfo struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	SizeGB       float64 `json:"size_gb"`
	MinRAMGB     float64 `json:"min_ram_gb"`
	SpeedFactor  float64 `json:"speed_factor"`
	Description  string `json:"description"`
	Filename     string `json:"filename"`
	URL          string `json:"url"`
	SHA256       string `json:"sha256"`
}

var AvailableModels = []ModelInfo{
	{
		Name:         "tiny",
		DisplayName:  "Tiny (~75 MB)",
		SizeGB:       0.075,
		MinRAMGB:     1.0,
		SpeedFactor:  10.0,
		Description:  "Fastest, lowest accuracy. Good for quick tests and low-resource devices.",
		Filename:     "ggml-tiny.bin",
		URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
		SHA256:       "",
	},
	{
		Name:         "base",
		DisplayName:  "Base (~150 MB)",
		SizeGB:       0.150,
		MinRAMGB:     1.0,
		SpeedFactor:  7.0,
		Description:  "Balanced speed and accuracy for simple dictation.",
		Filename:     "ggml-base.bin",
		URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
		SHA256:       "",
	},
	{
		Name:         "small",
		DisplayName:  "Small (~500 MB)",
		SizeGB:       0.500,
		MinRAMGB:     2.0,
		SpeedFactor:  4.0,
		Description:  "Good accuracy for most use cases. Requires 2+ GB RAM.",
		Filename:     "ggml-small.bin",
		URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		SHA256:       "",
	},
	{
		Name:         "medium",
		DisplayName:  "Medium (~1.5 GB)",
		SizeGB:       1.5,
		MinRAMGB:     4.0,
		SpeedFactor:  2.0,
		Description:  "High accuracy, good for professional transcription. Requires 4+ GB RAM.",
		Filename:     "ggml-medium.bin",
		URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin",
		SHA256:       "",
	},
	{
		Name:         "large-v3",
		DisplayName:  "Large v3 (~3 GB)",
		SizeGB:       3.0,
		MinRAMGB:     8.0,
		SpeedFactor:  0.8,
		Description:  "Maximum accuracy. Requires 8+ GB RAM, multi-core CPU recommended.",
		Filename:     "ggml-large-v3.bin",
		URL:          "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin",
		SHA256:       "",
	},
}

func AllModels() []ModelInfo {
	return AvailableModels
}

func RecommendModel() ModelInfo {
	specs := hardware.Detect()

	for i := len(AvailableModels) - 1; i >= 0; i-- {
		m := AvailableModels[i]

		if specs.TotalRAMGB < m.MinRAMGB+0.5 {
			continue
		}

		if m.SpeedFactor < 1.0 && specs.CPUThreads < 8 {
			continue
		}

		if specs.CPUThreads <= 2 && m.SizeGB > 0.5 {
			continue
		}
		if specs.CPUThreads <= 4 && m.SizeGB > 1.5 {
			continue
		}

		return m
	}
	return AvailableModels[0]
}

func ModelPath(storageDir, name string) string {
	for _, m := range AvailableModels {
		if m.Name == name {
			return filepath.Join(storageDir, m.Filename)
		}
	}
	return filepath.Join(storageDir, name)
}

func IsDownloaded(storageDir, name string) bool {
	p := ModelPath(storageDir, name)
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

func DownloadWithProgress(storageDir string, m ModelInfo, progress func(downloaded, total int64)) error {
	target := filepath.Join(storageDir, m.Filename)
	if m.URL == "" {
		return fmt.Errorf("no download URL for model %s", m.Name)
	}

	tmpTarget := target + ".part"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("create model dir: %w", err)
	}

	resp, err := http.Get(m.URL)
	if err != nil {
		return fmt.Errorf("download %s: %w", m.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", m.Name, resp.StatusCode)
	}

	total := resp.ContentLength

	out, err := os.Create(tmpTarget)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	hasher := sha256.New()
	pr := &progressReader{
		r:        resp.Body,
		total:    total,
		progress: progress,
	}

	wr := io.MultiWriter(out, hasher)
	_, err = io.Copy(wr, pr)
	out.Close()

	if err != nil {
		os.Remove(tmpTarget)
		return fmt.Errorf("download %s: %w", m.Name, err)
	}

	if m.SHA256 != "" {
		got := fmt.Sprintf("%x", hasher.Sum(nil))
		if got != m.SHA256 {
			os.Remove(tmpTarget)
			return fmt.Errorf("download %s: SHA256 mismatch (expected %s, got %s)", m.Name, m.SHA256, got)
		}
	}

	if err := os.Rename(tmpTarget, target); err != nil {
		os.Remove(tmpTarget)
		return fmt.Errorf("rename model file: %w", err)
	}

	return nil
}

type progressReader struct {
	r          io.Reader
	total      int64
	downloaded int64
	progress   func(downloaded, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.downloaded += int64(n)
	if pr.progress != nil && pr.total > 0 {
		pr.progress(pr.downloaded, pr.total)
	}
	return n, err
}
