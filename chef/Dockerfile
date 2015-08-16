from ubuntu
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get update
run apt-get install curl git -y

run curl -L https://www.opscode.com/chef/install.sh | sudo bash -s -- -P chefdk
run locale-gen en_US.UTF-8

# Make Chef DK the primary Ruby/Chef development environment.
run echo 'eval "$(chef shell-init bash)"' >> ~/.bash_profile
run echo 'eval "$(chef shell-init bash)"' >> ~/.bashrc

run gem install bundler

# copy the application files to the image
workdir /srv/capsulecd
run git clone https://github.com/AnalogJ/capsulecd.git .

run bundle install --without node ruby