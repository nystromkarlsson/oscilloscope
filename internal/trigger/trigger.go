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
	dir := float64(t.Polarity)

	l := dir * t.Lower
	u := dir * t.Upper

	lower := math.Min(l, u)
	upper := math.Max(l, u)

	if end <= start+1 {
		return Result{}, false
	}

	prev, ok := ring.ReadAt(start)
	if !ok {
		return Result{}, false
	}

	armed := false

	for i := start + 1; i < end; i++ {
		curr, ok := ring.ReadAt(i)
		if !ok {
			return Result{}, false
		}

		p := prev * dir
		c := curr * dir

		if !armed {
			if c <= lower {
				armed = true
			}
		} else {
			if c >= upper {
				armed = false
			}
		}

		if p < 0 && c >= 0 {
			slope := c - p
			offset := -p / slope

			return Result{
				Index:  i - 1,
				Offset: offset,
			}, true
		}

		prev = curr
	}

	return Result{}, false
}
