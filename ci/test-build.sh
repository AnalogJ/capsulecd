#!/usr/bin/env bash

# because capsulecd depends on compiled binaries (even for testing) we'll be building the test binaries first in the "build container" and then
# executing them in a "runtime container" to get coverage/profiling data.
#
# this script generates the test binaries in the "build container"

set -e

PACKAGE_NAME="capsulecd"

for d in $(go list ./... | grep -v vendor); do
    # determine the output path
    OUTPUT_PATH=$(echo "$d" | sed -e "s/^${PACKAGE_NAME}\///")

    go test -race -covermode=atomic -tags="static $1" -c -o=${OUTPUT_PATH}/test_binary $d
done