package signal

type Source interface {
	ValueAt(n uint64) float64
	SampleRate() int
}
