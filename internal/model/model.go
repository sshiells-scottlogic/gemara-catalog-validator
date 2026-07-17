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

// Requirement is an assessment-requirement id together with the control it is
// declared under, so checks can verify the two are consistent.
type Requirement struct {
	ID         string
	Location   Location
	ControlID  string // id of the enclosing control ("" if the control has none)
	ControlLoc Location
}

// Mapping is one outgoing relationship entry from an Item to another artifact
// (e.g. a control's threat mapping, a threat's capability mapping).
type Mapping struct {
	RefID    string // referenced id (entries[].reference-id)
	Group    string // the enclosing mapping-group reference-id (e.g. "CCC")
	Location Location
}

// Item is a top-level catalog entry (a capability, threat, or control) together
// with its outgoing mapping blocks, keyed by block name ("threats",
// "capabilities", "guidelines", …).
type Item struct {
	ID       string
	Kind     string // "capability" | "threat" | "control"
	Location Location
	Mappings map[string][]Mapping
}

// Catalog is one catalog directory (the folder holding capabilities/threats/
// controls assets and their metadata).
type Catalog struct {
	// Dir is the slash-normalized directory the catalog's files live in.
	Dir string
	// Prefix is the catalog's canonical id prefix, taken from `metadata.id`
	// (e.g. "CCC.Monitor"). Empty if no metadata.id was found.
	Prefix string
	// PrefixLoc is where `metadata.id` is defined, for diagnostics.
	PrefixLoc Location
	// HasAssets is true once a capabilities/threats/controls file is seen here.
	HasAssets bool
}

// Index is the whole loaded catalog set: every defined id, every reference, the
// catalogs they belong to, and the files they came from.
type Index struct {
	Files    []string
	Defs     map[string]Def      // id -> first definition
	AllDefs  []Def               // every `id:` occurrence, in file order
	DupDefs  []Def               // definitions of an id already seen (duplicates)
	Refs         []Ref               // every reference-id occurrence, in file order
	Requirements []Requirement       // assessment requirements with their control
	Items        []Item              // capabilities/threats/controls with mappings
	Catalogs     map[string]*Catalog // keyed by directory
}

func newIndex() *Index {
	return &Index{
		Defs:     make(map[string]Def),
		Catalogs: make(map[string]*Catalog),
	}
}

func (idx *Index) addDef(id string, loc Location) {
	idx.AllDefs = append(idx.AllDefs, Def{ID: id, Location: loc})
	if _, seen := idx.Defs[id]; seen {
		idx.DupDefs = append(idx.DupDefs, Def{ID: id, Location: loc})
		return
	}
	idx.Defs[id] = Def{ID: id, Location: loc}
}

func (idx *Index) addRef(id string, loc Location) {
	idx.Refs = append(idx.Refs, Ref{ID: id, Location: loc})
}
