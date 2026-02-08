package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// CRTConfig holds CRT effect parameters matching the reference shader
type CRTConfig struct {
	Enabled bool

	// Curvature (barrel distortion)
	CurvatureAmount float32 // 0.0 = flat, 1.0 = full curve like reference

	// Vignette (edge darkening)
	VignetteIntensity float32 // Multiplier for vignette effect

	// Scanlines
	ScanlineIntensity float32 // 0.0 = no scanlines, 1.0 = full scanlines

	// Chromatic aberration (RGB separation)
	ChromaticAberration float32 // Amount of color fringing

	// Brightness and color
	Brightness float32    // Overall brightness (reference uses 2.8)
	ColorTint  [3]float32 // RGB color tint
}

// DefaultCRTConfig returns settings matching the reference shader
func DefaultCRTConfig() *CRTConfig {
	return &CRTConfig{
		Enabled:             true,
		CurvatureAmount:     1.0,                          // Full curvature like reference
		VignetteIntensity:   1.0,                          // Full vignette
		ScanlineIntensity:   1.0,                          // Full scanlines
		ChromaticAberration: 1.0,                          // Default chromatic aberration
		Brightness:          2.8,                          // From reference shader
		ColorTint:           [3]float32{0.95, 1.05, 0.95}, // Slight green tint from reference
	}
}

// FlatCRTConfig returns settings for a flat screen with CRT effects
func FlatCRTConfig() *CRTConfig {
	cfg := DefaultCRTConfig()
	cfg.CurvatureAmount = 0.0 // No barrel distortion
	return cfg
}

// SubtleCRTConfig returns subtle CRT effects
func SubtleCRTConfig() *CRTConfig {
	return &CRTConfig{
		Enabled:             true,
		CurvatureAmount:     1.0,                          // Slight curve
		VignetteIntensity:   0.5,                          // Subtle vignette
		ScanlineIntensity:   0.1,                          // Subtle scanlines
		ChromaticAberration: 0.5,                          // Subtle aberration
		Brightness:          2.8,                          // From reference shader
		ColorTint:           [3]float32{0.95, 1.05, 0.95}, // Slight green tint from reference
	}
}

// applyCRT applies CRT effects using the pre-allocated crtBuffer
func (r *Renderer) applyCRT(src *ebiten.Image, crtCfg *CRTConfig) *ebiten.Image {
	if crtCfg == nil || !crtCfg.Enabled {
		return src
	}

	r.crtBuffer.Clear()

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]interface{}{
		"CurvatureAmount":     crtCfg.CurvatureAmount,
		"VignetteIntensity":   crtCfg.VignetteIntensity,
		"ScanlineIntensity":   crtCfg.ScanlineIntensity,
		"ChromaticAberration": crtCfg.ChromaticAberration,
		"Brightness":          crtCfg.Brightness,
		"ColorTint":           []float32{crtCfg.ColorTint[0], crtCfg.ColorTint[1], crtCfg.ColorTint[2]},
	}

	r.crtBuffer.DrawRectShader(r.config.Width, r.config.Height, r.crtShader, opts)

	return r.crtBuffer
}

// EnableCRT enables CRT effects with reference shader settings
func (r *Renderer) EnableCRT() {
	r.crtConfig = DefaultCRTConfig()
	r.crtPresetIndex = 3
}

// DisableCRT disables CRT effects
func (r *Renderer) DisableCRT() {
	if r.crtConfig != nil {
		r.crtConfig.Enabled = false
	}
	r.crtPresetIndex = 0
}

// SetCRTConfig sets custom CRT configuration
func (r *Renderer) SetCRTConfig(cfg *CRTConfig) {
	r.crtConfig = cfg
}

// GetCRTConfig returns current CRT configuration
func (r *Renderer) GetCRTConfig() *CRTConfig {
	return r.crtConfig
}

// CycleCRTPreset cycles through CRT presets: Off -> Subtle -> Flat -> Full -> Off
func (r *Renderer) CycleCRTPreset() {
	r.crtPresetIndex = (r.crtPresetIndex + 1) % 4
	switch r.crtPresetIndex {
	case 0: // Off
		r.crtConfig = &CRTConfig{Enabled: false}
	case 1: // Subtle
		r.crtConfig = SubtleCRTConfig()
	case 2: // Flat
		r.crtConfig = FlatCRTConfig()
	case 3: // Full
		r.crtConfig = DefaultCRTConfig()
	}
}

// AdjustCurvature adjusts barrel distortion at runtime
func (r *Renderer) AdjustCurvature(delta float32) {
	if r.crtConfig == nil {
		r.crtConfig = DefaultCRTConfig()
	}

	r.crtConfig.CurvatureAmount += delta

	if r.crtConfig.CurvatureAmount < 0.0 {
		r.crtConfig.CurvatureAmount = 0.0
	}
	if r.crtConfig.CurvatureAmount > 1.0 {
		r.crtConfig.CurvatureAmount = 1.0
	}
}

// AdjustScanlines adjusts scanline intensity at runtime
func (r *Renderer) AdjustScanlines(delta float32) {
	if r.crtConfig == nil {
		r.crtConfig = DefaultCRTConfig()
	}

	r.crtConfig.ScanlineIntensity += delta

	if r.crtConfig.ScanlineIntensity < 0.0 {
		r.crtConfig.ScanlineIntensity = 0.0
	}
	if r.crtConfig.ScanlineIntensity > 1.0 {
		r.crtConfig.ScanlineIntensity = 1.0
	}
}

// AdjustVignette adjusts vignette intensity at runtime
func (r *Renderer) AdjustVignette(delta float32) {
	if r.crtConfig == nil {
		r.crtConfig = DefaultCRTConfig()
	}

	r.crtConfig.VignetteIntensity += delta

	if r.crtConfig.VignetteIntensity < 0.0 {
		r.crtConfig.VignetteIntensity = 0.0
	}
	if r.crtConfig.VignetteIntensity > 1.0 {
		r.crtConfig.VignetteIntensity = 1.0
	}
}
