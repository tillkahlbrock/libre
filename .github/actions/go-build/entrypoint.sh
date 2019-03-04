#!/usr/bin/env bash

set -e

if [ -z "${IMPORT}" ]; then
  IMPORT="${GITHUB_REPOSITORY}"
fi
WORKDIR="${GOPATH}/src/github.com/${IMPORT}"

mkdir -p "`dirname "${WORKDIR}"`"
ln -s "${PWD}" "${WORKDIR}"
cd "${WORKDIR}"

exec "$@"