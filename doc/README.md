
# BAP Documentation

The main purpose is to easily build and store binaries and libraries for
huge system base that we want to support.

BringAuto Packager uses docker as a build environment and Git LFS as a storage backend.

Both of these technologies (Git, Docker) are well known and do not need any
special training for programmers whose want to use them.

## Basics

Package is a set of "data" mainly represented by files stored on the computer disk.

Each package has a config.

Each package belongs to a Package Group

Package Group is a set of packages that share common properties. \
The meaning of "common properties" term is context dependent. In most
cases it means "share the same code".

Each docker image has a name assigned

Docker images are referred from Package Config by it's name.

Context is a directory where all config files (including Dockerfiles for images) are stored.

BAP needs a Context directory to work.

Package repository is a storage (set) where all packages for a project are stored.

## Package identification - ID

Each package has a `package_name`, `version_tag` and `platform_string` represented as a string.

These three attributes unique identifies the package.

Let A, B are packages. We say that package A is equal to package B
If and only if

- `package_name` of A is equal to `package_name` of B,
- `version_tag` of A is equal to `version_tag` of B,
- `platform_string` of A is equal to `platform_string` of B.

Packages in the package repository are identified by Package ID.

## Package name

Package name is a string.

Each package name consist from three parts

`package_name` = <prefix><base_package_name><debug_suffix><library_type>

`prefix`, `base_package_name` and `debug_suffix` are strings.

`base_package_name` should contain only [a-zA-Z0-9-] chars.

### Library package name creation

- `prefix` = "lib"
- `debug_suffix` = "d" if the package is built as "Debug"
- `debug_suffix` = "" if the package is built as "Release"
- `library_type` = "-dev" if the package is development package (contains headers, ...)
- `library_type` = "" if the package is a package contains only runtime lib

### Executable package name creation

- `prefix` = ""
- `debug_suffix` = "d" if the package is built as "Debug"
- `debug_suffix` = "" if the package is built as "Release"
- `library_type` = ""

## Detail documentation

- [Context Directory Structure]
- [Package JSON Structure]
- [Docker Container Requirements]
- [Build a Reliable Package Source]
- [CMake Project Requirements]
- [Build Process]
- [Package Dependencies]
- [Use Case Scenarios]

[Context Directory Structure]:     ./ContextDirectoryStructure.md
[Package JSON Structure]:          ./PackageJSONStructure.md
[Docker Container Requirements]:   ./DockerContainerRequiremetns.md
[CMake Project Requirements]:      ./CMakeProjectRequirements.md
[Build a Reliable Package Source]: ./ReliablePackageSource.md
[Build Process]:                   ./BuildProcess.md
[Package Dependencies]:            ./PackageDependencies.md
[Git Repository]:                  ./UseCaseScenarios.md
