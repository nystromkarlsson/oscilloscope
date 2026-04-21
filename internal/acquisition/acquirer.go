package acquisition

import (
	"sync/atomic"

	"oscilloscope/internal/memory"
	"oscilloscope/internal/record"
	"oscilloscope/internal/trigger"
)

type Acquirer struct {
	Trigger          *trigger.Trigger
	HoldOff          atomic.Int64
	LastTriggerIndex int
}

type Result struct {
	Record record.Record
	Ready  bool
}

func New(trig *trigger.Trigger) *Acquirer {
	a := &Acquirer{
		Trigger:          trig,
		LastTriggerIndex: -1,
	}
	a.HoldOff.Store(int64(DefaultHoldOff))
	return a
}

func (a *Acquirer) AdjustHoldOff(delta int) {
	for {
		old := a.HoldOff.Load()
		next := max(old+int64(delta), 0)
		if a.HoldOff.CompareAndSwap(old, next) {
			return
		}
	}
}

func (a *Acquirer) Build(ring *memory.Ring) Result {
	if ring.Count() < SamplesPerRecord {
		return a.Empty()
	}

	searchStart := ring.OldestIndex() + PreSamples
	searchEnd := ring.NewestIndex() - PreSamples

	trig, ok := a.Trigger.Find(ring, searchStart, searchEnd)
	if !ok {
		return a.Empty()
	}

	if trig.Index-a.LastTriggerIndex < int(a.HoldOff.Load()) {
		return a.Empty()
	}

	recordStart := trig.Index - PreSamples
	recordEnd := recordStart + SamplesPerRecord

	if !ring.HasRange(recordStart, recordEnd) {
		return a.Empty()
	}

	samples, err := ring.ReadRange(recordStart, recordEnd)
	if err != nil {
		return a.Empty()
	}

	a.LastTriggerIndex = trig.Index

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

func (a *Acquirer) GetHoldOff() int {
	return int(a.HoldOff.Load())
}
