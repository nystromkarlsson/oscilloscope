package trigger

type Polarity int

const (
	Positive Polarity = 1
	Negative Polarity = -1
)

const hysteresis = 0.025

const (
	Epsilon        = hysteresis
	LowerThreshold = -hysteresis
	UpperThreshold = hysteresis
)
