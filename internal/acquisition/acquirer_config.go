package acquisition

import "oscilloscope/internal/source"

const milliSecond = source.SampleRate / 1000
const preTriggerRatio = 0.0

var (
	BPM           = 120.0
	QuarterBeatMs = convertBPM(BPM)

	SamplesPerRecord = milliSecond * QuarterBeatMs / 2

	DefaultHoldOff = SamplesPerRecord
	PreSamples     = SamplesPerRecord * preTriggerRatio
)

func convertBPM(bpm float64) float64 {
	return (60.0 / bpm) * 1000.0
}
