package acquisition

import "oscilloscope/internal/source"

const milliSecond = source.SampleRate / 1000
const preTriggerRatio = 0.25

const (
	HoldOff          = milliSecond * 10
	PreSamples       = SamplesPerRecord * preTriggerRatio
	SamplesPerRecord = milliSecond * 300
)
