---
name: go-documentation
description: Document an existing Go codebase on explicit request, following Go doc conventions while preserving behavior and clearly separating reviewed code from work in progress.
disable-model-invocation: true
---

# Go documentation

Use this skill only when explicitly requested. It documents an existing Go codebase; it does not run automatically after every code change.

## Scope and discovery

- Inspect the repository instructions, module layout, package names, existing comments, tests, generated files, and pending changes before editing.
- Limit the work to the requested packages or, when the request is project-wide, to production Go files. Do not rewrite tests or unrelated code.
- Preserve all logic, signatures, exported names, package names, struct tags, formatting semantics, and runtime behavior. Documentation changes should be comment-only unless a requested example requires a separate example file.
- Treat a file as reviewed only after its comments have been checked and the relevant validation has passed. Do not infer that undocumented or recently changed code is validated. If persistent status is needed, use a separate project-approved manifest or review note rather than adding non-idiomatic status markers to source comments.

## What to document

Document exported packages, types, interfaces, functions, methods, constants, variables, and exported fields when their purpose, contract, side effects, lifecycle, concurrency, ordering, error behavior, or important limitation is not obvious from the declaration.

- Every exported declaration that is part of the public API should normally have a concise Go doc comment beginning with its exact name.
- Skip comments that merely repeat a self-explanatory declaration or add no useful contract information.
- Document unexported declarations only when their behavior is subtle or needed to understand an exported API.
- Preserve useful existing comments and improve only inaccurate, unclear, or non-idiomatic text.

## Comment style

- Start the first sentence with the documented identifier, for example `// Set adds value to the set.`
- Prefer one or two clear lines in the source; add detail only for behavior users could otherwise misread.
- State observable behavior, not implementation narration. Mention nil handling, ordering, mutation, concurrency, error aggregation, time zones, and side effects when applicable.
- Use standard Go formatting and run `gofmt` on touched Go files. Avoid TODOs, speculative claims, and comments in languages inconsistent with the project unless requested.
- For a group of related constants, a group comment is sufficient when individual values are self-explanatory; add individual comments when values have distinct contracts.

## Examples

Add an example only when it teaches a non-obvious usage pattern, lifecycle, output, or constraint that a short doc comment cannot convey. Prefer an executable `Example...` test in the owning package when the API is stable and the output can be deterministic. Ensure identifiers, imports, output, and package names match the current API. Do not add examples merely to increase coverage or alter production code.

## Validation

After editing:

1. Review the diff and confirm that production behavior and public signatures are unchanged.
2. Use `go doc` or an equivalent AST-aware check to inspect exported declarations and verify that comments begin with the correct identifier.
3. Run `gofmt` on touched Go files and inspect the resulting diff.
4. Run the project's prescribed tests. If tests or examples fail, distinguish documentation mistakes from pre-existing production defects and do not silently change behavior.
5. Report the files reviewed, any intentionally undocumented declarations, examples added or omitted, and the validation performed.

Keep the reviewed scope explicit in the final report so later work can distinguish validated documentation from code still under development.
