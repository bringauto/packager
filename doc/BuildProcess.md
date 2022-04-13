
# Build Process

Each package has a Config file in the JSON format. (Config file details in [Context Directory Structure])

## Build All packages

**Config phase**

- the default Package config structure is created (with defaults filled in, look for `Config`
  in the config module).
- the JSON config for the package is read into Go structure.
- the default Package config is merged with the confile read from file. The config file
  data has a precedence over Default one.

**Build phase**

- the package configs are stored in linear list
- package config are topological sorted by `Dependencies`,
- the packages are build from first item of the list (head of the list) to the las package of the list.

During the build the package files installed by installation feature of the CMake are copied
to the `install_sysroot` directory located in the working directory of the builder.

Each package has a `IsDebug` flag. If the flag is true the package is considered as Debug package.
If the package is false the package is considered as Release.

Packages that are marked as Debug has separate sysroot dir in `install_sysroot` directory.  So the Debug and release
packages are not mixed together.

## Build single package

**Config phase**

Same as for All build except that only configs for given `package_group` are read and managed.

**Build phase**

Same as fo All build except that the dependencies are ignored. (so If you have not initialized `install_sysroot`
then the package build fail)





[Context Directory Structure]: ./ContextDirectoryStructure.md