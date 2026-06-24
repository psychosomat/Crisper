package diarize

import "Crisper/internal/vad"

const defaultNewSpeakerGap = 1.5

type Labeler struct {
	newSpeakerGap float64
}

func NewLabeler() *Labeler {
	return &Labeler{
		newSpeakerGap: defaultNewSpeakerGap,
	}
}

type LabeledSegment struct {
	Start     float64
	End       float64
	SpeakerID int
}

func (l *Labeler) Label(vadSegments []vad.Segment) []LabeledSegment {
	if len(vadSegments) == 0 {
		return nil
	}

	var labeled []LabeledSegment
	currentSpeaker := 0

	for i, seg := range vadSegments {
		if i > 0 {
			gap := seg.Start - vadSegments[i-1].End
			if gap > l.newSpeakerGap {
				currentSpeaker++
			}
		}

		labeled = append(labeled, LabeledSegment{
			Start:     seg.Start,
			End:       seg.End,
			SpeakerID: currentSpeaker,
		})
	}

	speakerCount := currentSpeaker + 1
	speakerMap := make([]int, speakerCount)
	for i := range speakerMap {
		speakerMap[i] = i
	}

	reorderSpeakers(labeled, speakerCount)

	for i := range labeled {
		labeled[i].SpeakerID++
	}

	return labeled
}

func reorderSpeakers(segments []LabeledSegment, total int) {}

func (l *Labeler) SetGap(gap float64) {
	l.newSpeakerGap = gap
}
