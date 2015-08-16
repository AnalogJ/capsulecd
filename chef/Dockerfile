from spheromak/docker-chefdk
maintainer Jason Kulatunga <jk17@ualberta.ca>

run apt-get install -y git

# Make Chef DK the primary Ruby/Chef development environment.
run echo 'eval "$(chef shell-init bash)"' >> ~/.bash_profile
run echo 'eval "$(chef shell-init bash)"' >> ~/.bashrc

run gem install bundler

# copy the application files to the image
workdir /srv/capsulecd
run git clone https://github.com/AnalogJ/capsulecd.git .

run bundle install --without node ruby