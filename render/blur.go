package render

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// applyPhosphorGlow creates multi-scale exponential glow for realistic phosphor rendering
func (r *Renderer) applyPhosphorGlow() {
	texelSize := []float32{
		1.0 / float32(r.config.Width),
		1.0 / float32(r.config.Height),
	}

	// Pass 1: Center (tight, sharp)
	r.applySeparableBlur(
		r.persistenceFront,
		r.centerBlurLayer,
		r.config.CenterScale,
		r.config.CenterRadius,
		texelSize,
	)

	// Pass 2: Near (medium glow)
	r.applySeparableBlur(
		r.persistenceFront,
		r.nearBlurLayer,
		r.config.NearScale,
		r.config.NearRadius,
		texelSize,
	)

	// Pass 3: Far (wide scatter)
	r.applySeparableBlur(
		r.persistenceFront,
		r.farBlurLayer,
		r.config.FarScale,
		r.config.FarRadius,
		texelSize,
	)

	// Combine all three scales
	r.compositeGlowLayers()
}

// applySeparableBlur performs horizontal + vertical exponential blur
func (r *Renderer) applySeparableBlur(
	src *ebiten.Image,
	dst *ebiten.Image,
	expScale float32,
	radius int,
	texelSize []float32,
) {
	// Horizontal pass
	r.blurTempH.Clear()
	optsH := &ebiten.DrawRectShaderOptions{}
	optsH.Images[0] = src
	optsH.Uniforms = map[string]interface{}{
		"Direction":        []float32{1.0, 0.0},
		"TexelSize":        texelSize,
		"BlurRadius":       float32(radius),
		"ExponentialScale": expScale,
	}
	r.blurTempH.DrawRectShader(r.config.Width, r.config.Height, r.blurShader, optsH)

	// Vertical pass
	dst.Clear()
	optsV := &ebiten.DrawRectShaderOptions{}
	optsV.Images[0] = r.blurTempH
	optsV.Uniforms = map[string]interface{}{
		"Direction":        []float32{0.0, 1.0},
		"TexelSize":        texelSize,
		"BlurRadius":       float32(radius),
		"ExponentialScale": expScale,
	}
	dst.DrawRectShader(r.config.Width, r.config.Height, r.blurShader, optsV)
}

// compositeGlowLayers combines three blur scales with weights
func (r *Renderer) compositeGlowLayers() {
	r.glowLayer.Clear()

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = r.centerBlurLayer
	opts.Images[1] = r.nearBlurLayer
	opts.Images[2] = r.farBlurLayer
	opts.Uniforms = map[string]interface{}{
		"CenterWeight": r.config.CenterWeight,
		"NearWeight":   r.config.NearWeight,
		"FarWeight":    r.config.FarWeight,
	}

	r.glowLayer.DrawRectShader(r.config.Width, r.config.Height, r.compositeShader, opts)
}
