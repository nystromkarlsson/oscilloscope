package acquisition

import (
	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
	"oscilloscope/internal/trigger"
)

type Acquirer struct {
	Trigger          *trigger.Trigger
	lastTriggerIndex int
}

type Result struct {
	Record record.Record
	Ready  bool
}

func New(trig *trigger.Trigger) *Acquirer {
	return &Acquirer{
		Trigger:          trig,
		lastTriggerIndex: -1,
	}
}

func (a *Acquirer) Build(ring *memory.Ring) Result {
	if ring.Count() < RecordLength {
		return a.Empty()
	}

	searchStart := ring.OldestIndex() + PreSamples
	searchEnd := ring.NewestIndex()

	trig, ok := a.Trigger.Find(ring, searchStart, searchEnd)
	if !ok {
		return a.Empty()
	}

	if trig.Index-a.lastTriggerIndex < HoldOff {
		return a.Empty()
	}

	recordStart := trig.Index - PreSamples
	recordEnd := recordStart + RecordLength

	// Implement once previous trigger intex is tracked
	// if recordStart < ring.OldestIndex() {
	//   fmt.Println("Acquirer: trigger index is too close to the start of the ring buffer")
	//   return a.Empty()
	// }

	if !ring.HasRange(int(recordStart), int(recordEnd)) {
		return a.Empty()
	}

	samples, err := ring.ReadRange(recordStart, recordEnd)
	if err != nil {
		return a.Empty()
	}

	a.lastTriggerIndex = trig.Index

	return Result{
		Record: record.Record{
			Samples:       samples,
			TriggerIndex:  PreSamples,
			TriggerOffset: trig.Offset,
		},
		Ready: true,
	}
}

func (a *Acquirer) Empty() Result {
	return Result{
		Record: record.Record{},
		Ready:  false,
	}
}
