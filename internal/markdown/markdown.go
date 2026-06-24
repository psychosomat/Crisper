package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Segment struct {
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	Text      string  `json:"text"`
	SpeakerID int     `json:"speaker_id"`
}

type Result struct {
	VideoName      string    `json:"video_name"`
	Segments       []Segment `json:"segments"`
	ModelName      string    `json:"model_name"`
	ShowTimestamps bool      `json:"show_timestamps"`
}

func Generate(outputPath string, r *Result) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	var sb strings.Builder

	name := r.VideoName
	ext := filepath.Ext(name)
	title := name[:len(name)-len(ext)]

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))

	if len(r.Segments) == 0 {
		sb.WriteString("*(no speech detected)*\n")
		return os.WriteFile(outputPath, []byte(sb.String()), 0644)
	}

	merged := mergeSpeakerBlocks(r.Segments)

	for _, blk := range merged {
		if r.ShowTimestamps {
			sb.WriteString(fmt.Sprintf("## Speaker %d (%s)\n\n", blk.SpeakerID, formatTime(blk.Start)))
		} else {
			sb.WriteString(fmt.Sprintf("## Speaker %d\n\n", blk.SpeakerID))
		}

		text := strings.TrimSpace(blk.Text)
		text = collapseWhitespace(text)
		sb.WriteString(text)
		sb.WriteString("\n\n")
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

type speakerBlock struct {
	SpeakerID int
	Start     float64
	Text      string
}

func mergeSpeakerBlocks(segs []Segment) []speakerBlock {
	if len(segs) == 0 {
		return nil
	}

	var blocks []speakerBlock
	current := speakerBlock{
		SpeakerID: segs[0].SpeakerID,
		Start:     segs[0].Start,
		Text:      segs[0].Text,
	}

	for i := 1; i < len(segs); i++ {
		s := segs[i]
		if s.SpeakerID == current.SpeakerID {
			current.Text += " " + s.Text
		} else {
			blocks = append(blocks, current)
			current = speakerBlock{
				SpeakerID: s.SpeakerID,
				Start:     s.Start,
				Text:      s.Text,
			}
		}
	}
	blocks = append(blocks, current)

	return blocks
}

func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var clean []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t != "" {
			clean = append(clean, t)
		}
	}
	return strings.Join(clean, "\n")
}

func formatTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
