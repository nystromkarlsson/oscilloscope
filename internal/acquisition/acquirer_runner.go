package acquisition

import (
	"sync"

	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
)

type AcquirerRunner struct {
	Ring     *memory.Ring
	Acquirer *Acquirer
	Cond     *sync.Cond

	Out  chan record.Record
	Done chan struct{}
}

func NewRunner(ring *memory.Ring, acquirer *Acquirer, cond *sync.Cond, out chan record.Record, done chan struct{}) *AcquirerRunner {
	return &AcquirerRunner{
		Ring:     ring,
		Acquirer: acquirer,
		Cond:     cond,
		Out:      out,
		Done:     done,
	}
}

func (ar *AcquirerRunner) Run() {
	for {
		ar.Cond.L.Lock()
		ar.Cond.Wait()
		ar.Cond.L.Unlock()

		select {
		case <-ar.Done:
			return
		default:
		}

		res := ar.Acquirer.Build(ar.Ring)
		if !res.Ready {
			continue
		}

		select {
		case ar.Out <- res.Record:
		default:
			<-ar.Out
			ar.Out <- res.Record
		}
	}
}
