from ruby:2.1
maintainer Jason Kulatunga <jk17@ualberta.ca>

run \
    gem install bundler


# copy the application files to the image
workdir /srv/capsulecd
copy . /srv/capsulecd/
run bundle install --path vendor/bundle