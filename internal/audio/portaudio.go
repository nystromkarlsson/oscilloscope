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

	stream *portaudio.Stream
	index  int
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
		Ring: ring,
		Cond: cond,
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

func (r *PortAudioRunner) process(out []float32) {
	r.Cond.L.Lock()
	defer r.Cond.L.Unlock()

	for _, v := range out {
		r.Ring.WriteAt(r.index, float64(v))
		r.index++
	}

	r.Cond.Broadcast()
}

func (r *PortAudioRunner) Start() error { return r.stream.Start() }
func (r *PortAudioRunner) Stop() error  { return r.stream.Stop() }
