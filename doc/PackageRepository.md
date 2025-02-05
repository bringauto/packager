# Package Repository

The Package Repository (or Git Lfs) is used for storage of built Packages. It is a git repository
at the same time. The git tool is used for managing consistent storage of Packages.

## Behaviour

### Package Repository consistency check

At the start of `build-package` and `create-sysroot` commands the Package Repository consistency
check is performed. The check consists of these steps:

1) Git status check
   - If the git status in Package Repository is not empty, the state of Repository is considered as
   non-consistent and the script ends with error. The user should then clean the Repository and
   continue. It is a bug if the non-consistent state is caused by `bap-builder` itself.
2) Comparing Configs currently in Context and built Packages in Package Repository
   - If in Package Repository there is any version of Package which is not in Context, the Package
   Repository is considered as non-consistent and the script ends with error. User needs to fix
   this problem manually and then continue.
   - If some Packages are in Context and are not in Package Repository, only warning is printed.
     
All files in `<DISTRO_NAME>/<DISTRO_VERSION/MACHINE_TYPE>` are checked, so any other files in this
directory (alongside Package directories) will be counted as an error. User can't add any files
here manually.

### Managing Packages in Package Repository

Following rules ans mechanisms ensures that the Package Repository is always consistent.

- Each succesfully built Package by `build-package` command is copied to Package Repository
(specified by cli flag) to specific path -
`<DISTRO_NAME>/<DISTRO_VERSION/MACHINE_TYPE/PACKAGE_NAME>`
- After all Packages are succesfully built, all copied files in Package Repository are git
committed and remain in Repository
- If any build fails or the script is interrupted, all copied Packages are removed from
Repository
