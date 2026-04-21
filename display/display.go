package display

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"oscilloscope/display/shaders"
	"oscilloscope/internal/acquisition"
	"oscilloscope/internal/record"
)

const (
	beamSpriteRadius = 4.0
	beamBrightness   = 0.75
	subsampleStep    = 0.25
)

type Config struct {
	WindowTitle  string
	WindowWidth  int
	WindowHeight int

	SweepDuration float64

	Phosphor          Phosphor
	PhosphorDecay     float64
	PhosphorThreshold float64

	CRTConfig shader.CRTConfig
}

func DefaultConfig() *Config {
	return &Config{
		WindowTitle:       "Oscilloscope",
		WindowWidth:       1000 * 1.15,
		WindowHeight:      800 * 1.15,
		Phosphor:          PhosphorP11,
		PhosphorThreshold: 0.025,
		CRTConfig:         shader.DefaultCRTConfig(),
	}
}

type Display struct {
	acquirer *acquisition.Acquirer
	recordCh <-chan record.Record
	done     <-chan struct{}
	shutdown func()

	layoutWidth   int
	layoutHeight  int
	sweepDuration float64

	phosphor          Phosphor
	phosphorDecay     float64
	phosphorThreshold float64

	phosphorA   *ebiten.Image
	phosphorB   *ebiten.Image
	beamSprite  *ebiten.Image
	decayShader *ebiten.Shader

	crtShader *ebiten.Shader
	crtConfig shader.CRTConfig
	crtCanvas *ebiten.Image

	currentRecord *record.Record
	sweeping      bool
	sweepPixelX   float64
	prevPixelX    float64
	prevPixelY    float64
}

func New(
	cfg *Config,
	acquirer *acquisition.Acquirer,
	recordCh <-chan record.Record,
	done <-chan struct{},
	shutdown func(),
) (*Display, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	decay, err := ebiten.NewShader(shader.DecayShaderSrc)
	if err != nil {
		return nil, fmt.Errorf("display: failed to compile decay shader: %w", err)
	}

	d := &Display{
		acquirer:          acquirer,
		recordCh:          recordCh,
		done:              done,
		shutdown:          shutdown,
		layoutWidth:       cfg.WindowWidth,
		layoutHeight:      cfg.WindowHeight,
		sweepDuration:     cfg.Phosphor.DecayTimeMs / 1000.0,
		phosphor:          cfg.Phosphor,
		phosphorDecay:     decayPerTick(cfg.Phosphor.DecayTimeMs*2, ebiten.TPS()),
		phosphorThreshold: cfg.PhosphorThreshold,
		phosphorA:         ebiten.NewImage(cfg.WindowWidth, cfg.WindowHeight),
		phosphorB:         ebiten.NewImage(cfg.WindowWidth, cfg.WindowHeight),
		beamSprite:        makeBeamSprite(beamSpriteRadius, cfg.Phosphor),
		decayShader:       decay,
	}

	crtShader, err := ebiten.NewShader(shader.CrtShaderSrc)
	if err != nil {
		return nil, fmt.Errorf("display: failed to compile CRT shader: %w", err)
	}

	d.crtShader = crtShader
	d.crtConfig = cfg.CRTConfig
	d.crtCanvas = ebiten.NewImage(cfg.WindowWidth, cfg.WindowHeight)

	ebiten.SetWindowSize(cfg.WindowWidth, cfg.WindowHeight)
	ebiten.SetWindowTitle(cfg.WindowTitle)

	return d, nil
}

func (d *Display) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		d.shutdown()
		return ebiten.Termination
	}

	d.phosphorB.Clear()

	shaderOp := &ebiten.DrawRectShaderOptions{}
	shaderOp.Images[0] = d.phosphorA
	shaderOp.Uniforms = map[string]any{
		"Decay":     d.phosphorDecay,
		"Threshold": d.phosphorThreshold,
	}

	d.phosphorB.DrawRectShader(d.layoutWidth, d.layoutHeight, d.decayShader, shaderOp)
	d.phosphorA, d.phosphorB = d.phosphorB, d.phosphorA

	if !d.sweeping {
		select {
		case rec, ok := <-d.recordCh:
			if ok {
				d.currentRecord = &rec
				d.sweepPixelX = 0
				d.prevPixelX = 0
				d.prevPixelY = float64(d.layoutHeight) / 2
				d.sweeping = true
			}
		default:
			return nil
		}
	}

	if d.sweeping {
		d.sweeping = d.depositSweepTick()
	}

	return nil
}

func (d *Display) Draw(screen *ebiten.Image) {
	d.crtCanvas.Fill(d.phosphor.Background)

	phosphorOp := &ebiten.DrawImageOptions{}
	phosphorOp.Blend = ebiten.BlendLighter
	d.crtCanvas.DrawImage(d.phosphorA, phosphorOp)

	drawGrid(d.crtCanvas, d.layoutWidth, d.layoutHeight)

	crtOp := &ebiten.DrawRectShaderOptions{}
	crtOp.Images[0] = d.crtCanvas
	crtOp.Uniforms = map[string]any{
		"CurvatureAmount":     d.crtConfig.CurvatureAmount,
		"VignetteIntensity":   d.crtConfig.VignetteIntensity,
		"ScanlineIntensity":   d.crtConfig.ScanlineIntensity,
		"ChromaticAberration": d.crtConfig.ChromaticAberration,
		"Brightness":          d.crtConfig.Brightness,
		"ColorTint":           d.crtConfig.ColorTint[:],
	}

	screen.DrawRectShader(d.layoutWidth, d.layoutHeight, d.crtShader, crtOp)
}

func (d *Display) Layout(outsideWidth, outsideHeight int) (int, int) {
	return d.layoutWidth, d.layoutHeight
}

func (d *Display) depositSweepTick() bool {
	samples := d.currentRecord.Samples
	total := len(samples)
	if total == 0 {
		return false
	}

	screenW := float64(d.layoutWidth)
	screenH := float64(d.layoutHeight)

	ticksPerSweep := d.sweepDuration * float64(ebiten.TPS())
	pxPerTick := screenW / ticksPerSweep

	curX := d.sweepPixelX + pxPerTick
	if curX > screenW {
		curX = screenW
	}

	// Maps a screen X pixel position to the nearest sample index.
	toSampleIdx := func(px float64) int {
		idx := int((px / screenW) * float64(total-1))
		if idx < 0 {
			return 0
		}
		if idx >= total {
			return total - 1
		}
		return idx
	}

	dx := curX - d.prevPixelX
	dy := sampleToScreenY(float64(samples[toSampleIdx(curX)]), screenH) - d.prevPixelY
	dist := math.Sqrt(dx*dx + dy*dy)
	steps := int(dist/subsampleStep) + 1

	depositOp := &ebiten.DrawImageOptions{}
	depositOp.Blend = ebiten.BlendLighter

	for s := 0; s <= steps; s++ {
		t := float64(s) / float64(steps)
		px := d.prevPixelX + dx*t

		// Re-sample Y at each substep's X rather than linearly interpolating
		// between endpoints. This preserves signal detail at slow timebases
		// where a single tick spans many samples.
		py := sampleToScreenY(float64(samples[toSampleIdx(px)]), screenH)

		depositBeam(d.phosphorA, d.beamSprite, px, py, depositOp)
	}

	d.prevPixelX = curX
	d.prevPixelY = sampleToScreenY(float64(samples[toSampleIdx(curX)]), screenH)
	d.sweepPixelX = curX

	return curX < screenW
}

func sampleToScreenY(sample, screenH float64) float64 {
	return (screenH / 2) - (sample * screenH / 2)
}

func depositBeam(dst, sprite *ebiten.Image, x, y float64, op *ebiten.DrawImageOptions) {
	r := float64(sprite.Bounds().Dx() / 2)
	op.GeoM.Reset()
	op.GeoM.Translate(x-r, y-r)
	dst.DrawImage(sprite, op)
}

func makeBeamSprite(radius int, p Phosphor) *ebiten.Image {
	size := radius * 2
	img := ebiten.NewImage(size, size)
	cx, cy := float64(radius), float64(radius)

	for py := range size {
		for px := range size {
			dx := float64(px) - cx
			dy := float64(py) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > float64(radius) {
				continue
			}
			t := 1.0 - (dist / float64(radius))
			brightness := beamBrightness * t * t
			img.Set(px, py, color.RGBA{
				R: uint8(float64(p.BeamColor.R) * brightness),
				G: uint8(float64(p.BeamColor.G) * brightness),
				B: uint8(float64(p.BeamColor.B) * brightness),
				A: uint8(float64(p.BeamColor.A) * brightness),
			})
		}
	}
	return img
}

func drawGrid(screen *ebiten.Image, w, h int) {
	const cols, rows = 10, 8
	gridCol := color.RGBA{R: 0, G: 0, B: 0, A: 100}

	cellW := float64(w) / cols
	cellH := float64(h) / rows

	// Vertical lines.
	for i := 1; i < cols; i++ {
		x := float32(float64(i) * cellW)
		vector.StrokeLine(screen, x, 0, x, float32(h), 2, gridCol, true)
	}

	// Horizontal lines.
	for i := 1; i < rows; i++ {
		y := float32(float64(i) * cellH)
		vector.StrokeLine(screen, 0, y, float32(w), y, 2, gridCol, true)
	}
}

func decayPerTick(decayMs float64, tps int) float64 {
	n := decayMs / 1000.0 * float64(tps)
	return 1.0 - math.Pow(0.1, 1.0/n)
}
