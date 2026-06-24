package whisper

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const defaultVersion = "v1.9.1"

var httpClient = &http.Client{Timeout: 30 * time.Second}

type archiveSpec struct {
	asset      string
	binaryName string
	archiveFmt string
}

var platformArchives = map[string]archiveSpec{
	"linux/amd64":   {"whisper-bin-ubuntu-x64.tar.gz", "whisper-cli", "tar.gz"},
	"linux/arm64":   {"whisper-bin-ubuntu-arm64.tar.gz", "whisper-cli", "tar.gz"},
	"windows/amd64": {"whisper-bin-x64.zip", "whisper-cli.exe", "zip"},
}

var versionCache struct {
	mu  sync.Mutex
	ver string
	ok  bool
}

func latestVersion(defaultVer string) string {
	versionCache.mu.Lock()
	defer versionCache.mu.Unlock()

	if versionCache.ok {
		return versionCache.ver
	}

	req, err := http.NewRequest("GET", "https://api.github.com/repos/ggml-org/whisper.cpp/releases/latest", nil)
	if err != nil {
		return defaultVer
	}
	req.Header.Set("User-Agent", "Crisper")

	resp, err := httpClient.Do(req)
	if err != nil {
		return defaultVer
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return defaultVer
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil || release.TagName == "" {
		return defaultVer
	}

	versionCache.ver = release.TagName
	versionCache.ok = true
	return release.TagName
}

func supportedPlatform() (archiveSpec, bool) {
	s, ok := platformArchives[runtime.GOOS+"/"+runtime.GOARCH]
	return s, ok
}

func downloadURL(asset string) string {
	ver := latestVersion(defaultVersion)
	return fmt.Sprintf("https://github.com/ggml-org/whisper.cpp/releases/download/%s/%s", ver, asset)
}

func DownloadWhisperCLI(destDir string, progress func(downloaded, total int64)) (string, error) {
	spec, ok := supportedPlatform()
	if !ok {
		return "", fmt.Errorf("no pre-built whisper-cli for %s/%s\n"+
			"Install manually:\n"+
			"  macOS: brew install whisper-cpp\n"+
			"  Other: https://github.com/ggml-org/whisper.cpp", runtime.GOOS, runtime.GOARCH)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("create bin dir: %w", err)
	}

	url := downloadURL(spec.asset)
	tmpArchive := filepath.Join(destDir, spec.asset+".part")

	if err := downloadFile(url, tmpArchive, progress); err != nil {
		os.Remove(tmpArchive)
		return "", err
	}

	binaryPath := filepath.Join(destDir, spec.binaryName)

	switch spec.archiveFmt {
	case "tar.gz":
		if err := extractTarGz(tmpArchive, destDir); err != nil {
			os.Remove(tmpArchive)
			return "", fmt.Errorf("extract tar.gz: %w", err)
		}
	case "zip":
		if err := extractZip(tmpArchive, destDir); err != nil {
			os.Remove(tmpArchive)
			return "", fmt.Errorf("extract zip: %w", err)
		}
	}

	os.Remove(tmpArchive)

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return "", fmt.Errorf("chmod: %w", err)
		}
	}

	if err := checkBinary(binaryPath); err != nil {
		return "", fmt.Errorf("downloaded whisper-cli is broken: %w", err)
	}

	return binaryPath, nil
}

func downloadFile(url, dest string, progress func(downloaded, total int64)) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("http new request: %w", err)
	}
	req.Header.Set("User-Agent", "Crisper")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	total := resp.ContentLength
	var read int64
	buf := make([]byte, 32*1024)
	writeErr := error(nil)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				writeErr = werr
				break
			}
			read += int64(n)
			if progress != nil && total > 0 {
				progress(read, total)
			}
		}
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			out.Close()
			return err
		}
	}

	if cerr := out.Close(); cerr != nil && writeErr == nil {
		writeErr = cerr
	}

	if writeErr != nil {
		return fmt.Errorf("write file: %w", writeErr)
	}

	if total > 0 && read != total {
		return fmt.Errorf("incomplete download: %d/%d bytes", read, total)
	}

	return nil
}

func extractTarGz(archive, destDir string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	extracted := false
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			dest := filepath.Join(destDir, filepath.Base(header.Name))
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			extracted = true
		}
	}

	if !extracted {
		return fmt.Errorf("empty archive")
	}
	return nil
}

func extractZip(archive, destDir string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	extracted := false
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		dest := filepath.Join(destDir, filepath.Base(f.Name))
		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		extracted = true
	}

	if !extracted {
		return fmt.Errorf("empty archive")
	}
	return nil
}

func needBinary(name string) bool {
	base := filepath.Base(name)
	return base == "whisper-cli" || base == "whisper-cli.exe"
}
