package render

import (
	"oscilloscope/config"
)

// ApplyPreset applies a phosphor preset to the renderer
func (r *Renderer) ApplyPreset(preset string) {
	var cfg *config.RenderConfig

	switch preset {
	case "P1":
		cfg = config.P1Config()
	case "P7":
		cfg = config.P7Config()
	case "P31":
		cfg = config.DefaultConfig()
	default:
		return
	}

	// Update configuration
	r.config = cfg

	// Note: This doesn't recreate buffers, only updates parameters
	// For full reconfiguration, use Resize() or recreate the Renderer
}

// UpdateConfig allows runtime configuration updates
func (r *Renderer) UpdateConfig(cfg *config.RenderConfig) {
	r.config = cfg
}

// GetConfig returns the current configuration
func (r *Renderer) GetConfig() *config.RenderConfig {
	return r.config
}
