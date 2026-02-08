package source

import "math"

type Signal struct {
	frequency  float64
	amplitude  float64
	sampleRate int
	bufferSize int
}

func Sine(frequency, amplitude float64, sampleRate, bufferSize int) Signal {
	return Signal{
		frequency:  frequency,
		amplitude:  amplitude,
		sampleRate: sampleRate,
		bufferSize: bufferSize,
	}
}

func (s Signal) ValueAt(n int) float64 {
	t := float64(n) / float64(s.sampleRate)
	return s.amplitude * math.Sin(2*math.Pi*s.frequency*t)
}

func (s Signal) SampleRate() int { return s.sampleRate }
func (s Signal) BufferSize() int { return s.bufferSize }
