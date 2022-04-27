#!/usr/bin/env bash

set -e

go get bringauto/bap-builder
go get bringauto/tools/lsb_release
go get bringauto/tools/uname

pushd bap-builder
  echo "Compile bap_builder"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w'
popd

pushd tools/uname
  echo "Compile tools/uname"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w'
popd

pushd tools/lsb_release
  echo "Compile tools/lsb_release"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w'
popd

if [ -d install_dir ]; then
  echo "install_dir already exist. Delete it pls" >&2
  exit 1
fi
mkdir -p install_dir
mkdir -p install_dir/tools

cp bap-builder/bap-builder                 install_dir/
cp -r doc                                  install_dir/
cp README.md                               install_dir/
cp LICENSE                                 install_dir/
cp tools/lsb_release/lsb_release           install_dir/tools/
cp tools/lsb_release/lsb_release.txt       install_dir/tools/
cp tools/lsb_release/lsb_release_README.md install_dir/tools/
cp tools/uname/uname_README.md             install_dir/tools/
cp tools/uname/uname                       install_dir/tools/
cp tools/uname/uname.txt                   install_dir/tools/

