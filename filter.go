package ecs

import "sort"

// Filter ...
type Filter struct {
	include     []int
	exclude     []int
	entities    []Entity
	entitiesMap map[Entity]int
}

// NewFilter ...
func NewFilter(include []int, exclude []int, capacity int) *Filter {
	return &Filter{
		include:     include,
		exclude:     exclude,
		entities:    make([]Entity, 0, capacity),
		entitiesMap: make(map[Entity]int, capacity),
	}
}

// Entities ...
func (f *Filter) Entities() []Entity {
	return f.entities
}

// Count ...
func (f *Filter) Count() int {
	return len(f.entities)
}

func (f *Filter) add(e Entity) {
	if DEBUG {
		if _, ok := f.entitiesMap[e]; ok {
			panic("entity already in filter")
		}
	}
	f.entitiesMap[e] = len(f.entities)
	f.entities = append(f.entities, e)
}

func (f *Filter) remove(e Entity) {
	if DEBUG {
		if _, ok := f.entitiesMap[e]; !ok {
			panic("entity not in filter")
		}
	}
	idx := f.entitiesMap[e]

	// without order.
	lastIdx := len(f.entities) - 1
	if idx < lastIdx {
		f.entities[idx] = f.entities[lastIdx]
	}
	f.entities = f.entities[:lastIdx]

	// preserve order.
	// copy(f.entities[idx:], f.entities[idx+1:])

	delete(f.entitiesMap, e)
}

func (f *Filter) isCompatible(entityData *EntityData) bool {
	mask := entityData.Mask
	maskLen := len(mask)
	for _, id := range f.include {
		idx := sort.Search(maskLen, func(i int) bool { return mask[i] >= id })
		if idx >= maskLen || mask[idx] != id {
			return false
		}
	}
	for _, id := range f.exclude {
		idx := sort.Search(maskLen, func(i int) bool { return mask[i] >= id })
		if idx < maskLen && mask[idx] == id {
			return false
		}
	}
	return true
}

func (f *Filter) isCompatibleWithout(entityData *EntityData, typeID int) bool {
	mask := entityData.Mask
	maskLen := len(mask)
	for _, id := range f.include {
		if id == typeID {
			return false
		}
		idx := sort.Search(maskLen, func(i int) bool { return mask[i] >= id })
		if idx >= maskLen || mask[idx] != id {
			return false
		}
	}
	for _, id := range f.exclude {
		if id == typeID {
			continue
		}
		idx := sort.Search(maskLen, func(i int) bool { return mask[i] >= id })
		if idx < maskLen && mask[idx] == id {
			return false
		}
	}
	return true
}
