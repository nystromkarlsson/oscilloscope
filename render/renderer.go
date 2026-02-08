package render

import (
	_ "embed"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"oscilloscope/config"
)

// Embed shader sources
//
//go:embed shaders/blur.kage
var blurShaderSource []byte

//go:embed shaders/persistence.kage
var persistenceShaderSource []byte

//go:embed shaders/composite.kage
var compositeShaderSource []byte

//go:embed shaders/crt.kage
var crtShaderSource []byte

// Renderer handles all oscilloscope rendering with phosphor simulation
type Renderer struct {
	config *config.RenderConfig

	// Rendering surfaces
	traceLayer *ebiten.Image

	// Persistence double-buffer (ping-pong for zero-allocation updates)
	persistenceFront *ebiten.Image
	persistenceBack  *ebiten.Image

	// Three blur scales for multi-scale phosphor glow
	centerBlurLayer *ebiten.Image
	nearBlurLayer   *ebiten.Image
	farBlurLayer    *ebiten.Image
	glowLayer       *ebiten.Image

	// Shared temporary buffer for separable blur passes
	blurTempH *ebiten.Image

	// CRT effects buffer
	crtBuffer *ebiten.Image

	// Trace management
	traceBuffer *TraceBuffer

	// Pre-allocated 1x1 white sprite for batched rendering
	sprite *ebiten.Image

	// Reusable vertex/index buffers for DrawTriangles
	vertices []ebiten.Vertex
	indices  []uint16

	// Shaders
	blurShader        *ebiten.Shader
	persistenceShader *ebiten.Shader
	compositeShader   *ebiten.Shader
	crtShader         *ebiten.Shader

	// CRT configuration
	crtConfig      *CRTConfig
	crtPresetIndex int // 0=off, 1=subtle, 2=flat, 3=full

	// Frame counter
	frameCount uint64
}

// NewRenderer creates a new oscilloscope renderer
func NewRenderer(cfg *config.RenderConfig) (*Renderer, error) {
	sprite := ebiten.NewImage(1, 1)
	sprite.Fill(color.White)

	r := &Renderer{
		config:           cfg,
		traceBuffer:      NewTraceBuffer(cfg.FadingFrames, cfg.MaxPointsPerFrame),
		traceLayer:       ebiten.NewImage(cfg.Width, cfg.Height),
		persistenceFront: ebiten.NewImage(cfg.Width, cfg.Height),
		persistenceBack:  ebiten.NewImage(cfg.Width, cfg.Height),
		centerBlurLayer:  ebiten.NewImage(cfg.Width, cfg.Height),
		nearBlurLayer:    ebiten.NewImage(cfg.Width, cfg.Height),
		farBlurLayer:     ebiten.NewImage(cfg.Width, cfg.Height),
		glowLayer:        ebiten.NewImage(cfg.Width, cfg.Height),
		blurTempH:        ebiten.NewImage(cfg.Width, cfg.Height),
		crtBuffer:        ebiten.NewImage(cfg.Width, cfg.Height),
		crtConfig:        SubtleCRTConfig(),
		crtPresetIndex:   1, // subtle
		sprite:           sprite,
	}

	if err := r.loadShaders(); err != nil {
		return nil, err
	}

	return r, nil
}

// loadShaders compiles all Kage shaders
func (r *Renderer) loadShaders() error {
	var err error

	r.blurShader, err = ebiten.NewShader(blurShaderSource)
	if err != nil {
		return fmt.Errorf("blur shader: %w", err)
	}

	r.persistenceShader, err = ebiten.NewShader(persistenceShaderSource)
	if err != nil {
		return fmt.Errorf("persistence shader: %w", err)
	}

	r.compositeShader, err = ebiten.NewShader(compositeShaderSource)
	if err != nil {
		return fmt.Errorf("composite shader: %w", err)
	}

	r.crtShader, err = ebiten.NewShader(crtShaderSource)
	if err != nil {
		return fmt.Errorf("crt shader: %w", err)
	}

	return nil
}

// SubmitFrame receives new trace points from the acquisition pipeline
func (r *Renderer) SubmitFrame(points []TracePoint) {
	r.traceBuffer.AddFrame(points)
}

// Update is called by Ebitengine every frame
func (r *Renderer) Update() error {
	r.frameCount++

	// Age trace buffer points every frame so points fade even when no new data arrives
	r.traceBuffer.AgePoints()

	return nil
}

// Draw renders the oscilloscope display with CRT effects
func (r *Renderer) Draw(screen *ebiten.Image) {
	// Step 1: Clear and draw current trace points
	r.traceLayer.Clear()
	r.drawTracePoints()

	// Step 2: Apply phosphor persistence (temporal decay)
	r.applyPersistence()

	// Step 3: Apply three-scale exponential blur (spatial decay)
	r.applyPhosphorGlow()

	// Step 4: Apply CRT effects
	finalImage := r.applyCRT(r.glowLayer, r.crtConfig)

	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendLighter
	screen.DrawImage(finalImage, opts)
}

// drawTracePoints renders all active points with age-based intensity
func (r *Renderer) drawTracePoints() {
	points := r.traceBuffer.GetAllPoints()
	if len(points) == 0 {
		return
	}

	r.drawPointsBatched(points)
}

// drawPointsBatched renders points using batched DrawTriangles for minimal draw calls
func (r *Renderer) drawPointsBatched(points []TracePoint) {
	// Trace color normalized to [0,1]
	cr := float32(r.config.TraceColor.R) / 255.0
	cg := float32(r.config.TraceColor.G) / 255.0
	cb := float32(r.config.TraceColor.B) / 255.0
	brightness := r.config.Brightness
	fadedIntensity := r.config.FadedIntensity
	w := r.config.Width
	h := r.config.Height

	// Ensure buffer capacity
	needed := len(points)
	if cap(r.vertices) < needed*4 {
		r.vertices = make([]ebiten.Vertex, 0, needed*4)
		r.indices = make([]uint16, 0, needed*6)
	}
	r.vertices = r.vertices[:0]
	r.indices = r.indices[:0]

	// Max 16383 points per batch (4 vertices each, must fit uint16 indices)
	const batchLimit = 16383
	pointsInBatch := 0

	for _, pt := range points {
		if pt.Intensity < fadedIntensity {
			continue
		}

		x, y := NormalizedToScreen(pt.X, pt.Y, w, h)
		if x < 0 || x >= w || y < 0 || y >= h {
			continue
		}

		// Flush batch if full
		if pointsInBatch >= batchLimit {
			r.flushBatch()
			pointsInBatch = 0
		}

		alpha := pt.Intensity * brightness
		if alpha > 1.0 {
			alpha = 1.0
		}

		fx := float32(x)
		fy := float32(y)
		vi := uint16(pointsInBatch * 4)

		r.vertices = append(r.vertices,
			ebiten.Vertex{DstX: fx, DstY: fy, SrcX: 0, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: alpha},
			ebiten.Vertex{DstX: fx + 1, DstY: fy, SrcX: 1, SrcY: 0, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: alpha},
			ebiten.Vertex{DstX: fx, DstY: fy + 1, SrcX: 0, SrcY: 1, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: alpha},
			ebiten.Vertex{DstX: fx + 1, DstY: fy + 1, SrcX: 1, SrcY: 1, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: alpha},
		)
		r.indices = append(r.indices,
			vi, vi+1, vi+2,
			vi+1, vi+3, vi+2,
		)
		pointsInBatch++
	}

	if pointsInBatch > 0 {
		r.flushBatch()
	}
}

// flushBatch submits the current vertex/index batch to the GPU
func (r *Renderer) flushBatch() {
	opts := &ebiten.DrawTrianglesOptions{}
	opts.Blend = ebiten.BlendLighter
	r.traceLayer.DrawTriangles(r.vertices, r.indices, r.sprite, opts)
	r.vertices = r.vertices[:0]
	r.indices = r.indices[:0]
}

// Resize handles window/screen size changes
func (r *Renderer) Resize(width, height int) {
	if width < 2 || height < 2 {
		return
	}
	if width == r.config.Width && height == r.config.Height {
		return
	}

	r.config.Width = width
	r.config.Height = height

	// Recreate all render targets
	r.traceLayer = ebiten.NewImage(width, height)
	r.persistenceFront = ebiten.NewImage(width, height)
	r.persistenceBack = ebiten.NewImage(width, height)
	r.centerBlurLayer = ebiten.NewImage(width, height)
	r.nearBlurLayer = ebiten.NewImage(width, height)
	r.farBlurLayer = ebiten.NewImage(width, height)
	r.glowLayer = ebiten.NewImage(width, height)
	r.blurTempH = ebiten.NewImage(width, height)
	r.crtBuffer = ebiten.NewImage(width, height)
}

// GetStats returns rendering statistics for debugging
func (r *Renderer) GetStats() RenderStats {
	return RenderStats{
		FrameCount:     r.frameCount,
		ActivePoints:   len(r.traceBuffer.points),
		BufferCapacity: r.traceBuffer.capacity,
	}
}

// RenderStats contains performance and state information
type RenderStats struct {
	FrameCount     uint64
	ActivePoints   int
	BufferCapacity int
}
