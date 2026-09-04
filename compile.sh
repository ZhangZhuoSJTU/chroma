#!/bin/bash
# chroma's CLI is a nested Go module under cmd/chroma (its own go.mod).
set -e
cd "$(dirname "$0")/cmd/chroma"
go build -buildvcs=false -o ../../executable .
