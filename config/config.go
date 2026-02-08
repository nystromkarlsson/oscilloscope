package config

import "image/color"

// RenderConfig holds all rendering parameters
type RenderConfig struct {
	// Display dimensions
	Width  int
	Height int

	// Trace appearance
	Brightness float32
	TraceColor color.RGBA

	// Phosphor persistence (temporal decay)
	EnableFade     bool
	FadingFrames   int     // Number of frames in history
	FadedIntensity float32 // Minimum intensity for oldest traces
	DecayFactor    float32 // Exponential decay factor (0.0-1.0) for shader-based fading

	// Exponential blur parameters (spatial decay)
	// Three scales model different scattering distances in phosphor
	CenterScale  float32 // Tight glow around beam
	CenterRadius int     // Sample radius for center
	NearScale    float32 // Medium glow
	NearRadius   int     // Sample radius for near
	FarScale     float32 // Long-range scatter
	FarRadius    int     // Sample radius for far

	// Glow weights (relative brightness of each scale)
	CenterWeight float32
	NearWeight   float32
	FarWeight    float32

	// Performance
	MaxPointsPerFrame int

	// Scaling (from acquisition to screen coordinates)
	XScale float32
	YScale float32
}

// DefaultConfig returns physically-based defaults for P31 phosphor
// (the green phosphor commonly used in analog oscilloscopes)
func DefaultConfig() *RenderConfig {
	return &RenderConfig{
		Width:      2000,
		Height:     1600,
		Brightness: 0.6,
		TraceColor: color.RGBA{R: 0, G: 255, B: 128, A: 255},

		EnableFade:     true,
		FadingFrames:   60,    // 2 seconds (less trails)
		DecayFactor:    0.001, // Moderate exponential decay for smoother fading
		FadedIntensity: 0.001,

		// Tighter blur for sharper lines
		CenterScale:  1.0,
		CenterRadius: 3,
		NearScale:    3.0,
		NearRadius:   10,
		FarScale:     8.0,
		FarRadius:    20,

		// More weight on center for sharper trace
		CenterWeight: 0.6,
		NearWeight:   0.3,
		FarWeight:    0.4,

		MaxPointsPerFrame: 100000,
		XScale:            0.73125,
		YScale:            1.3,
	}
}

// P1Config returns config for P1 phosphor (green, very fast decay)
func P1Config() *RenderConfig {
	cfg := DefaultConfig()
	cfg.TraceColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	cfg.FadingFrames = 60 // 1 second
	cfg.FarWeight = 0.30  // Dimmer halo
	return cfg
}

// P7Config returns config for P7 phosphor (blue-white, long persistence)
func P7Config() *RenderConfig {
	cfg := DefaultConfig()
	cfg.TraceColor = color.RGBA{R: 180, G: 200, B: 255, A: 255}
	cfg.FadingFrames = 480 // 8 seconds
	cfg.FarScale = 16.0    // More scatter
	cfg.FarWeight = 0.70   // Brighter halo
	return cfg
}
