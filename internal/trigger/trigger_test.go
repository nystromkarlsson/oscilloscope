package trigger

import (
	"testing"

	"oscilloscope/internal/memory"
)

func TestRisingTrigger(t *testing.T) {
	ring := memory.NewRing(4)

	ring.WriteAt(0, -0.5)
	ring.WriteAt(1, -0.2)
	ring.WriteAt(2, 0.2)
	ring.WriteAt(3, 0.5)

	trig := &Trigger{
		Polarity: Positive,
		Lower:    LowerThreshold,
		Upper:    UpperThreshold,
		Epsilon:  Epsilon,
	}

	res, ok := trig.Find(ring, 0, 4)
	if !ok {
		t.Fatalf("expected rising trigger")
	}

	if res.Index != 1 {
		t.Fatalf("index = %d, want 2", res.Index)
	}

	if res.Offset < 0 || res.Offset >= 0.5 {
		t.Fatalf("offset = %f, want in (0,1)", res.Offset)
	}
}

func TestFallingTrigger(t *testing.T) {
	ring := memory.NewRing(4)

	ring.WriteAt(0, 0.5)
	ring.WriteAt(1, 0.2)
	ring.WriteAt(2, -0.2)
	ring.WriteAt(3, -0.5)

	trig := &Trigger{
		Polarity: Negative,
		Lower:    LowerThreshold,
		Upper:    UpperThreshold,
		Epsilon:  Epsilon,
	}

	res, ok := trig.Find(ring, 0, 4)
	if !ok {
		t.Fatalf("expected falling trigger")
	}

	if res.Index != 1 {
		t.Fatalf("index = %d, want 2", res.Index)
	}

	if res.Offset <= 0 || res.Offset >= 0.5 {
		t.Fatalf("offset = %f, want in (0,1)", res.Offset)
	}
}
