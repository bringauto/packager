
# Reliable Package Source Rules

## Problem

We want to build reliable, robust application
with predictive closed dependency source, but we want to have an
ability to update our dependencies in case of need (so we do not rely on
many years obsolete package versions).

## Solution

If we build new authoritative package repository we need to know how to add package.

If we already have an authoritative package repository we need to know how to update, delete and add packages.

Let's have three actions

- Add package
- Update package
- Remove package

### Package dependencies

Each package can have a dependency.

Dependency can be a library, executable or any other set of files on which the package depends on.\
In our case each dependency is represented by a Package.

Let's call all of these dependencies `Set of package dependencies`.

Because each dependency is a Package, the `Set of package dependencies` is Set of Packages.

## Add Package

We assume the package is not present in the package repository.

Package with dependencies:

- Get `Set of dependencies` for a given Package.
- For each package in `Set of dependencies` verify if the package is present in the
  repository and has a correct version and platform string.
  - if the package is present in the package repository, check that the version
    and platform string are the same as for the given package. If not you need to apply
    `update package` or `add package` action on the given package dependency.
  - if the package is not present in the package repository, apply `add package` action
    on the given package.
- If all package dependencies are verified (eq. are present in the Package repository),
  create sysroot for all package dependencies - unzip and copy package dependencies into a specified directory.
  We call this directory as a `SYSROOT`
- During the build of the Package, the Package dependencies are looked in `SYSROOT` first

Package without dependencies:

- if the package has no dependencies build package as is.

**NOTE**: currently there is no command for "Create sysroot from already built packages".\
If you want to build only your package you need to build all dependencies one by one or
build complete package repository from scratch.

## Update Package

We assume the package is present in the package repository.

- delete old version of the Package.
- delete all packages that have the Package in the `Package dependency set`
- add all packages by `Add Package` action.

NOTE: if we do not rebuild all packages that depend on the
updated package we will have inconsistency and cannot guarantee the Package repository consistency.

## Remove package

We assume the package is present in the package repository.

- remove package from the Package repository.
- remove all packages that have the Package in the `Package dependency set`.
