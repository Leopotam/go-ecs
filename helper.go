// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

// indexPool - pool for entity IDs.
type indexPool []Entity

// newIndexPool returns new instance of IndexPool.
func newIndexPool(cap int) *indexPool {
	var pool indexPool = make([]Entity, 0, cap)
	return &pool
}

// Push saves index in pool for use later.
func (p *indexPool) Push(idx Entity) {
	*p = append(*p, idx)
}

// Pop returns saved index from pool.
func (p *indexPool) Pop() (Entity, bool) {
	lastIdx := len(*p) - 1
	if lastIdx < 0 {
		return 0, false
	}
	lastV := (*p)[lastIdx]
	*p = (*p)[:lastIdx]
	return lastV, true
}

const chunkSize = 64

type chunkType uint64

// BitSet is collection of bits.
type BitSet []chunkType

// NewBitSet creates new BitSet instance.
func NewBitSet(cap uint16) BitSet {
	return make([]chunkType, (cap-1)/chunkSize+1)
}

// Clear sets all bits to 0.
func (s *BitSet) Clear() {
	for i, iMax := 0, len(*s); i < iMax; i++ {
		(*s)[i] = 0
	}
}

// Set ensures that the given bit is set in the BitSet.
func (s *BitSet) Set(i uint16) {
	(*s)[i/chunkSize] |= 1 << (i % chunkSize)
}

// Unset ensures that the given bit is cleared (not set) in the BitSet.
func (s *BitSet) Unset(i uint16) {
	if len(*s) >= int(i/chunkSize+1) {
		(*s)[i/chunkSize] &^= 1 << (i % chunkSize)
	}
}

// Get returns true if the given bit is set, false if it is cleared.
func (s *BitSet) Get(i uint16) bool {
	return (*s)[i/chunkSize]&(1<<(i%chunkSize)) != 0
}
