package sample

import (
	"testing"

	"oscilloscope/internal/signal"
)

func TestSamplerProducesCorrectNumberOfSamples(t *testing.T) {
	sine := signal.Sine{
		Frequency: 440,
		Amplitude: 1.0,
		Fs:        44100,
	}

	sampler := NewSampler(sine, BufferSize)
	samples := sampler.Step()

	if len(samples) != BufferSize {
		t.Fatalf("got %d samples, want %d", len(samples), BufferSize)
	}
}

func TestSamplerAdvancesAbsoluteIndex(t *testing.T) {
	sine := signal.Sine{
		Frequency: 440,
		Amplitude: 1.0,
		Fs:        44100,
	}

	sampler := NewSampler(sine, BufferSize)
	sampler.Step()

	if sampler.NextIndex() != BufferSize {
		t.Fatalf("after first step, index = %d, want %d", sampler.NextIndex(), BufferSize)
	}

	sampler.Step()

	if sampler.NextIndex() != BufferSize*2 {
		t.Fatalf("after second step, index = %d, want %d", sampler.NextIndex(), BufferSize*2)
	}
}

func TestSamplerProducesContiguousSamples(t *testing.T) {
	sine := signal.Sine{
		Frequency: 440,
		Amplitude: 1.0,
		Fs:        44100,
	}

	sampler := NewSampler(sine, BufferSize)
	sampler.Step()
	second := sampler.Step()

	// the first sample of the second batch must equal
	// the value at the correct absolute index.
	expected := sine.ValueAt(BufferSize)

	if second[0] != expected {
		t.Fatalf("non-contiguous sampling: got %f, want %f", second[0], expected)
	}
}

func TestSamplerIsDeterministic(t *testing.T) {
	sine := signal.Sine{
		Frequency: 440,
		Amplitude: 1.0,
		Fs:        44100,
	}

	sampler := NewSampler(sine, BufferSize)
	a := sampler.Step()
	b := sampler.Step()

	// reset and re-run
	sampler2 := NewSampler(sine, BufferSize)
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
