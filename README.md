# capsulecd
[![Circle CI](https://circleci.com/gh/AnalogJ/capsulecd.svg?style=svg)](https://circleci.com/gh/AnalogJ/capsulecd)
Continuous Delivery scripts for automating package releases (npm, chef, ruby, python, crate)

#capsule.yml
capsume.yml file should have the following sections:

    # docker image which will be used to build and package this release. 
    # Image should have all required tools installed to validate, build, package and release 
    image: chef 
    
    # source should be where the platform where the package source is stored. ie. github/bitbucket/gitlab (only github supported to start)
    source: 
      github:
        # a github access_token is required to identify the user who has "push" rights to the repo.
        access_token: 
        # additional github specific configuration which can be used to configure enterprise github connections
    
    # type specifies which capsule script is run against the code
    # can only be: node, chef, ruby, python, crate, puppet, docker?,  general, (more to come)
    type: 
    
    build:
      config:
      
    
    # validate contains the configuration for the validation step. 
    validate:
      config:
        #flags is an array of options used to enable and disable steps in the default package validate script.
        flags: []
        #options which are used to 
      
      # the pre hook takes a multiline script which allows you to prepare the environment before the package validation script is run. This could include downloading additional test libraries, . This custom pre step will be run before the default package script is run. 
      pre:
      
      # the post hook takes a multiline ruby script which allows you to do any cleanup or post processing after the package validation script is run. This could include steps like running a custom test/validation suite, deleting test folders, commiting added files
      post:
      
    # the build hook generates the acutal package??
    package: 
      config:
        flags: []
      pre:
      post:
      
    # deploy the package to the package source (npm publish, gem push, etc)
    release:
      config:
        flags: []
      pre:
      post:


# Hosted CI Providers


Hosted Service | Pricing | Docker | Pull Request | Secrets |  Comments
------------ | ------------- | ------------- | ------------- | -------------  | -------------
Appveyor | Free Tier | No | [Yes](https://www.appveyor.com/docs/environment-variables) | Yes | Windows only
Codeship | Free Tier | [No](http://pages.codeship.com/docker?utm_source=CodeshipNavBar) | [Yes](https://codeship.com/documentation/continuous-integration/set-environment-variables/) | Yes | Docker support not publically available. 
Circleci | Free Tier | [Yes](https://circleci.com/integrations/docker/) | [Yes](https://circleci.com/docs/environment-variables) | Yes | n/a
Drone.io | Free Tier | No | [Yes](http://docs.drone.io/env.html) | Yes | Opensource Drone lets you specify a .drone.yml file and Docker image, but hosted Drone does not. 
Shippable | Free Tier | Yes | [Yes](http://docs.shippable.com/yml_reference/) | Yes | Terrible UI. 
Travis.ci | Free Tier | [Yes](https://docs.travis-ci.com/user/docker/) | [Yes](https://docs.travis-ci.com/user/pull-requests) | Yes | Pull requests have a [security restriction](https://docs.travis-ci.com/user/pull-requests#Security-Restrictions-when-testing-Pull-Requests). Secrets arn't available
Wercker | Free Tier | [Yes](http://devcenter.wercker.com/docs/containers/private-containers.html) | [No](https://github.com/wercker/support/issues/19) | Yes | Pull requests do not specify the PR , or even that they are a [Pull Request](https://github.com/wercker/support/issues/19)

# Environmental Variables

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
