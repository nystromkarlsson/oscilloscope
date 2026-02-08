package acquisition

import "oscilloscope/internal/source"

const milliSecond = source.SampleRate / 1000
const preTriggerRatio = 0.5

const (
	HoldOff          = milliSecond * 1
	PreSamples       = SamplesPerRecord * preTriggerRatio
	SamplesPerRecord = milliSecond * 20
)
