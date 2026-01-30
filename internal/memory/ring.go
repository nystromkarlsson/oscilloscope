package memory

type Ring struct {
	buf  []float64
	size uint64
	mask uint64

	oldest uint64
	newest uint64
	empty  bool
}

func NewRing(size int) *Ring {
	if size <= 0 || (size&(size-1)) != 0 {
		panic("ring size must be a power of two")
	}

	s := uint64(size)

	return &Ring{
		buf:    make([]float64, size),
		size:   s,
		mask:   s - 1,
		oldest: 0,
		newest: 0,
		empty:  true,
	}
}

func (r *Ring) WriteAt(index uint64, value float64) {
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

func (r *Ring) ReadAt(index uint64) (float64, bool) {
	if r.empty {
		return 0, false
	}

	if index < r.oldest || index > r.newest {
		return 0, false
	}

	return r.buf[index&r.mask], true
}

func (r *Ring) OldestIndex() uint64 {
	return r.oldest
}

func (r *Ring) NewestIndex() uint64 {
	return r.newest
}
