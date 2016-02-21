require 'octokit'
require 'uri'
require 'git'
require 'capsulecd'
require 'pp'

module CapsuleCD
  module Source
    module Github
      # all of these instance variables are available for use within hooks
      attr_accessor :source_client
      attr_accessor :source_git_base_info
      attr_accessor :source_git_head_info
      attr_accessor :source_git_parent_path
      attr_accessor :source_git_local_path
      attr_accessor :source_git_local_branch
      attr_accessor :source_git_remote
      attr_accessor :source_release_commit
      attr_accessor :source_release_artifacts

      # define the Source API methods

      # configure method will generate an authenticated client that can be used to comunicate with Github
      # MUST set @source_git_parent_path
      # MUST set @source_client
      def source_configure
        puts 'github source_configure'
        fail CapsuleCD::Error::SourceAuthenticationFailed, 'Missing "CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN" env variable' unless ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN']

        @source_release_commit = nil
        @source_release_artifacts = []

        @source_git_parent_path = Dir.mktmpdir
        Octokit.auto_paginate = true
        @source_client = Octokit::Client.new(access_token: ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN'])
      end

      # all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the webhook data.
      # should check if the pull request opener even has permissions to create a release.
      # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
      # MUST set source_git_local_path
      # MUST set source_git_local_branch
      # MUST set source_git_head_info
      # REQUIRES source_git_parent_path
      def source_process_push_payload(payload)
        # set the processed head info
        @source_git_head_info = payload['head']
        CapsuleCD::ValidationUtils.validate_repo_payload(@source_git_head_info)

        # set the remote url, with embedded token
        uri = URI.parse(@source_git_head_info['repo']['clone_url'])
        uri.user = ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN']
        @source_git_remote = uri.to_s
        @source_git_local_branch = @source_git_head_info['repo']['ref']
        # clone the merged branch
        # https://sethvargo.com/checkout-a-github-pull-request/
        # https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
        @source_git_local_path = CapsuleCD::GitUtils.clone(@source_git_parent_path, @source_git_head_info['repo']['name'], @source_git_remote)
        CapsuleCD::GitUtils.checkout(@source_git_local_path, @source_git_head_info['repo']['sha1'])
      end

      # all capsule CD processing will be kicked off via a payload. In Github's case, the payload is the pull request data.
      # should check if the pull request opener even has permissions to create a release.
      # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
      # MUST set source_git_local_path
      # MUST set source_git_local_branch
      # MUST set source_git_base_info
      # MUST set source_git_head_info
      # REQUIRES source_client
      # REQUIRES source_git_parent_path
      def source_process_pull_request_payload(payload)
        puts 'github source_process_payload'

        # validate the github specific payload options
        unless (payload['state'] == 'open')
          fail CapsuleCD::Error::SourcePayloadUnsupported, 'Pull request has an invalid action'
        end
        unless (payload['base']['repo']['default_branch'] == payload['base']['ref'])
          fail CapsuleCD::Error::SourcePayloadUnsupported, 'Pull request is not being created against the default branch of this repository (usually master)'
        end
        # check the payload push user.

        unless @source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])

          @source_client.add_comment(payload['base']['repo']['full_name'], payload['number'], CapsuleCD::BotUtils.pull_request_comment)
          fail CapsuleCD::Error::SourceUnauthorizedUser, 'Pull request was opened by an unauthorized user'
        end

        # set the processed base/head info,
        @source_git_base_info = payload['base']
        @source_git_head_info = payload['head']
        CapsuleCD::ValidationUtils.validate_repo_payload(@source_git_base_info)
        CapsuleCD::ValidationUtils.validate_repo_payload(@source_git_head_info)

        # set the remote url, with embedded token
        uri = URI.parse(payload['base']['repo']['clone_url'])
        uri.user = ENV['CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN']
        @source_git_remote = uri.to_s

        # clone the merged branch
        # https://sethvargo.com/checkout-a-github-pull-request/
        # https://coderwall.com/p/z5rkga/github-checkout-a-pull-request-as-a-branch
        @source_git_local_path = CapsuleCD::GitUtils.clone(@source_git_parent_path, @source_git_head_info['repo']['name'], @source_git_remote)
        @source_git_local_branch = "pr_#{payload['number']}"
        CapsuleCD::GitUtils.fetch(@source_git_local_path, "refs/pull/#{payload['number']}/head", @source_git_local_branch)
        CapsuleCD::GitUtils.checkout(@source_git_local_path, @source_git_local_branch)

        # show a processing message on the github PR.
        @source_client.create_status(payload['base']['repo']['full_name'], @source_git_head_info['sha'], 'pending',
                                     target_url: 'http://www.github.com/AnalogJ/capsulecd',
                                     description: 'CapsuleCD has started processing cookbook. Pull request will be merged automatically when complete.')
      end

      # REQUIRES source_client
      # REQUIRES source_release_commit
      # REQUIRES source_git_local_path
      # REQUIRES source_git_local_branch
      # REQUIRES source_git_base_info
      # REQUIRES source_git_head_info
      # REQUIRES source_release_artifacts
      # REQUIRES source_git_parent_path
      def source_release
        puts 'github source_release'

        # push the version bumped metadata file + newly created files to
        CapsuleCD::GitUtils.push(@source_git_local_path, @source_git_local_branch, @source_git_base_info['ref'])
        # sleept because github needs time to process the new tag.
        sleep 5

        # calculate the release sha
        release_sha = ('0' * (40 - @source_release_commit.sha.strip.length)) + @source_release_commit.sha.strip

        # get the release changelog
        release_body = CapsuleCD::GitUtils.generate_changelog(@source_git_local_path, @source_git_base_info['sha'], @source_git_head_info['sha'], @source_git_base_info['repo']['full_name'])

        release = @source_client.create_release(@source_git_base_info['repo']['full_name'], @source_release_commit.name,       target_commitish: release_sha,
                                                                                                                               name: @source_release_commit.name,
                                                                                                                               body: release_body)

        @source_release_artifacts.each do |release_artifact|
          @source_client.upload_asset(release[:url], release_artifact[:path], name: release_artifact[:name])
        end

        FileUtils.remove_entry_secure @source_git_parent_path
        # set the pull request status
        @source_client.create_status(@source_git_base_info['repo']['full_name'], @source_git_head_info['sha'], 'success',      target_url: 'http://www.github.com/AnalogJ/capsulecd',
                                                                                                                               description: 'pull-request was successfully merged, new release created.')
      end

      def source_process_failure(ex)
        puts 'github source_process_failure'
        FileUtils.remove_entry_secure @source_git_parent_path
        @source_client.create_status(@source_git_base_info['repo']['full_name'], @source_git_head_info['sha'], 'failure',       target_url: 'http://www.github.com/AnalogJ/capsulecd',
                                                                                                                                description: ex.message.slice!(0..135))
      end
    end
  end
end
