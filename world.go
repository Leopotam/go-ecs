package ecs

// EntityData ...
type EntityData struct {
	Gen        int16
	Components []int32
	Mask       []int
}

// World ...
type World struct {
	Pools            []Component
	filters          []Filter
	filtersByInclude [][]*Filter
	filtersByExclude [][]*Filter
	componentsCount  int
	Entities         []EntityData
	recycledEntities IndexPool
}

// Component ...
type Component interface {
	Recycle(idx Entity)
}

// NewWorld ...
func NewWorld(pools []Component, filters []Filter) *World {
	componentsCount := len(pools)
	w := World{
		Pools:            pools,
		filters:          filters,
		filtersByInclude: make([][]*Filter, componentsCount),
		filtersByExclude: make([][]*Filter, componentsCount),
		componentsCount:  componentsCount,
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

// Destroy ...
func (w *World) Destroy() {
	for i := 0; i < len(w.Entities); i++ {
		if w.Entities[i].Gen > 0 {
			w.DelEntity(Entity(i))
		}
	}
}

// Pool ...
func (w *World) Pool(idx int) Component {
	return w.Pools[idx]
}

// Filter ...
func (w *World) Filter(idx int) *Filter {
	return &w.filters[idx]
}

// NewEntity ...
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
	return entity
}

// DelEntity ...
func (w *World) DelEntity(entity Entity) {
	entityData := &w.Entities[entity]
	gen := entityData.Gen
	for _, typeID := range entityData.Mask {
		w.UpdateFilters(entity, typeID, false)
		itemIdx := &entityData.Components[typeID]
		w.Pools[typeID].Recycle(*itemIdx)
		*itemIdx = 0
	}
	entityData.Mask = entityData.Mask[:0]
	gen++
	if gen == 0 {
		gen = 1
	}
	entityData.Gen = -gen
	w.recycledEntities.Push(entity)
}

// PackEntity ...
func (w *World) PackEntity(entity Entity) PackedEntity {
	entityData := &w.Entities[entity]
	return PackedEntity{gen: entityData.Gen, idx: entity}
}

// UnpackEntity ...
func (w *World) UnpackEntity(packedEntity PackedEntity) (Entity, bool) {
	entityData := &w.Entities[packedEntity.idx]
	if packedEntity.gen != entityData.Gen {
		return -1, false
	}
	return packedEntity.idx, true
}

// UpdateFilters ...
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
