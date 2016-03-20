FROM ruby:2.1.8-alpine
MAINTAINER Jason Kulatunga <jason@thesparktree.com>

RUN mkdir -p /srv/capsulecd
COPY . /srv/capsulecd
workdir /srv/capsulecd

RUN apk --update --no-cache add build_deps \
    build-base ruby-dev libc-dev linux-headers \
    openssl-dev libxml2-dev libxslt-dev \
	git && \
    bundle install --without test chef

CMD ["sh"]
#CMD ["capsulecd", "start", "--runner", "circleci", "--source", "github", "--package_type", "ruby"]
