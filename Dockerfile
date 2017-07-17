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
	go get github.com/Masterminds/glide

#RUN cd /tmp && \
#	git clone -b "maint/v0.25" https://github.com/libgit2/libgit2.git && \
#	mkdir -p libgit2/build && \
#	cd libgit2/build && \
#	cmake .. && \
#	cmake --build . --target install && \
#	cp /usr/local/lib/libgit2.so* /usr/lib

COPY . .

# download glide deps
RUN glide install --strip-vcs

## build libgit2 from source, because the libgit2-dev package on jessie is ancient.
RUN cd vendor/gopkg.in/libgit2/git2go.v25 && \
	make install-static


#RUN cd vendor/gopkg.in/libgit2/git2go.v25/ && \
#	git submodule update --init && \
#	cd vendor/libgit2 && \
#	mkdir -p install/lib && \
#	mkdir -p build && \
#	cd build && \
#	cmake -DTHREADSAFE=ON \
#		  -DBUILD_CLAR=OFF \
#		  -DBUILD_SHARED_LIBS=OFF \
#		  -DCMAKE_C_FLAGS=-fPIC \
#		  -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
#		  -DCMAKE_INSTALL_PREFIX=../install \
#		  .. && \
#	cmake --build . --target install
#
# build capsulecd executable
RUN go build -tags 'static' cmd/capsulecd/capsulecd.go && \
	./capsulecd --version



# CGO_ENABLED=0
# -ldflags '-extldflags "-static"'
# -ldflags "-linkmode external -extldflags -static"
#	file capsulecd && \
#	ldd capsulecd

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
