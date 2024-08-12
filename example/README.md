# BringAuto Package Example Context Directory

This example context directory is used for testing and guidance.

## Helper Scripts

- **`add_docker_to_matrix.sh`** - Adds a new Docker image to all package JSON files in the context directory. In the example, it adds the `ubuntu2310` Docker image to all package JSON files.

- **`change_docker_name.sh`** - Changes the Docker image name in all package JSON files in the context directory. In the example, it changes `fleet-os` to `fleet-os-2`.

## Package JSON Structure

- **`Env`**: Environment variables used for the project.

- **`DependsOn`**: List of external dependencies required for this project.

- **`Git`**: Details about the Git repository for fetching the project source code.
  - **`URI`**: The URL of the Git repository.
  - **`Revision`**: The specific version tag or commit identifier.

- **`Build`**: Contains build configuration settings using CMake, including various CMake options.

- **`Package`**: Defines metadata for the project package.
  - **`Name`**: The name of the package.
  - **`VersionTag`**: The version tag for the package.
  - **`PlatformString`**: Specifies the platform compatibility mode.
  - **`IsLibrary`**: Indicates if the project is a library.
  - **`IsDevLib`**: Specifies if the library is a development library.
  - **`IsDebug`**: Indicates if the build is intended for debugging.

- **`DockerMatrix`**: Lists Docker images for different platforms used for building or testing the project.
