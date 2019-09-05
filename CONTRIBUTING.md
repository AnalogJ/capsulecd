# How to contribute

Thanks! There are tons of different programming languages & SCM's, making it difficult to develop and keep everything up
to date. We want to keep it as easy as possible to
contribute to `capsulecd`, so that you can automate package management for your favorite language.
There are a few guidelines that we need contributors to follow so that 
we can keep on top of things.

## Getting Started

Fork, then clone the repo:

    $ git clone git@github.com:your-username/capsulecd.git

Ensure you have docker installed. 

	$ docker version 
	Client:
     Version:           18.06.0-ce
     API version:       1.38
     Go version:        go1.10.3
     Git commit:        0ffa825
     Built:             Wed Jul 18 19:05:26 2018
     OS/Arch:           darwin/amd64
     Experimental:      false
    
    Server:
     Engine:
      Version:          18.06.0-ce
      API version:      1.38 (minimum version 1.12)
      Go version:       go1.10.3
      Git commit:       0ffa825
      Built:            Wed Jul 18 19:13:46 2018
      OS/Arch:          linux/amd64
      Experimental:     true

Build the CapsuleCD docker development environment:

    $ docker build -f Dockerfile.build --tag capsulecd-development .

Run the docker development environment

    $ docker run --rm -it -v `pwd`:/go/src/github.com/analogj/capsulecd capsulecd-development /scripts/development.sh

Now we should be inside the development container. Lets run the test suite. 

    $ go test -v -tags "static" ./...
    
Once you've validated that the test suite has passed, you can now begin making changes to the capsulecd source.

# Adding a SCM

    $ go test -v -tags "static" pkg/scm/scm_github_test.go

# Adding an Engine ()Language/Package Manager)
