
# Context Structure

Context structure is a directory structure that gathers Configs needed for BAP to
work.

In the Context the definitions of Packages (Configs) and Docker images are stored.

``` plaintext
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

You can use `bap-builder build-image` feature to build docker images instead of directly invoke `docker` command.

## Package Group Name

Each Package Group can have multiple Configs.

Each Config represents one Package.

Each Config is a json file.

Each Config must have '.json' extension.

The Config format is described by [ConfigStructure]

[ConfigStructure]: ./ConfigStructure.md
