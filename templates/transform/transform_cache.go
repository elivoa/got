/*
Template cache that contains structor path of protons.
e.g.: /got/status contains TemplateStatus (default id is TemplateStatus).
*/
package transform

import ()

// TODO: Monitor Tempaltes and reload it.

var transcache = &TransformCache{
	m: make(map[string]*TransformCacheInfo, 8),
}

type TransformCache struct {
	// l sync.RWMutex
	m map[string]*TransformCacheInfo
}

type TransformCacheInfo struct {
	ID     string // component id
	InLoop bool   // is in loop?
	Inner  map[string]*TransformCacheInfo
}
