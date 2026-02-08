package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/ebiten/v2"

	"oscilloscope/display"
	"oscilloscope/internal/acquisition"
	"oscilloscope/internal/audio"
	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
	"oscilloscope/internal/source"
	"oscilloscope/internal/trigger"
)

func main() {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("PortAudio init:", err)
	}
	defer func() {
		if err := portaudio.Terminate(); err != nil {
			log.Printf("PortAudio terminate: %v", err)
		}
	}()

	// Setup audio acquisition pipeline
	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ring := memory.New(memory.MemoryBufferSize)
	trig := trigger.New()
	acquirer := acquisition.New(trig)

	done := make(chan struct{})
	recordCh := make(chan record.Record, 4)

	// Safe shutdown function (can be called from multiple goroutines)
	var closeOnce sync.Once
	shutdown := func() {
		closeOnce.Do(func() {
			close(done)
		})
	}

	stream, err := audio.NewPortAudioRunner(
		ring,
		cond,
		float64(source.SampleRate),
		source.BufferSize,
	)
	if err != nil {
		log.Fatal("Audio stream:", err)
	}

	if err := stream.Start(); err != nil {
		log.Fatal("Stream start:", err)
	}
	defer func() {
		if err := stream.Stop(); err != nil {
			log.Printf("Stream stop: %v", err)
		}
	}()

	// Start acquisition goroutine
	acquirerRunner := acquisition.NewRunner(ring, acquirer, cond, recordCh, done)
	go acquirerRunner.Run()

	// Setup signal handling for graceful shutdown
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigch
		fmt.Println("\nShutting down...")
		shutdown()
		cond.Broadcast()
	}()

	// Create display
	displayCfg := display.DefaultConfig()
	game, err := display.New(recordCh, done, shutdown, displayCfg)
	if err != nil {
		log.Fatal("Display init:", err)
	}
	defer game.Close()

	// Configure Ebiten window
	ebiten.SetWindowSize(displayCfg.WindowWidth, displayCfg.WindowHeight)
	ebiten.SetWindowTitle(displayCfg.WindowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	// Run the game loop (blocks until window closes or error)
	if err := ebiten.RunGame(game); err != nil {
		log.Printf("Game error: %v", err)
	}

	// Ensure cleanup
	shutdown()
	cond.Broadcast()
}
