#!/usr/bin/env bash

set -e
mkdir -p /coverage
echo "" > /coverage/coverage-$1.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic -tags="static $1" $d
    if [ -f profile.out ]; then
        cat profile.out >> /coverage/coverage-$1.txt
        rm profile.out
    fi
done

ls /coverage