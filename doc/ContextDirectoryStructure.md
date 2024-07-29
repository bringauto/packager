
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

You can use `bap-builder build-image` feature to build docker images instead of directly invoke `docker` command.

## package_group_name

Each Package Group can have multiple configuration.

Each configuration represents one package for set of Docker Images and Build Types.

Each package config is a JSON TXT file.

Each package config mas have '.json' extension.

The Package JSON format is described by [PackageJSONStructure]


[PackageJSONStructure]: ./PackageJSONStructure.md