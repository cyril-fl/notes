---
name: plum-process
description: Write or refactor a multi-step, fallible workflow (load, resolve, generate, create...) as a plum `primitive.Process` with ordered guarded steps and a cached result.
---

# Plum Process

Use this skill when code runs ordered steps that can each fail and must stop at the first failure (load then resolve then save, create tables then indexes, ...).

## Process vs Mill

- **Process** (`primitive.Process`): one entry point `Run()`, ordered steps, optional single result via `GetResult()`. Example: `resolvearchives`, `generatesession`.
- **Mill** (`primitive.Mill`): a reusable tool exposing several independent methods that each collect errors. Example: plum `millload`, `millentries`.

If callers would only ever call one method in a fixed sequence, it is a Process.

## Layout

- Package in `app/<scope>/services/process/<verb><noun>/` (e.g. `resolvearchives`, `createschema`), lowercase, no underscore.
- Import plum as `plumprimitive "github.com/cyril-fl/plum/primitive"`.

## Template

```go
package resolvesomething

import plumprimitive "github.com/cyril-fl/plum/primitive"

type run struct {
	plumprimitive.Process
	cache

	// dependencies, injected by Process(...)
	db *Db
}

// cache holds intermediate values and the final result.
type cache struct {
	result *Something
}

func Process(db *Db) *run {
	return &run{
		db: db,
	}
}

func (p *run) Run() {
	p.loadSomething()
	p.ensureSomething()
}

// Getters & Setters
func (p *run) GetResult() *Something {
	return p.cache.result
}

// Methods
func (p *run) loadSomething() {
	result, err := load(p.db)
	if err != nil {
		p.SetErr(err)
		return
	}

	p.cache.result = result
}

func (p *run) ensureSomething() {
	if p.HasErr() {
		return
	}

	// ...
}

// Helpers
```

## Rules

1. `Process(...)` only assigns dependencies: no logic, no I/O.
2. `Run()` is a list of step calls, one per line, no logic. Step names: `makeX`, `loadX`, `ensureX`, `runX`, `createX`...
3. Every fallible step after the first starts with `if p.HasErr() { return }`. Steps that cannot fail and do not depend on earlier results may skip the guard.
4. Steps communicate only through `p.cache`, never through return values.
5. Record errors with `p.SetErr(err)` (nil is ignored); never return them from a step.
6. A step that must run even after an error (cleanup, rollback) checks `p.HasErr()` to choose its branch instead of returning early.
7. `GetResult()` returns the cached result. Omit it when there is nothing to return (`Processable` only needs `Run` and the error methods).
8. Keep the section comments `// Getters & Setters`, `// Methods`, `// Helpers`. Helpers are stateless free functions.

## Sub-process

```go
func (p *run) makeArchives() {
	if p.HasErr() {
		return
	}

	process := resolvearchives.Process(p.db, p.cache.campaignID)
	process.Run()

	p.cache.archives = process.GetResult()
	p.SetErr(process.GetErr())
}
```

## Calling a process

```go
process := createschema.Process(ctx, db)
process.Run()
if err := process.GetErr(); err != nil {
	return err
}
```

To chain several processes without results, use `plumprimitive.HandleProcess([]plumprimitive.Processable{...})`.
