from phusion/passenger-full
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get install -y git curl
run gem install bundler
run npm install -g bower

#install pip
run curl -o /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py
run python /tmp/get-pip.py

# copy the application files to the image
#run git clone https://github.com/AnalogJ/capsulecd.git .

RUN mkdir -p /srv/capsulecd/bin
RUN mkdir -p /srv/capsulecd/lib
RUN mkdir -p /srv/capsulecd/spec
RUN mkdir -p /srv/capsulecd/.git
COPY bin/ /srv/capsulecd/bin/
COPY lib/ /srv/capsulecd/lib/
COPY .git/ /srv/capsulecd/.git/

COPY capsulecd.gemspec /srv/capsulecd/
COPY .rspec /srv/capsulecd/
COPY Gemfile /srv/capsulecd/
COPY Rakefile /srv/capsulecd/
#COPY Gemfile.lock /srv/capsulecd/

RUN ls -alt /srv/capsulecd

workdir /srv/capsulecd

run bundle install --path vendor/bundle --with github

CMD ["bash"]
#CMD ["bundle", "exec", "capsulecd", "start", "--runner", "circleci", "--source", "github", "--package_type", "node"]
