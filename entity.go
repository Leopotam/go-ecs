// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

// Entity - ID of container with data,
// cant be cached somehow, use PackedEntity instead!
type Entity = int32

// PackedEntity - packed version of Entity,
// useful for saving as cached data somewhere.
type PackedEntity struct {
	gen int16
	idx Entity
}
