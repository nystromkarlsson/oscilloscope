package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// applyPersistence handles phosphor decay using GPU shader with ping-pong buffering
func (r *Renderer) applyPersistence() {
	if !r.config.EnableFade {
		r.persistenceFront.Clear()
		r.persistenceFront.DrawImage(r.traceLayer, nil)
		return
	}

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = r.persistenceFront
	opts.Images[1] = r.traceLayer
	opts.Uniforms = map[string]interface{}{
		"DecayFactor": r.config.DecayFactor,
		"Brightness":  r.config.Brightness,
	}

	// Render to back buffer
	r.persistenceBack.Clear()
	r.persistenceBack.DrawRectShader(
		r.config.Width,
		r.config.Height,
		r.persistenceShader,
		opts,
	)

	// Swap buffers (zero-cost pointer swap)
	r.persistenceFront, r.persistenceBack = r.persistenceBack, r.persistenceFront
}

// ClearPersistence clears all accumulated phosphor
func (r *Renderer) ClearPersistence() {
	r.persistenceFront.Clear()
	r.persistenceBack.Clear()
}
