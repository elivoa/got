package register

import (
	"fmt"
	"sync"

	"github.com/elivoa/got/core"
)

// Use Module instead.
// TODO:
//   - DONE| Cache module
//   - Remove APP
//
var Modules = &ModuleCache{m: map[string]*core.Module{}}

type ModuleCache struct {
	l sync.RWMutex
	m map[string]*core.Module
}

func RegisterModule(modules ...*core.Module) {
	for _, m := range modules {
		Modules.Add(m)
	}
}

func (mc *ModuleCache) Add(module *core.Module) {
	mc.l.Lock()
	key := module.PackageName //module.Key()
	if dupm, ok := mc.m[key]; ok {
		panic(fmt.Sprint("Duplicated model when register! ", key,
			"\n\t", dupm.BasePath,
			"\n\t", module.BasePath,
		))
	}
	mc.m[key] = module // use package path as package key
	mc.l.Unlock()
}

// TODO no one use this?
func (mc *ModuleCache) Get(name string) *core.Module {
	mc.l.RLock()
	module := mc.m[name]
	mc.l.RUnlock()
	return module
}

func (mc *ModuleCache) Map() map[string]*core.Module {
	return mc.m
}

// ----  Printing  -----------------------------------------------------------------------------------

func (mc *ModuleCache) PrintALL() {
	fmt.Println("---- Modules ---------------------")
	mc.l.RLock()
	for _, module := range mc.m {
		fmt.Printf("  %v\n", module.String())
	}
	mc.l.RUnlock()
}
