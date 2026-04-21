package display

import "image/color"

type Phosphor struct {
	Name        string
	BeamColor   color.RGBA
	Background  color.RGBA
	DecayTimeMs float64
}

var decayTime = 90.0

var (
	PhosphorP31 = Phosphor{
		Name:        "P31 Green",
		BeamColor:   color.RGBA{R: 64, G: 255, B: 64, A: 255},
		Background:  color.RGBA{R: 0, G: 18, B: 8, A: 255},
		DecayTimeMs: decayTime,
	}
	PhosphorP11 = Phosphor{
		Name:        "P11 Blue",
		BeamColor:   color.RGBA{R: 64, G: 128, B: 255, A: 255},
		Background:  color.RGBA{R: 0, G: 4, B: 18, A: 255},
		DecayTimeMs: decayTime,
	}
	PhosphorP7 = Phosphor{
		Name:        "P7 Blue/Yellow",
		BeamColor:   color.RGBA{R: 180, G: 255, B: 128, A: 255},
		Background:  color.RGBA{R: 6, G: 18, B: 4, A: 255},
		DecayTimeMs: decayTime,
	}
	PhosphorP39 = Phosphor{
		Name:        "P39 Long Green",
		BeamColor:   color.RGBA{R: 32, G: 255, B: 96, A: 255},
		Background:  color.RGBA{R: 0, G: 18, B: 4, A: 255},
		DecayTimeMs: decayTime,
	}
	PhosphorAmber = Phosphor{
		Name:        "Amber",
		BeamColor:   color.RGBA{R: 255, G: 176, B: 0, A: 255},
		Background:  color.RGBA{R: 18, G: 10, B: 0, A: 255},
		DecayTimeMs: decayTime,
	}
)
