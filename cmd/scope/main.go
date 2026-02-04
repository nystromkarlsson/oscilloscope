package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"oscilloscope/internal/acquisition"
	"oscilloscope/internal/audio"
	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
	// "oscilloscope/internal/sampler"
	"oscilloscope/internal/source"
	"oscilloscope/internal/trigger"
)

func main() {
	if err := portaudio.Initialize(); err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	// freq := 20.0
	// sine := source.Sine(
	// freq,
	// 1.0,
	// source.SampleRate,
	// source.BufferSize,
	// )

	ring := memory.New(memory.RingBufferSize)
	// samp := sampler.New(sine, source.BufferSize)

	var mu sync.Mutex
	cond := sync.NewCond(&mu)

	trig := trigger.New()
	acq := acquisition.New(trig)

	done := make(chan struct{})
	out := make(chan record.Record, 1)

	// samplerRunner := &sampler.SamplerRunner{
	// Sampler: samp,
	// Ring:    ring,
	// Cond:    cond,
	// Done:    done,
	// }
	pa, err := audio.NewPortAudioRunner(
		ring,
		cond,
		float64(source.SampleRate),
		source.BufferSize,
	)
	if err != nil {
		panic(err)
	}

	acquirerRunner := &acquisition.AcquirerRunner{
		Ring:     ring,
		Acquirer: acq,
		Cond:     cond,
		Out:      out,
		Done:     done,
	}

	if err := pa.Start(); err != nil {
		panic(err)
	}
	defer pa.Stop()

	go acquirerRunner.Run()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigch
		close(done)
		cond.Broadcast()
	}()

	for {
		select {
		case rec := <-out:
			RenderASCII(rec, SpringGreen)
		case <-done:
			return
		}
	}

}

type WaveColor string

const (
	Reset                 = "\x1b[0m"
	SpringGreen WaveColor = "\x1b[38;2;0;238;105m" // #00ee69
	Tangerine   WaveColor = "\x1b[38;2;249;131;0m" // #f98300
)

func RenderASCII(rec record.Record, color WaveColor) {
	const (
		width  = 188
		height = 48
	)

	var b strings.Builder
	b.Grow(width*height*2 + 128)
	b.WriteString("\033[H\033[2J")

	samples := rec.Samples
	step := float64(len(samples)-1) / float64(width-1)

	canvas := make([][]rune, height)
	for y := range canvas {
		row := make([]rune, width)
		for x := range row {
			row[x] = ' '
		}
		canvas[y] = row
	}

	prevY := -1
	maxSample := len(samples) - 1

	for x := 0; x < width; x++ {
		pos := float64(x) * step
		i := int(pos)

		if i >= maxSample {
			break
		}

		frac := pos - float64(i)
		v := samples[i]*(1-frac) + samples[i+1]*frac

		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}

		y := int((1 - (v+1)/2) * float64(height-1))

		if y < 0 {
			y = 0
		} else if y >= height {
			y = height - 1
		}

		if prevY >= 0 {
			from, to := prevY, y
			if from > to {
				from, to = to, from
			}

			for yy := from; yy <= to; yy++ {
				canvas[yy][x] = '*'
			}
		} else {
			canvas[y][x] = '*'
		}

		prevY = y
	}

	for _, row := range canvas {
		for _, ch := range row {
			if ch == '*' {
				b.WriteString(string(color))
				b.WriteByte('*')
				b.WriteString(Reset)
			} else {
				b.WriteRune(ch)
			}
		}
		b.WriteByte('\n')
	}

	fmt.Print(b.String())
}
