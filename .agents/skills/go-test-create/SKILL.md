---
name: go-test-create
description: "Create or update Go unit tests when requested, especially for files containing a // TEST marker."
---

Go Tests

Objective

Write simple, readable, and useful Go unit tests that follow the language’s conventions and the project’s existing practices.

Procedure

1. Identify the Relevant Files

* Search the source code for // TEST comments.
* Examine the relevant files and pending changes to understand the scope of the work.
* Do not automatically assume that every modified file needs to be tested.

2. Understand the Behavior

* Read the source code and identify its essential behaviors.
* Identify the expected behavior, relevant edge cases, and expected errors.
* Avoid testing implementation details that have no meaningful functional impact.

3. Write the Tests

* Place each test file alongside the source file it tests.
* Follow Go’s *_test.go naming convention.
* Use the existing package and follow idiomatic Go conventions.
* Prefer Go’s standard testing package.
* Test only essential behaviors without unnecessarily multiplying test cases or introducing abstractions.

4. Preserve Production Code

* Do not modify production code unless there is a genuine need.
* Do not refactor code solely to make it easier to test.
* If a defect is discovered in the production code, clearly report it rather than silently changing its behavior.

5. Remove Test Markers

* After completing the tests for a file, remove its // TEST comment if present.
* Remove only the marker comment, preserving any surrounding code and documentation.
* Do not remove markers from files whose tests have not yet been completed.

6. Verify the Results

* Consult the Makefile and project documentation to confirm the recommended test command.
* Run tests from the project root with `make test`, or target a package with `make test ARGS="..."`.
* Verify that the new tests pass and distinguish test failures from potential defects in the production code.
* Confirm that all completed files have had their // TEST markers removed.

Principles

* Prioritize simplicity over artificial comprehensiveness.
* Avoid overengineering, unnecessary mocks, and premature abstractions.
* Do not introduce external dependencies without justification.
* Follow the project’s existing style and conventions.
* Do not make changes unrelated to the requested tests.
