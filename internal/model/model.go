// Package model loads Gemara source catalogs into a position-aware in-memory
// index. It deliberately keys off the stable Gemara authoring conventions
// (`id` and `reference-id`) rather than decoding into typed structs, so it is
// agnostic to the exact source "sugar" shape and to the Gemara schema version.
package model

// Location is a source position for a finding. Line/Col are 1-based; a zero
// Line means the position is unknown.
type Location struct {
	File string
	Line int
	Col  int
}

// Def is a defined identifier (the value of an `id:` key).
type Def struct {
	ID       string
	Location Location
}

// Ref is a reference to an identifier (the value of a `reference-id:` key).
type Ref struct {
	ID       string
	Location Location
}

// Index is the whole loaded catalog set: every defined id, every reference,
// and the files they came from.
type Index struct {
	Files   []string
	Defs    map[string]Def // id -> first definition
	DupDefs []Def          // definitions of an id already seen (duplicates)
	Refs    []Ref          // every reference-id occurrence, in file order
}

func newIndex() *Index {
	return &Index{Defs: make(map[string]Def)}
}

func (idx *Index) addDef(id string, loc Location) {
	if _, seen := idx.Defs[id]; seen {
		idx.DupDefs = append(idx.DupDefs, Def{ID: id, Location: loc})
		return
	}
	idx.Defs[id] = Def{ID: id, Location: loc}
}

func (idx *Index) addRef(id string, loc Location) {
	idx.Refs = append(idx.Refs, Ref{ID: id, Location: loc})
}
