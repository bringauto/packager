
# BAP - BringAuto Packager

Build and track C/C++ project dependencies for apps for TODO any Linux distro!

BAM - simple way how to build and maintain our dependencies with almost zero learning curve and out-of-the
box integration into our workflows.

## Usage

bap-builder build and stores all dependencies in the git repository
(LFS enabled recommended)

1. create git repository (optionally with LFS)

```
mkdir lfsrepo && cd lfsrepo
git init
cd ../
```

2. Build all docker images needed for build

```
bap-builder build-image --context ./example --name debian11
```

3. Build all packages for given distro

```
bap-builder build-package --context ./example --image-name debian11 --output-dir ./lfsrepo --all
```

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

## Requirements

- Docker >= 20.10 (installed by official Docker documentation)
- git >= 2.25

Standalone binaries are built for Linux kernel >= 5.10.0-amd64

## Build

The project requires go >= 1.22.

```
go get bringauto/bap-builder
cd bap-builder
go build bringauto/bap-builder
```

## Build standalone binaries

There is a script `build.sh` by that we can build a complete release package.

Additional requirements for `build.sh`:

- zip
- uname
- sed

## FAQ

### Q: I have got a wierd error

Many errors are caused by problem with SSH connection to the Docker container
or impossibility to start Docker container itself.

In this case

- check if there are no running docker container that 
are instance of one of the Docker Images in the context directory
- check if there are no other container that uses port `1122`

## TODO

- SFTP copy and package creation are too slow (~20 minutes for Boost)
- Refactor `error` handling (use `Log` library, ...) and error messages

