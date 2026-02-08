// display/adapter.go
package display

import (
	"oscilloscope/internal/record"
	"oscilloscope/render"
)

// RecordAdapter converts acquisition records to trace points for raster display
type RecordAdapter struct {
	xScale              float32
	yScale              float32
	maxPoints           int
	interpolationFactor int // How many points to create between each sample
}

// NewRecordAdapter creates a new adapter for raster mode display
func NewRecordAdapter(xScale, yScale float32, maxPoints int) *RecordAdapter {
	return &RecordAdapter{
		xScale:              xScale,
		yScale:              yScale,
		maxPoints:           maxPoints,
		interpolationFactor: 4,
	}
}

// SetInterpolation allows runtime adjustment of interpolation
func (a *RecordAdapter) SetInterpolation(factor int) {
	if factor < 1 {
		factor = 1
	}
	a.interpolationFactor = factor
}

// Convert transforms a Record into TracePoints with interpolation
func (a *RecordAdapter) Convert(rec record.Record) []render.TracePoint {
	numSamples := len(rec.Samples)
	if numSamples == 0 {
		return nil
	}

	if numSamples < 2 {
		// Can't interpolate with less than 2 points
		return a.convertDirect(rec)
	}

	// Calculate total points after interpolation
	totalPoints := (numSamples-1)*a.interpolationFactor + 1

	// Check if we need to downsample even after interpolation
	step := 1
	if totalPoints > a.maxPoints {
		step = totalPoints / a.maxPoints
	}

	points := make([]render.TracePoint, 0, totalPoints/step)

	// Interpolate between each pair of samples
	for i := 0; i < numSamples-1; i++ {
		y0 := float32(rec.Samples[i])
		y1 := float32(rec.Samples[i+1])

		// Create interpolated points between sample i and i+1
		for j := 0; j < a.interpolationFactor; j++ {
			if (i*a.interpolationFactor+j)%step != 0 {
				continue // Skip if downsampling
			}

			// Interpolation factor [0.0, 1.0]
			t := float32(j) / float32(a.interpolationFactor)

			// Linear interpolation
			y := y0 + t*(y1-y0)

			// X position in normalized space
			samplePos := float32(i) + t
			x := (samplePos/float32(numSamples-1))*2.0 - 1.0

			points = append(points, render.TracePoint{
				X:         x * a.xScale,
				Y:         y * a.yScale,
				Intensity: 1.0,
				Age:       0,
			})
		}
	}

	// Add the last sample
	if (numSamples-1)%step == 0 {
		x := float32(1.0)
		y := float32(rec.Samples[numSamples-1])
		points = append(points, render.TracePoint{
			X:         x * a.xScale,
			Y:         y * a.yScale,
			Intensity: 1.0,
			Age:       0,
		})
	}

	return points
}

// convertDirect converts without interpolation (fallback)
func (a *RecordAdapter) convertDirect(rec record.Record) []render.TracePoint {
	numSamples := len(rec.Samples)

	step := 1
	if numSamples > a.maxPoints {
		step = numSamples / a.maxPoints
	}

	points := make([]render.TracePoint, 0, numSamples/step)

	for i := 0; i < numSamples; i += step {
		x := (float32(i)/float32(numSamples-1))*2.0 - 1.0
		y := float32(rec.Samples[i])

		points = append(points, render.TracePoint{
			X:         x * a.xScale,
			Y:         y * a.yScale,
			Intensity: 1.0,
			Age:       0,
		})
	}

	return points
}
