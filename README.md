[![discord](https://img.shields.io/discord/404358247621853185.svg?label=enter%20to%20discord%20server&style=for-the-badge&logo=discord)](https://discord.gg/5GZVde6)

# LeoECS GO - Simple golang Entity Component System framework.

> **Important!** Don't forget to use `DEBUG` (default) builds for development and `RELEASE` (with `-tags RELEASE`) builds in production: all internal sanitize checks works only in `DEBUG` builds and eleminated for performance reasons in `RELEASE`.

> **Important!** Ecs core is **not goroutine-friendly** and will never be! If you need multithread-processing - you should implement it on your side as part of ecs-system.

> **Important!** In development stage, not recommended to production use!

# Installation

`go get github.com/leopotam/go-ecs`

# Api generation
For creating ecs-world with proper api you should describe all components and filters that should exist:
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

// GameWorld is new type that will be used as target for api generation.
type Game1World struct {
    // It should contains private "world" field with type "*ecs.World" - its important!
    // Meta tag should contains name of proto-interface type in form `ecs:"ProtoInterfaceTypeName"`
	world *ecs.World `ecs:"game1WorldInfo"`
}

// MainWorld2Name is optional world name that can be used
// later as key in Systems worlds storage.
const Game1WorldName = "Game1"

// IMPORTANT! Next "go:generate" line should be added at any line of file where world type + world proto-interface placed.

//go:generate go run github.com/leopotam/go-ecs/cmd/world-gen
```
Then you can run `go generate ./...` shell command from root of project and api-file will be generated.

# Example

```go
type testSystem struct {
	// You can store local to system data here.
}

// SystemTypes describes supported ecs-system types.
func (s *testSystem) SystemTypes() ecs.SystemType {
	return ecs.InitSystemType | ecs.RunSystemType
}

func (s *testSystem) Init(systems *ecs.Systems) {
	world := *systems.World(Game1WorldName).(*Game1World)
	entity := world.NewEntity()
	c1 := world.SetC1(entity)
	c1.ID = 123
	c2 := world.SetC2(entity)
	c2.ID = 456
}

func (s *testSystem) Run(systems *ecs.Systems) {
	world := *systems.World(Game1WorldName).(*Game1World)
	for _, entity := range world.WithC1C2().Entities() {
		c1 := world.GetC1(entity)
		c2 := world.GetC2(entity)
		fmt.Printf("c1.ID=%d\n", c1.ID)
		fmt.Printf("c2.ID=%d\n", c2.ID)
	}
}
```

# License
The software released under the terms of the [MIT license](./LICENSE.md).

No support or any guarantees, no personal help. 