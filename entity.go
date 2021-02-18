package ecs

// Entity ...
type Entity = int32

// PackedEntity ...
type PackedEntity struct {
	gen int16
	idx Entity
}
