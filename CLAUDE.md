# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A standalone Go CLI (and composite GitHub Action) that validates the **semantic data correctness** of [Gemara](https://github.com/gemaraproj/go-gemara) source catalogs — the cross-reference/convention rules that JSON-schema or CUE validation does not cover (referential integrity, id uniqueness, prefix conventions, mapping completeness). It is **project-agnostic**: every project-specific convention lives in `.gemara-validate.yaml`, never in code.

## Commands

```bash
go build -o gemara-validate .     # build the CLI
go vet ./...                      # vet
go test ./...                     # all tests
go test ./internal/check/checks -run TestARNesting   # a single test
./gemara-validate                 # run: uses .gemara-validate.yaml + defaults
./gemara-validate -format github  # GitHub Actions inline-annotation output
./gemara-validate path/to/catalogs  # positional args override configured paths
./gemara-validate -orphans        # also report orphan stats (warnings, opt-in)
./gemara-validate -fail-on never  # report but never exit non-zero (monitor mode)
```

CI (`.github/workflows/ci.yml`) runs `go mod tidy` then **fails if `go.mod`/`go.sum` changed** — keep them tidy before pushing. It then runs vet, test, and build.

## Architecture

The pipeline in `main.go` is: `config.Load` → `model.Load` (builds a position-aware `*model.Index`) → each registered `check.Check.Run(ctx)` collects `[]Finding` → `check.Report` writes them. Exit codes: **2** on internal error (bad config, unreadable path), **1** when error-severity findings exist *and* `fail-on == "error"`, **0** otherwise.

- **`internal/model`** — Loads catalogs by walking the **raw `yaml.Node` tree**, deliberately *not* decoding into `go-gemara` typed structs. Two reasons, both load-bearing: (1) it keeps line/column positions for annotations, and (2) the catalog *source* files are an authoring "sugar" shape that differs from Gemara's published types. The `Index` exposes `Defs` (id → first definition), `AllDefs`, `DupDefs`, `Refs`, `Requirements`, `Items` (with their outgoing `Mappings`), and per-directory `Catalogs`. A file is treated as a catalog asset only if its root has a `capabilities`/`threats`/`controls` key; `metadata.id` supplies a catalog's canonical prefix. This layer is the seam for a future typed/compiled-artifact checker.
- **`internal/config`** — `.gemara-validate.yaml`. The key concept is `IDPattern` / `InternalRef`: a `reference-id` is only required to resolve if it **matches the id-pattern regexp**. Non-matching refs (mapping-group shorthands like `CCC`, external framework references) are intentionally ignored by referential checks. Defaults apply when no config file exists.
- **`internal/check`** — the `Check` interface, the self-registration `registry`, the `Finding` type (carries `File`/`Line`/`Col`), and `report.go`. Note: the `github` reporter folds the `file:line:col` location into the **visible message text** as well as the annotation metadata, because GitHub renders the raw job log as a bare `Error: <message>` and drops the metadata there.
- **`internal/check/checks`** — one file per rule.

### Adding a check

Create a file in `internal/check/checks/` with a type implementing `check.Check` (`ID`, `Description`, `Run(*check.Context) []check.Finding`) and register it in an `init()`:

- `check.Register(c)` — runs by default.
- `check.RegisterOptional("name", c)` — opt-in, enabled via a flag (e.g. `-orphans`); conventionally used for noisy "stats" checks that must never affect the exit code.

Registration works because `main.go` blank-imports the `checks` package for its side effects. Populate `Finding.File/Line/Col` from the relevant `model` element's `Location` so findings are locatable.

## Releasing

Pushing a `v*` tag triggers `.github/workflows/release.yml` (GoReleaser). Conventions:

- Tags are **strict semver with a `v` prefix and no dot after it**: `v1.0.5`, never `v.1.0.5` (the latter fails GoReleaser's semver parse).
- The moving **`v1`** major tag is force-updated to each new release so `action@v1` consumers get fixes: `git tag -f v1 <commit> && git push origin v1 --force`.
- The action's top-level `description` in `action.yml` must stay **under 125 characters** or the GitHub Marketplace release is rejected.
- The composite action builds the validator with `working-directory: ${{ github.action_path }}` so Go resolves *this* module's `go.mod` rather than the consumer repo's checkout.
