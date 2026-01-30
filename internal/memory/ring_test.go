package memory

import "testing"

func TestRingWriteAndRead(t *testing.T) {
	r := NewRing(RingBufferSize)

	a, b := uint64(0), uint64(1)

	r.WriteAt(a, float64(a))
	r.WriteAt(b, float64(b))

	v, ok := r.ReadAt(a)
	if !ok || v != float64(a) {
		t.Fatalf("readAt(0) = %v, %v", v, ok)
	}

	v, ok = r.ReadAt(b)
	if !ok || v != float64(b) {
		t.Fatalf("readAt(1) = %v, %v", v, ok)
	}
}

func TestRingOverwrite(t *testing.T) {
	r := NewRing(RingBufferSize)

	for i := uint64(0); i < RingBufferSize*2; i++ {
		r.WriteAt(i, float64(i))
	}

	if r.OldestIndex() != RingBufferSize {
		t.Fatalf("oldest = %d, want %d", r.OldestIndex(), RingBufferSize)
	}

	for i := uint64(RingBufferSize); i < RingBufferSize*2; i++ {
		v, ok := r.ReadAt(i)
		if !ok || v != float64(i) {
			t.Fatalf("readAt(%d) = %v, %v", i, v, ok)
		}
	}

	_, ok := r.ReadAt(RingBufferSize - 1)
	if ok {
		t.Fatalf("expected index to be overwritten")
	}
}

func TestRingMasking(t *testing.T) {
	r := NewRing(RingBufferSize)

	r.WriteAt(RingBufferSize, 10)

	v, ok := r.ReadAt(RingBufferSize)
	if !ok || v != 10 {
		t.Fatalf("masking failed: %v, %v", v, ok)
	}
}
