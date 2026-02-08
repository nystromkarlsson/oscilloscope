package render

import "math"

// TracePoint represents a single point in the oscilloscope trace
type TracePoint struct {
	X, Y      float32 // Normalized coordinates [-1, 1]
	Intensity float32 // Brightness [0, 1]
	Age       int     // Frames since creation
}

// TraceBuffer manages point history with age-based expiry
type TraceBuffer struct {
	points       []TracePoint
	maxAge       int
	currentFrame int
	capacity     int
	fadeLUT      []float32    // precomputed fade intensity per age
	resultBuf    []TracePoint // reusable buffer for GetAllPoints
}

// NewTraceBuffer creates a buffer that maintains N frames of history
func NewTraceBuffer(framesOfHistory, maxPointsPerFrame int) *TraceBuffer {
	capacity := framesOfHistory * maxPointsPerFrame

	// Precompute fade lookup table (only maxAge distinct values)
	fadeLUT := make([]float32, framesOfHistory+1)
	for i := 0; i <= framesOfHistory; i++ {
		fadeLUT[i] = ComputeFadeIntensity(i, framesOfHistory)
	}

	return &TraceBuffer{
		points:    make([]TracePoint, 0, capacity),
		maxAge:    framesOfHistory,
		capacity:  capacity,
		fadeLUT:   fadeLUT,
		resultBuf: make([]TracePoint, 0, capacity),
	}
}

// Add new method to age points without adding new ones
func (tb *TraceBuffer) AgePoints() {
	tb.currentFrame++

	// Age all existing points and remove expired ones
	writePos := 0
	for i := range tb.points {
		tb.points[i].Age++
		if tb.points[i].Age < tb.maxAge {
			tb.points[writePos] = tb.points[i]
			writePos++
		}
	}
	tb.points = tb.points[:writePos]
}

// Modified AddFrame - now only adds new points (aging happens separately)
func (tb *TraceBuffer) AddFrame(newPoints []TracePoint) {
	// Don't age existing points here anymore - that's done in AgePoints()
	// Just add new points with age 0

	for _, pt := range newPoints {
		pt.Age = 0
		pt.Intensity = 1.0
		tb.points = append(tb.points, pt)

		// Prevent buffer overflow
		if len(tb.points) >= tb.capacity {
			break
		}
	}
}

// GetAllPoints returns all active points with precomputed fade intensity.
// The returned slice is reused between calls; do not retain it.
func (tb *TraceBuffer) GetAllPoints() []TracePoint {
	tb.resultBuf = tb.resultBuf[:0]
	for _, pt := range tb.points {
		if pt.Age < len(tb.fadeLUT) {
			pt.Intensity *= tb.fadeLUT[pt.Age]
		} else {
			pt.Intensity = 0
		}
		tb.resultBuf = append(tb.resultBuf, pt)
	}
	return tb.resultBuf
}

// Clear removes all points
func (tb *TraceBuffer) Clear() {
	tb.points = tb.points[:0]
	tb.currentFrame = 0
}

// NormalizedToScreen converts normalized coordinates to screen pixels
func NormalizedToScreen(nx, ny float32, width, height int) (int, int) {
	x := int((nx + 1.0) * float32(width) / 2.0)
	y := int((1.0 - ny) * float32(height) / 2.0) // Flip Y axis
	return x, y
}

// ComputeFadeIntensity calculates the fade multiplier for a given age
// Implements dual exponential decay from MATLAB for realistic phosphor behavior
func ComputeFadeIntensity(age, maxAge int) float32 {
	if age == 0 {
		return 1.0
	}

	// Normalize age to [0, 1]
	t := float32(age) / float32(maxAge)

	// Fast exponential decay (bright initial glow)
	fastDecay := float32(math.Pow(10, -2.0*float64(t)*2.4))

	// Slow exponential decay (long afterglow)
	slowDecay := float32(math.Pow(10, -6.0+4.4*float64(t)))

	// Combine both decay curves
	intensity := fastDecay + slowDecay

	// Clamp to reasonable range
	if intensity < 0.001 {
		return 0.001
	}
	if intensity > 1.0 {
		return 1.0
	}

	return intensity
}
