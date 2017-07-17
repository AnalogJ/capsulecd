FROM golang:1.8 AS build
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

#################################################
#
# Build
#
#################################################
WORKDIR /go/src/capsulecd

RUN apt-get update && apt-get install -y --no-install-recommends git curl cmake && \
	rm -rf /var/lib/apt/lists/* && \
	curl https://glide.sh/get | sh

# build libgit2 from source, because the libgit2-dev package on jessie is ancient.
RUN cd /tmp && \
	git clone -b "maint/v0.25" https://github.com/libgit2/libgit2.git && \
	mkdir -p libgit2/build && \
	cd libgit2/build && \
	cmake .. && \
	cmake --build . --target install && \
	cp /usr/local/lib/libgit2.so* /usr/lib

COPY . .

# download glide deps
RUN glide install

# build capsulecd executable
RUN go build --ldflags '-extldflags "-static"' cmd/capsulecd/capsulecd.go && \
	./capsulecd --version

#CMD ["capsulecd", "start", "--runner", "circleci", "--source", "github", "--package_type", "ruby"]

#################################################
#
# Dist
#
#################################################

FROM debian:jessie AS dist
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

COPY --from=build /go/src/capsulecd/capsulecd /usr/local/bin/capsulecd

CMD ["capsulecd"]
