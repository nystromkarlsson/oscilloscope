package audio

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gordonklaus/portaudio"
	"oscilloscope/internal/memory"
)

type PortAudioRunner struct {
	Ring *memory.Ring
	Cond *sync.Cond

	stream  *portaudio.Stream
	index   int
	convBuf []float64 // pre-allocated float32→float64 conversion buffer
}

func NewPortAudioRunner(
	ring *memory.Ring,
	cond *sync.Cond,
	sampleRate float64,
	bufferSize int,
) (*PortAudioRunner, error) {
	device, err := findBlackHoleDevice()
	if err != nil {
		return nil, err
	}

	runner := &PortAudioRunner{
		Ring:    ring,
		Cond:    cond,
		convBuf: make([]float64, bufferSize),
	}

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: 1,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      sampleRate,
		FramesPerBuffer: bufferSize,
	}

	stream, err := portaudio.OpenStream(
		params,
		func(in []float32) {
			runner.process(in)
		},
	)
	if err != nil {
		return nil, err
	}

	runner.stream = stream
	return runner, nil
}

func findBlackHoleDevice() (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if strings.Contains(d.Name, "BlackHole") && d.MaxInputChannels > 0 {
			return d, nil
		}
	}

	return nil, fmt.Errorf("BlackHole device not found")
}

func (r *PortAudioRunner) process(in []float32) {
	// Convert float32→float64 into pre-allocated buffer
	buf := r.convBuf[:len(in)]
	for i, v := range in {
		buf[i] = float64(v)
	}

	// Single lock acquisition for all samples
	r.Ring.WriteBatch(r.index, buf)
	r.index += len(in)

	// Signal the acquirer that new data is available
	r.Cond.L.Lock()
	r.Cond.Broadcast()
	r.Cond.L.Unlock()
}

func (r *PortAudioRunner) Start() error { return r.stream.Start() }
func (r *PortAudioRunner) Stop() error  { return r.stream.Stop() }
