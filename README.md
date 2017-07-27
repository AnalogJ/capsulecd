# CapsuleCD

<p align="center">
  <img width="300" alt="portfolio_view" src="https://cdn.rawgit.com/AnalogJ/capsulecd/master/logo.svg">
</p>


[![Circle CI](https://img.shields.io/circleci/project/github/AnalogJ/capsulecd.svg?style=flat-square)](https://circleci.com/gh/AnalogJ/capsulecd)
[![Coverage Status](https://img.shields.io/codecov/c/github/AnalogJ/capsulecd.svg?style=flat-square)](https://codecov.io/gh/AnalogJ/capsulecd)
[![GitHub license](https://img.shields.io/github/license/AnalogJ/capsulecd.svg?style=flat-square)](https://github.com/AnalogJ/capsulecd/blob/master/LICENSE)
[![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/analogj/capsulecd)
[![Go Report Card](https://goreportcard.com/badge/github.com/AnalogJ/capsulecd?style=flat-square)](https://goreportcard.com/report/github.com/AnalogJ/capsulecd)
[![GitHub release](http://img.shields.io/github/release/AnalogJ/capsulecd.svg?style=flat-square)](https://github.com/AnalogJ/capsulecd/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/analogj/capsulecd.svg?style=flat-square)](https://hub.docker.com/r/analogj/capsulecd)
[![Github All Releases](https://img.shields.io/github/downloads/analogj/capsulecd/total.svg?style=flat-square)](https://github.com/AnalogJ/capsulecd/releases)

<!-- 
[![Gemnasium](https://img.shields.io/gemnasium/analogj/capsulecd.svg)]()
-->

CapsuleCD is a generic Continuous Delivery pipeline for versioned artifacts and libraries written in any language. 
It's goal is to bring automation to the packaging and deployment stage of your library release cycle.
CapsuleCD is incredibly flexible, and works best when implemented side-by-side with a CI pipeline.

A short list of the features...

* Supports libraries written in any language. Has built-in support for 
	* Chef Cookbooks
	* Python Pip
	* NodeJS Npm Packages
	* Ruby Gems
	* Golang Packages
* Highly configurable
* Follows language/library best practices. Including things like:
	* automatically bumping the semvar version number
	* regenerating any `*.lock` files/ shrinkwrap files with new version
	* creating any recommended files (eg. `.gitignore`) 
	* validates all dependencies exist (by vendoring locally)
	* vulnerbility scanning in dependencies
	* running unit tests
	* linting library syntax
	* source formatting
	* generating code coverage reports
	* updating changelog
	* uploading versioned artifact to community hosting service (rubygems/supermarket/pypi/etc)
	* creating a new git tag 
	* pushing changes back to source control (github)
	* creating a new release in source control (github) and attaching any common artifacts
	
## Links

* Source: <http://github.com/AnalogJ/capsulecd>
* Bugs:   <http://github.com/AnalogJ/capsulecd/issues>

# Introduction

## What is CapsuleCD

CapsuleCD is a generic Continuous Delivery pipeline for versioned artifacts and libraries written in any language. 
It's goal is to bring automation to the packaging and deployment stage of your library release cycle.
It automates away all the common steps required when creating a new version of your library.

## Why use CapsuleCD
At first glance, it seems simple to publish a new library version. Just bump the version number and publish, right?
Well, not always:

- If you're library includes a Gemfile.lock, Berksfile.lock or other common lock files, you'll need to regenerate them as the old version number is embedded inside. 
- Everyone runs their library unit tests before creating a new release (right?!), but what about validating that your [library dependencies exist](http://www.theregister.co.uk/2016/03/23/npm_left_pad_chaos/) (maybe in your Company's private repo)?
- How about linting your source, to ensure that it follows common/team conventions? 
- Who owns the gem? Is there one developer who has the credentials to push to RubyGems.org? Are they still on your team/on vacation? 
- Did you remember to tag your source when the new version was created (making it easy to determine what's changed between versions?)
- Did you update your changelog?

CapsuleCD handles all of that (and more!) for you. It pretty much guarantees that your library will have proper and consistent releases every time. 
CapsuleCD is well structured and fully tested, unlike the release scripts you've manually cobbled together for each library and language.
It can be customized as needed without rewriting from scratch.
The best part is that CapsuleCD uses CapsuleCD to automate its releases.
We [dogfood](https://en.wikipedia.org/wiki/Eating_your_own_dog_food) it so we're the first ones to find any issues with a new release.

## How do I start?
You can use CapsuleCD to automate creating a new release from a pull request __or__ from the latest code on your default branch.

### Automated pull request processing:

Here's how to use __docker__ to merge a pull request to your Ruby library

    CAPSULE_SCM_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
    CAPSULE_SCM_REPO_FULL_NAME=AnalogJ/gem_analogj_test \
    CAPSULE_SCM_PULL_REQUEST=4 \
    CAPSULE_RUBYGEMS_API_KEY=ASDF12345F \
    docker run AnalogJ/capsulecd:ruby capsulecd start --scm github --package_type ruby

Or you could download the latest linux [release](https://github.com/AnalogJ/capsulecd/releases), and call CapsuleCD directly to merge a pull request to your Python library:

	CAPSULE_SCM_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
	CAPSULE_RUNNER_REPO_FULL_NAME=AnalogJ/pip_analogj_test \
	CAPSULE_RUNNER_PULL_REQUEST=2 \
	CAPSULE_PYPI_USERNAME=AnalogJ \
	CAPSULE_PYPI_PASSWORD=mysupersecurepassword \
	capsulecd start --scm github --package_type python
	
### Creating a branch release

	TODO: add documentation on how to create a release from the master branch without a pull request. Specify the env variables required. 
	
# Engine
Every package type is mapped to an engine class which inherits from a `EngineScm` class, ie `EnginePython`, `EngineNode`, `EngineRuby` etc.
Every scm type is mapped to a scm class, ie `ScmGithub`. When CapsuleCD starts, it initializes the specified Engine, and loads the correct Scm module.
Then it begins processing your source code step by step.

Step | Description
------------ | ------------ 
scm_init_step | This will initialize the scm client, ensuring that we can authenticate with the git server
scm_retrieve_payload_step | If a Pull Request # is specified, the payload is retrieved from Scm api, otherwise the repo default branch HEAD info is retrived.
scm_process_pull_request_payload __or__ scm_process_push_payload | Depending on the retrieve_payload step, the merged pull request is cloned, or the default branch is cloned locally
assemble_step | Code is built, which includes adding any missing files/default structure, version bumping, etc.
dependencies_step | Download package dependencies
compile_step | Optional compilation of source into binaries
test_step | Run the package test runner(s) (eg. npm test, rake test, kitchen test, tox), linter, formatter & dependency vulnerbility scanner
package_step | Clean any unnecessary files, commit any local changes and create a git tag. Nothing should be pushed to remote repository
dist_step | Push the release to the package repository (ie. npm, chef supermarket, rubygems)
scm_publish | Push the merged, tested and version updated code up to the source code repository. Also do any source specific releases (github release, asset uploading, etc)

# Configuration
Specifying your `GITHUB_ACCESS_TOKEN` and `PYPI_PASSWORD` via an environmental variable might make sense, but do you 
really want to specify the `PYPI_USERNAME`, `REPO_FULL_NAME` each time? Probably not. 

CapsuleCD has you covered. We support a global YAML configuration file (that can be specified using the `--config_file` flag), and a repo specific YAML configuration file stored as `capsule.yml` inside the repo itself.

## Setting Inheritance/Overrides
CapsuleCD settings are determined by loading configuration in the following order (where the last value specified is used)

- system YAML config file (`--config-file`)
- repo YAML config file (`capsule.yml`)
- environmental variables (setting in capital letters and prefixed with `CAPSULE_`)

## Configuration Settings

Check the [`example.capsule.yml`](example.capsule.yml) file for a full list of all the available coniguration options.

As mentioned above, all settings can be specified via Environmental variable. All you need to do is convert the setting to uppercase
and then prefix it with `CAPSULE_`. So `pypi_password` can be set with `CAPSULE_PYPI_PASSWORD` and `engine_cmd_test` with `CAPSULE_ENGINE_CMD_TEST`

### Example System Configuration File

Here's what an example system configuration file might look like:

```
scm_git_parent_path: /srv/myclonefolder
scm_github_api_endpoint: https://git.mycorpsubnet.example.com/v2
scm_github_web_endpoint: https://git.mycorpsubnet.example.com/v2
```

## Step pre/post hooks and overrides

CapsuleCD is completely customizable, to the extent that you can run your own Ruby code as `pre` and `post` hooks before every step. 
To add a `pre`/`post` hook or override a step, just modify your config `yml` file by adding the step you want to modify, and
specify `pre` or `post` as a subkey. Then specify your shell commands as a list

	---
      scm_init:
        pre:
          - echo "override pre_scm_configure"
          - `git clone ...`
        post: |
          # do additional cleanup or anything else you want.
          - echo "override post_scm_configure"
      assemble_step:
        post: |
          # this post hook runs after the assemble_step runs
          - echo "override post_build_step"

# Testing

## Test suite and continuous integration

CapsuleCD provides an extensive test-suite based on `go test` and a full integration suite which uses `go-vcr`.
You can run all the integration & unit tests with `go test $(glide novendor)`

CircleCI is used for continuous integration testing: <https://circleci.com/gh/AnalogJ/capsulecd>

# Contributing

If you'd like to help improve CapsuleCD, clone the project with Git by running:

    $ git clone git://github.com/AnalogJ/capsulecd
    
Work your magic and then submit a pull request. We love pull requests!

If you find the documentation lacking, help us out and update this README.md. If you don't have the time to work on CapsuleCD, but found something we should know about, please submit an issue.

## To-do List

We're actively looking for pull requests in the following areas:

- CapsuleCD Engines for other languages
	- C#
	- Objective C
	- Dash
	- Java
	- Lua
	- Rust
	- Scala
	- Swift
	- [Any others you can think of](https://libraries.io/)
- CapsuleCD Sources
	- GitLab
	- Bitbucket
	- Beanstalk
	- Kiln
	- Any others you can think of


# Versioning

We use SemVer for versioning. For the versions available, see the tags on this repository.

# Authors

Jason Kulatunga - Initial Development -  [@AnalogJ](https://github.com/AnalogJ)

# License

CapsuleCD is licensed under the MIT License - see the [LICENSE.md](https://github.com/AnalogJ/capsulecd/blob/master/LICENSE.md) file for details






# References
- https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1
- http://matthewbrown.io/2016/01/23/factory-pattern-in-golang/
- https://medium.com/@matryer/5-simple-tips-and-tricks-for-writing-unit-tests-in-golang-619653f90742
- http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/
- https://stackoverflow.com/questions/6395076/using-reflect-how-do-you-set-the-value-of-a-struct-field
- https://medium.com/@skdomino/writing-better-clis-one-snake-at-a-time-d22e50e60056
- https://stackoverflow.com/questions/15148331/test-naming-conventions-in-golang
- https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
- https://blog.golang.org/error-handling-and-go
- http://blog.hashbangbash.com/2014/04/linking-golang-statically/
- https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f
- https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
- https://stackoverflow.com/questions/37026399/git2go-after-createcommit-all-files-appear-like-being-added-for-deletion
- https://stackoverflow.com/questions/25965584/separating-unit-tests-and-integration-tests-in-go
- https://peter.bourgon.org/go-best-practices-2016/
- http://golangcookbook.com/chapters/running/cross-compiling/
- https://github.com/kelseyhightower/confd/blob/20b3d37da7aaa2c176c0612202c06c5ba4f7d987/docs/release-checklist.md
- https://gist.github.com/Ehekatl/93b4ac1621771f2889cd99c7b7cfc2ec
- https://github.com/sithembiso/git2go-build
- https://gist.github.com/Ehekatl/93b4ac1621771f2889cd99c7b7cfc2ec
- https://github.com/danielfbm/docker-go-libgit2/blob/master/Dockerfile
- https://github.com/thockin/go-build-template
- https://peter.bourgon.org/go-best-practices-2016/
- https://golang.org/cmd/gofmt/
- https://github.com/weaveworks/mesh/blob/master/lint
- https://github.com/alecthomas/gometalinter
- https://golang.org/cmd/vet/
- https://medium.com/statuscode/the-9-most-popular-golang-links-from-2016-c49287d99448
- https://medium.com/@sebdah/go-best-practices-testing-3448165a0e18
- https://buildroot.org/
- https://github.com/multiarch/crossbuild
- https://github.com/Cimpress-MCP/go-git2consul/tree/crossbuild/build-multi
- https://docs.codecov.io/docs/testing-with-docker
- https://github.com/codecov/example-go
- https://github.com/codecov/support/wiki/Codecov-Yaml#how-to-disable-a-single-ci-provider
- https://github.com/gopheracademy/gopheracademy-web/blob/master/content/advent-2014/git2go-tutorial.md
- https://stackoverflow.com/questions/2381665/list-tags-contained-by-a-branch
- https://dmitri.shuralyov.com/blog/18
- http://www.ryanday.net/2012/10/01/installing-go-and-gopath/
- http://craigwickesser.com/2015/02/golang-cmd-with-custom-environment/