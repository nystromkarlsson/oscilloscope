package trigger

import (
	"math"

	"oscilloscope/internal/memory"
)

type Result struct {
	Index  int
	Offset float64
}

type Trigger struct {
	Polarity Polarity
	Lower    float64
	Upper    float64
}

func New() *Trigger {
	return &Trigger{
		Polarity: Positive,
		Lower:    LowerThreshold,
		Upper:    UpperThreshold,
	}
}

func (t *Trigger) Find(
	ring *memory.Ring,
	start int,
	end int,
) (Result, bool) {
	if end <= start+1 {
		return Result{}, false
	}

	dir := float64(t.Polarity)

	l := dir * t.Lower
	u := dir * t.Upper

	lower := math.Min(l, u)
	upper := math.Max(l, u)

	// Bulk read from ring buffer (single lock acquisition)
	samples, err := ring.ReadRange(start, end)
	if err != nil {
		return Result{}, false
	}

	armed := false

	for i := 1; i < len(samples); i++ {
		prev := samples[i-1]
		curr := samples[i]

		if !armed {
			if prev <= lower {
				armed = true
			}
		} else {
			if curr >= upper {
				armed = false
			}
		}

		// Gate zero-crossing on armed state for correct hysteresis
		if armed && prev < 0 && curr >= 0 {
			slope := curr - prev
			offset := -prev / slope

			return Result{
				Index:  start + i - 1, // Map back to ring buffer index
				Offset: offset,
			}, true
		}
	}

	return Result{}, false
}
