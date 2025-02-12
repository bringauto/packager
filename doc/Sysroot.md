# Sysroot

BAP is creating its own sysroot directories when building Packages. The separated sysroot
directories are created for both debug and release Packages. All Package build files are copied to
these directories. The desired behaviour is to ensure that no Package build files are being
ovewritten by another Package build files. To ensure this, following mechanisms have been
introduced to BAP.

## Sysroot consistency mechanisms

- At the start of `build-package` command the both debug and release sysroots are checked. If it
isn't empty, the warning is printed.

- During Package builds, the Package build files are copied to the sysroot directory
(`install_sysroot`). If any of the file is already present in the sysroot, the error is printed
that the Package tries to overwrite files in sysroot, which would corrupt consistency of sysroot
directory. If Package doesn't try to overwrite any files, the build proceeds and Package files are
added to the Package Repository.

- Copied Package names are added to `built_packages.json` file in `install_sysroot` directory. When
the `build-package` command with `--build-deps-on` option is used, it is expected that the Package
with its dependencies are already in sysroot. If it is not (the Package is not in `built_packages.json` file), the error is printed and build fails.

- When `create-sysroot` command is used, all Packages in Package Repository for given target platform
files are copied to new sysroot directory. Because of the sysroot consistency mechanism this new
sysroot will also be consistent.

## Notes

- The `install_sysroot` directory is not being deleted at the end of BAP execution (for a backup
reason). If the user wants to build the same Packages again, it is a common practice to delete this
directory manually. If it is not deleted and same Packages are build, the build obviously fails
because same Package wants to overwrite its own files, which are already present in sysroot.


