package record

import "math"

func (r *Record) HighPass(sampleRate float64, cutoffHz float64) {
	if len(r.Samples) < 2 || cutoffHz <= 0 {
		return
	}

	dt := 1.0 / sampleRate
	rc := 1.0 / (2 * math.Pi * cutoffHz)
	alpha := rc / (rc + dt)

	prevX := r.Samples[0]
	prevY := r.Samples[0]

	r.Samples[0] = prevY

	for i := 1; i < len(r.Samples); i++ {
		x := r.Samples[i]

		y := alpha * (prevY + x - prevX)

		r.Samples[i] = y

		prevX = x
		prevY = y
	}
}

