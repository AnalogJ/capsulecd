[![Circle CI](https://circleci.com/gh/AnalogJ/capsulecd.svg?style=shield)](https://circleci.com/gh/AnalogJ/capsulecd)
[![Coverage Status](https://coveralls.io/repos/github/AnalogJ/capsulecd/badge.svg)](https://coveralls.io/github/AnalogJ/capsulecd)

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
