package transcribe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Crisper/internal/audio"
	"Crisper/internal/diarize"
	"Crisper/internal/markdown"
	"Crisper/internal/vad"
	"Crisper/internal/whisper"
)

type Config struct {
	Model          *whisper.Model
	Language       string
	Threads        int
	ShowTimestamps bool
	OutputDir      string
	TmpDir         string
}

type ProgressFn func(phase string, progress float64)

func Run(ctx context.Context, videoPath string, cfg Config, progress ProgressFn) (*markdown.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if cfg.TmpDir == "" {
		cfg.TmpDir = os.TempDir()
	}
	if err := os.MkdirAll(cfg.TmpDir, 0755); err != nil {
		return nil, fmt.Errorf("tmp dir: %w", err)
	}

	report(progress, "extracting", 0)
	wavPath, err := audio.ExtractAudio(ctx, videoPath, cfg.TmpDir)
	if err != nil {
		return nil, fmt.Errorf("extract audio: %w", err)
	}
	defer audio.CleanupWAV(wavPath)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	report(progress, "reading", 0.05)
	samples, _, err := audio.ReadWAV(wavPath)
	if err != nil {
		return nil, fmt.Errorf("read wav: %w", err)
	}

	report(progress, "detecting speech", 0.08)
	vadDetector := vad.NewDetector()
	vadSegments := vadDetector.Detect(samples)
	if len(vadSegments) == 0 {
		return nil, fmt.Errorf("no speech detected")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	report(progress, "labeling speakers", 0.1)
	labeler := diarize.NewLabeler()
	labeled := labeler.Label(vadSegments)

	report(progress, "transcribing", 0.15)
	transcriber := whisper.NewTranscriber(cfg.Model, cfg.Language, cfg.Threads)
	whisperSegments, err := transcriber.Transcribe(ctx, wavPath, func(whisperPct float64) {
		pct := 0.15 + whisperPct*0.55
		report(progress, "transcribing", pct)
	})
	if err != nil {
		return nil, fmt.Errorf("transcribe: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	report(progress, "filtering", 0.68)
	whisperSegments = filterByVAD(whisperSegments, vadSegments)

	report(progress, "stabilizing", 0.7)
	stable := stabilize(whisperSegments)

	report(progress, "assembling", 0.85)
	result := assemble(stable, labeled, videoPath, cfg)

	report(progress, "saving", 0.95)
	outputPath := outputFilePath(videoPath, cfg.OutputDir)
	if err := markdown.Generate(outputPath, result); err != nil {
		return nil, fmt.Errorf("save markdown: %w", err)
	}

	report(progress, "done", 1.0)
	return result, nil
}

func report(progress ProgressFn, phase string, pct float64) {
	if progress == nil {
		return
	}
	progress(phase, pct)
}

func outputFilePath(videoPath, outputDir string) string {
	base := filepath.Base(videoPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	if outputDir == "" {
		outputDir = filepath.Dir(videoPath)
	}

	return filepath.Join(outputDir, name+".md")
}

func assemble(whisperSegs []whisper.Segment, labeled []diarize.LabeledSegment, videoPath string, cfg Config) *markdown.Result {
	mdSegs := make([]markdown.Segment, len(whisperSegs))

	for i, ws := range whisperSegs {
		speakerID := 1
		segMid := (ws.Start + ws.End) / 2

		for _, ls := range labeled {
			if segMid >= ls.Start && segMid < ls.End {
				speakerID = ls.SpeakerID
				break
			}
		}

		mdSegs[i] = markdown.Segment{
			Start:     ws.Start,
			End:       ws.End,
			Text:      ws.Text,
			SpeakerID: speakerID,
		}
	}

	name := filepath.Base(videoPath)
	return &markdown.Result{
		VideoName:      name,
		Segments:       mdSegs,
		ModelName:      cfg.Model.Name,
		ShowTimestamps: cfg.ShowTimestamps,
	}
}

func stabilize(segments []whisper.Segment) []whisper.Segment {
	if len(segments) <= 1 {
		return segments
	}

	deduped := dedupOverlaps(segments)
	for i := range deduped {
		deduped[i].Text = strings.TrimSpace(deduped[i].Text)
	}
	return deduped
}

func dedupOverlaps(segments []whisper.Segment) []whisper.Segment {
	if len(segments) <= 1 {
		return segments
	}

	var result []whisper.Segment
	current := segments[0]

	for i := 1; i < len(segments); i++ {
		next := segments[i]
		overlap := current.End - next.Start

		if overlap > 0 && overlap < 3.0 {
			similarity := textSimilarity(current.Text, next.Text)
			if similarity > 0.8 {
				if len(current.Text) < len(next.Text) {
					current = next
				}
				continue
			}
		}

		result = append(result, current)
		current = next
	}
	result = append(result, current)

	return result
}

func textSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	return jaccardSimilarity(a, b)
}

func jaccardSimilarity(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	setA := make(map[string]struct{})
	for _, w := range wordsA {
		setA[w] = struct{}{}
	}

	intersection := 0
	for _, w := range wordsB {
		if _, ok := setA[w]; ok {
			intersection++
		}
	}

	union := len(setA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func filterByVAD(whisperSegs []whisper.Segment, vadSegs []vad.Segment) []whisper.Segment {
	if len(vadSegs) == 0 {
		return whisperSegs
	}

	var filtered []whisper.Segment
	for _, ws := range whisperSegs {
		wsLen := ws.End - ws.Start
		if wsLen <= 0 {
			continue
		}

		kept := false
		for _, vs := range vadSegs {
			overlapStart := max(ws.Start, vs.Start)
			overlapEnd := min(ws.End, vs.End)
			overlap := overlapEnd - overlapStart
			if overlap > 0 && overlap/wsLen > 0.3 {
				kept = true
				break
			}
		}

		if kept {
			filtered = append(filtered, ws)
		}
	}

	return filtered
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
