#!/usr/bin/env bash

set -eo pipefail

echo "============"
go version
echo "============"
echo

retdir=$(pwd)
cd ./source
go test -v ./...
cd ${retdir}