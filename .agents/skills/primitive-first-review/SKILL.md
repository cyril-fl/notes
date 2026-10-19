---
name: primitive-first-review
description: Inspect shared primitives before adding reusable state, errors, timestamps, collections, descriptions, or other cross-cutting behavior.
---

# Primitive First Review

Use this skill when changing a reusable component or introducing fields for state, errors, timestamps, collections, descriptions, or other cross-cutting behavior.

Before adding a new field, interface, helper, or error container:

1. Search the shared primitive packages for an existing type that already models the behavior. Inspect both its implementation and tests.
2. Prefer embedding or composing the existing primitive when its semantics match. Preserve its public methods and invariants instead of duplicating fields such as `Err`, timestamps, list accessors, or description accessors.
3. Check pointer and value receiver behavior, zero-value behavior, serialization hooks, and whether the primitive accumulates or replaces state.
4. Add a focused regression test for the primitive-backed behavior, then run the relevant package tests and the affected callers.

If no primitive matches, explain why a new abstraction is needed before introducing one. Keep this review focused on shared behavior; ordinary local fields do not require a new primitive.
