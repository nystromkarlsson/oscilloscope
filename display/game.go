package display

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"oscilloscope/config"
	"oscilloscope/internal/record"
	"oscilloscope/render"
)

// Game implements ebiten.Game interface for the oscilloscope display
type Game struct {
	renderer *render.Renderer
	adapter  *RecordAdapter

	// Communication with acquisition
	recordCh <-chan record.Record
	done     <-chan struct{}
	shutdown func() // safe close via sync.Once

	// Current state
	currentRecord *record.Record
	mu            sync.RWMutex

	// Layout tracking for responsive resize
	layoutWidth  int
	layoutHeight int

	// UI state
	showStats bool
	showHelp  bool
}

// Config holds display configuration
type Config struct {
	RenderConfig *config.RenderConfig
	ShowStats    bool
	ShowHelp     bool
	WindowWidth  int
	WindowHeight int
	WindowTitle  string
}

// DefaultConfig returns sensible defaults for the display
func DefaultConfig() *Config {
	return &Config{
		RenderConfig: config.DefaultConfig(),
		ShowStats:    false,
		ShowHelp:     false,
		WindowWidth:  2000,
		WindowHeight: 1600,
		WindowTitle:  "Oscilloscope",
	}
}

// New creates a new Game instance
func New(recordCh <-chan record.Record, done <-chan struct{}, shutdown func(), cfg *Config) (*Game, error) {
	renderer, err := render.NewRenderer(cfg.RenderConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	adapter := NewRecordAdapter(
		cfg.RenderConfig.XScale,
		cfg.RenderConfig.YScale,
		cfg.RenderConfig.MaxPointsPerFrame,
	)

	return &Game{
		renderer:     renderer,
		adapter:      adapter,
		recordCh:     recordCh,
		done:         done,
		shutdown:     shutdown,
		showStats:    cfg.ShowStats,
		showHelp:     cfg.ShowHelp,
		layoutWidth:  cfg.RenderConfig.Width,
		layoutHeight: cfg.RenderConfig.Height,
	}, nil
}

// Update implements ebiten.Game
func (g *Game) Update() error {
	g.handleInput()

	// Non-blocking receive of new records
	select {
	case rec := <-g.recordCh:
		points := g.adapter.Convert(rec)

		g.renderer.SubmitFrame(points)

		g.mu.Lock()
		g.currentRecord = &rec
		g.mu.Unlock()

	case <-g.done:
		return fmt.Errorf("shutdown requested")

	default:
		// No new record - continue rendering phosphor decay
	}

	return g.renderer.Update()
}

// Draw implements ebiten.Game
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen)

	if g.showStats {
		g.drawStats(screen)
	}

	if g.showHelp {
		g.drawHelp(screen)
	}
}

// Layout implements ebiten.Game with responsive resize
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != g.layoutWidth || outsideHeight != g.layoutHeight {
		g.layoutWidth = outsideWidth
		g.layoutHeight = outsideHeight
		g.renderer.Resize(outsideWidth, outsideHeight)
	}
	return outsideWidth, outsideHeight
}

// handleInput processes keyboard controls
func (g *Game) handleInput() {
	// Toggle stats
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.showStats = !g.showStats
	}

	// Toggle help
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.showHelp = !g.showHelp
	}

	// Toggle CRT effects (cycle presets)
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.renderer.CycleCRTPreset()
	}

	// Adjust scanlines (while holding C)
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.renderer.AdjustScanlines(0.01)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.renderer.AdjustScanlines(-0.01)
		}
	}

	// Adjust vignette (while holding V)
	if ebiten.IsKeyPressed(ebiten.KeyV) {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.renderer.AdjustVignette(0.01)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.renderer.AdjustVignette(-0.01)
		}
	}

	// Adjust curvature (while holding B)
	if ebiten.IsKeyPressed(ebiten.KeyB) {
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.renderer.AdjustCurvature(0.01)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.renderer.AdjustCurvature(-0.01)
		}
	}

	// Quit
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.shutdown()
	}
}

// drawStats renders basic statistics
func (g *Game) drawStats(screen *ebiten.Image) {
	renderStats := g.renderer.GetStats()

	g.mu.RLock()
	var recordInfo string
	if g.currentRecord != nil {
		recordInfo = fmt.Sprintf("Samples: %d", len(g.currentRecord.Samples))
	} else {
		recordInfo = "Waiting for signal..."
	}
	g.mu.RUnlock()

	crtCfg := g.renderer.GetCRTConfig()
	crtStatus := "Off"
	if crtCfg != nil && crtCfg.Enabled {
		crtStatus = fmt.Sprintf("On (curve:%.1f)", crtCfg.CurvatureAmount)
	}

	msg := fmt.Sprintf(
		"FPS: %.1f\n"+
			"Points: %d / %d\n"+
			"%s\n"+
			"CRT: %s\n"+
			"Frame: %d",
		ebiten.ActualFPS(),
		renderStats.ActivePoints,
		renderStats.BufferCapacity,
		recordInfo,
		crtStatus,
		renderStats.FrameCount,
	)

	ebitenutil.DebugPrint(screen, msg)
}

// drawHelp renders keyboard shortcuts
func (g *Game) drawHelp(screen *ebiten.Image) {
	help := `
Controls:
  H - Toggle this help
  S - Toggle stats
  C - Cycle CRT presets
  C + Up/Down - Adjust scanlines
  V + Up/Down - Adjust vignette
  B + Up/Down - Adjust curvature
  Q/ESC - Quit
`
	ebitenutil.DebugPrintAt(screen, help, 10, 150)
}

// Close performs cleanup
func (g *Game) Close() error {
	return nil
}
