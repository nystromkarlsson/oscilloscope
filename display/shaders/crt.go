package shader

type CRTConfig struct {
	CurvatureAmount     float64
	VignetteIntensity   float64
	ScanlineIntensity   float64
	ChromaticAberration float64
	Brightness          float64
	ColorTint           [3]float64
}

func DefaultCRTConfig() CRTConfig {
	return CRTConfig{
		CurvatureAmount:     1.0,
		VignetteIntensity:   0.75,
		ScanlineIntensity:   0.75,
		ChromaticAberration: 0.25,
		Brightness:          5.0,
		ColorTint:           [3]float64{1.0, 1.0, 1.0},
	}
}

var CrtShaderSrc = []byte(`
//kage:unit pixels
package main

var CurvatureAmount float
var VignetteIntensity float
var ScanlineIntensity float
var ChromaticAberration float
var Brightness float
var ColorTint vec3

func curve(uv vec2, amount float) vec2 {
    if amount == 0.0 {
        return uv
    }
    uv = (uv - 0.5) * 2.0
    uv *= 1.0 + (0.1 * amount)
    uv.x *= (1.0 + pow(abs(uv.y)/5.0, 2.0)*amount)
    uv.y *= (1.0 + pow(abs(uv.x)/4.0, 2.0)*amount)
    uv = uv/2.0 + 0.5
    uv = uv*0.92 + 0.04
    return uv
}

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
    origin, size := imageSrcRegionOnTexture()
    q := (texCoord - origin) / size
    uv := curve(q, CurvatureAmount)

    var col vec3
    aberration := ChromaticAberration * 0.001
    col.r = imageSrc0At((uv+vec2(aberration, aberration))*size + origin).r
    col.g = imageSrc0At((uv+vec2(0.0, -aberration*2))*size + origin).g
    col.b = imageSrc0At((uv+vec2(-aberration*2, 0.0))*size + origin).b

    col.r += 0.08 * imageSrc0At((0.75*vec2(0.025, -0.027)+uv+vec2(aberration, aberration))*size+origin).r
    col.g += 0.05 * imageSrc0At((0.75*vec2(-0.022, -0.02)+uv+vec2(0.0, -aberration*2))*size+origin).g
    col.b += 0.08 * imageSrc0At((0.75*vec2(-0.02, -0.018)+uv+vec2(-aberration*2, 0.0))*size+origin).b

    col = clamp(col*0.6+0.4*col*col, 0.0, 1.0)

    vig := 16.0 * uv.x * uv.y * (1.0 - uv.x) * (1.0 - uv.y)
    col *= vec3(pow(vig, 0.3*VignetteIntensity))

    col *= ColorTint
    col *= Brightness

    scans := clamp(0.35+0.35*sin(uv.y*size.y*1.5), 0.0, 1.0)
    s := pow(scans, 1.7)
    col *= vec3(0.4 + 0.7*s*ScanlineIntensity)

    if uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0 {
        col = vec3(0.0)
    }

    shadowMask := 1.0 - 0.65*clamp((mod(texCoord.x, 2.0)-1.0)*2.0, 0.0, 1.0)
    col *= shadowMask

    return vec4(col, 1.0)
}
`)
