# Config Structure

This document provides an example of the Config file. Note that this example does not follow the
JSON format strictly because it contains comments and the values are just sample values. For a
better understanding of the JSON format, check the `example/package` directory in this repository.

Not all fields are required, and some fields have default values.

``` json
{
  "Env": { // Environment variables used for the project
    "ENV_A": "Value A",
    "ENV_B": "Value B"
  },
  "DependsOn": [ // List of external dependencies required for this project
    "protobuf",
    "fleet-protocol-interface",
    "zlib"
  ],
  "Git": { // Details about the Git repository for fetching the project source code
    "URI": "https://github.com/bringauto/example-repo.git", // Valid Git URI that can be used with the "git clone" command
    "Revision": "v1.2.0" // Valid git hash, tag, or branch
  },
  "Build": { // Build configuration settings using CMake, including various CMake options
    "CMake": {
      "CMakeListDir": "/cmake", // Directory where the CMakeLists.txt is located. Default value is "./", path is relative to the module's Git root
      "Defines": { // CMake variables passed with the CMake -D switch
        "CMAKE_BUILD_TYPE": "Debug",
        "MY_NICE_VAR": "VarValue"
      }
    }
  },
  "Package": { // Metadata for the project Package
    "Name": "PackageName", // Package name used to construct the Package archive name
    "VersionTag": "v5.98.0", // Detailed in the VersionTag section
    "PlatformString": {
      "Mode": "auto", // Detailed in the Platform_String_Mode section
    },
    "IsLibrary": true, // If true, adds 'lib' prefix to the Package name
    "IsDevLib": true,  // If true, adds '-dev' suffix to the Package name
    "IsDebug": true    // If true, adds 'd' to the Package name (but before the -dev suffix)
  },
  "DockerMatrix": { // Specifies the Docker images from the "docker/" directory used to build this Package
    "ImageNames":  [ "ubuntu1804", "ubuntu2004", "debian11" ]
  }
}
```

## Version_Tag

`VersionTag` represents a version in normalized form.

``` plaintext
VersionTag = 'v'x'.'y'.'z
where x, y, z are from { 0, 1, 2, ... }
```

Examples:

- v1.5.9
- v0.0.5
- v5.98.0

## Platform_String_Mode

Platform String mode determines how the PlatformString is constructed.

``` plaintext
Mode = "auto" | "explicit"
```

**auto** - Constructs the PlatformString for a target machine automatically.
Example for Debian 11 (x86_64): "x86-64-debian-11"

**explicit** - User must fill in all three parts manually:

``` json
...
    "PlatformString": {
      "Mode": "explicit",
      "String": {
        "DistroName": "Debian",
        "DistroRelease": "11",
        "Machine": "x86-64"
      }
    }
...
```
