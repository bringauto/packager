
# Use Case List

List of use-cases and all usage-related options needed to understand intent and full-feature set
of Packager

## Build Package

### Dependencies

Each use-case is described by a simple mermaid diagram which describes dependency graph of noted
packages.

Green color indicates which packages will be built.

Arrows indicate dependency (build deps) between packages. When package A depends on package B the
following is written.

```mermaid
graph TD
    A
    A --> B
```

### Build Package - without Dependencies

Build single package (F) without any dependencies.

It expects all package dependencies are already build and installed into build sysroot directory.

```mermaid
graph TD
    A
    A --> B
    A --> C
    B --> D
    B --> E
    C --> F
    C --> G
    D --> H
    E --> I
    F --> J
    G --> K
    J --> L
    J --> M

    style F color:green;
```

**Command**

```bash
packager --context ./example \
  --image-name debian \
  --name F \
  --output-dir ./git-lfs-repo
```

### Build Package - with Dependencies

Build all dependencies (J, K, L) of the package (F) before the package (F) is build.

```mermaid
graph TD
    A
    A --> B
    A --> C
    B --> D
    B --> E
    C --> F
    C --> G
    D --> H
    E --> I
    F --> J
    G --> K
    J --> L
    J --> M

    style F color:green;
    style J color:green;
    style L color:green;
    style M color:green;
    linkStyle 8 stroke:green,stroke-width:2px;
    linkStyle 10 stroke:green,stroke-width:2px;
    linkStyle 11 stroke:green,stroke-width:2px;
```

**Command**

```bash
packager --context ./example \
  --image-name debian \
  --name F \
  --build-deps \
  --output ./git-lfs-repo
```

### Build Package - with Depends on Packages

Build packages (C) which depends on package (F) with its dependencies (G, K) without package (F)
and its dependencies (J, L, M).

```mermaid
graph TD
    A
    A --> B
    A --> C
    B --> D
    B --> E
    C --> F
    C --> G
    D --> H
    E --> I
    F --> J
    G --> K
    J --> L
    J --> M

    style C color:green;
    style G color:green;
    style K color:green;
    linkStyle 5 stroke:green,stroke-width:2px;
    linkStyle 9 stroke:green,stroke-width:2px;
```

**Command**

```bash
packager --context ./example \
  --image-name debian \
  --name F \
  --build-deps-on \
  --output ./git-lfs-repo
```


### Build Package - with Depends on Packages Recursive

Build packages (C, A) which depends on package (F) recursively with its dependencies
(G, K, B, D, E, H, I) without package (F) and its dependencies (J, L, M).

```mermaid
graph TD
    A
    A --> B
    A --> C
    B --> D
    B --> E
    C --> F
    C --> G
    D --> H
    E --> I
    F --> J
    G --> K
    J --> L
    J --> M

    style C color:green;
    style G color:green;
    style K color:green;
    style A color:green;
    style B color:green;
    style D color:green;
    style E color:green;
    style H color:green;
    style I color:green;
    linkStyle 1 stroke:green,stroke-width:2px;
    linkStyle 4 stroke:green,stroke-width:2px;
```

**Command**

```bash
packager --context ./example \
  --image-name debian \
  --name F \
  --build-deps-on-recursive \
  --output ./git-lfs-repo
```

## Create Sysroot

When all packages are build and stored as part of `--output-dir` directory the sysroot can be
created.

The packager takes all archives for a given Image name and Architecture
and unzip them into the specified directory.

TODO: Does it preserve UNIX permissions?

**Command**

Creates sysroot from packages in Git Lfs for `debian` image in `new_sysroot/` directory.

```bash
packager create-sysroot
  --context ./example \
  --image-name debian \
  --git-lfs ./git-lfs-repo \
  --sysroot-dir new_sysroot
```
