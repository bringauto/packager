#!/bin/bash

set -e

shopt -s nullglob
for package_json in $(find . -name '*.json')
do
  jq '.DockerMatrix.ImageNames += [ "debian12"  ]' ${package_json} > ${package_json}.test
  mv ${package_json}.test ${package_json}
done
shopt -u nullglob