FROM capsulecd-linux-amd64:latest AS libgit2
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

# Install Go
ENV GO_VERSION 1.8.3
RUN curl -fsSL "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" \
	| tar -xzC /usr/local

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go


#################################################
#
# Build
#
#################################################
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
	&& cp -r /tmp/libgit2/build/. /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/

#RUN GOOS=linux GOARCH=amd64 go build -tags 'static' cmd/capsulecd/capsulecd.go \
#	&& mv capsulecd capsulecd-linux-amd64 \
#	&& ./capsulecd-linux-amd64 --version
#RUN glide install \
#	&& mkdir -p /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
#	&& cp /root/libgit2/libgit2.pc /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
#	&& GOOS=linux GOARCH=amd64 go build -tags 'static' cmd/capsulecd/capsulecd.go \
#	&& mv capsulecd capsulecd-linux-amd64 \
#	&& ./capsulecd-linux-amd64 --version