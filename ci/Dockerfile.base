ARG PACKAGE_TYPE=base

#################################################
#
# Base
# This container should not be used as a runtime environment.
# It is based off a massive build image (crossbuild) which has lots of unnecessary build tools
# It does not actually build the capsulecd executable
# It runs unit tests for each supported engine type.
#
# Use the docker containers in https://github.com/AnalogJ/capsulecd-docker as an example of what a
# proper runtime-environment for CapsuleCD looks like.
#
#################################################

FROM analogj/libgit2-crossbuild:linux-amd64 AS base
MAINTAINER Jason Kulatunga <jason@thesparktree.com>
WORKDIR /go/src/capsulecd

RUN apt-get update && apt-get install -y --no-install-recommends \
 	apt-transport-https \
    ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& go get github.com/Masterminds/glide

COPY . .

## download glide deps & move libgit2 library into expected location.
RUN glide install \
	&& mkdir -p /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
	&& cp -r /usr/local/lib/libgit2/lib/pkgconfig/. /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

RUN ci/test-build.sh ${PACKAGE_TYPE}

CMD ci/test-coverage.sh