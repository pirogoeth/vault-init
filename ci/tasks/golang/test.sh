#!/usr/bin/env bash

set -eo pipefail

echo "============"
go version
echo "============"
echo

parent=$(dirname "${PACKAGE}")

mkdir -p /go/${parent}
ln -s ./source /go/${PACKAGE}

retdir=$(pwd)
cd /go/${PACKAGE}/
go mod tidy -v
go mod verify
go test -v
cd ${retdir}