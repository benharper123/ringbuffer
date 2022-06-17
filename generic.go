package ringbuffer

// example
// length: 8
// tail:   1
// head:   3
// number of elements in buffer: 2
//
// 0      1      2      3      4      5      6      7
//      tail          head

// RingT is a generic ring buffer that holds points to a generic type T
type RingT[T any] struct {
	MaxWeight int   // we guarantee that weight <= MaxWeight
	weight    int   // current weight
	items     []*T  // len(items) == len(weights). len(items) is a power of 2.
	weights   []int // weights
	tail      uint  // read from tail
	head      uint  // write into head
}

// NewRingT creates a new ring buffer with the specified maximum weight
func NewRingT[T any](maxWeight int) RingT[T] {
	return RingT[T]{
		MaxWeight: maxWeight,
	}
}

// Len returns the number of elements in the buffer
func (r *RingT[T]) Len() int {
	return int((r.head - r.tail) & r.mask())
}

// Weight returns the total weight of all items in the ring buffer
func (r *RingT[T]) Weight() int {
	return r.weight
}

// Next returns the next item in the ring
func (r *RingT[T]) Next() (haveItem bool, item *T, weight int) {
	if r.Len() == 0 {
		return false, nil, 0
	}
	t := r.tail
	r.tail = (r.tail + 1) & r.mask()
	r.weight -= r.weights[t]
	haveItem, item, weight = true, r.items[t], r.weights[t]
	r.items[t] = nil // erase item, so that the garbage collector can do it's job
	return
}

// peek provides the Tail+i element from the buffer.
// This is here for unit tests.
func (r *RingT[T]) peek(i uint) (item *T, weight int) {
	j := (r.tail + i) & r.mask()
	return r.items[j], r.weights[j]
}

// Add an item to the buffer.
// Before adding, delete enough items so that we can store this new one.
func (r *RingT[T]) Add(weight int, item *T) {
	if len(r.items) == 0 || r.Len() == len(r.items)-1 {
		// need to grow array
		newSize := len(r.items) * 2
		if newSize < 4 {
			newSize = 4
		}
		newItems := make([]*T, newSize, newSize)
		newWeights := make([]int, newSize, newSize)
		orgWeight := r.weight
		n := r.Len()
		for i := 0; i < n; i++ {
			_, item, w := r.Next()
			newItems[i] = item
			newWeights[i] = w
		}
		r.items = newItems
		r.weights = newWeights
		r.tail = 0
		r.head = uint(n)
		r.weight = orgWeight
	}

	// erase old items until we're no longer overweight
	// If this new item size exceeds MaxWeight, then we store only this item.
	for r.weight+weight > r.MaxWeight && r.Len() != 0 {
		r.Next()
	}

	r.items[r.head] = item
	r.weights[r.head] = weight
	r.weight += weight
	r.head = (r.head + 1) & r.mask()
}

func (r *RingT[T]) mask() uint {
	return uint(len(r.items)) - 1
}
