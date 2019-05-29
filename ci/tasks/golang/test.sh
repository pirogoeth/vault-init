#!/usr/bin/env bash

set -eo pipefail

mkdir -p /go/${PACKAGE}
cp -rv source/* /go/${PACKAGE}/

retdir=$(pwd)
cd /go/${PACKAGE}/
go test -v
cd ${retdir}