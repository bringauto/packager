
# BAP Documentation

The main purpose of BringAuto Packager is to easily build and store binaries and libraries for huge
system base that we want to support.

BringAuto Packager (BAP) uses docker as a build environment and Git LFS as a storage backend.

Both of these technologies (Git, Docker) are well known and do not need any
special training for programmers who want to use them.

## Basics

Following terms are reffered in this documentation:
 - **Package** - is a set of "data" mainly represented by files stored on the computer disk
 - **Package Group** - is a set of Packages that share common properties; the meaning of "common
  properties" term is context dependent, in most cases it means "share the same code"
 - **Config** - is a json file representing one Package
 - **Context** - is a directory where all Config files (including Dockerfiles for images) are stored
- **Package Repository** - is a storage (directory) where built Packages are stored

Each Package has exactly one Config.

Each Package belongs to a Package Group.

Each docker image has a name assigned.

Docker images are referred from Config by it's name.

## Package identification - ID

Each Package has a `package_name`, `version_tag` and `platform_string` represented as a string.

These three attributes are unique identification of the Package.

Let A, B are Packages. We say that A is equal to B if and only if

- `package_name` of A is equal to `package_name` of B,
- `version_tag` of A is equal to `version_tag` of B,
- `platform_string` of A is equal to `platform_string` of B.

Packages in the Package Repository are identified by Package ID.

## Package name

Package name is a string.

Each Package name consist from three parts:

- `package_name` = <prefix><base_package_name><debug_suffix><library_type>
- `prefix`, `base_package_name` and `debug_suffix` are strings.
- `base_package_name` should contain only [a-zA-Z0-9-] characters.

### Library Package name creation

- `prefix` = "lib"
- `debug_suffix` = "d" if the Package is built as "Debug"
- `debug_suffix` = "" if the Package is built as "Release"
- `library_type` = "-dev" if the Package is development package (contains headers, ...)
- `library_type` = "" if the Package contains only runtime lib

### Executable Package name creation

- `prefix` = ""
- `debug_suffix` = "d" if the Package is built as "Debug"
- `debug_suffix` = "" if the Package is built as "Release"
- `library_type` = ""

## Detail documentation

- [Context Structure]
- [Config Structure]
- [Docker Container Requirements]
- [Build a Reliable Package Source]
- [CMake Project Requirements]
- [Build Process]
- [Package Dependencies]
- [Use Case Scenarios]

[Context Structure]:               ./ContextStructure.md
[Config Structure]:                ./ConfigStructure.md
[Docker Container Requirements]:   ./DockerContainerRequiremetns.md
[CMake Project Requirements]:      ./CMakeProjectRequirements.md
[Build a Reliable Package Source]: ./ReliablePackageSource.md
[Build Process]:                   ./BuildProcess.md
[Package Dependencies]:            ./PackageDependencies.md
[Use Case Scenarios]:              ./UseCaseScenarios.md
[Package Repository]:              ./PackageRepository.md
[Sysroot]:                         ./Sysroot.md
