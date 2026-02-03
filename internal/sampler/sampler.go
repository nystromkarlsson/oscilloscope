package sampler

import "oscilloscope/internal/source"

type Sampler struct {
	signal source.Signal
	index  int
}

func New(signal source.Signal, bufferSize int) *Sampler {
	return &Sampler{
		signal: signal,
		index:  0,
	}
}

func (s *Sampler) Step() []float64 {
	out := make([]float64, s.signal.BufferSize())

	for i := 0; i < s.signal.BufferSize(); i++ {
		out[i] = s.signal.ValueAt(s.index)
		s.index++
	}

	return out
}

func (s *Sampler) SampleRate() int { return s.signal.SampleRate() }
func (s *Sampler) BufferSize() int { return s.signal.BufferSize() }
func (s *Sampler) Index() int      { return s.index }
