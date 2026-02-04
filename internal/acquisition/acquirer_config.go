package acquisition

import "oscilloscope/internal/source"

const (
	SamplesPerRecord = (source.BufferSize * 8)
	PreSamples       = SamplesPerRecord / 3
	RecordLength     = SamplesPerRecord + PreSamples
	HoldOff          = source.SampleRate / 90
)
