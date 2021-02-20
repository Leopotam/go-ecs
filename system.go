// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package ecs

import "fmt"

// System - base interface for all systems.
type System interface {
	SystemTypes() SystemType
}

// PreInitSystem - interface for PreInit() systems.
type PreInitSystem interface {
	PreInit(systems *Systems)
}

// InitSystem - interface for Init() systems.
type InitSystem interface {
	Init(systems *Systems)
}

// RunSystem - interface for Run() systems.
type RunSystem interface {
	Run(systems *Systems)
}

// DestroySystem - interface for Destroy() systems.
type DestroySystem interface {
	Destroy(systems *Systems)
}

// PostDestroySystem - interface for PostDestroy() systems.
type PostDestroySystem interface {
	PostDestroy(systems *Systems)
}

// SystemType - bit flags for
// supported types definition.
type SystemType uint8

const (
	// PreInitSystemType declares PreInitSystem support.
	PreInitSystemType SystemType = 1 << iota
	// InitSystemType declares InitSystem support.
	InitSystemType
	// RunSystemType declares RunSystem support.
	RunSystemType
	// DestroySystemType declares DestroySystem support.
	DestroySystemType
	// PostDestroySystemType declares PostDestroySystem support.
	PostDestroySystemType
)

// Systems - container for systems, worlds, shared data.
type Systems struct {
	preInitSystems     []PreInitSystem
	initSystems        []InitSystem
	runSystems         []RunSystem
	destroySystems     []DestroySystem
	postDestroySystems []PostDestroySystem
	worlds             map[string]CustomWorld
	shared             interface{}
}

// NewSystems returns new instance of Systems.
func NewSystems(shared interface{}) *Systems {
	return &Systems{worlds: make(map[string]CustomWorld), shared: shared}
}

// World returns instance of user world saved with SetWorld().
func (s *Systems) World(key string) CustomWorld {
	w, _ := s.worlds[key]
	return w
}

// SetWorld saves instance of user world to use later inside systems.
func (s *Systems) SetWorld(key string, world CustomWorld) *Systems {
	if world != nil {
		s.worlds[key] = world
	} else {
		delete(s.worlds, key)
	}
	return s
}

// Shared returns optional shared user data.
func (s *Systems) Shared() interface{} {
	return s.shared
}

// Add registers user system based on SystemType flags.
// System should implements interface for all requested types.
func (s *Systems) Add(system System) *Systems {
	types := system.SystemTypes()
	if DEBUG && types == 0 {
		panic("system doesnt requested any SystemType support")
	}
	if types&PreInitSystemType != 0 {
		if DEBUG {
			if _, ok := system.(PreInitSystem); !ok {
				panic(`system requested PreInitSystemType but not implemented "PreInitSystem"`)
			}
		}
		s.preInitSystems = append(s.preInitSystems, system.(PreInitSystem))
	}
	if types&InitSystemType != 0 {
		if DEBUG {
			if _, ok := system.(InitSystem); !ok {
				panic(`system requested InitSystemType but not implemented "InitSystem"`)
			}
		}
		s.initSystems = append(s.initSystems, system.(InitSystem))
	}
	if types&RunSystemType != 0 {
		if DEBUG {
			if _, ok := system.(RunSystem); !ok {
				panic(`system requested RunSystemType but not implemented "RunSystem"`)
			}
		}
		s.runSystems = append(s.runSystems, system.(RunSystem))
	}
	if types&DestroySystemType != 0 {
		if DEBUG {
			if _, ok := system.(DestroySystem); !ok {
				panic(`system requested DestroySystemType but not implemented "DestroySystem"`)
			}
		}
		s.destroySystems = append(s.destroySystems, system.(DestroySystem))
	}
	if types&PostDestroySystemType != 0 {
		if DEBUG {
			if _, ok := system.(PostDestroySystem); !ok {
				panic(`system requested PostDestroySystemType but not implemented "PostDestroySystem"`)
			}
		}
		s.postDestroySystems = append(s.postDestroySystems, system.(PostDestroySystem))
	}
	return s
}

// Init processes PreInitSystem / InitSystem systems execution.
func (s *Systems) Init() {
	for _, system := range s.preInitSystems {
		system.PreInit(s)
		if DEBUG {
			for _, w := range s.worlds {
				if w.InternalWorld().checkLeakedEntities() {
					panic(fmt.Sprintf("entity leak detected after %T.PreInit()", system))
				}
				if w.InternalWorld().checkLeakedFilters() {
					panic(fmt.Sprintf("filter invalid lock/unlock detected after %T.PreInit()", system))
				}
			}
		}
	}
	for _, system := range s.initSystems {
		system.Init(s)
		if DEBUG {
			for _, w := range s.worlds {
				if w.InternalWorld().checkLeakedEntities() {
					panic(fmt.Sprintf("entity leak detected after %T.Init()", system))
				}
				if w.InternalWorld().checkLeakedFilters() {
					panic(fmt.Sprintf("filter invalid lock/unlock detected after %T.Init()", system))
				}
			}
		}
	}
}

// Run processes RunSystem systems execution.
func (s *Systems) Run() {
	for _, system := range s.runSystems {
		system.Run(s)
		if DEBUG {
			for _, w := range s.worlds {
				if w.InternalWorld().checkLeakedEntities() {
					panic(fmt.Sprintf("entity leak detected after %T.Run()", system))
				}
				if w.InternalWorld().checkLeakedFilters() {
					panic(fmt.Sprintf("filter invalid lock/unlock detected after %T.Run()", system))
				}
			}
		}
	}
}

// Destroy processes DestroySystem / PostDestroySystem systems execution.
func (s *Systems) Destroy() {
	for _, system := range s.destroySystems {
		system.Destroy(s)
		if DEBUG {
			for _, w := range s.worlds {
				if w.InternalWorld().checkLeakedEntities() {
					panic(fmt.Sprintf("entity leak detected after %T.Destroy()", system))
				}
				if w.InternalWorld().checkLeakedFilters() {
					panic(fmt.Sprintf("filter invalid lock/unlock detected after %T.Destroy()", system))
				}
			}
		}
	}
	for _, system := range s.postDestroySystems {
		system.PostDestroy(s)
		if DEBUG {
			for _, w := range s.worlds {
				if w.InternalWorld().checkLeakedEntities() {
					panic(fmt.Sprintf("entity leak detected after %T.PostDestroy()", system))
				}
				if w.InternalWorld().checkLeakedFilters() {
					panic(fmt.Sprintf("filter invalid lock/unlock detected after %T.PostDestroy()", system))
				}
			}
		}
	}
}
