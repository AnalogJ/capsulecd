#!/usr/bin/env bash


# because capsulecd depends on compiled binaries (even for testing) we'll be building the test binaries first in the "build container" and then
# executing them in a "runtime container" to get coverage/profiling data.
#
# this script executes the test binaries in the "runtime container"

set -e
mkdir -p /coverage

echo "" > /coverage/coverage-$1.txt

PACKAGE_NAME="capsulecd"

for d in $(go list ./... | grep -v vendor); do
    # determine the output path
    TEST_BINARY_PATH=$(echo "$d" | sed -e "s/^${PACKAGE_NAME}\///")
    TEST_BINARY="${TEST_BINARY_PATH}/test_binary"

    if [ -f "${TEST_BINARY}" ]; then
        pushd ${TEST_BINARY_PATH}
        ./test_binary -test.coverprofile=profile.out
        if [ -f profile.out ]; then
            cat profile.out >> /coverage/coverage-$1.txt
            rm profile.out
        fi
        popd
    fi
done

ls /coverage