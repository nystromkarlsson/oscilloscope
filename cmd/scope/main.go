package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"golang.org/x/term"
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
	defer func() {
		if err := portaudio.Terminate(); err != nil {
			panic(err)
		}
	}()

	// freq := 20.0
	// sine := source.Sine(
	// freq,
	// 1.0,
	// source.SampleRate,
	// source.BufferSize,
	// )

	ring := memory.New(memory.MemoryBufferSize)
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
	defer func() {
		if err := pa.Stop(); err != nil {
			panic(err)
		}
	}()

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
			fmt.Print("\033[H\033[2J")
			RenderASCII(rec, White)
		case <-done:
			return
		}
	}
}

type WaveColor string

const Reset = "\033[0m"
const (
	AmberTrace    WaveColor = "\033[38;2;255;176;0m"   // #ffb000
	BrightLime    WaveColor = "\033[38;2;140;255;0m"   // #8cff00
	CyanTrace     WaveColor = "\033[38;2;0;255;238m"   // #00ffee
	DeepBlue      WaveColor = "\033[38;2;0;14;238m"    // #000eee
	ElectricBlue  WaveColor = "\033[38;2;0;153;255m"   // #0099ff
	Glow          WaveColor = "\033[38;2;27;253;156m"  // #1bfd9c
	MagentaDebug  WaveColor = "\033[38;2;255;68;255m"  // #ff44ff
	PhosphorGreen WaveColor = "\033[38;2;51;255;51m"   // #33ff33
	PlasmaRed     WaveColor = "\033[38;2;255;51;85m"   // #ff3355
	SoftGreen     WaveColor = "\033[38;2;102;255;153m" // #66ff99
	WarmYellow    WaveColor = "\033[38;2;255;238;85m"  // #ffee55
	White         WaveColor = "\033[38;2;241;241;241m" // #f1f1f1
)

func RenderASCII(rec record.Record, color WaveColor) {
	width, h, err := term.GetSize(int(os.Stdout.Fd()))
	height := h - 2

	if err != nil {
		panic(err)
	}

	var b strings.Builder
	b.Grow(width*height*2 + 128)

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
