package sample

import "oscilloscope/internal/signal"

type Sampler struct {
	src        signal.Source
	bufferSize int
	nextIndex  uint64
}

func NewSampler(src signal.Source, bufferSize int) *Sampler {
	if bufferSize <= 0 {
		panic("bufferSize must be > 0")
	}

	return &Sampler{
		src:        src,
		bufferSize: bufferSize,
		nextIndex:  0,
	}
}

func (s *Sampler) Step() []float64 {
	out := make([]float64, s.bufferSize)

	for i := 0; i < s.bufferSize; i++ {
		out[i] = s.src.ValueAt(s.nextIndex)
		s.nextIndex++
	}

	return out
}

func (s *Sampler) SampleRate() int {
	return s.src.SampleRate()
}

func (s *Sampler) NextIndex() uint64 {
	return s.nextIndex
}
