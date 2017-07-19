#!/bin/sh
set -x

# Set temp environment vars
export LIBGIT2REPO=https://github.com/libgit2/libgit2.git
export LIBGIT2BRANCH=v0.25.0
export LIBGIT2PATH=/tmp/libgit2

# Env used during libgit2 install
export OPENSSL_SSL_LIBRARY=/tmp/openssl/install/lib
export OPENSSL_CRYPTO_LIBRARY=/tmp/openssl/install/lib
export LIBSSH2_FOUND=true
export LIBSSH2_INCLUDE_DIRS=/tmp/libssh2/install/include
export LIBSSH2_LIBRARY_DIRS=/tmp/libssh/install/lib64

# Compile & Install libgit2 (v0.23)
git clone -b ${LIBGIT2BRANCH} --depth 1 -- ${LIBGIT2REPO} ${LIBGIT2PATH}

mkdir -p ${LIBGIT2PATH}/build
cd ${LIBGIT2PATH}/build
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      ..
cmake --build . --target install

# Cleanup
# rm -r ${LIBGIT2PATH}
