#!/bin/sh
set -x

# Set temp environment vars
export REPO=https://github.com/libssh2/libssh2
export BRANCH=libssh2-1.7.0
export REPO_PATH=/tmp/libssh2
export PKG_CONFIG_PATH="/usr/lib/pkgconfig/:/usr/local/lib/pkgconfig/"
export PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:/tmp/libgit2/install/lib/pkgconfig:/tmp/openssl/install/lib/pkgconfig:/tmp/libssh2/build/src"

# Env used during libssh2 install
export OPENSSL_ROOT_DIR=/tmp/openssl/install
export OPENSSL_LIBRARIES=/tmp/openssl/install/lib
export OPENSSL_INCLUDE_DIR=/tmp/openssl/install/include

# Compile & Install libgit2 (v0.23)
git clone -b ${BRANCH} --depth 1 -- ${REPO} ${REPO_PATH}

mkdir -p ${REPO_PATH}/build
cd ${REPO_PATH}/build
cmake -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DBUILD_EXAMPLES=OFF \
      -DBUILD_TESTING=OFF \
      -DCMAKE_INSTALL_PREFIX=../install \
      ..
cmake --build . --target install
