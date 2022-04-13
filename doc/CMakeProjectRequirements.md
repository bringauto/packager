
# CMake Project Requirements

BAP Builder can build and maintain only CMake based projects.

Requirements

- Project must be able to be installed by GNU Make - `make install`
- Project must NOT override `CMAKE_INSTALL_PREFIX` CMake variable - it's used for the project installation to
  a given directory and package creation. If you override it the build fail!