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
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("PortAudio init:", err)
	}
	defer func() {
		if err := portaudio.Terminate(); err != nil {
			log.Printf("PortAudio terminate: %v", err)
		}
	}()

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	ring := memory.New(memory.MemoryBufferSize)
	trig := trigger.New()
	acquirer := acquisition.New(trig)

	done := make(chan struct{})
	recordCh := make(chan record.Record, 1)

	var closeOnce sync.Once
	shutdown := func() {
		closeOnce.Do(func() {
			close(done)
		})
	}

	stream, err := audio.NewPortAudioRunner(
		ring,
		cond,
		source.SampleRate,
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

	acquirerRunner := acquisition.NewRunner(ring, acquirer, cond, recordCh, done)
	go acquirerRunner.Run()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigch
		fmt.Println("\nShutting down...")
		shutdown()
		cond.Broadcast()
	}()

	cfg := display.DefaultConfig()
	d, err := display.New(cfg, acquirer, recordCh, done, shutdown)
	if err != nil {
		log.Fatal("Display init:", err)
	}

	if err := ebiten.RunGame(d); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}

	shutdown()
	cond.Broadcast()
}
