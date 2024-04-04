#!/usr/bin/env bash

set -e

INSTALL_DIR="$1"
TOOLS_INSTALL_DIR="$2"
TMP_DIR="/tmp/toolchain-install"

TOOLS_PACKAGE_URI="https://github.com/bringauto/packager/releases/download/v0.3.0/bringauto-packager-tools_v0.3.0_x86-64-linux.zip"

if [[ ${INSTALL_DIR} = "" ]]
then
  echo "Specify toolchain absolute install dir path as a first argument!" >&2
  exit 1
fi

if [[ ${TOOLS_INSTALL_DIR} = "" ]]
then
  echo "Specify tools install dir absolute path as a second argument!" >&2
  exit 1
fi

if ! [[ -d ${INSTALL_DIR} ]]
then
  echo "Install dir '${INSTALL_DIR}' does not exist"
fi


function install_tools() {
  if [[ -d ${TMP_DIR} ]]
  then
    echo "TMP dir '${TMP_DIR}' exist"
  fi
  mkdir -p "${TMP_DIR}"
  mkdir -p "${TOOLS_INSTALL_DIR}"

  pushd "${TMP_DIR}"
    wget ${TOOLS_PACKAGE_URI} \
      -O "bringauto-packager-tools.zip"
    unzip bringauto-packager-tools.zip
    rm bringauto-packager-tools.zip
    directory_name="$(echo ./*)"
    mv "${directory_name}"/* "${TOOLS_INSTALL_DIR}"
    rm -r "${directory_name}"
  popd
  rm -r "${TMP_DIR}"

  chmod +x "${TOOLS_INSTALL_DIR}/lsb_release"
  chmod +x "${TOOLS_INSTALL_DIR}/uname"
  echo 'PATH='"${TOOLS_INSTALL_DIR}"'/:$PATH' >> /root/.bashrc
}


install_tools
