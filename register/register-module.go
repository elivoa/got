package register

import (
	"fmt"
	"got/core"
	"sync"
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
	mc.m[module.PackagePath] = module // use package path as package key
	mc.l.Unlock()
}

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