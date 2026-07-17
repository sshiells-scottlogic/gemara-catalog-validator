package model

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// catalogKeys are the top-level keys that identify a file as a source catalog
// asset. A file lacking all of them (metadata.yaml, groups.yaml, categories.yaml,
// …) is skipped.
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
			out = append(out, p)
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

func (idx *Index) addFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %q: %w", path, err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parsing %q: %w", path, err)
	}
	root := rootMapping(&doc)
	if root == nil || !isCatalog(root) {
		return nil // not a catalog asset — skip
	}
	idx.Files = append(idx.Files, path)
	idx.walk(path, &doc)
	return nil
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

// walk records every `id:` and `reference-id:` scalar in the node tree, with
// its source position.
func (idx *Index) walk(file string, n *yaml.Node) {
	switch n.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, c := range n.Content {
			idx.walk(file, c)
		}
	case yaml.MappingNode:
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
