---
name: plum-integration
description: Migrate Solopg to a newer Plum checkout or release by consuming Plum APIs directly and removing duplicated local utilities.
---

# Plum integration

Use this skill when updating Solopg to a newer version of `github.com/cyril-fl/plum`.

## Workflow

1. Inspect the Plum checkout, its `go.mod`, recent commits, public packages, and its `AGENTS.md` before editing Solopg.
2. Consider only Plum files whose names end in `.go`. Ignore code stored in files without the `.go` extension, even if the content looks like Go source.
3. Confirm that the eligible `.go` files compile as part of their declared Plum package. A `.go` file with a package-name conflict, missing imports, or another Plum-side compile error is not a usable API yet; report the Plum issue instead of reproducing or wrapping it in Solopg.
4. Compare the usable Plum public symbols with local utilities, primitives, interfaces, and types. Replace duplicate Solopg imports with direct Plum imports only for those usable packages.
5. Remove local copies once all callers use a usable Plum package. If a Plum feature exists only in a non-`.go` file, leave the Solopg implementation in place; do not import, copy, wrap, or reimplement that non-`.go` source as part of the Plum migration.
6. Alias every Plum import explicitly with the `plum<package>` convention: `plumprimitive`, `plumtypes`, `pluminterface`, `plumtransform`, `plumcleanui`, or `plumdebug`. Do not use the package's bare default name or an unrelated alias.
7. Do not add re-exports, compatibility facades, or new wrappers around Plum APIs.
8. Keep Solopg-specific abstractions only when Plum has no usable equivalent. Existing domain embeddings of Plum primitives may remain when they provide the model's required promoted behavior; do not introduce new ones just to hide Plum.
9. Update the required Plum pseudo-version in `go.mod` to the exact commit. Preserve the local `replace` when the repository is intentionally developing against the sibling Plum checkout.

For shared caching, use `plumprimitive.NewCache[T]`, migrate both production callers and their tests, then remove only the duplicated local cache implementation. Keep unrelated local primitives such as `List` or `Notifier` when Plum does not provide their equivalent.

## Source validity

Before importing a Plum package, verify that its implementation is in one or more `.go` files:

- ignore non-`.go` files completely;
- do not modify the Plum repository;
- keep the existing Solopg implementation when the equivalent Plum source is not a `.go` file.
- for example, use Plum's `primitive/process.go` directly for Process APIs, but keep Solopg's Mill primitive when Plum exposes Mill only from an extensionless `primitive/mill` file.

## Verification

- When touching a Go file containing `// TEST`, invoke `$go-test-create` before changing it.
- Invoke `$go-test-run` before running tests, then run `make test` from the project root.
- Run `make build`; if the host Go cache is inaccessible, use `GOCACHE=/tmp/solopg-go-cache make build`.
- Search for removed local import paths and `git diff --check` before finishing.
