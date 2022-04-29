#!/usr/bin/env bash

set -e

INSTALL_DIR="./install_dir"

if [[ -d ${INSTALL_DIR} ]]; then
  echo "${INSTALL_DIR} already exist. Delete it pls" >&2
  exit 1
fi

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

mkdir -p "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/tools"

cp bap-builder/bap-builder                 "${INSTALL_DIR}/"
cp -r doc                                  "${INSTALL_DIR}/"
cp README.md                               "${INSTALL_DIR}/"
cp LICENSE                                 "${INSTALL_DIR}/"
cp tools/lsb_release/lsb_release           "${INSTALL_DIR}/tools/"
cp tools/lsb_release/lsb_release.txt       "${INSTALL_DIR}/tools/"
cp tools/lsb_release/lsb_release_README.md "${INSTALL_DIR}/tools/"
cp tools/uname/uname_README.md             "${INSTALL_DIR}/tools/"
cp tools/uname/uname                       "${INSTALL_DIR}/tools/"
cp tools/uname/uname.txt                   "${INSTALL_DIR}/tools/"


VERSION=$(sed -E -n 's/version=([^=]+)/\1/p' < version.txt)
MACHINE=$(uname -m | sed -E 's/_/-/')
zip -r "bringauto-packager_v${VERSION}_${MACHINE}-linux.zip" ${INSTALL_DIR}/

rm -fr install_dir