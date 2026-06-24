package audio

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Video struct {
	Path string
	Name string
	Size int64
}

func ExtractAudio(ctx context.Context, videoPath string, tmpDir string) (string, error) {
	base := filepath.Base(videoPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	wavPath := filepath.Join(tmpDir, name+".wav")

	args := []string{
		"-y",
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
	}

	if strings.HasSuffix(strings.ToLower(ext), ".ts") {
		args = append(args, "-fflags", "+genpts")
	}

	args = append(args, wavPath)

	cmd := exec.Command("ffmpeg", args...)
	setSysProcAttr(cmd)

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ffmpeg start: %w", err)
	}

	errc := make(chan error, 1)
	var errData []byte
	go func() {
		errData, _ = io.ReadAll(stderr)
		errc <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		killPg(cmd)
		return "", ctx.Err()
	case err := <-errc:
		if err != nil {
			return "", fmt.Errorf("ffmpeg: %w\n%s", err, string(errData))
		}
	}

	return wavPath, nil
}

func ReadWAV(path string) ([]float32, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("read wav: %w", err)
	}
	if len(data) < 44 {
		return nil, 0, fmt.Errorf("wav too short: %d bytes", len(data))
	}

	bitsPerSample := int(data[34]) | int(data[35])<<8
	if bitsPerSample != 16 {
		return nil, 0, fmt.Errorf("unsupported bits per sample: %d (need 16)", bitsPerSample)
	}

	sampleRate := int(data[24]) | int(data[25])<<8 | int(data[26])<<16 | int(data[27])<<24

	numSamples := (len(data) - 44) / 2
	samples := make([]float32, numSamples)
	for i := 0; i < numSamples; i++ {
		offset := 44 + i*2
		s := int16(data[offset]) | int16(data[offset+1])<<8
		samples[i] = float32(s) / 32768.0
	}

	return samples, sampleRate, nil
}

func ReadWAVSimple(path string) ([]float32, error) {
	samples, _, err := ReadWAV(path)
	return samples, err
}

func CleanupWAV(path string) error {
	return os.Remove(path)
}


