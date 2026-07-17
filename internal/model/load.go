package model

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// catalogKeys are the top-level keys that identify a file as a source catalog
// asset. A file lacking all of them (metadata.yaml, groups.yaml, categories.yaml,
// …) is not walked for ids/refs — though metadata.yaml is still read for its
// catalog prefix.
var catalogKeys = map[string]bool{
	"capabilities": true,
	"threats":      true,
	"controls":     true,
}

// Load discovers YAML files under the given paths (files or directories),
// parses each catalog asset, and returns a position-aware Index.
func Load(paths []string) (*Index, error) {
	files, err := discover(paths)
	if err != nil {
		return nil, err
	}
	idx := newIndex()
	for _, f := range files {
		if err := idx.addFile(f); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

func discover(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", p, err)
		}
		if !info.IsDir() {
			out = append(out, filepath.ToSlash(p))
			continue
		}
		err = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			switch filepath.Ext(path) {
			case ".yaml", ".yml":
				out = append(out, filepath.ToSlash(path))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (idx *Index) addFile(p string) error {
	p = filepath.ToSlash(p)
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("reading %q: %w", p, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing %q: %w", p, err)
	}
	root := rootMapping(&doc)
	if root == nil {
		return nil
	}
	dir := path.Dir(p)

	// A metadata.yaml carries the catalog's canonical id prefix.
	if id, loc, ok := metadataID(root); ok {
		loc.File = p
		c := idx.catalog(dir)
		c.Prefix = id
		c.PrefixLoc = loc
	}

	// A catalog asset (capabilities/threats/controls) contributes ids and refs.
	if isCatalog(root) {
		idx.catalog(dir).HasAssets = true
		idx.Files = append(idx.Files, p)
		idx.walk(p, &doc)
	}
	return nil
}

// catalog returns the Catalog for dir, creating it on first use.
func (idx *Index) catalog(dir string) *Catalog {
	c := idx.Catalogs[dir]
	if c == nil {
		c = &Catalog{Dir: dir}
		idx.Catalogs[dir] = c
	}
	return c
}

// CatalogFor returns the catalog a file belongs to, or nil if none is known.
func (idx *Index) CatalogFor(file string) *Catalog {
	return idx.Catalogs[path.Dir(filepath.ToSlash(file))]
}

func rootMapping(doc *yaml.Node) *yaml.Node {
	n := doc
	if n.Kind == yaml.DocumentNode {
		if len(n.Content) == 0 {
			return nil
		}
		n = n.Content[0]
	}
	if n.Kind != yaml.MappingNode {
		return nil
	}
	return n
}

func isCatalog(m *yaml.Node) bool {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if catalogKeys[m.Content[i].Value] {
			return true
		}
	}
	return false
}

// metadataID extracts `metadata.id` from a document root, if present.
func metadataID(root *yaml.Node) (string, Location, bool) {
	md := mapValue(root, "metadata")
	if md == nil || md.Kind != yaml.MappingNode {
		return "", Location{}, false
	}
	id := mapValue(md, "id")
	if id == nil || id.Kind != yaml.ScalarNode || id.Value == "" {
		return "", Location{}, false
	}
	return id.Value, Location{Line: id.Line, Col: id.Column}, true
}

// mapValue returns the value node for key in a mapping node, or nil.
func mapValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// walk records every `id:` and `reference-id:` scalar in the node tree, with
// its source position.
func (idx *Index) walk(file string, n *yaml.Node) {
	switch n.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, c := range n.Content {
			idx.walk(file, c)
		}
	case yaml.MappingNode:
		idx.recordRequirements(file, n)
		for i := 0; i+1 < len(n.Content); i += 2 {
			key, val := n.Content[i], n.Content[i+1]
			if val.Kind == yaml.ScalarNode {
				loc := Location{File: file, Line: val.Line, Col: val.Column}
				switch key.Value {
				case "id":
					idx.addDef(val.Value, loc)
				case "reference-id":
					idx.addRef(val.Value, loc)
				}
			}
			idx.walk(file, val)
		}
	}
}

// recordRequirements detects a control mapping (one bearing an
// `assessment-requirements` list) and records each requirement id together
// with its parent control id.
func (idx *Index) recordRequirements(file string, ctrl *yaml.Node) {
	ars := mapValue(ctrl, "assessment-requirements")
	if ars == nil || ars.Kind != yaml.SequenceNode {
		return
	}
	var ctrlID string
	var ctrlLoc Location
	if idNode := mapValue(ctrl, "id"); idNode != nil && idNode.Kind == yaml.ScalarNode {
		ctrlID = idNode.Value
		ctrlLoc = Location{File: file, Line: idNode.Line, Col: idNode.Column}
	}
	for _, ar := range ars.Content {
		idNode := mapValue(ar, "id")
		if idNode == nil || idNode.Kind != yaml.ScalarNode {
			continue
		}
		idx.Requirements = append(idx.Requirements, Requirement{
			ID:         idNode.Value,
			Location:   Location{File: file, Line: idNode.Line, Col: idNode.Column},
			ControlID:  ctrlID,
			ControlLoc: ctrlLoc,
		})
	}
}
