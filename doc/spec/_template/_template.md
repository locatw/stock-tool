<!--
  This template is a starting structure, not a rigid format.
  Add sections, subsections, tables, or detail blocks as needed
  to preserve information fidelity from source material.
-->

# <Spec Name>

## Background

<!--
  Describe the product-level motivation for this feature:
  why the product needs this capability regardless of the
  current implementation state. Avoid referencing the current
  implementation — it changes over time and makes the background
  fragile.
-->

Why the product needs this capability.

## Purpose

What problem this feature solves.

## User Stories

- As a <role>, I want <capability> so that <benefit>

## Requirements

<!--
  Flat bullets work for simple requirements:
    - FR-1: Files stored byte-for-byte without transformation

  For complex requirement areas, use a ### subsection with tables,
  sub-bullets, or structured detail. Preserve design decision notes
  and rationale from source material — do not compress structured
  content into a single line.

  Deferred items should appear under the relevant requirement with
  rationale, not only in Out of Scope.

  Example — structured requirement subsection:

  ### FR-3: Acquisition Metadata

  Every acquired file must have associated metadata.

  #### Data Identity

  | Item | Description |
  |---|---|
  | Source | Data source identifier |
  | Data format | Format of stored file (JSON, CSV) |

  #### Design Notes

  - Record count excluded — requires parsing raw data, conflicts with landing raw-preservation.
    File size serves as proxy; record count deferred to Bronze.
  - Storage mechanism (DB columns, S3 metadata, sidecar files) is a design decision — deferred.
-->

- FR-1: <functional requirement>
- NFR-1: <non-functional requirement>

## Acceptance Criteria

- [ ] <Criterion that defines "done">

## Design Direction

High-level approach.

## Use Cases

- [<usecase>](usecase/<usecase>.md) — <one-line description>

## Constraints

- <Technical or business constraint>

## Out of Scope

<!--
  Each item should include rationale:
    - <item> -- <why it is excluded or deferred>
-->

- <What this feature does not cover>
