package trigger

import (
	"math"

	"oscilloscope/internal/memory"
)

type Result struct {
	Index  uint64
	Offset float64
}

type Trigger struct {
	Polarity Polarity
	Lower    float64
	Upper    float64
	Epsilon  float64
}

func NewTrigger() *Trigger {
	return &Trigger{
		Polarity: Positive,
		Lower:    LowerThreshold,
		Upper:    UpperThreshold,
		Epsilon:  Epsilon,
	}
}

func (t *Trigger) Find(
	ring *memory.Ring,
	start uint64,
	end uint64,
) (Result, bool) {
	if end <= start+1 {
		return Result{}, false
	}

	prev, ok := ring.ReadAt(start)
	if !ok {
		return Result{}, false
	}

	dir := float64(t.Polarity)

	l := dir * t.Lower
	u := dir * t.Upper

	lower := math.Min(l, u)
	upper := math.Max(l, u)

	for i := start + 1; i < end; i++ {
		curr, ok := ring.ReadAt(i)
		if !ok {
			return Result{}, false
		}

		p := prev * dir
		c := curr * dir

		if p < lower && c >= upper {
			slope := c - p

			if slope < t.Epsilon {
				prev = curr
				continue
			}

			alpha := (lower - p) / slope

			return Result{
				Index:  i - 1,
				Offset: alpha,
			}, true
		}

		prev = curr
	}

	return Result{}, false
}
