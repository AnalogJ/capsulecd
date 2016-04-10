# CapsuleCD
---

[![Circle CI](https://circleci.com/gh/AnalogJ/capsulecd.svg?style=shield)](https://circleci.com/gh/AnalogJ/capsulecd)
[![Coverage Status](https://coveralls.io/repos/github/AnalogJ/capsulecd/badge.svg)](https://coveralls.io/github/AnalogJ/capsulecd)
[![GitHub license](https://img.shields.io/github/license/AnalogJ/capsulecd.svg)](https://github.com/AnalogJ/capsulecd/blob/master/LICENSE)
[![Gratipay User](https://img.shields.io/gratipay/user/analogj.svg)](https://gratipay.com/~AnalogJ/)

<!-- 
[![Gem](https://img.shields.io/gem/dt/capsulecd.svg)]()
[![Gem](https://img.shields.io/gem/v/capsulecd.svg)]()
[![Gemnasium](https://img.shields.io/gemnasium/analogj/capsulecd.svg)]()
[![Docker Pulls](https://img.shields.io/docker/pulls/analogj/capsulecd.svg)]()
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
	* Vanilla Javascript Bower/Npm Packages
* Highly configurable
* Follows language/library best practices. Including things like:
	* automatically bumping the semvar version number
	* regenerating any `*.lock` files/ shrinkwrap files with new version
	* creating any recommended files (eg. `.gitignore`) 
	* validates all dependencies exist (by vendoring locally)
	* running unit tests
	* source minification
	* linting library syntax
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
---

## What is CapsuleCD

CapsuleCD is a generic Continuous Delivery pipeline for versioned artifacts and libraries written in any language. 
It's goal is to bring automation to the packaging and deployment stage of your library release cycle.
It automates away all the common steps required when creating a new version of your library.

## Why use CapsuleCD
At first glance, it seems simple to publish a new library version. Just bump the version number and publish, right?
Well, not always:

- If you're library includes a Gemfile.lock, Berksfile.lock or other most common lock files, you'll need to regenerate them. 
- Everyone runs their library unit tests before creating a new release, but hat about validating that your library dependencies exist (maybe in your Company's private repo)?
- How about linting your source, to ensure that it follows common/team conventions? 
- Who owns the gem? Is there one developer who has the credentials to push to RubyGems.org? Are they still on your team? 
- Did you remember to tag your source when the new version was created (making it easy to determine what's changed between versions?)
- Did you update your changelog?

CapsuleCD handles all of that (and more!) for you. It pretty much guarantees that your library will have consitent and correct releases every time. 
CapsuleCD is simple and fully tested, unlike the release scripts you've manually cobbled together for each library. 

## How do I start?
You can use CapsuleCD to automate creating a new release from a pull request __or__ from the latest code on your default branch.

### Automated pull request processing:

Here's how to use __docker__ to merge a pull request to your Ruby library

    CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
    CAPSULE_RUNNER_REPO_FULL_NAME=AnalogJ/gem_analogj_test \
    CAPSULE_RUNNER_PULL_REQUEST=4 \
    CAPSULE_RUBYGEMS_API_KEY=ASDF12345F \
    docker run AnalogJ/capsulecd:ruby capsulecd start --source github --package_type ruby

Or you could install and call CapsuleCD directly to merge a pull request to your Python library:

	gem install capsulecd
	CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
	CAPSULE_RUNNER_REPO_FULL_NAME=AnalogJ/pip_analogj_test \
	CAPSULE_RUNNER_PULL_REQUEST=2 \
	CAPSULE_PYPI_USERNAME=AnalogJ \
	CAPSULE_PYPI_PASSWORD=mysupersecurepassword \
	capsulecd start --source github --package_type python
	
### Creating a branch release

TODO: add documentation on how to create a release from the master branch without a pull request. Specify the env variables required. 
	
# Configuration
Specifying your `GITHUB_ACCESS_TOKEN` and `PYPI_PASSWORD` via an environmental variable might make sense, but do you 
really want to specify the `PYPI_USERNAME`, `REPO_FULL_NAME` each time? Probably not. 

CapsuleCD has you covered. We support a global YAML configuration file (that can be specified using the `--config-file` flag), and a repo specific YAML configuration file stored as `capsule.yml` inside the repo itself.

## Setting Inheritance/Overrides
CapsuleCD settings are determined by loading configuration in the following order (where the last value specified is used)

- system YAML config file (`--config-file`)
- repo YAML config file (`capsule.yml`)
- environmental variables (setting in capital letters and prefixed with `CAPSULE_`)

## Configuration Settings

Setting | System Config | Repo Config | Notes
------------ | ------------- | ------------- | -------------
package_type | No | No | Must be set by `--package-type` flag
source | No | No | Must be set by `--source` flag
runner | No | No | Must be set by `--runner` flag
dry_run | Yes | No | Can be `YES` or `NO`
source_git_parent_path | Yes | No | Specifies the location where the git repo will be cloned, defaults to tmp directory
source_github_api_endpoint | Yes | No | Specifies the Github api endpoint to use (for use with Enterprise Github)
source_github_web_endpoint | Yes | No | Specifies the Github web endpoint to use (for use with Enterprise Github)
source_github_access_token | Yes | No | Specifies the access token to use when cloning from and committing to Github
runner_pull_request | Yes | No | Specifies the repo pull request number to clone from  Github
runner_repo_full_name | Yes | No | Specifies the repo name to clone from Github
chef_supermarket_username | Yes | Yes | Specifies the Chef Supermarket username to use when creating public release for Chef cookbook
chef_supermarket_key | Yes | Yes | Specifies the Base64 encoded Chef Supermarket private key to use when creating public release for Chef cookbook
chef_supermarket_type | Yes | Yes | Specifies the Chef Supermarket cookbook type to use when creating public release for Chef cookbook
npm_auth_token | Yes | Yes | Specifies the NPM auth to use when creating public release for NPM package
pypi_username | Yes | Yes | Specifies the PYPI username to use when creating public release for Pypi package
pypi_password | Yes | Yes | Specifies the PYPI password to use when creating public release for Pypi package
engine_disable_test | Yes | Yes | Disables test_step before releasing package
engine_disable_minification | Yes | Yes | Disables source minification (if applicable) before releasing package
engine_disable_lint | Yes | Yes | Disables source linting before releasing package
engine_cmd_test | Yes | Yes | Specifies the test command to before releasing package
engine_cmd_minification | Yes | Yes | Specifies the minification command to before releasing package
engine_cmd_lint | Yes | Yes | Specifies the lint command to before releasing package
engine_version_bump_type | Yes | Yes | Specifies the Semvar segment (`major`, `minor`, `patch`) to bump before releasing package

TODO: specify the missing `BRANCH` release style settings.

As mentioned above, all settings can be specified via Environmental variable. All you need to do is convert the setting to uppercase
and then prefix it with `CAPSULE_`. So `pypi_password` can be set with `CAPSULE_PYPI_PASSWORD` and `engine_cmd_test` with `CAPSULE_ENGINE_CMD_TEST`

### Example System Configuration File

Here's what an example system configuration file might look like:

```
source_git_parent_path: /srv/myclonefolder
source_github_api_endpoint: https://git.mycorpsubnet.example.com/v2
source_github_web_endpoint: https://git.mycorpsubnet.example.com/v2
```

# Testing
---

##Test suite and continuous integration

CapsuleCD provides an extensive test-suite based on rspec and a full integration suite which uses VCR. 
You can run the unit tests with `rake  test`. The integration tests can be run by `rake 'spec:<package_type>'`. 
So to run the Python integration tests you would call `rake 'spec:python'`.

Travis-CI is used for continuous integration testing: http://travis-ci.org/slim-template/slim

Slim is working well on all major Ruby implementations:

Ruby 1.9.3, 2.0.0, 2.1.0 and 2.2.0
Ruby EE
JRuby 1.9 mode
Rubinius 2.0






# capsulecd
Continuous Delivery scripts for automating package releases (npm, chef, ruby, python, crate)

# Support
The current languages, and their packages, are supported:

- Chef (Cookbook)	
- Javascript (Bower)
- NodeJS (Npm, Bower)
- Python (Pip)

Support for the following languages will be added in the future (feel free to open a PR) 

- C#
- Objective C
- Dash
- Go
- Java
- Lua
- Rust
- Ruby
- Scala
- Swift
- Any others you can think of. 

#capsule.yml
capsume.yml file should have the following sections:

	# TODO:

# Hosted CI Providers

# Environmental Variables


Hosted Service | Pricing | Docker | Pull Request | Secrets |  Comments
------------ | ------------- | ------------- | ------------- | -------------  | -------------
Appveyor | Free Tier | No | [Yes](https://www.appveyor.com/docs/environment-variables) | Yes | Windows only
Codeship | Free Tier | [No](http://pages.codeship.com/docker?utm_source=CodeshipNavBar) | [Yes](https://codeship.com/documentation/continuous-integration/set-environment-variables/) | Yes | Docker support not publically available. 
Circleci | Free Tier | [Yes](https://circleci.com/integrations/docker/) | [Yes](https://circleci.com/docs/environment-variables) | Yes | n/a
Drone.io | Free Tier | No | [Yes](http://docs.drone.io/env.html) | Yes | Opensource Drone lets you specify a .drone.yml file and Docker image, but hosted Drone does not. 
Shippable | Free Tier | Yes | [Yes](http://docs.shippable.com/yml_reference/) | Yes | Terrible UI. 
Travis.ci | Free Tier | [Yes](https://docs.travis-ci.com/user/docker/) | [Yes](https://docs.travis-ci.com/user/pull-requests) | Yes | Pull requests have a [security restriction](https://docs.travis-ci.com/user/pull-requests#Security-Restrictions-when-testing-Pull-Requests). Secrets arn't available
Wercker | Free Tier | [Yes](http://devcenter.wercker.com/docs/containers/private-containers.html) | [No](https://github.com/wercker/support/issues/19) | Yes | Pull requests do not specify the PR , or even that they are a [Pull Request](https://github.com/wercker/support/issues/19)


## Runner

Runner Name | Required Variables
------------ | -------------
CircleCI | `CI_PULL_REQUEST`, `CIRCLE_PROJECT_USERNAME`, `CIRCLE_PROJECT_REPONAME`

## Source

Source Name | Required Variables
------------ | -------------
Github | `CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN`

## Type

Package Type | Required Variables
------------ | -------------
Chef | `CAPSULE_CHEF_SUPERMARKET_USERNAME`, `CAPSULE_CHEF_SUPERMARKET_KEY` (base64 encoded), `CAPSULE_CHEF_SUPERMARKET_TYPE`
Node | `CAPSULE_NPM_AUTH_TOKEN`
Javascript | `CAPSULE_NPM_AUTH_TOKEN`
Python | `CAPSULE_PYPI_USERNAME`, `CAPSULE_PYPI_PASSWORD`
