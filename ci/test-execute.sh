#!/usr/bin/env bash


# because capsulecd depends on compiled binaries (even for testing) we'll be building the test binaries first in the "build container" and then
# executing them in a "runtime container" to get coverage/profiling data.
#
# this script executes the test binaries in the "runtime container"

set -e

echo "Args: $@"
echo  "/coverage/coverage-${1}.txt"

mkdir -p /coverage
echo "" > "/coverage/coverage-${1}.txt"


echo "Printing current folder structure"

for d in $(go list ./...); do
    # determine the output path
    TEST_BINARY_PATH=$(echo "$d" | sed -e "s/^github.com\/analogj\/capsulecd\///")
    TEST_BINARY="test_binary_${1}"

    echo "Looking for TEST BINARY: ${TEST_BINARY_PATH}/${TEST_BINARY}"

    if [ -f "${TEST_BINARY_PATH}/${TEST_BINARY}" ]; then
        echo "Found TEST BINARY"
        pushd ${TEST_BINARY_PATH}

        eval "./${TEST_BINARY} -test.coverprofile=profile.out"
        if [ -f profile.out ]; then
            cat profile.out >> "/coverage/coverage-${1}.txt"
            rm profile.out
        fi
        popd
    fi
done

ls -alt /coverage
