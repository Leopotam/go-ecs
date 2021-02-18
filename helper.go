package ecs

// IndexPool ...
type IndexPool struct {
	items []Entity
}

// NewIndexPool ...
func NewIndexPool(cap int) *IndexPool {
	return &IndexPool{
		items: make([]Entity, 0, cap),
	}
}

// Push ...
func (p *IndexPool) Push(idx Entity) {
	p.items = append(p.items, idx)
}

// Pop ...
func (p *IndexPool) Pop() Entity {
	lastIdx := len(p.items) - 1
	if lastIdx < 0 {
		return -1
	}
	lastV := p.items[lastIdx]
	p.items = p.items[:lastIdx]
	return lastV
}
