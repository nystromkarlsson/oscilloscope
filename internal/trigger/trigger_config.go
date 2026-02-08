package trigger

type Polarity float64

const (
	Positive Polarity = 1
	Negative Polarity = -1
)

const hysteresis = 0.0625

const (
	Epsilon        = hysteresis / 8
	LowerThreshold = -hysteresis
	UpperThreshold = hysteresis
)
