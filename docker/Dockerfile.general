from phusion/passenger-full
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get install -y git
run gem install bundler

# copy the application files to the image
#run git clone https://github.com/AnalogJ/capsulecd.git .

RUN mkdir -p /srv/capsulecd/base
RUN mkdir -p /srv/capsulecd/chef
RUN mkdir -p /srv/capsulecd/node
RUN mkdir -p /srv/capsulecd/ruby
COPY base/ /srv/capsulecd/base/
COPY chef/ /srv/capsulecd/chef/
COPY node/ /srv/capsulecd/node/
COPY ruby/ /srv/capsulecd/ruby/
COPY cli.rb /srv/capsulecd/
COPY Gemfile /srv/capsulecd/
#COPY Gemfile.lock /srv/capsulecd/

RUN ls -alt /srv/capsulecd

workdir /srv/capsulecd

run bundle install --path vendor/bundle --without ruby chef --with github


#CMD ["bash"]
CMD ["bundle", "exec", "ruby", "cli.rb", "--source", "github", "--type", "node"]
# bundle exec ruby cli.rb --runner circleci --source github --type node
