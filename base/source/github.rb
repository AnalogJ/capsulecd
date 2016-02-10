require 'octokit'
require 'uri'
require 'git'
require_relative '../../base/common/git_utils'

module GithubSource



  @source_client
  @source_git_base_info
  @source_git_head_info
  @source_git_parent_path = '/var/capsulecd/' # should be the parent folder of the cloned repository. /var/capsule-cd/
  @source_git_local_path
  @source_git_local_branch
  @source_git_remote

  @source_release_commit = nil
  @source_release_artifacts = []

  #define the Source API methods

  # configure method will generate an authenticated client that can be used to comunicate with Github
    def source_configure
      puts 'github source_configure'
      Octokit.auto_paginate = true
      @source_client = Octokit::Client.new(:access_token => ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN'])
    end

  # all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
  # should check if the pull request opener even has permissions to create a release.
  # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
  # MUST set source_git_local_path
  def source_process_payload(payload)
    puts 'github source_process_payload'

    #TODO: ensure that this is a pullrequest

    print payload

    # check the payload action
    unless(payload['state'] == 'open')
      raise 'pull request has an invalid action'
    end


    if(payload['base']['repo']['default_branch'] != payload['base']['ref'])
      raise 'pull request is not being created against the default branch of this repository (usually master)'
    end

    # check the payload push user.
    if !@source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])
      raise 'pull request was opened by a unauthorized user'
    end

    #set the remote url, with embedded token
    uri = URI.parse(payload['base']['repo']['clone_url'])
    uri.user = ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN']
    @source_git_remote = uri.to_s

    #set the base/head info,
    @source_git_base_info = payload['base']
    @source_git_head_info = payload['head']

    # clone the merged branch
    # https://sethvargo.com/checkout-a-github-pull-request/
    # https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
    @source_git_local_path = GitUtils.clone(@source_git_parent_path,@source_git_head_info['repo']['name'],@source_git_remote)
    @source_git_local_branch = "pr_#{payload['number']}"
    GitUtils.fetch(@source_git_local_path, "refs/pull/#{payload['number']}/head", @source_git_local_branch)
    GitUtils.checkout(@source_git_local_path, @source_git_local_branch)

    #show a processing message on the github PR.
    @source_client.create_status(payload['base']['repo']['full_name'], @source_git_head_info['sha'], 'pending',
     {
      :target_url => 'http://www.github.com/AnalogJ/capsulecd',
      :description => 'CapsuleCD has started processing cookbook. Pull request will be merged automatically when complete.'
     })
  end

  def source_release
    puts 'github source_release'

    #push the version bumped metadata file + newly created files to
    GitUtils.push(@source_git_local_path, @source_git_local_branch, @source_git_base_info['ref'])
    #sleept because github needs time to process the new tag.
    sleep 5

    #calculate the release sha
    release_sha = ('0'*(40 - @source_release_commit.sha.strip.length)) + @source_release_commit.sha.strip

    #get the release changelog
    release_body = generate_changelog(@source_git_local_path, @source_git_base_info['sha'], @source_git_head_info['sha'], @source_git_base_info['repo']['full_name'])

    release = @source_client.create_release(@source_git_base_info[:full_name], @source_release_commit.name, {
      :target_commitish => release_sha,
      :name => @source_release_commit.name,
      :body => release_body
    })

    @source_release_artifacts.each { |release_artifact|
      @source_client.upload_asset(release[:url], release_artifact[:path], {:name => release_artifact[:name]})
    }

    #set the pull request status
    @source_client.create_status(@source_git_base_info['repo']['full_name'], @source_git_head_info['sha'], 'success',{
      :target_url => 'http://www.github.com/AnalogJ/capsulecd',
      :description => 'pull-request was successfully merged, new release created.'
    })
  end

  def source_process_failure(ex)
    puts 'github source_process_failure'

    @source_client.create_status(@source_git_base_info['repo']['full_name'], @source_git_head_info['sha'], 'failure',{
       :target_url => 'http://www.github.com/AnalogJ/capsulecd',
       :description => ex.message.slice!(0..135)
    })
  end


  ######################################################################################################################
  # Utilities
  ######################################################################################################################

  def generate_changelog(repo_path, base_sha, head_sha, full_name)
    repo = Git.open(repo_path)
    markdown = "Timestamp |  SHA | Message | Author \n"
    markdown += "------------- | ------------- | ------------- | ------------- \n"
    repo.log.between(base_sha, head_sha).each { |commit|
      markdown += "#{commit.date.strftime('%Y-%m-%d %H:%M:%S%z')} | [`#{commit.sha.slice 0..8}`](https://github.com/#{full_name}/commit/#{commit.sha} | #{commit.message.gsub!('|','!')} | #{commit.author.name}) \n"
    }
    return markdown
  end
end