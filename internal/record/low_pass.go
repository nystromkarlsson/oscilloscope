package record

import "math"

func (r *Record) LowPass(sampleRate float64, cutoffHz float64) {
	if len(r.Samples) < 2 || cutoffHz <= 0 {
		return
	}

	dt := 1.0 / sampleRate
	rc := 1.0 / (2 * math.Pi * cutoffHz)
	alpha := dt / (rc + dt)

	y := r.Samples[0]
	r.Samples[0] = y

	for i := 1; i < len(r.Samples); i++ {
		x := r.Samples[i]
		y = y + alpha*(x-y)
		r.Samples[i] = y
	}
}
