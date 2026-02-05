package record

import "math"

func (r *Record) Boost(boost float64) {
	if len(r.Samples) == 0 {
		return
	}

	if boost <= 0 {
		return
	}

	multiplier := 1 + boost

	for i := range r.Samples {
		if (math.Abs(r.Samples[i]) * multiplier) > 1 {
			r.Samples[i] = 1
			continue
		}

		r.Samples[i] = r.Samples[i] * multiplier
	}
}
