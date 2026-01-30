package signal

import "math"

type Sine struct {
	Frequency float64
	Amplitude float64
	Fs        int
}

func (s Sine) ValueAt(n uint64) float64 {
	t := float64(n) / float64(s.Fs)
	return s.Amplitude * math.Sin(2*math.Pi*s.Frequency*t)
}

func (s Sine) SampleRate() int {
	return s.Fs
}
