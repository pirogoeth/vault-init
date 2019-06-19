#!/usr/bin/env bash

set -eo pipefail

echo "============"
go version
echo "============"
echo

retdir=$(pwd)
cd ./source
go vet -v ./...
cd ${retdir}