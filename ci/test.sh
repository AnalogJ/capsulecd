#!/usr/bin/env bash

# because capsulecd depends on compiled binaries (even for testing) we'll be building the test binaries first in the "build container" and then
# executing them in a "runtime container" to get coverage/profiling data.
#
# this script generates the test binaries in the "build container"

set -e

echo "Args: $@"
echo  "writing coverage data to /coverage/coverage-${1}.txt"
mkdir -p /coverage/
go test -race -covermode=atomic -test.coverprofile=/coverage/coverage-${1}.txt -tags="static $1" ./...
