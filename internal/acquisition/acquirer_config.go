package acquisition

import "oscilloscope/internal/source"

const milliSecond = source.SampleRate / 1000
const preTriggerRatio = 0.1

const (
	DefaultHoldOff   = SamplesPerRecord - PreSamples
	Length           = 60.0
	PreSamples       = SamplesPerRecord * preTriggerRatio
	SamplesPerRecord = milliSecond * Length
)
