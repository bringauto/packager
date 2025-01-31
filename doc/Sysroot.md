# Sysroot

BAP is creating its own sysroot directory when building Packages. All Package build files are
copied to this directory. The desired behaviour is to ensure that no Package build files are being
ovewritten by another Package build files. To ensure this, below described mechanisms have been
introduced to BAP.

## Behaviour

At the start of `build-package` command the `install_sysroot` directory is checked. If it isn't empty, the warning is printed.

During Package builds, the Package build files are copied to sysroot directory (`install_sysroot`).
If any of the file is already in the sysroot, the error is printed that the Package tries to
overwrite files in sysroot, which would corrupt consistency of sysroot directory. If Package
doesn't try to ovewrite any files, the build proceeds and Package files are added to Package
Repository.

When `create-sysroot` command is used, all Packages in Package Repository for given target platform
files are copied to new sysroot directory. Because of the sysroot consistency mechanism this new sysroot will also be consistent.

## Notes

- The `install_sysroot` directory is not being deleted at the end of BAP execution (for a backup
reason). If the user wants to build the same Packages again, it is a common practice to delete this
directory manually. If it is not deleted and same Packages are build, the build obviously fails
because same Package wants to overwrite its own files, which are already present in sysroot.


