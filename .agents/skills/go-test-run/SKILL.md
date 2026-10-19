---
name: go-test-run
description: "Run Go unit tests using this project's Makefile and report failures accurately."
---

# Go Tests

Mandatory Command

The project includes a Makefile that configures the environment required to run Go tests. Always use the Makefile to execute tests.

* From the project root, run `make test` to execute all tests.
* To target a specific package or individual tests, use make test ARGS="...".
* Never run `go test ./...` or `go test` directly, as these commands do not configure the project's specific test environment.

Procedure

1. Consult the Makefile if you need to understand how the test commands work.
2. Run all tests from the project root directory.
3. Use make test or its arguments to target the relevant tests.
4. If tests fail, analyze the error and fix its root cause. Do not bypass the Makefile to try to make the tests pass.
5. After modifying any production code or tests, run the relevant tests and verify the results before finishing.

Examples

Run all tests:

make test

Run all tests in a specific package:

make test ARGS="./app/shared/types/primitive"

Run a specific test:

make test ARGS="./app/shared/types/primitive -run TestNotifierStartsInactive"

Important

The Makefile is the authoritative way to execute tests in this project. Do not manually reproduce its internal command or remove any of its environment variables.