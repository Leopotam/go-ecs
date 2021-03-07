# LecsGO - Simple Entity Component System framework powered by Golang

Framework name is consonant with phrase "Let's GO", but with "ecs" acronym for "EntityComponentSystem".

> **Important!** Don't forget to use `DEBUG` (default) builds for development and `RELEASE` (with `-tags RELEASE`) builds in production: all internal sanitize checks works only in `DEBUG` builds and eleminated for performance reasons in `RELEASE`.

> **Important!** Ecs core is **not goroutine-friendly** and will never be! If you need multithread-processing - you should implement it on your side as part of ecs-system.

> **Important!** In development stage, not recommended for production use!

# Socials
[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# Installation

`go get github.com/leopotam/go-ecs`

# Main parts of ecs

## Component
Container for user data without / with small logic inside:
```go
type WeaponComponent struct {
    Ammo int
    GunName string
}
```

> **Important!** Don't forget to manually init all fields for each new component - they will be reset to default values on recycling to pool.

## Entity
Сontainer for components. Implemented as `ecs.Entity` for addressing internal data. You can add / remove / request components on entity through world api:
```csharp
// Creates new entity in world context.
entity := world.NewEntity ()

// GetXX returns exist component on entity or nil.
var c1 Component1
var c2 Component2
c1 = world.GetComponent1 (entity) // nil.
c2 = world.GetComponent1 (entity) // nil.

// SetXX returns exist component on entity or creates new one.
c1 = world.SetComponent1 (entity) // not nil.
c2 = world.SetComponent1 (entity) // not nil.

// DelXX removes component from entity if exist.
// If it was last component - entity will be destroyed automatically.
world.DelComponent2 (entity)

// Destroy() removes all components on entity and destroy it.
entity.Destroy ();
```

> **Important!** Entities can't be alive without components, you will get panic in `DEBUG` build if create entity and forget to add any component on it.

## System
Сontainer for logic for processing filtered entities. System struct should be compatible with `ecs.System` and one / many iterface types as `ecs.PreInitSystem`, `ecs.InitSystem`, `ecs.RunSystem`, `ecs.DestroySystem`, `ecs.PostDestroySystem`, or you will get panic in `DEBUG` build:
```go
type MySystsem struct {}

// Compatibility with ecs.System interface.
func (s *MySystem) SystemTypes() ecs.SystemType {
	// System should returns bitmask of compatible types.
	return ecs.PreInitSystemType |
		ecs.InitSystemType |
		ecs.RunSystemType |
		ecs.DestroySystemType |
		ecs.PostDestroySystemType
}

func (s *MySystem) PreInit (systems *ecs.Systems) {
	// Will be called once during ecs.Systems.Init() call and before InitSystem.Init().
}

func (s *MySystem) Init (systems *ecs.Systems) {
	// Will be called once during ecs.Systems.Init() call and after PreInitSystem.PreInit().
}

func (s *MySystem) Run (systems *ecs.Systems) {
	// Will be called on each ecs.Systems.Run() call.
}

func (s *MySystem) Destroy (systems *ecs.Systems) {
	// Will be called once during Systems.Destroy() call and before PostDestroySystem.PostDestroy().
}

func (s *MySystem) PostDestroy (systems *ecs.Systems) {
	// Will be called once during Systems.Destroy() call and after DestroySystem.Destroy().
}
```

# Special classes

## World
Root level container for all entities / components, works like isolated environment.
> Important: You should'nt touch `ecs.World` directly, but generate your custom world with required components, filters, etc. Check "[Api generation](#api-generation)" section for this.

> Important: Do not forget to call `Destroy()` method on your world when instance will not be used anymore.

## Filter
Container for keeping links to constraints-filtered entities. Any entity will be added to filter or removed from it based on components list and filter constraints. Constrains can be 2 types:
* Include - "included" component should be attached to entity.
* Exclude - "excluded" component should not be attached to entity.

Name of filter and component constraints can be declared during "[Api generation](#api-generation)".

Inside systems you can iterate over filtered entities in this way:
```go
struct Unit struct {
	Health float32
}
// Filter declared in world scheme as "Units(Unit)".
func (s *MySystem) Run(systems *ecs.Systems) {
	world := systems.World("MyWorldName").(*MyWorld)
	for _, entity := world.Units().Entities() {
		unit := world.GetUnit(entity)
		// GetUnitUnsafe(entity) can be used here
		// for performance reason due to Unit 100%
		// present on entity.
		fmt.Printf("user health: %v", unit.Health)
	}
}
```
> Important: If you know that filter entities list will be changed inside loop body (on add or remove component from constraint lists) filter should be locked before loop and unlocked after:
```go
func (s *MySystem) Run(systems *ecs.Systems) {
	world := systems.World("MyWorldName").(*MyWorld)
	for _, entity := world.Units().EntitiesWithLock() {
		unit := world.GetUnit(entity)
		fmt.Printf("user health: %v", unit.Health)
	}
	world.Units.Unlock()
	// locked filter will be updated right after last Unlock() call.
}
```

## Systems
Container for systems, and shared data between them. It's main entry point for registration and execution for systems:
```go
// Create world with reserved space for 100 entities.
world := NewGame1World(128)
// Create logic group for systems.
systems := ecs.NewSystems(nil)
systems.
	// Register system.
	Add(&EnvironmentInitSystem{}).
	Add(&UnitInitSystem{}).
	Add(&UnitMoveSystem{}).
	// etc for other systems.

	// Init all registered systems.
	Init()
// ...
// Update loop.
systems.Run()
// ...
// Destroy all registered systems.
systems.Destroy()
// Destroy all entities.
world.Destroy()
```

# Data sharing
Systems can be used as storage for shared user data and world instances:
```go
// Shared data.
type SharedData struct {
	ClientID int
}

// world name key for keeping inside ecs.Systems.
const MyWorldName = "MyWorld"

// startup init.
shared := SharedData{ClientID: 1234567}
world := NewMyWorld(128)
// Keep SharedData instance inside ecs.Systems instance.
systems := ecs.NewSystems(&shared)
systems.
	// Keep world instance inside ecs.Systems instance.
	.SetWorld(&world)
	Add(&LogInitSystem{}).
	Init()

// System.
func (s *LogInitSystem) Init(systems *ecs.Systems) {
	// Get access to SharedData.
	shared := systems.Shared().(*SharedData)
	fmt.Printf("ClientID: %d", shared.ClientID)
	// Get access to MyWorld instance.
	world := systems.World(MyWorldName).(*MyWorld)
	e := world.NewEntity()
	// ...
}
```

# Api generation
For working with custom components and filters you should create custom world type. It can be done with creating world scheme description:
```go
package game1

import "github.com/leopotam/go-ecs"

// C1 component.
type C1 struct {
	ID int
}

// C2 component.
type C2 struct {
	ID int
}

// C3 component.
type C3 struct {
	ID int
}

// game1WorldInfo is proto-interface for api generation.
// Type can be private - will not be used after generation.
type game1WorldInfo interface {
	// Only one private method should exist and should
	// return all component types that this world will contain.
	components() (
		C1,
		C2,
		C3)

	// Each public method describe filter name (you can name it as you want).
	WithC1(c1 C1)
	// In-parameters works as "Include" components constraint.
	WithC1C2(c1 C1, c2 C2)
	// Out-parameters works as "Exclude" components constraint.
	WithC1WithoutC2(c1 C1) C2
	WithC1WithoutC2C3(c1 C1) (C2, C3)
}

// Game1World is user world type that will be used later, should be public.
type Game1World struct {
    // It should contains private "world" field with type "*ecs.World" - it's important!
    // Meta tag should contains name of proto-interface type in form `ecs:"ProtoInterfaceTypeName"`
	world *ecs.World `ecs:"game1WorldInfo"`
}

// Game1WorldName is optional world name that can be used
// later as key in Systems worlds storage.
const Game1WorldName = "Game1"

// IMPORTANT! Next "go:generate" line should be added at any line of file where world type + world proto-interface placed.

//go:generate go run github.com/leopotam/go-ecs/cmd/world-gen
```

Then you can run `go generate ./...` shell command from root of project and api-file will be generated.

# License
The software released under the terms of the [MIT license](./LICENSE.md).

No support or any guarantees, no personal help. 