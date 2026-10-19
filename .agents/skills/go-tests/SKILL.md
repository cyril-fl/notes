---
name: go-tests
description: "Create, update, or run Go unit tests while preserving production behavior and following this project's conventions."
---

# Go tests

Use this skill for broad Go test work. For a focused operation, use `go-test-create` to write tests or `go-test-run` to execute them.

## Create or update tests

Before editing tests:

1. Search the relevant source for `// TEST` markers and inspect pending changes.
2. Read the implementation and identify observable behavior, meaningful edge cases, and expected errors.
3. Follow the existing package layout and prefer the standard `testing` package.

Write focused tests alongside the production package. Test behavior rather than implementation details, and avoid unnecessary mocks, abstractions, or dependencies. Do not change production code or refactor it solely to make it testable; report defects separately. Remove only a completed file's `// TEST` marker, preserving surrounding comments and code.

## Run tests

Run the project's prescribed command from its root:

```sh
make test
```

Use a package path or `-run` only when targeting a specific test. After changing production code or tests, rerun the relevant tests and distinguish test failures from pre-existing defects.

## Boundaries

Do not create or modify tests merely because a production file changed. Keep test changes scoped to the requested behavior, preserve production signatures and behavior, and report which markers, packages, and commands were completed.
