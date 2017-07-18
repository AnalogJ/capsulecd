FROM golang:1.8 AS build
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

#################################################
#
# Build
#
#################################################
WORKDIR /go/src/capsulecd

RUN apt-get update && apt-get install -y --no-install-recommends \
 	openssl apt-transport-https ca-certificates curl g++ gcc libc6-dev cmake make pkg-config git && \
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
RUN GOOS=linux GOARCH=amd64 go build -tags 'static' cmd/capsulecd/capsulecd.go && \
	mv capsulecd capsulecd-linux-amd64 && \
	./capsulecd-linux-amd64 --version



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

COPY --from=build /go/src/capsulecd/capsulecd-linux-amd64 /usr/local/bin/capsulecd

RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
	curl -k -O https://packages.chef.io/files/stable/chefdk/1.5.0/debian/7/chefdk_1.5.0-1_amd64.deb && \
	dpkg -i chefdk_1.5.0-1_amd64.deb && \
	echo 'eval "$(chef shell-init bash)"' >> ~/.bash_profile

CMD ["capsulecd"]


#FROM alpine AS dist
#MAINTAINER Jason Kulatunga <jason@thesparktree.com>
#
#COPY --from=build /go/src/capsulecd/capsulecd /usr/local/bin/capsulecd
#
#CMD ["capsulecd"]