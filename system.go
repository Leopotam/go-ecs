package ecs

// System ...
type System interface {
	SystemTypes() SystemType
}

// PreInitSystem ...
type PreInitSystem interface {
	PreInit(systems *Systems)
}

// InitSystem ...
type InitSystem interface {
	Init(systems *Systems)
}

// RunSystem ...
type RunSystem interface {
	Run(systems *Systems)
}

// DestroySystem ...
type DestroySystem interface {
	Destroy(systems *Systems)
}

// PostDestroySystem ...
type PostDestroySystem interface {
	PostDestroy(systems *Systems)
}

// SystemType ...
type SystemType uint8

const (
	// PreInitSystemType ...
	PreInitSystemType SystemType = 1 << iota
	// InitSystemType ...
	InitSystemType
	// RunSystemType ...
	RunSystemType
	// DestroySystemType ...
	DestroySystemType
	// PostDestroySystemType ...
	PostDestroySystemType
)

// Systems ...
type Systems struct {
	preInitSystems     []PreInitSystem
	initSystems        []InitSystem
	runSystems         []RunSystem
	destroySystems     []DestroySystem
	postDestroySystems []PostDestroySystem
	worlds             map[string]interface{}
	shared             interface{}
}

// NewSystems ...
func NewSystems(shared interface{}) *Systems {
	return &Systems{worlds: make(map[string]interface{}), shared: shared}
}

// World ...
func (s *Systems) World(key string) interface{} {
	w, _ := s.worlds[key]
	return w
}

// SetWorld ...
func (s *Systems) SetWorld(key string, world interface{}) *Systems {
	if world != nil {
		s.worlds[key] = world
	} else {
		delete(s.worlds, key)
	}
	return s
}

// Shared ...
func (s *Systems) Shared() interface{} {
	return s.shared
}

// Add ...
func (s *Systems) Add(system System) *Systems {
	types := system.SystemTypes()
	if DEBUG && types == 0 {
		panic("system doesnt requested any SystemType support")
	}
	if types&PreInitSystemType != 0 {
		if DEBUG {
			if _, ok := system.(PreInitSystem); !ok {
				panic("system requested PreInitSystemType but not implemented it")
			}
		}
		s.preInitSystems = append(s.preInitSystems, system.(PreInitSystem))
	}
	if types&InitSystemType != 0 {
		if DEBUG {
			if _, ok := system.(InitSystem); !ok {
				panic("system requested InitSystemType but not implemented it")
			}
		}
		s.initSystems = append(s.initSystems, system.(InitSystem))
	}
	if types&RunSystemType != 0 {
		if DEBUG {
			if _, ok := system.(RunSystem); !ok {
				panic("system requested RunSystemType but not implemented it")
			}
		}
		s.runSystems = append(s.runSystems, system.(RunSystem))
	}
	if types&DestroySystemType != 0 {
		if DEBUG {
			if _, ok := system.(DestroySystem); !ok {
				panic("system requested DestroySystemType but not implemented it")
			}
		}
		s.destroySystems = append(s.destroySystems, system.(DestroySystem))
	}
	if types&PostDestroySystemType != 0 {
		if DEBUG {
			if _, ok := system.(PostDestroySystem); !ok {
				panic("system requested PostDestroySystemType but not implemented it")
			}
		}
		s.postDestroySystems = append(s.postDestroySystems, system.(PostDestroySystem))
	}
	return s
}

// Init ...
func (s *Systems) Init() {
	for _, system := range s.preInitSystems {
		system.PreInit(s)
	}
	for _, system := range s.initSystems {
		system.Init(s)
	}
}

// Run ...
func (s *Systems) Run() {
	for _, system := range s.runSystems {
		system.Run(s)
	}
}

// Destroy ...
func (s *Systems) Destroy() {
	for _, system := range s.destroySystems {
		system.Destroy(s)
	}
	for _, system := range s.postDestroySystems {
		system.PostDestroy(s)
	}
}
