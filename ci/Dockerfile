ARG engine_type="golang"
FROM analogj/libgit2-crossbuild:linux-amd64-slim as base

ARG go_version=1.10.8
ARG engine_type="golang"


WORKDIR /go/src/github.com/analogj/capsulecd

# Install build tooling.
RUN echo "go version: $go_version" \
    && apt-get update \
	&& apt-get install -y gcc git build-essential binutils curl apt-transport-https ca-certificates pkg-config --no-install-recommends \
	&& rm -rf /usr/share/doc && rm -rf /usr/share/man \
	&& rm -rf /var/lib/apt/lists/* \
    && apt-get clean


ENV PATH="/go/bin:/usr/local/go/bin:${PATH}" \
	GOPATH="/go:${GOPATH}" \
	SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
    PKG_CONFIG_PATH="/usr/lib/pkgconfig/:/usr/local/lib/pkgconfig/:/usr/local/lib/libgit2/lib/pkgconfig:/usr/local/lib/openssl/lib/pkgconfig:/usr/local/lib/libssh2"

# install go and dep
RUN curl -fsSL "https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz" | tar -xzC /usr/local \
    && mkdir -p /go/bin \
    && mkdir -p /go/src \
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . .

## download deps & move libgit2 library into expected location.
RUN git --version \
    && dep ensure \
	&& mkdir -p /go/src/github.com/analogj/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build \
	&& cp -r /usr/local/lib/libgit2/lib/pkgconfig/. /go/src/github.com/analogj/capsulecd/vendor/gopkg.in/libgit2/git2go.v25/vendor/libgit2/build/ \
	&& ./ci/test-build.sh ${engine_type}

##################################################
##
## Dynamically selected runtime container using Build Arg
## engine_type
##
##################################################
FROM analogj/capsulecd:$engine_type

ARG go_version=1.10.8
ARG engine_type="golang"

WORKDIR /go/src/github.com/analogj/capsulecd

## Install build tooling.
#RUN echo "go version: $go_version" \
#    && apt-get update \
#	&& apt-get install -y curl git --no-install-recommends \
#	&& rm -rf /usr/share/doc && rm -rf /usr/share/man \
#	&& rm -rf /var/lib/apt/lists/* \
#    && apt-get clean


ENV PATH="/go/bin:/usr/local/go/bin:${PATH}" \
	GOPATH="/go:${GOPATH}"

RUN go || curl -fsSL "https://storage.googleapis.com/golang/go${go_version}.linux-amd64.tar.gz" | tar -xzC /usr/local



COPY --from=base /go/src/github.com/analogj/capsulecd /go/src/github.com/analogj/capsulecd

ENTRYPOINT ["ci/test-execute.sh"]
