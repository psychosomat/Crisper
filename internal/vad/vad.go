package vad

import "math"

type Segment struct {
	Start float64
	End   float64
}

const (
	defaultSampleRate    = 16000
	defaultFrameMs       = 30
	defaultThreshold     = 0.01
	minSpeechFrames      = 3
	minSilenceFrames     = 15
	maxSilenceGap        = 0.8
)

type Detector struct {
	sampleRate       int
	frameMs          int
	threshold        float64
	minSpeechFrames  int
	minSilenceFrames int
	maxSilenceGap    float64
}

func NewDetector() *Detector {
	return &Detector{
		sampleRate:       defaultSampleRate,
		frameMs:          defaultFrameMs,
		threshold:        defaultThreshold,
		minSpeechFrames:  minSpeechFrames,
		minSilenceFrames: minSilenceFrames,
		maxSilenceGap:    maxSilenceGap,
	}
}

func (d *Detector) SetThreshold(t float64) {
	d.threshold = t
}

func (d *Detector) Detect(samples []float32) []Segment {
	frameSize := d.sampleRate * d.frameMs / 1000
	if frameSize <= 0 {
		frameSize = 480
	}

	numFrames := len(samples) / frameSize
	if numFrames == 0 {
		return nil
	}

	speechFlags := make([]bool, numFrames)
	for i := 0; i < numFrames; i++ {
		start := i * frameSize
		end := start + frameSize
		if end > len(samples) {
			end = len(samples)
		}

		energy := 0.0
		for j := start; j < end; j++ {
			energy += float64(samples[j]) * float64(samples[j])
		}
		energy /= float64(end - start)
		energy = math.Sqrt(energy)

		speechFlags[i] = energy > d.threshold
	}

	smoothed := smooth(speechFlags, d.minSpeechFrames, d.minSilenceFrames)

	segments := extractSegments(smoothed, frameSize, d.sampleRate)
	segments = mergeCloseSegments(segments, d.maxSilenceGap)

	return segments
}

func smooth(flags []bool, minSpeech, minSilence int) []bool {
	n := len(flags)
	out := make([]bool, n)
	copy(out, flags)

	for i := 0; i < n; i++ {
		if out[i] {
			speechCount := 0
			for j := i; j < n && j < i+minSpeech; j++ {
				if flags[j] {
					speechCount++
				}
			}
			if speechCount < minSpeech {
				out[i] = false
			}
		}
	}

	for i := 0; i < n; i++ {
		if !out[i] {
			silCount := 0
			for j := i; j < n && j < i+minSilence; j++ {
				if !flags[j] {
					silCount++
				}
			}
			if silCount < minSilence {
				out[i] = true
			}
		}
	}

	return out
}

func extractSegments(flags []bool, frameSize int, sampleRate int) []Segment {
	var segments []Segment
	inSpeech := false
	var start int

	for i, speech := range flags {
		if speech && !inSpeech {
			start = i
			inSpeech = true
		}
		if !speech && inSpeech {
			segments = append(segments, Segment{
				Start: float64(start*frameSize) / float64(sampleRate),
				End:   float64(i*frameSize) / float64(sampleRate),
			})
			inSpeech = false
		}
	}

	if inSpeech {
		segments = append(segments, Segment{
			Start: float64(start*frameSize) / float64(sampleRate),
			End:   float64(len(flags)*frameSize) / float64(sampleRate),
		})
	}

	return segments
}

func mergeCloseSegments(segments []Segment, maxGap float64) []Segment {
	if len(segments) <= 1 {
		return segments
	}

	var merged []Segment
	current := segments[0]

	for i := 1; i < len(segments); i++ {
		gap := segments[i].Start - current.End
		if gap < maxGap {
			current.End = segments[i].End
		} else {
			merged = append(merged, current)
			current = segments[i]
		}
	}
	merged = append(merged, current)

	return merged
}
