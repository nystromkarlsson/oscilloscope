package trigger

type Polarity float64

// trigger polarity.
// +1 = rising edge
// -1 = falling edge
const (
	Positive Polarity = 1
	Negative Polarity = -1
)

// window thresholds.
const (
	Hysteresis     = 0.0625
	UpperThreshold = Hysteresis
	LowerThreshold = -Hysteresis
)

const Epsilon = 1e-9
