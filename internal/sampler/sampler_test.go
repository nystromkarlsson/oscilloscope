package sampler

import (
	"testing"

	"oscilloscope/internal/source"
)

func TestSamplerProducesCorrectNumberOfSamples(t *testing.T) {
	sine := source.Sine(440, 1.0, 44100, 4)

	sampler := New(sine, 4)
	samples := sampler.Step()

	if len(samples) != 4 {
		t.Fatalf("got %d samples, want %d", len(samples), 4)
	}
}

func TestSamplerAdvancesAbsoluteIndex(t *testing.T) {
	sine := source.Sine(440, 1.0, 44100, 4)

	sampler := New(sine, 4)
	sampler.Step()

	if sampler.Index() != 4 {
		t.Fatalf("after first step, index = %d, want %d", sampler.Index(), 4)
	}

	sampler.Step()

	if sampler.Index() != 8 {
		t.Fatalf("after second step, index = %d, want %d", sampler.Index(), 8)
	}
}

func TestSamplerProducesContiguousSamples(t *testing.T) {
	sine := source.Sine(440, 1.0, 44100, 4)

	sampler := New(sine, 4)
	sampler.Step()
	second := sampler.Step()

	// the first sample of the second batch must equal
	// the value at the correct absolute index.
	expected := sine.ValueAt(4)

	if second[0] != expected {
		t.Fatalf("non-contiguous sampling: got %f, want %f", second[0], expected)
	}
}

func TestSamplerIsDeterministic(t *testing.T) {
	sine := source.Sine(440, 1.0, 44100, 4)

	sampler := New(sine, 4)
	a := sampler.Step()
	b := sampler.Step()

	// reset and re-run.
	sampler2 := New(sine, 4)
	a2 := sampler2.Step()
	b2 := sampler2.Step()

	for i := range a {
		if a[i] != a2[i] {
			t.Fatalf("determinism failure in first batch at %d", i)
		}
	}

	for i := range b {
		if b[i] != b2[i] {
			t.Fatalf("determinism failure in second batch at %d", i)
		}
	}
}
