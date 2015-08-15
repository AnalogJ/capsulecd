require 'hooks'
require 'octokit'
require 'uri'
require_relative '../../base/common/git_utils'

module GithubSource



  @source_client
  @source_git_base_info
  @source_git_head_info
  @source_git_parent_path = '/var/capsule-cd/' # should be the parent folder of the cloned repository. /var/capsule-cd/
  @source_git_local_path
  @source_git_local_branch
  @source_git_remote

  #define the Source API methods

  # configure method will generate an authenticated client that can be used to comunicate with Github
  def source_configure
    Octokit.auto_paginate = true
    @source_client = Octokit::Client.new(:access_token => ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN'])
  end

  # all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
  # should check if the pull request opener even has permissions to create a release.
  # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
  # MUST set source_git_local_path
  def source_process_payload(payload)
    # check the payload action
    if(!(payload['action'] == 'opened' || payload['action'] == 'reopened'))
      raise 'pull request has an invalid action'
    end

    if(payload['repository']['default_branch'] != payload['pull_request']['base']['ref'])
      raise 'pull request is not being created against the default branch of this repository (usually master)'
    end

    if(payload['repository']['full_name'] != payload['pull_request']['base']['full_name'])
      raise 'pull request is not being created against the primary repository (probably a pull request against a fork)'
    end

    # check the payload push user.
    if !@source_client.collaborator?(payload['repository']['full_name'], payload['pull_request']['user']['login'])
      raise 'pull request was opened by a unauthorized user'
    end

    #set the remote url, with embedded token
    uri = URI.parse(payload['repository']['clone_url'])
    uri.user = ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN']
    @source_git_remote = uri.to_s

    #set the base/head info,
    @source_git_base_info = payload['pull_request']['base']
    @source_git_head_info = payload['pull_request']['head']

    # clone the merged branch
    # https://sethvargo.com/checkout-a-github-pull-request/
    # https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
    @source_git_local_path = GitUtils.clone_repository(@source_git_parent_path,@source_git_head_info['repo']['name'],@source_git_remote)
    @source_git_local_branch = "pr_#{payload['number']}"
    GitUtils.fetch(@source_git_local_path, "refs/pull/#{payload['number']}/head", @source_git_local_branch)
    GitUtils.checkout(@source_git_local_path, @source_git_local_branch)

  end

end