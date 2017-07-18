FROM golang:1.8 AS build
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

#################################################
#
# Build
#
#################################################
WORKDIR /go/src/capsulecd

RUN apt-get update && apt-get install -y --no-install-recommends \
 	apt-transport-https ca-certificates curl g++ gcc libc6-dev cmake make pkg-config git libssl-dev && \
	rm -rf /var/lib/apt/lists/* && \
	go get github.com/Masterminds/glide


# Build openSSL from source
#RUN cd $HOME && git clone https://github.com/openssl/openssl.git \
#	&& cd openssl \
#	&& git checkout tags/OpenSSL_1_0_2f \
#	&& mkdir -p build \
#	&& ./config threads no-shared --prefix=/root/openssl/install -fPIC -DOPENSSL_PIC \
#	&& make depend \
#	&& make \
#	&& make install \
#	&& ldconfig

#ENV OPENSSL_STATIC=1 \
#    OPENSSL_ROOT_DIR=/root/openssl \
#    OPENSSL_LIB_DIR=/root/openssl/lib \
#    OPENSSL_INCLUDE_DIR=/root/openssl/include


# Build libssh2 from source
RUN cd $HOME && git clone -b "libssh2-1.7.0" https://github.com/libssh2/libssh2.git \
	&& cd libssh2 \
	&& cmake -DBUILD_SHARED_LIBS=OFF . \
	&& cmake --build . \
	&& make \
	&& make install \
	&& ldconfig


# Build libgit2 from source
RUN cd $HOME && git clone -b "maint/v0.25" https://github.com/libgit2/libgit2.git \
	&& cd libgit2 \
	&& cmake -DTHREADSAFE=ON \
			-DBUILD_CLAR=OFF \
			-DBUILD_SHARED_LIBS=OFF \
			-DCMAKE_C_FLAGS=-fPIC \
			-DCMAKE_INSTALL_PREFIX=/usr \
			-DCMAKE_BUILD_TYPE="RelWithDebInfo" \
			. \
	&& cmake --build . \
	&& make \
	&& make install \
	&& ldconfig

# set env variables
#ENV LIBSSH2="$HOME/libssh2" \
#	OPENSSL="$HOME/openssl" \
#	LIBGIT2="$HOME/libgit2"
#
#ENV PKG_CONFIG_PATH="$PKG_CONFIG_PATH:$OPENSSL/install/lib/pkgconfig:$LIBSSH2/src:$LIBGIT2" \
#	BUILD="/go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build" \
#	FLAGS="-I/usr/local/include -L/usr/local/lib -L/usr/lib/x86_64-linux-gnu -L/usr/local/lib -lssh2 -lssl -lcrypto -ldl -lcrypto -ldl -lgit2 -lssh2 -lrt -lpthread -lssl -lcrypto -ldl -lz"
#
#ENV CGO_LDFLAGS="$LIBGIT2/libgit2.a $OPENSSL/libcrypto.a $OPENSSL/libssl.a $LIBSSH2/build/src/libssh2.a -L$BUILD $FLAGS" \
#    CGO_CFLAGS="-I/go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/include"
COPY . .


## download glide deps & move libgit2 library into expected location.
RUN glide install \
	&& mkdir -p /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
	&& cp /root/libgit2/libgit2.pc /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
	&& GOOS=linux GOARCH=amd64 go build -tags 'static' cmd/capsulecd/capsulecd.go \
	&& mv capsulecd capsulecd-linux-amd64 \
	&& ./capsulecd-linux-amd64 --version
#
#RUN glide install && \
#	ln -s /tmp/libgit2/build /go/src/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build
#
## build capsulecd executable
#RUN GOOS=linux GOARCH=amd64 go build -tags 'static' cmd/capsulecd/capsulecd.go && \
#GOOS=linux GOARCH=amd64 go build cmd/capsulecd/capsulecd.go
#RUN GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" cmd/capsulecd/capsulecd.go && \



# pkg-config --list-all
# CGO_ENABLED=0
# -ldflags '-extldflags "-static"'
# -ldflags "-linkmode external -extldflags -static"
#	file capsulecd && \
#	ldd capsulecd

##################################################
##
## Dist
##
##################################################
#
#FROM debian:jessie AS dist
#MAINTAINER Jason Kulatunga <jason@thesparktree.com>
#
#COPY --from=build /go/src/capsulecd/capsulecd-linux-amd64 /usr/local/bin/capsulecd
#
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates openssl && \
	curl -k -O https://packages.chef.io/files/stable/chefdk/1.5.0/debian/7/chefdk_1.5.0-1_amd64.deb && \
	dpkg -i chefdk_1.5.0-1_amd64.deb && \
	echo 'eval "$(chef shell-init bash)"' >> ~/.bashrc

#CMD ["capsulecd"]
