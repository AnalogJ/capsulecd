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
You can use CapsuleCD in a few ways:

To release your Ruby package with the CapsuleCD Docker image:

    CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
    CAPSULE_RUNNER_REPO_FULL_NAME=AnalogJ/gem_analogj_test \
    CAPSULE_RUNNER_PULL_REQUEST=4 \
    CAPSULE_RUBYGEMS_API_KEY=ASDF12345F \
    docker run AnalogJ/capsulecd:ruby capsulecd start --source github --package_type ruby

Or you could manually add CapsuleCD to your existing Python library release script:

	gem install capsulecd
	CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN=123456789ABCDEF \
	CAPSULE_RUNNER_REPO_FULL_NAME=AnalogJ/pip_analogj_test \
	CAPSULE_RUNNER_PULL_REQUEST=2 \
	CAPSULE_PYPI_USERNAME=AnalogJ \
	CAPSULE_PYPI_PASSWORD=mysupersecurepassword \
	capsulecd start --source github --package_type python
	
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

Setting | System Config | Repo Config | Environmental Variable | Notes
------------ | ------------- | ------------- | ------------- | -------------
`package_type` | No | No | -- | Must be set by `--package-type` flag
`source` | No | No | -- | Must be set by `--source` flag
`runner` | No | No | -- | Must be set by `--runner` flag
`dry_run` | Yes | No | CAPSULE_DRY_RUN | Can be `YES` or `NO`
`source_git_parent_path` | Yes | No | `CAPSULE_SOURCE_GIT_PARENT_PATH` | Specifies the location where the git repo will be cloned
`source_github_api_endpoint` | Yes | No | `CAPSULE_SOURCE_GITHUB_API_ENDPOINT` | Specifies the Github api endpoint to use (for use with Enterprise Github)
`source_github_web_endpoint` | Yes | No | `CAPSULE_SOURCE_GITHUB_WEB_ENDPOINT` | Specifies the Github web endpoint to use (for use with Enterprise Github)
`source_github_access_token` | Yes | No | `CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN` | Specifies the access token to use when cloning from and committing to Github
`runner_pull_request` | Yes | No | `CAPSULE_RUNNER_PULL_REQUEST` | Specifies the repo pull request number to clone from  Github
`runner_repo_full_name` | Yes | No | `CAPSULE_RUNNER_REPO_FULL_NAME` | Specifies the repo name to clone from Github
`chef_supermarket_username` | Yes | Yes | `CAPSULE_CHEF_SUPERMARKET_USERNAME` | Specifies the Chef Supermarket username to use when creating public release for Chef cookbook
`chef_supermarket_key` | Yes | Yes | `CAPSULE_CHEF_SUPERMARKET_KEY` | Specifies the Base64 encoded Chef Supermarket private key to use when creating public release for Chef cookbook
`chef_supermarket_type` | Yes | Yes | `CAPSULE_CHEF_SUPERMARKET_TYPE` | Specifies the Chef Supermarket cookbook type to use when creating public release for Chef cookbook
`npm_auth_token` | Yes | Yes | `CAPSULE_NPM_AUTH_TOKEN` | Specifies the NPM auth to use when creating public release for NPM package
`pypi_username` | Yes | Yes | `CAPSULE_PYPI_USERNAME` | Specifies the PYPI username to use when creating public release for Pypi package
`pypi_password` | Yes | Yes | `CAPSULE_PYPI_PASSWORD` | Specifies the PYPI password to use when creating public release for Pypi package
`engine_disable_test` | Yes | Yes | `CAPSULE_ENGINE_DISABLE_TEST` | Disables test_step before releasing package
`engine_disable_minification` | Yes | Yes | `CAPSULE_ENGINE_DISABLE_MINIFICATION` | Disables source minification (if applicable) before releasing package
`engine_disable_lint` | Yes | Yes | `CAPSULE_ENGINE_DISABLE_LINT` | Disables source linting before releasing package
`engine_cmd_test` | Yes | Yes | `CAPSULE_ENGINE_CMD_TEST` | Specifies the test command to before releasing package
`engine_cmd_minification` | Yes | Yes | `CAPSULE_ENGINE_CMD_MINIFICATION` | Specifies the minification command to before releasing package
`engine_cmd_lint` | Yes | Yes | `CAPSULE_ENGINE_CMD_LINT` | Specifies the lint command to before releasing package
`engine_version_bump_type` | Yes | Yes | `CAPSULE_ENGINE_VERSION_BUMP_TYPE` | Specifies the Semvar segment (`major`, `minor`, `patch`) to bump before releasing package

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
