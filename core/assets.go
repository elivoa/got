package core

import (
	"fmt"
	"sort"
)

// interface
type Asset interface {
	SourcePath() string
	DupCount() int // 重复次数
	IncDupCount()  //增加一次重复次数
	Order() int    // 出现的顺序
}

// base asset object
type AssetBase struct {
	dupCount int // 重复次数
	order    int // 顺序
}

func (a *AssetBase) DupCount() int { return a.dupCount }
func (a *AssetBase) IncDupCount()  { a.dupCount += 1 }
func (a *AssetBase) Order() int    { return a.order }

// Stylesheet Link
type StyleLink struct {
	AssetBase
	Href string
	Type string
	Rel  string
}

func (a *StyleLink) SourcePath() string { return a.Href }

// Javascript
type Script struct {
	AssetBase
	Type string
	Src  string
}

func (a *Script) SourcePath() string { return a.Src }

// collection of Assets
type AssetSet struct {
	Scripts     map[string]*Script
	StyleSheets map[string]*StyleLink
}

func NewAssetSet() *AssetSet {
	return &AssetSet{
		Scripts:     map[string]*Script{},
		StyleSheets: map[string]*StyleLink{},
	}
}

// 使用SourcePath来去重
func (as *AssetSet) AddScripts(script *Script) {
	if nil == script {
		return
	}
	if m, ok := as.Scripts[script.SourcePath()]; ok {
		if nil != m {
			m.IncDupCount()
		}
	} else {
		// set index
		if script.order == 0 {
			script.order = len(as.Scripts) + 1 // 使用size+1来创建按添加顺序的order。
		}
		as.Scripts[script.SourcePath()] = script
	}
}

func (as *AssetSet) AddStyleSheet(style *StyleLink) {
	if nil == style {
		return
	}
	if m, ok := as.StyleSheets[style.SourcePath()]; ok {
		if nil != m {
			m.IncDupCount()
		}
	} else {
		// set index
		if style.order == 0 {
			style.order = len(as.StyleSheets) + 1 // 使用size+1来创建按添加顺序的order。
		}
		as.StyleSheets[style.SourcePath()] = style
	}
}

func (as *AssetSet) DebugPrintAll() {
	if as.Scripts != nil {
		for _, s := range as.Scripts {
			fmt.Println("Asset: js  - ", s.SourcePath(), ", dup:", s.dupCount)
		}
	}
	if as.StyleSheets != nil {
		for _, s := range as.StyleSheets {
			fmt.Println("Asset: css - ", s.SourcePath(), ", dup:", s.dupCount)
		}
	}
}

// combine AssetSet into current AssetSet.
func (as *AssetSet) CombineAssets(assetset *AssetSet) {
	if nil != assetset {
		if nil != assetset.Scripts {
			for _, v := range assetset.Scripts {
				as.AddScripts(v)
			}
		}
		if nil != assetset.StyleSheets {
			for _, v := range assetset.StyleSheets {
				as.AddStyleSheet(v)
			}
		}
	}
}

// for sort
type SortableAsset struct {
	core []Asset
}

func (p SortableAsset) Len() int           { return len(p.core) }
func (p SortableAsset) Less(i, j int) bool { return p.core[i].Order() < p.core[j].Order() }
func (p SortableAsset) Swap(i, j int)      { p.core[i], p.core[j] = p.core[j], p.core[i] }

func (as *AssetSet) OrderedScripts() []Asset {
	var sa = new(SortableAsset) // , len(as.Scripts))
	if nil != as.Scripts {
		for _, v := range as.Scripts {
			sa.core = append(sa.core, v)
		}
	}
	sort.Sort(sa)
	return sa.core
}

func (as *AssetSet) OrderedStyles() []Asset {
	var sa = new(SortableAsset) // , len(as.Scripts))
	if nil != as.StyleSheets {
		for _, v := range as.StyleSheets {
			sa.core = append(sa.core, v)
		}
	}
	sort.Sort(sa)
	return sa.core
}
