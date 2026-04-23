package shader

var BlurShaderSrc = []byte(`
//kage:unit pixels
package main

var Radius float
var Intensity float

func gaussian(x float, sigma float) float {
	return exp(-(x * x) / (2.0 * sigma * sigma))
}

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	c := vec4(0)
	total := 0.0
	sigma := Radius / 2.0

	// Fixed loop bound of 16 covers any radius up to 16px.
	// Samples beyond Radius are skipped via weight zeroing.
	for i := -16; i <= 16; i++ {
		d := float(i)
		if abs(d) > Radius {
			continue
		}
		w := gaussian(d, sigma)
		c += imageSrc0At(src + vec2(d, 0)) * w
		total += w
	}

	c /= total
	return c * Intensity
}
`)
