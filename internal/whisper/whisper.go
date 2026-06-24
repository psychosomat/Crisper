package whisper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const dllNotFoundExitCode = 0xc0000135

func isDLLNotFound(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return exitErr.ExitCode() == int(dllNotFoundExitCode)
}

type Model struct {
	Path string
	Name string
}

func NewModel(modelPath string, modelName string) *Model {
	return &Model{Path: modelPath, Name: modelName}
}

type Segment struct {
	Start float64
	End   float64
	Text  string
}

type Transcriber struct {
	mu       sync.Mutex
	model    *Model
	language string
	threads  int
}

func NewTranscriber(model *Model, language string, threads int) *Transcriber {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	return &Transcriber{
		model:    model,
		language: language,
		threads:  threads,
	}
}

var cliCache struct {
	mu  sync.Mutex
	path string
	ok   bool
}

func FindWhisperCLI() (string, error) {
	return findOrInstallWhisper()
}

func InvalidateCache() {
	cliCache.mu.Lock()
	cliCache.ok = false
	cliCache.path = ""
	cliCache.mu.Unlock()
}

func checkBinary(path string) error {
	cmd := exec.Command(path, "--help")
	if err := cmd.Run(); err != nil {
		if isDLLNotFound(err) {
			return fmt.Errorf(
				"whisper-cli requires Microsoft Visual C++ Redistributable.\n" +
					"Download from: https://aka.ms/vs/17/release/vc_redist.x64.exe")
		}
		return fmt.Errorf("whisper-cli at %s cannot be executed: %w", path, err)
	}
	return nil
}

func findOrInstallWhisper() (string, error) {
	cliCache.mu.Lock()
	defer cliCache.mu.Unlock()

	if cliCache.ok {
		return cliCache.path, nil
	}

	paths := []string{"whisper-cli", "whisper"}
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			if err := checkBinary(p); err != nil {
				return "", err
			}
			cliCache.path = p
			cliCache.ok = true
			return p, nil
		}
	}

	home, _ := os.UserHomeDir()
	cfgDir, _ := os.UserConfigDir()
	exe, _ := os.Executable()
	appDir := filepath.Dir(exe)

	binName := "whisper-cli"
	if runtime.GOOS == "windows" {
		binName = "whisper-cli.exe"
	}

	candidates := []string{
		filepath.Join(appDir, binName),
		filepath.Join(cfgDir, "Crisper", "bin", binName),
		filepath.Join(home, ".local", "bin", "whisper-cli"),
		filepath.Join(home, "whisper.cpp", "build", "bin", "whisper-cli"),
		"/usr/local/bin/whisper-cli",
		"/opt/homebrew/bin/whisper-cli",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := checkBinary(p); err != nil {
				return "", err
			}
			cliCache.path = p
			cliCache.ok = true
			return p, nil
		}
	}

	return "", whisperNotFoundErr()
}

func whisperNotFoundErr() error {
	if runtime.GOOS == "darwin" {
		return fmt.Errorf(
			"whisper-cli not found.\n\n" +
				"Run:\n" +
				"  brew install whisper-cpp")
	}
	return fmt.Errorf(
		"whisper-cli not found.\n\n" +
			"Install from https://github.com/ggerganov/whisper.cpp")
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "whisper-cli.exe"
	}
	return "whisper-cli"
}

func (t *Transcriber) Transcribe(ctx context.Context, wavPath string, progress func(pct float64)) ([]Segment, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := os.Stat(t.model.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found: %s", t.model.Path)
	}

	cli, err := findOrInstallWhisper()
	if err != nil {
		return nil, err
	}

	outDir := filepath.Dir(wavPath)
	outBase := filepath.Join(outDir, "whisper_out")

	args := []string{
		"-m", t.model.Path,
		"-f", wavPath,
		"-t", fmt.Sprintf("%d", t.threads),
		"-osrt",
		"-of", outBase,
		"-pp",
	}
	if t.language != "" {
		args = append(args, "-l", t.language)
	}

	cmd := exec.Command(cli, args...)
	setSysProcAttr(cmd)

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("whisper-cli start: %w", err)
	}

	errc := make(chan error, 1)
	var errData []byte

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			errData = append(errData, line...)
			errData = append(errData, '\n')

			if progress != nil && strings.Contains(line, "progress =") {
				if pct := parseProgress(line); pct >= 0 {
					progress(pct)
				}
			}
		}
		errc <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		killPg(cmd)
		return nil, ctx.Err()
	case err := <-errc:
		if err != nil {
			return nil, fmt.Errorf("whisper error: %w\n%s", err, string(errData))
		}
	}

	srtPath := outBase + ".srt"
	if _, err := os.Stat(srtPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("whisper produced no output\n%s", string(errData))
	}

	segments, err := parseSRT(srtPath)
	os.Remove(srtPath)
	if err != nil {
		return nil, fmt.Errorf("parse srt: %w", err)
	}

	return segments, nil
}

func parseProgress(line string) float64 {
	parts := strings.Split(line, "progress =")
	if len(parts) < 2 {
		return -1
	}
	s := strings.TrimSpace(parts[1])
	s = strings.TrimRight(s, "% ")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return v / 100.0
}



func parseSRT(path string) ([]Segment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var segments []Segment
	var current Segment
	var readingText bool
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if current.Text != "" {
				segments = append(segments, current)
				current = Segment{}
			}
			readingText = false
			continue
		}

		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "-->")
			if len(parts) == 2 {
				current.Start = parseSRTTime(strings.TrimSpace(parts[0]))
				current.End = parseSRTTime(strings.TrimSpace(parts[1]))
			}
			readingText = true
			continue
		}

		if readingText && line != "" {
			if _, err := fmt.Sscanf(line, "%d", new(int)); err == nil {
				continue
			}
			if current.Text != "" {
				current.Text += " "
			}
			current.Text += line
		}
	}

	if current.Text != "" {
		segments = append(segments, current)
	}

	return segments, nil
}

func parseSRTTime(s string) float64 {
	var h, m, sec, ms int
	fmt.Sscanf(s, "%d:%d:%d,%d", &h, &m, &sec, &ms)
	return float64(h*3600) + float64(m*60) + float64(sec) + float64(ms)/1000.0
}
