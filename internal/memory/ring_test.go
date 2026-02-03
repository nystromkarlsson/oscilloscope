package memory

import "testing"

func TestRingWriteAndRead(t *testing.T) {
	r := New(4)

	a, b := 0, 1

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
	r := New(4)

	for i := 0; i < 4*2; i++ {
		r.WriteAt(i, float64(i))
	}

	if r.OldestIndex() != 4 {
		t.Fatalf("oldest = %d, want %d", r.OldestIndex(), 4)
	}

	for i := 4; i < 4*2; i++ {
		v, ok := r.ReadAt(i)
		if !ok || v != float64(i) {
			t.Fatalf("readAt(%d) = %v, %v", i, v, ok)
		}
	}

	_, ok := r.ReadAt(0)
	if ok {
		t.Fatalf("expected index to be overwritten")
	}
}

func TestRingMasking(t *testing.T) {
	r := New(4)

	r.WriteAt(8, 8)

	v, ok := r.ReadAt(8)
	if !ok || v != 8 {
		t.Fatalf("masking failed: %v, %v", v, ok)
	}
}
