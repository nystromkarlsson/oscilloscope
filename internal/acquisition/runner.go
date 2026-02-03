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

func (l *AcquirerRunner) Run() {
	for {
		l.Cond.L.Lock()
		l.Cond.Wait()
		l.Cond.L.Unlock()

		select {
		case <-l.Done:
			return
		default:
		}

		res := l.Acquirer.Build(l.Ring)
		if !res.Ready {
			continue
		}

		select {
		case l.Out <- res.Record:
		case <-l.Done:
			return
		}
	}
}
