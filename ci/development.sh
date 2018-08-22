#!/usr/bin/env bash

# instructions taken from capsule.yml (we only care about a linux development environment right now)
cd /go/src/capsulecd
rm -rf vendor
glide install

mkdir -p vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/
cp /usr/local/linux/lib/pkgconfig/libgit2.pc vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/libgit2.pc
. /scripts/toolchains/linux/linux-build-env.sh

export DEV_MODE=true

/bin/bash