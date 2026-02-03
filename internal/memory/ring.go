package memory

import (
	"errors"
	"sync"
)

type Ring struct {
	buf  []float64
	size int
	mask int

	newest int
	oldest int
	empty  bool

	mu sync.Mutex
}

func New(size int) *Ring {
	if size <= 0 || (size&(size-1)) != 0 {
		panic("ring size must be a power of two")
	}

	return &Ring{
		buf:    make([]float64, size),
		size:   size,
		mask:   size - 1,
		oldest: 0,
		newest: 0,
		empty:  true,
	}
}

func (r *Ring) WriteAt(index int, value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf[index&r.mask] = value

	if r.empty {
		r.oldest = index
		r.newest = index
		r.empty = false
		return
	}

	if index > r.newest {
		r.newest = index
	}

	if r.newest-r.oldest+1 > r.size {
		r.oldest = r.newest - r.size + 1
	}
}

func (r *Ring) ReadAt(index int) (float64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.empty {
		return 0, false
	}

	if index < r.oldest || index > r.newest {
		return 0, false
	}

	return r.buf[index&r.mask], true
}

func (r *Ring) ReadRange(start, end int) ([]float64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.empty || start < r.oldest || end > r.newest {
		return nil, errors.New("requested range is out of bounds")
	}

	size := end - start
	samples := make([]float64, size)

	for i := 0; i < size; i++ {
		samples[i] = r.buf[(start+i)&r.mask]
	}

	return samples, nil
}

func (r *Ring) HasRange(start, end int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.empty {
		return false
	}

	if start < r.oldest || end > r.newest {
		return false
	}

	return true
}

func (r *Ring) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.empty {
		return 0
	}

	return r.newest - r.oldest + 1
}

func (r *Ring) Size() int        { return r.size }
func (r *Ring) NewestIndex() int { return r.newest }
func (r *Ring) OldestIndex() int { return r.oldest }
