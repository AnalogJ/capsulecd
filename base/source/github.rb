require 'hooks'
require 'octokit'

module GithubSource



  @source_client
  @source_git_base_info
  @source_git_head_info
  @source_git_local_path # should be /var/..
  @source_git_remote

  #define the Source API methods

  # configure method will generate an authenticated client that can be used to comunicate with Github
  def source_configure
    Octokit.auto_paginate = true
    @source_client = Octokit::Client.new(:access_token => ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN'])
  end

  # all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
  # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
  # MUST set source_git_local_path
  def source_process_payload(payload)
  end


end