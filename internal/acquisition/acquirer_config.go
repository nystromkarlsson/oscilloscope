package acquisition

const (
	SamplesPerRecord = 1024
	PreSamples       = 0 // SamplesPerRecord / 2
	RecordLength     = SamplesPerRecord + PreSamples
)
