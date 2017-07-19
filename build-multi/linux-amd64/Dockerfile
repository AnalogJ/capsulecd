FROM multiarch/crossbuild

RUN apt-get update \
 && apt-get install -q -y pkg-config

COPY scripts /scripts
WORKDIR /app

RUN /scripts/configure.sh
