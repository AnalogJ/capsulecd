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

WORKDIR /srv/capsulecd

ENV PATH="/srv/capsulecd:/go/bin:${PATH}" \
	SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

RUN apt-get update && apt-get install -y --no-install-recommends \
 	apt-transport-https \
    ca-certificates \
    git \
    curl \
	&& rm -rf /var/lib/apt/lists/* \
	&& go get -u gopkg.in/alecthomas/gometalinter.v2 \
	&& gometalinter.v2 --install \
	&& go get github.com/Masterminds/glide

COPY ./ci/capsulecd.sh /scripts/capsulecd.sh
COPY ./ci/development.sh /scripts/development.sh

RUN /scripts/capsulecd.sh