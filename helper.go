// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

// IndexPool - pool for entity IDs.
type IndexPool struct {
	items []Entity
}

// NewIndexPool returns new instance of IndexPool.
func NewIndexPool(cap int) *IndexPool {
	return &IndexPool{
		items: make([]Entity, 0, cap),
	}
}

// Push saves index in pool for use later.
func (p *IndexPool) Push(idx Entity) {
	p.items = append(p.items, idx)
}

// Pop returns saved index from pool
// or -1 if pool is empty.
func (p *IndexPool) Pop() Entity {
	lastIdx := len(p.items) - 1
	if lastIdx < 0 {
		return -1
	}
	lastV := p.items[lastIdx]
	p.items = p.items[:lastIdx]
	return lastV
}
