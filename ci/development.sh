#!/usr/bin/env bash

# instructions taken from capsule.yml (we only care about a linux development environment right now)
cd /go/src/github.com/analogj/capsulecd
rm -rf vendor
dep ensure

mkdir -p /go/src/github.com/analogj/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build
cp -r /usr/local/lib/libgit2/lib/pkgconfig/. /go/src/github.com/analogj/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/


export DEV_MODE=true

/bin/bash
