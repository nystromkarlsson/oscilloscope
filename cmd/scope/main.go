package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"oscilloscope/internal/acquisition"
	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
	"oscilloscope/internal/sampler"
	"oscilloscope/internal/source"
	"oscilloscope/internal/trigger"
)

func main() {
	freq := 20.0
	sine := source.Sine(
		freq,
		1.0,
		source.SampleRate,
		source.BufferSize,
	)

	ring := memory.New(memory.RingBufferSize)
	samp := sampler.New(sine, source.BufferSize)

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	trig := trigger.New()
	acq := acquisition.New(trig)

	done := make(chan struct{})
	out := make(chan record.Record, 1)

	samplerRunner := &sampler.SamplerRunner{
		Sampler: samp,
		Ring:    ring,
		Cond:    cond,
		Done:    done,
	}
	acquirerRunner := &acquisition.AcquirerRunner{
		Ring:     ring,
		Acquirer: acq,
		Cond:     cond,
		Out:      out,
		Done:     done,
	}

	go samplerRunner.Run()
	go acquirerRunner.Run()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigch
		close(done)      // signal intent
		cond.Broadcast() // wake sleepers
	}()

	for {
		select {
		case rec := <-out:
			RenderASCII(rec, 0)
		case <-done:
			return
		}
	}

}

const (
	asciiWidth  = 85
	asciiHeight = 48
)

func RenderASCII(rec record.Record, preTriggerSamples int) {
	fmt.Print("\033[H\033[2J")

	const (
		width  = 85
		height = 48
	)

	samples := rec.Samples
	step := float64(len(samples)) / float64(width)

	if step < 1 {
		step = 1
	}

	canvas := make([][]rune, height)
	for y := 0; y < height; y++ {
		canvas[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			canvas[y][x] = ' '
		}
	}

	prevY := -1
	for x := 0; x < width; x++ {
		pos := float64(x) * step
		i := int(pos)
		if i >= len(samples)-1 {
			break
		}

		frac := pos - float64(i)
		v := samples[i]*(1-frac) + samples[i+1]*frac

		if v > 1 {
			v = 1
		}
		if v < -1 {
			v = -1
		}

		y := int((1 - (v+1)/2) * float64(height-1))
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}

		canvas[y][x] = '*'

		if prevY >= 0 {
			from, to := prevY, y
			if from > to {
				from, to = to, from
			}
			for yy := from; yy <= to; yy++ {
				if canvas[yy][x] == ' ' {
					canvas[yy][x] = '|'
				}
			}
		}

		prevY = y
	}

	for _, row := range canvas {
		fmt.Println(string(row))
	}
}
