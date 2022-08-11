
# Context Directory Structure

Context Directory structure is a directory structure that gathers
configurations needed for BAM to work.

In the Content Directory the definitions of packages, Docker images, ... are stored.

```
<context_directory>/
	docker/
		<docker_name>/
			Dockerfile
		...
	package/
		<package_group_name>/
			<package_config_a>.json
			<package_config_b>.json
			...
		...
```


## Docker Name

The image name is recognized by a name of a directory in the `docker/` directory.

Docker image built by Dockerfile in <docker_name> directory must be tagged by <docker_name>.

You can use `bam-builder build-image` feature to build docker images instead of directly invoke `docker` command.

## package_group_name

Each Package Group can have multiple configuration.

Each configuration represents one package for set of Docker Images and Build Types.

Each package config is a JSON TXT file.

Each package config mas have '.json' extension.

### package_config

```
{
  "Env": {
    "ENV_A": "Value A",
    "ENV_B": "Value B",
  },
  "DependsOn": [
    <package_group_name>
  ],
  "Git": {
    "URI": <uri>,
    "Revision": <GitRevision_HashTagBranch>
  },
  "Build": {
    "CMake": {
      "CMakeListDir": "<path_to_cmake_list_dir>",
      "Defines": { // CMake variables passed to CMake -D switch
        "CMAKE_BUILD_TYPE": "Debug",
        "MY_NICE_VAR": "VarValue"
      }
    }
  },
  "Package": {
    "Name": "PackageName", // package name from which to construct the package archive name
    "VersionTag": <version_tag>,
    "PlatformString": {
      "Mode": <platform_string_mode>
    },
    "IsLibrary": true, // if true add 'lib' prefix to the package name
    "IsDevLib": true,  // if true add '-dev' suffix to package name
    "IsDebug": true    // if true add 'd' to the package name (but before -dev prefix)
  },
  "DockerMatrix": { // from which Docker images names from docker/ repository this package wil be built
    "ImageNames":  [ "ubuntu1804", "ubuntu2004", "debian11" ]
  }
}
```



### uri

valid Git URi that can be used by `git clone` command

### GitRevision_HashTagBranch

Valid git Hash, Tag or branch

### path_to_cmake_list_dir

Directory where the CMakeLists.txt is located. Default value is "./"

Path is relative against project git root.

### version_tag

`version_tag` represents a version in normalized form.

```
version_tag = 'v'x'.'y'.'z
where x, y, z are from { 0, 1, 2, ... }
```

Examples:

- v1.5.9
- v0.0.5
- v5.98.0

### platform_string_mode

Platform String mode determine how the PlatformString is constructed 

```
platform_string_mode = "auto" | "any_machine" | "explicit"
```

**auto** - construct PlatformString for a target machine by an automatic way.
Debian 11 (x86_64) generated example: "x86-64-debian-11"

**any_machine** - construct platform string that is specific only for Distro name and distro release.
as a machine part of the PlatformString "any". Debian 11 (x86_64) generated example: "any-debian-11"

**explicit** - user must fill all three parts manually:

```
...
    "PlatformString": {
      "Mode": "Explicit",
      "String": {
        "DistroName": "Debian",
        "DistroRelease": "11",
        "Machine": "x86-64"
      }
    },
...
```
