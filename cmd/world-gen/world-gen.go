// ----------------------------------------------------------------------------
// The MIT License
// LecsGO - Entity Component System framework powered by Golang.
// Url: https://github.com/Leopotam/go-ecs
// Copyright (c) 2021 Leopotam <leopotam@gmail.com>
// ----------------------------------------------------------------------------

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type componentInfo struct {
	Name string
	Type string
}

type filterInfo struct {
	Name           string
	IncludeTypes   []string
	ExcludeTypes   []string
	IncludeIndices []string
	ExcludeIndices []string
}

type worldInfo struct {
	Name         string
	InfoTypeName string
	Components   []componentInfo
	Filters      []filterInfo
}

func newComponentInfo(typeName string) componentInfo {
	return componentInfo{
		Name: strings.ReplaceAll(strings.Title(typeName), ".", ""),
		Type: typeName,
	}
}

func main() {
	fset := token.NewFileSet()
	inPackage := os.Getenv("GOPACKAGE")
	inFileName := os.Getenv("GOFILE")
	src, err := ioutil.ReadFile(inFileName)
	if err != nil {
		panic(err)
	}
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	// imports.
	var imports []string
	for _, item := range f.Imports {
		var importData string
		if item.Name != nil {
			importData = fmt.Sprintf("%s %s", item.Name.Name, item.Path.Value)
		} else {
			importData = item.Path.Value
		}
		imports = append(imports, importData)
	}

	// find worlds.
	worlds := scanWorlds(f)
	for _, info := range worlds {
		fmt.Printf("world: %s => %s\n", info.Name, info.InfoTypeName)
	}
	for i := range worlds {
		w := &worlds[i]
		scanWorldInfo(f, w)
		validateFilters(w)
	}

	var buf bytes.Buffer
	if err := packageTemplate.Execute(&buf, struct {
		Package string
		Imports []string
		Worlds  []worldInfo
	}{
		Package: inPackage,
		Imports: imports,
		Worlds:  worlds,
	}); err != nil {
		panic(err)
	}
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(inFileName)
	outFileName := filepath.Join(dir,
		fmt.Sprintf("%s-gen.go", inFileName[:len(inFileName)-len(filepath.Ext(inFileName))]))
	w, err := os.Create(outFileName)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	w.Write(formattedCode)
}

func scanWorlds(f *ast.File) []worldInfo {
	var worlds []worldInfo
	ast.Inspect(f, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.TypeSpec:
			switch t.Type.(type) {
			case *ast.StructType:
				for _, field := range t.Type.(*ast.StructType).Fields.List {
					if field.Tag != nil {
						if tag, err := strconv.Unquote(field.Tag.Value); err == nil {
							if meta, ok := reflect.StructTag(tag).Lookup("ecs"); ok /*&& meta == "world"*/ {
								worldInfo := worldInfo{
									Name: t.Name.Name,
									// InfoTypeName: field.Type.(*ast.Ident).Name,
									InfoTypeName: meta,
								}
								worlds = append(worlds, worldInfo)
								break
							}
						}
					}
				}
			}
		}
		return true
	})
	return worlds
}

func scanWorldInfo(f *ast.File, worldInfo *worldInfo) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.TypeSpec:
			switch t.Type.(type) {
			case *ast.InterfaceType:
				if t.Name.Name == worldInfo.InfoTypeName {
					fmt.Printf("world-info found: %s\n", worldInfo.InfoTypeName)
					componentsFound := false
					for _, method := range t.Type.(*ast.InterfaceType).Methods.List {
						if len(method.Names) == 0 {
							continue
						}
						fnName := method.Names[0]
						fn := method.Type.(*ast.FuncType)
						if !fnName.IsExported() {
							if componentsFound {
								panic(fmt.Sprintf(`only one private func should be present in world "%s"`, worldInfo.Name))
							}
							worldInfo.Components = scanComponents(worldInfo, fnName.Name, fn)
							componentsFound = true
							for _, ci := range worldInfo.Components {
								fmt.Printf("component: name=%s, type=%s\n", ci.Name, ci.Type)
							}
							continue
						}
						filter := scanFilterConstraints(fn)
						filter.Name = fnName.Name
						worldInfo.Filters = append(worldInfo.Filters, filter)
					}
				}
			}
		}
		return true
	})
}

func scanComponents(w *worldInfo, name string, fn *ast.FuncType) []componentInfo {
	var components []componentInfo
	if len(fn.Params.List) > 0 {
		panic(fmt.Sprintf(`private func "%s" cant get parameters in world "%s"`, name, w.Name))
	}
	if fn.Results == nil {
		panic(fmt.Sprintf(`private func "%s" should returns components in world "%s"`, name, w.Name))
	}
	for _, par := range fn.Results.List {
		var typeName string
		switch par.Type.(type) {
		case *ast.SelectorExpr:
			sel := par.Type.(*ast.SelectorExpr)
			typeName = fmt.Sprintf("%s.%s", sel.X.(*ast.Ident).Name, sel.Sel)
		case *ast.Ident:
			typeName = par.Type.(*ast.Ident).Name
		}
		if idx := findComponentByType(components, typeName); idx != -1 {
			panic(fmt.Sprintf(`component "%s" already declared in world "%s"`, typeName, w.Name))
		}
		components = append(components, newComponentInfo(typeName))
	}
	return components
}

func scanFilterConstraints(fn *ast.FuncType) filterInfo {
	filter := filterInfo{}
	for _, par := range fn.Params.List {
		// fmt.Printf("filter-include: %s\n", par.Type.(*ast.Ident).Name)
		var typeName string
		switch par.Type.(type) {
		case *ast.SelectorExpr:
			sel := par.Type.(*ast.SelectorExpr)
			typeName = fmt.Sprintf("%s.%s", sel.X.(*ast.Ident).Name, sel.Sel)
		case *ast.Ident:
			typeName = par.Type.(*ast.Ident).Name
		}
		filter.IncludeTypes = append(filter.IncludeTypes, typeName)
	}
	if fn.Results != nil {
		for _, par := range fn.Results.List {
			// fmt.Printf("filter-exclude: %v\n", par.Type.(*ast.Ident))
			var typeName string
			switch par.Type.(type) {
			case *ast.SelectorExpr:
				sel := par.Type.(*ast.SelectorExpr)
				typeName = fmt.Sprintf("%s.%s", sel.X.(*ast.Ident).Name, sel.Sel)
			case *ast.Ident:
				typeName = par.Type.(*ast.Ident).Name
			}
			filter.ExcludeTypes = append(filter.ExcludeTypes, typeName)
		}
	}
	return filter
}

func findComponentByType(c []componentInfo, typeName string) int {
	for i := range c {
		if c[i].Type == typeName {
			return i
		}
	}
	return -1
}

func validateFilters(w *worldInfo) {
	for fIdx := range w.Filters {
		f := &w.Filters[fIdx]
		for _, inc := range f.IncludeTypes {
			i := findComponentByType(w.Components, inc)
			if i == -1 {
				panic(fmt.Sprintf(`filter "%s" requested "%s" as include constraint that not exist in world "%s"`,
					f.Name, inc, w.Name))
			}
			f.IncludeIndices = append(f.IncludeIndices, strconv.Itoa(i))
		}
		for _, exc := range f.ExcludeTypes {
			i := findComponentByType(w.Components, exc)
			if i == -1 {
				panic(fmt.Sprintf(`filter "%s" requested "%s" as exclude constraint that not exist in world "%s"`,
					f.Name, exc, w.Name))
			}
			f.ExcludeIndices = append(f.ExcludeIndices, strconv.Itoa(i))
		}
	}
	fmt.Printf("world \"%s\" info:\n", w.Name)
	var cNames []string
	for _, c := range w.Components {
		cNames = append(cNames, c.Name)
	}
	fmt.Printf("components: %v\n", cNames)
	for _, f := range w.Filters {
		fmt.Printf("filter \"%s\": include=%v, exclude=%v\n", f.Name, f.IncludeTypes, f.ExcludeTypes)
	}
}

func joinSlice(s []string) string {
	res := strings.Join(s, ",")
	if len(res) > 0 {
		res += ","
	}
	return res
}

var templateFuncs = template.FuncMap{
	"joinSlice": joinSlice,
}

var packageTemplate = template.Must(template.New("").Funcs(templateFuncs).Parse(
	`// Code generated by "go generate", DO NOT EDIT.
package {{ .Package }}

import (
	"sort"
{{ range $i,$import := .Imports }}	
	{{$import}}
{{- end}}
)
{{ range $worldIdx,$world := .Worlds }}
{{- $worldName := $world.Name }}
// New{{$worldName}} returns new instance of {{$worldName}}.
func New{{$worldName}}(entitiesCount uint32) *{{$worldName}} {
	return &{{$worldName}}{
		world: ecs.NewWorld(entitiesCount, []ecs.ComponentPool{
{{- range $i,$c := $world.Components }}
		new{{$c.Name}}Pool(entitiesCount),
{{- end}}
		},[]ecs.Filter{
{{- range $i,$f := $world.Filters }}
		*ecs.NewFilter([]uint16{ {{ joinSlice $f.IncludeIndices }} }, []uint16{ {{ joinSlice $f.ExcludeIndices }} }, 512),
{{- end}}
		}),
	}
}

// InternalWorld returns internal ecs.World instance.
func (w {{$worldName}}) InternalWorld() *ecs.World { return w.world }

// Destroy processes cleanup of data inside world.
func (w *{{$worldName}}) Destroy() { w.world.Destroy(); w.world = nil }

// NewEntity creates and returns new entity inside world.
func (w {{$worldName}}) NewEntity() ecs.Entity {
	return w.world.NewEntity()
}

// DelEntity removes entity from world if exists. All attached components will be removed first.
func (w {{$worldName}}) DelEntity(entity ecs.Entity) { w.world.DelEntity(entity) }

// PackEntity packs Entity to save outside from world.
func (w {{$worldName}}) PackEntity(entity ecs.Entity) ecs.PackedEntity { return w.world.PackEntity(entity) }

// UnpackEntity tries to unpack data to Entity, returns unpacked entity and success of operation.
func (w {{$worldName}}) UnpackEntity(packedEntity ecs.PackedEntity) (ecs.Entity, bool) {
	return w.world.UnpackEntity(packedEntity)
}

{{ range $i,$c := $world.Components }}
type pool{{$c.Name}} []{{$c.Type}}

func new{{$c.Name}}Pool(cap uint32) *pool{{$c.Name}} {
	var pool pool{{$c.Name}} = make([]{{$c.Type}}, 0, cap)
	return &pool
}

func (p *pool{{$c.Name}}) New() {
	*p = append(*p, {{$c.Type}}{})
}

func (p *pool{{$c.Name}}) Recycle(idx uint32) {
	(*p)[idx] = {{$c.Type}}{}
}

// Set{{$c.Name}} adds or returns exist {{$c.Name}} component on entity.
func (w {{$worldName}}) Set{{$c.Name}}(entity ecs.Entity) *{{$c.Type}} {
	entityData := &w.world.Entities[entity]
	pool := w.world.Pools[{{$i}}].(*pool{{$c.Name}})
	if !entityData.BitMask.Get({{$i}}) {
		entityData.BitMask.Set({{$i}})
		maskIdx := sort.Search(len(entityData.Mask), func(i int) bool { return entityData.Mask[i] > {{$i}} })
		entityData.Mask = append(entityData.Mask, 0)
		copy(entityData.Mask[maskIdx+1:], entityData.Mask[maskIdx:])
		entityData.Mask[maskIdx] = {{$i}}
		w.world.UpdateFilters(entity, {{$i}}, true)
	}
	return &(*pool)[entity]
}

// Get{{$c.Name}} returns exist {{$c.Name}} component on entity or nil.
func (w {{$worldName}}) Get{{$c.Name}}(entity ecs.Entity) *{{$c.Type}} {
	if !w.world.Entities[entity].BitMask.Get({{$i}}) {
		return nil
	}
	return &(*w.world.Pools[{{$i}}].(*pool{{$c.Name}}))[entity]
}

// Get{{$c.Name}}Unsafe returns exist {{$c.Name}} component on entity or nil.
func (w {{$worldName}}) Get{{$c.Name}}Unsafe(entity ecs.Entity) *{{$c.Type}} {
	return &(*w.world.Pools[{{$i}}].(*pool{{$c.Name}}))[entity]
}

// Del{{$c.Name}} removes {{$c.Name}} component or do nothing.
// If entity is empty after removing - it will be destroyed automatically.
func (w {{$worldName}}) Del{{$c.Name}}(entity ecs.Entity) {
	entityData := &w.world.Entities[entity]
	if entityData.BitMask.Get({{$i}}) {
		if len(entityData.Mask) > 1 {
			w.world.UpdateFilters(entity, {{$i}}, false)
			w.world.Pools[{{$i}}].(*pool{{$c.Name}}).Recycle(entity)
			maskLen := len(entityData.Mask)
			maskIdx := sort.Search(maskLen, func(i int) bool { return entityData.Mask[i] >= {{$i}} })
			copy(entityData.Mask[maskIdx:], entityData.Mask[maskIdx+1:])
			entityData.Mask = entityData.Mask[:maskLen-1]
			entityData.BitMask.Unset({{$i}})
		} else {
			w.DelEntity(entity)
		}
	}
}
{{- end}}
{{- range $i,$f := $world.Filters }}
// {{$f.Name}} returns user filter.
func (w {{$worldName}}) {{$f.Name}}() *ecs.Filter {
	return w.world.Filter({{$i}})
}
{{- end}}
{{- end}}
`))
