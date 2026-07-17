# Gemara Catalog Validator

A standalone tool that validates the **data correctness** of [Gemara](https://github.com/gemaraproj/go-gemara)
source catalogs — the semantic rules that schema validation doesn't cover — and
reports findings inline on a pull request so they can be fixed before merge.

It is **project-agnostic**: it keys off the stable Gemara authoring conventions
(`id`, `reference-id`, imports/mappings), so it works for any repo that authors
Gemara catalogs, not just one. All project-specific conventions live in a config
file, not in code.

## What it checks

This is **not** schema validation (structure / types / required fields) — pair
it with your JSON-schema or CUE check for that. This tool catches semantic
errors those miss:

- **`reference-resolves`** — every internal `reference-id` (imports, threat→capability
  and control→threat mappings) resolves to a defined `id` in the catalog set.
- **`unique-ids`** — every `id` is defined exactly once.
- **`id-prefix`** — every defined `id` carries its catalog's canonical prefix,
  taken from that catalog's `metadata.id` (e.g. a catalog whose `metadata.id` is
  `CCC.Monitor` must define `CCC.Monitor.*` ids). Catches casing/typo/copy-paste
  prefix drift.
- **`ar-nesting`** — every `assessment-requirement` id nests under its parent
  control id as `<controlId>.AR<n>` (e.g. control `CCC.GenAI.CN01` →
  `CCC.GenAI.CN01.AR01`). Catches requirements copy-pasted under the wrong
  control and malformed requirement ids.

More checks are easy to add (see below).

> Dogfooded against the FINOS common-cloud-controls catalogs, it surfaced real
> bugs including a `CCC.Monitor.*` vs `CCC.Monitoring.*` prefix mismatch and a
> reference to a non-existent `CCC.Core.CN12`.

## Usage

Local:

```bash
go build -o gemara-validate .
./gemara-validate                       # uses .gemara-validate.yaml + defaults
./gemara-validate -format github        # inline annotations for CI
./gemara-validate path/to/catalogs      # override configured paths
```

In GitHub Actions (blocks merge as a required status check):

```yaml
on: { pull_request: { paths: ['**/*.yaml', '**/*.yml'] } }
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go run github.com/sshiells-scottlogic/gemara-catalog-validator@latest -format github
```

(For production, publish tagged binaries via GoReleaser and ship a thin
composite action that downloads the right one — faster than building per run.)

## Configuration

`.gemara-validate.yaml` at the repo root carries everything project-specific:

```yaml
paths:
  - catalogs
# Regexp for an internal, resolvable reference-id. Non-matching refs
# (mapping-group shorthands like "CCC", external frameworks) are ignored.
id-pattern: '^CCC\.[A-Za-z0-9]+\.(CP|CN|TH)[0-9]+$'
fail-on: error   # or "never" to report without failing
```

With no config file, sensible defaults apply (`paths: [catalogs]`, a
CCC-compatible id pattern).

## Adding a check

Each check is a small type implementing `check.Check`, self-registered via
`init()`. Add a file under [`internal/check/checks/`](internal/check/checks/):

```go
type myCheck struct{}
func (myCheck) ID() string          { return "my-check" }
func (myCheck) Description() string  { return "…" }
func (myCheck) Run(ctx *check.Context) []check.Finding { /* inspect ctx.Index */ }
func init() { check.Register(myCheck{}) }
```

`ctx.Index` gives you every defined id (`Defs`), duplicate definitions
(`DupDefs`), and every reference with its source position (`Refs`).

## Architecture

- [`internal/model`](internal/model) — loads catalogs into a position-aware
  index by walking the YAML node tree (keeps line/column for annotations).
- [`internal/config`](internal/config) — project conventions.
- [`internal/check`](internal/check) — `Check` interface, registry, reporters
  (`text`, `github`).
- [`internal/check/checks`](internal/check/checks) — the rules.

Typed / compiled-artifact checks can later decode via `go-gemara`; the model
layer is the seam for that. The current checks intentionally avoid it because
the **source** files are an authoring "sugar" shape that differs from Gemara's
published types.

## Development

```bash
go vet ./...
go test ./...
go build -o gemara-validate .
```
