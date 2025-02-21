
# Build Process

Each Package has a Config file in the JSON format (Config file details in [Context Structure]).

## Build All Packages

### Config phase for all Packages

- the default Package Config structure is created (with defaults filled in, look for `Config`
  in the bringauto_config module).
- the JSON Config for the Package is read into Go structure.
- the default Package Config is merged with the Config read from file. The Config file
  data has a precedence over Default one.

### Build phase for all Packages

- All Package Configs are stored in linear list
- All Configs are topological sorted by `Dependencies`,
- the Packages are build from first item of the list (head of the list) to the last Package of the
  list.

During the build the Package files installed by installation feature of the CMake are copied
to the `install_sysroot` directory located in the working directory of the builder. If Package
build files would overwrite any files already present in sysroot, the build fails (more in
[Sysroot]).

Each Package has a `IsDebug` flag. If the flag is true the Package is considered as Debug Package.
If the Package is false the Package is considered as Release.

Packages that are marked as Debug has separate sysroot dir in `install_sysroot` directory.  So the
Debug and Release Packages are not mixed together.

If there is any circular dependency between Packages in build list, the build fails.

## Build single Package

### Config phase for single Package

Same as for All build except that only Configs for given Package Group are read and managed.

### Build phase for single Package

Same as for All build except that the dependencies are ignored. (so if you have not initialized
`install_sysroot` then the Package build may fail)

If you want to build Package with all dependencies, you can add `--build-deps` option to script
call.

## Build single Package with its dependencies

### Config phase for single Package

Same as for All build except that only Configs for given Package Group and all dependencies in
these Configs recursively are read and managed.

### Build phase for single Package

Same as for All build.

## Build Packages which depends on a Package

### Config phase for single Package

Same as for All build except that only Configs of Packages which depends on given Package and its
dependencies are read and managed. If the recursive option is triggered, the process is done
recursively (More in [Use Case Scenarios]).

### Build phase for single Package

Same as for All build.

[Context Structure]: ./ContextStructure.md
[Sysroot]: ./Sysroot.md
[Use Case Scenarios]: ./UseCaseScenarios.md
