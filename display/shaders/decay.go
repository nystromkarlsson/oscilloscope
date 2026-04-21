package shader

var DecayShaderSrc = []byte(`
//kage:unit pixels
package main

var Decay float
var Threshold float

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	c := imageSrc0At(src)
	c = c * (1.0 - Decay);
	if c.r < Threshold && c.g < Threshold && c.b < Threshold {
		return vec4(0)
	}
	return c
}
`)
