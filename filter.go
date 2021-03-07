// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

import "sort"

type lockedChange struct {
	Entity Entity
	Add    bool
}

// Filter - container for keeping constraints-filtered entities.
type Filter struct {
	include       []uint16
	exclude       []uint16
	entities      []Entity
	entitiesMap   map[Entity]uint32
	lockedChanges []lockedChange
	lockCount     int
}

// NewFilter returns new instance of Filter.
func NewFilter(include []uint16, exclude []uint16, capacity uint32) *Filter {
	return &Filter{
		include:       include,
		exclude:       exclude,
		entities:      make([]Entity, 0, capacity),
		entitiesMap:   make(map[Entity]uint32, capacity),
		lockedChanges: make([]lockedChange, 0, capacity),
		lockCount:     0,
	}
}

// Count returns count of filtered entities.
func (f *Filter) Count() uint32 {
	return uint32(len(f.entities))
}

// EntitiesWithLock increases counter for protect filter from changes and returns filtered entities.
func (f *Filter) EntitiesWithLock() []Entity {
	f.lockCount++
	return f.entities
}

// Unlock decreases lock counter and update filter if not locked.
func (f *Filter) Unlock() {
	f.lockCount--
	if DEBUG {
		if f.lockCount < 0 {
			panic("filter lock/unlock balance broken")
		}
	}
	if f.lockCount == 0 {
		for i := 0; i < len(f.lockedChanges); i++ {
			v := f.lockedChanges[i]
			if v.Add {
				f.add(v.Entity)
			} else {
				f.remove(v.Entity)
			}
		}
		f.lockedChanges = f.lockedChanges[:0]
	}
}

// Entities returns filtered entities.
func (f *Filter) Entities() []Entity {
	return f.entities
}

func (f *Filter) add(e Entity) {
	if f.lockCount > 0 {
		f.lockedChanges = append(f.lockedChanges, lockedChange{Entity: e, Add: true})
	} else {
		if DEBUG {
			if _, ok := f.entitiesMap[e]; ok {
				panic("entity already in filter")
			}
		}
		f.entitiesMap[e] = uint32(len(f.entities))
		f.entities = append(f.entities, e)
	}
}

func (f *Filter) remove(e Entity) {
	if f.lockCount > 0 {
		f.lockedChanges = append(f.lockedChanges, lockedChange{Entity: e, Add: false})
	} else {
		if DEBUG {
			if _, ok := f.entitiesMap[e]; !ok {
				panic("entity not in filter")
			}
		}
		idx := f.entitiesMap[e]

		// without order.
		lastIdx := uint32(len(f.entities)) - 1
		if idx < lastIdx {
			f.entities[idx] = f.entities[lastIdx]
		}
		f.entities = f.entities[:lastIdx]

		// preserve order.
		// copy(f.entities[idx:], f.entities[idx+1:])

		delete(f.entitiesMap, e)
	}
}

func (f *Filter) isCompatible(entityData *EntityData) bool {
	maskLen := len(entityData.Mask)
	for _, id := range f.include {
		idx := sort.Search(maskLen, func(i int) bool { return entityData.Mask[i] >= id })
		if idx >= maskLen || entityData.Mask[idx] != id {
			return false
		}
	}
	for _, id := range f.exclude {
		idx := sort.Search(maskLen, func(i int) bool { return entityData.Mask[i] >= id })
		if idx < maskLen && entityData.Mask[idx] == id {
			return false
		}
	}
	return true
}

func (f *Filter) isCompatibleWithout(entityData *EntityData, typeID uint16) bool {
	maskLen := len(entityData.Mask)
	for _, id := range f.include {
		if id == typeID {
			return false
		}
		idx := sort.Search(maskLen, func(i int) bool { return entityData.Mask[i] >= id })
		if idx >= maskLen || entityData.Mask[idx] != id {
			return false
		}
	}
	for _, id := range f.exclude {
		if id == typeID {
			continue
		}
		idx := sort.Search(maskLen, func(i int) bool { return entityData.Mask[i] >= id })
		if idx < maskLen && entityData.Mask[idx] == id {
			return false
		}
	}
	return true
}
