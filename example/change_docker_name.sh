#!/bin/bash

set -e

command='( .DockerMatrix.ImageNames[] | select(. == $imageNameOld) ) |= $imageNameNew'

shopt -s nullglob
for package_json in $(find . -name '*.json')
do
  jq "${command}" --arg imageNameOld ubuntu1804 --arg imageNameNew ubuntu2204 ${package_json} > ${package_json}.test
  mv ${package_json}.test ${package_json}
done
shopt -u nullglob