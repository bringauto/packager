
# Reliable Package Source Rules

## Problem

We want to build reliable, robust application with predictive closed dependency source, but we want
to have an ability to update our dependencies in case of need (so we do not rely on many years
obsolete Package versions).

## Solution

If we build new authoritative Package Repository we need to know how to add Package.

If we already have an authoritative Package Repository we need to know how to update, delete and add Packages.

Let's have three actions

- Add Package
- Update Package
- Remove Package

### Package dependencies

Each Package can have a dependency.

Dependency can be a library, executable or any other set of files on which the Package depends on.\
In our case each dependency is represented by a Package.

Let's call all of these dependencies `Set of Package dependencies`.

Because each dependency is a Package, the `Set of Package dependencies` is Set of Packages.

## Add Package

We assume the Package is not present in the Package repository.

Package with dependencies:

- Get `Set of dependencies` for a given Package.
- For each Package in `Set of dependencies` verify if the Package is present in the
  repository and has a correct version and platform string.
  - if the Package is present in the Package repository, check that the version
    and platform string are the same as for the given Package. If not you need to apply
    `update Package` or `add Package` action on the given Package dependency.
  - if the Package is not present in the Package repository, apply `add Package` action
    on the given Package.
- If all Package dependencies are verified (eq. are present in the Package repository),
  create sysroot for all Package dependencies - unzip and copy Package dependencies into a specified directory.
  We call this directory as a `SYSROOT`
- During the build of the Package, the Package dependencies are looked in `SYSROOT` first

Package without dependencies:

- if the Package has no dependencies build Package as is.

**NOTE**: currently there is a command for "Create sysroot from already built Packages" -
`create-sysroot`.

## Update Package

We assume the Package is present in the Package repository.

- delete old version of the Package.
- delete all Packages that have the Package in the `Package dependency set`
- add all Packages by `Add Package` action.

NOTE: if we do not rebuild all Packages that depend on the
updated Package we will have inconsistency and cannot guarantee the Package repository consistency.

## Remove Package

We assume the Package is present in the Package repository.

- remove Package from the Package repository.
- remove all Packages that have the Package in the `Package dependency set`.
