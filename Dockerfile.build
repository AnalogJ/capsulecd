###############################################################################
#
# This Dockerfile should only be used to cross-compile capsulecd for various
# OS's and Architectures. Its massive, and should not be used as a base image
# for your Dockerfiles.
#
# Usable Docker Images and Dockerfiles for different languages are located:
# - https://github.com/AnalogJ/capsulecd-docker
# - https://hub.docker.com/r/analogj/capsulecd
#
# Use `docker pull analogj/capsulecd:<language>`
#
###############################################################################
FROM analogj/libgit2-xgo
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

ARG go_version=1.13.4
ARG engine_type="golang"


WORKDIR /go/src/github.com/analogj/capsulecd

ENV PATH="/go/src/github.com/analogj/capsulecd:/go/bin:${PATH}" \
	SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Install build tooling.
RUN echo "go version: $go_version" \
    && apt-get update \
	&& apt-get install -y gcc git build-essential binutils curl apt-transport-https ca-certificates pkg-config --no-install-recommends \
	&& rm -rf /usr/share/doc && rm -rf /usr/share/man \
	&& rm -rf /var/lib/apt/lists/* \
    && apt-get clean


ENV PATH="/go/bin:/usr/local/go/bin:${PATH}" \
	GOPATH="/go:${GOPATH}" \
	SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# install go and dep
RUN which go  || (curl -fsSL "https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz" | tar -xzC /usr/local) \
    && mkdir -p /go/bin \
    && mkdir -p /go/src \
    && go get -u gopkg.in/alecthomas/gometalinter.v2 \
    && gometalinter.v2 --install \
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY ./ci/capsulecd.sh /scripts/capsulecd.sh
COPY ./ci/development.sh /scripts/development.sh

RUN /scripts/capsulecd.sh
