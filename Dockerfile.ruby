from ruby:2.1
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get install -y git
run gem install bundler

# copy the application files to the image
workdir /srv/capsulecd
run git clone https://github.com/AnalogJ/capsulecd.git .

run bundle install --with test --without chef node python
run bundle exec rake spec:ruby
CMD ["bundle", "exec", "capsulecd", "start", "--runner", "circleci", "--source", "github", "--package_type", "ruby"]