package sampler

import (
	"sync"
	"time"

	"oscilloscope/internal/memory"
)

type SamplerRunner struct {
	Sampler *Sampler
	Ring    *memory.Ring

	Cond *sync.Cond
	Done chan struct{}
}

func (r *SamplerRunner) Run() {
	stepDuration := time.Duration(r.Sampler.BufferSize()) * time.Second / time.Duration(r.Sampler.SampleRate())

	ticker := time.NewTicker(stepDuration)
	defer ticker.Stop()

	for {
		select {
		case <-r.Done:
			return
		case <-ticker.C:
		}

		buf := r.Sampler.Step()
		base := r.Sampler.Index() - len(buf)

		for i, v := range buf {
			r.Ring.WriteAt(base+i, v)
		}

		r.Cond.L.Lock()
		r.Cond.Broadcast()
		r.Cond.L.Unlock()
	}
}
