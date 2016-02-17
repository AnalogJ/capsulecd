from ubuntu
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get update
run apt-get install curl wget git make build-essential -y

run wget https://opscode-omnibus-packages.s3.amazonaws.com/ubuntu/12.04/x86_64/chefdk_0.6.2-1_amd64.deb
run dpkg -i chefdk_0.6.2-1_amd64.deb
run locale-gen en_US.UTF-8

# Make Chef DK the primary Ruby/Chef development environment.
env PATH /opt/chefdk/bin:/root/.chefdk/gem/ruby/2.1.0/bin:/opt/chefdk/embedded/bin:$PATH
env GEM_ROOT /opt/chefdk/embedded/lib/ruby/gems/2.1.0
env GEM_HOME /root/.chefdk/gem/ruby/2.1.0
env GEM_PATH /root/.chefdk/gem/ruby/2.1.0:/opt/chefdk/embedded/lib/ruby/gems/2.1.0

run which ruby
run gem install bundler

# copy the application files to the image
workdir /srv/capsulecd
run git clone https://github.com/AnalogJ/capsulecd.git .

run bundle install --without node ruby python

cmd ["bundle", "exec", "ruby", "cli.rb", "--type", "chef"]
