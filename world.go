// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

// CustomWorld - interface for all user worlds.
type CustomWorld interface {
	NewEntity() Entity
	Destroy()
	DelEntity(entity Entity)
	PackEntity(entity Entity) PackedEntity
	UnpackEntity(packedEntity PackedEntity) (Entity, bool)
	InternalWorld() *World
}

// EntityData - container for keeping internal entity data.
type EntityData struct {
	Gen        int16
	Components []int32
	Mask       []int
}

// World - container for all data.
type World struct {
	Pools            []ComponentPool
	filters          []Filter
	filtersByInclude [][]*Filter
	filtersByExclude [][]*Filter
	componentsCount  int
	Entities         []EntityData
	recycledEntities IndexPool
	leakedEntities   []Entity
}

// ComponentPool - interface for all user component pools.
type ComponentPool interface {
	Recycle(idx Entity)
}

// NewWorld returns new instance of World.
func NewWorld(pools []ComponentPool, filters []Filter) *World {
	componentsCount := len(pools)
	w := World{
		Pools:            pools,
		filters:          filters,
		filtersByInclude: make([][]*Filter, componentsCount),
		filtersByExclude: make([][]*Filter, componentsCount),
		componentsCount:  componentsCount,
		Entities:         make([]EntityData, 0, 512),
		recycledEntities: *NewIndexPool(512),
	}
	if DEBUG {
		w.leakedEntities = make([]Entity, 0, 256)
	}
	for i := range w.filters {
		f := &w.filters[i]
		for _, inc := range f.include {
			list := &w.filtersByInclude[inc]
			if *list == nil {
				*list = make([]*Filter, 0, 8)
			}
			*list = append(*list, f)
		}
		for _, exc := range f.exclude {
			list := &w.filtersByExclude[exc]
			if *list == nil {
				*list = make([]*Filter, 0, 8)
			}
			*list = append(*list, f)
		}
	}
	return &w
}

// Destroy processes cleanup of data inside world.
func (w *World) Destroy() {
	for i := 0; i < len(w.Entities); i++ {
		if w.Entities[i].Gen > 0 {
			w.DelEntity(Entity(i))
		}
	}
}

// Filter returns registered filter by index.
func (w *World) Filter(idx int) *Filter {
	return &w.filters[idx]
}

// NewEntity creates and returns new entity inside world.
func (w *World) NewEntity() Entity {
	var entity Entity = w.recycledEntities.Pop()
	if entity >= 0 {
		// use recycled entity.
		entityData := &w.Entities[entity]
		entityData.Gen = -entityData.Gen
	} else {
		// create new entity.
		entity = Entity(len(w.Entities))
		entityData := EntityData{
			Gen:        1,
			Components: make([]int32, w.componentsCount),
			Mask:       make([]int, 0, w.componentsCount),
		}
		w.Entities = append(w.Entities, entityData)
	}
	if DEBUG {
		w.leakedEntities = append(w.leakedEntities, entity)
	}
	return entity
}

// DelEntity removes exist entity from world.
// All attached components will be removed first.
func (w *World) DelEntity(entity Entity) {
	entityData := &w.Entities[entity]
	gen := entityData.Gen
	for i := len(entityData.Mask) - 1; i >= 0; i-- {
		typeID := entityData.Mask[i]
		w.UpdateFilters(entity, typeID, false)
		itemIdx := &entityData.Components[typeID]
		w.Pools[typeID].Recycle(*itemIdx)
		*itemIdx = 0
		entityData.Mask = entityData.Mask[:i]
	}
	entityData.Mask = entityData.Mask[:0]
	gen++
	if gen == 0 {
		gen = 1
	}
	entityData.Gen = -gen
	w.recycledEntities.Push(entity)
}

// PackEntity packs Entity to save outside from world.
func (w *World) PackEntity(entity Entity) PackedEntity {
	entityData := &w.Entities[entity]
	return PackedEntity{gen: entityData.Gen, idx: entity}
}

// UnpackEntity tries to unpack data to Entity,
// returns unpacked entity and success of operation.
func (w *World) UnpackEntity(packedEntity PackedEntity) (Entity, bool) {
	entityData := &w.Entities[packedEntity.idx]
	if packedEntity.gen != entityData.Gen {
		return -1, false
	}
	return packedEntity.idx, true
}

// UpdateFilters updates all compatible with requested component filters.
func (w *World) UpdateFilters(e Entity, componentType int, add bool) {
	entityData := &w.Entities[e]
	includeList := w.filtersByInclude[componentType]
	excludeList := w.filtersByExclude[componentType]
	if add {
		// add component.
		for _, f := range includeList {
			if f.isCompatible(entityData) {
				if DEBUG {
					if _, ok := f.entitiesMap[e]; ok {
						panic("entity already in filter")
					}
				}
				f.add(e)
			}
		}
		for _, f := range excludeList {
			if f.isCompatibleWithout(entityData, componentType) {
				if DEBUG {
					if _, ok := f.entitiesMap[e]; !ok {
						panic("entity not in filter")
					}
				}
				f.remove(e)
			}
		}
	} else {
		// remove component.
		for _, f := range includeList {
			if f.isCompatible(entityData) {
				if DEBUG {
					if _, ok := f.entitiesMap[e]; !ok {
						panic("entity not in filter")
					}
				}
				f.remove(e)
			}
		}
		for _, f := range excludeList {
			if f.isCompatibleWithout(entityData, componentType) {
				if DEBUG {
					if _, ok := f.entitiesMap[e]; ok {
						panic("entity already in filter")
					}
				}
				f.add(e)
			}
		}
	}
}

func (w *World) checkLeakedEntities() bool {
	if len(w.leakedEntities) > 0 {
		for _, e := range w.leakedEntities {
			if w.Entities[e].Gen > 0 && len(w.Entities[e].Mask) == 0 {
				return true
			}
		}
		w.leakedEntities = w.leakedEntities[:0]
	}
	return false
}

func (w *World) checkLeakedFilters() bool {
	for _, f := range w.filters {
		if f.lockCount > 0 {
			return true
		}
	}
	return false
}
