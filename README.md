# BAP - BringAuto Packager

Build and track C/C++ project dependencies for apps on any Linux distro!

BAM provides a simple way to build and maintain dependencies with almost zero learning curve and out-of-the-box integration into your workflows.

## Requirements

- Docker >= 20.10 (installed according to the official Docker documentation)
- git >= 2.25

Standalone binaries are built for Linux kernel >= 5.10.0-amd64.

## Build

### Requirements for Build

- Go >= 1.22

### Build from Source

Clone the repository and, in the repository root directory, run:

```bash
go get bringauto/bap-builder
cd bap-builder
go build bringauto/bap-builder
```

## Build Standalone Binaries

To build a complete release package, use the `build.sh` script.

Additional requirements for `build.sh`:

- zip
- uname
- sed

## Usage

The `packager` (`bap-builder`) has these commands:
 - `build-image` for building Docker images
 - `build-package` for building packages
 - `create-sysroot` for creating sysroot from already built packages

The `build-package` and `create-sysroot` commands are using Git Repository as storage for built
packages. Given Git Repository must be created before usage.

**NOTE:** Detailed use case scenarios is decribed in [UseCaseScenarios](./doc/UseCaseScenarios.md) document.

### Example

1. Create a git repository (optionally with LFS):

    ```bash
    mkdir lfsrepo && cd lfsrepo
    git init
    cd ../
    ```

2. Build all Docker images needed for the build:

    ```bash
    bap-builder build-image --context ./example --name debian12
    ```

3. Build all packages for the given distro:

    ```bash
    bap-builder build-package --context ./example --image-name debian12 --output-dir ./lfsrepo --all
    ```

4. Create sysroot for built packages:

    ```bash
    bap-builder create-sysroot --context ./example --image-name debian12 --git-lfs ./lfsrepo --sysroot-dir ./new_sysroot
    ```

**Note:** If you do not have `bap-builder` in your system path, you need to use `./bap-builder/bap-builder` instead of `bap-builder`.

## Documentation

The detailed documentation is available in the `doc` directory.

## Example

In the `example` directory, there is a simple example of how to track dependencies for the `example` project.

## Motivation

If you want to run your application directly on the host system you need to ensure that every dependency
is consistent and work well.
That's really hard to do because every disto has a different version of the same library. If you want to release
you app to a new Linux distro you need to test every dependency, solve bugs, maintain your app in time...

There is no simple way how to track, maintain and pack consistent dependency set for C/C++ project without worrying
about different version of libraries, tools and the whole environment.

The needed requirement is

- **You need to know how your dependencies works, how to compile them,
  how to update them and how to use them.**

If you want to use one of existing solutions (Connan, Nix, ...) to develop reliable system you need to

- **Learn how the package system works - how to create own package, how to integrate the system with you build workflow,**
- **You need to train everyone in your Organization to use the system and solve pitfalls...**

In other ways the public repository of the system is not authoritative you cannot rely on the package pushed
to the public repository.\
If you develop a reliable application you need to have authoritative source of the dependencies

- **You need to create your own instance of the system (or pay to someone to host instance for you)**
- **You need to create and push your own packages to obtain authoritative source of dependencies.**

as we can see the existing package managers adds huge sort of problems and requirements which must
be fulfilled.

There is no  reason for that! As a C++ programmers we already know how to compile the problem, how to link
all parts together. We have own workflows and systems to prepare a build environment.

**We just need simple way how to build and maintain our dependencies with almost zero learning curve and out-of-the
box integration into our workflows.**

BAM solve all these problems! It uses technologies and workflows that are known by every C/C++ programmer!

You can easily build and track dependencies for your project, download then and use them.

## FAQ

### Q: I am encountering a weird error

Many errors are caused by issues with the SSH connection to the Docker container or the inability to start the Docker container itself.

In this case:

- Check if there are no running Docker containers that are instances of one of the Docker images in the context directory.
- Check if there are no other containers using port `1122`.
