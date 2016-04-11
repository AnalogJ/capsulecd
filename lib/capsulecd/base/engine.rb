require_relative 'runner/default'
require_relative 'configuration'
require 'pp'
module CapsuleCD
  class Engine
    attr_reader :config

    def initialize(options)
      @config = CapsuleCD::Configuration.new(options)
      if @config.source == :github
        require_relative 'source/github'
        self.class.send(:include, CapsuleCD::Source::Github)
      else
        fail CapsuleCD::Error::SourceUnspecifiedError, 'No source defined.'
      end

      if @config.runner == :circleci
        require_relative 'runner/circleci'
        self.class.send(:include, CapsuleCD::Runner::Circleci)
      else
        self.class.send(:include, CapsuleCD::Runner::Default)
      end
    end

    def start
      # start the source, and whatever work needs to be done there.
      # MUST set @source_git_parent_path
      # MUST set @source_client
      pre_source_configure
      source_configure
      post_source_configure

      # runner must determine if this is a pull request or a push.
      # if it's a pull request the runner must retrieve the pull request payload and return it
      # if its a push, the runner must retrieve the push payload and return it
      # the variable @runner_is_pullrequest will be true if a pull request was created.
      # MUST set runner_is_pullrequest
      # REQUIRES source_client
      pre_runner_retrieve_payload
      payload = runner_retrieve_payload(@options)
      post_runner_retrieve_payload

      if @runner_is_pullrequest
        # all capsule CD processing will be kicked off via a payload. In this case the payload is the pull request data.
        # should check if the pull request opener even has permissions to create a release.
        # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
        # MUST set source_git_local_path
        # MUST set source_git_local_branch
        # MUST set source_git_base_info
        # MUST set source_git_head_info
        # REQUIRES source_client
        pre_source_process_pull_request_payload
        source_process_pull_request_payload(payload)
        post_source_process_pull_request_payload
      else
        # start processing the payload, which should result in a local git repository that we
        # can begin to test. Since this is a push, no packaging is required
        # MUST set source_git_local_path
        # MUST set source_git_local_branch
        # MUST set source_git_head_info
        # REQUIRES source_client
        pre_source_process_push_payload
        source_process_push_payload(payload)
        post_source_process_push_payload
      end

      # now that the payload has been processed we can begin by building the code.
      # this may be creating missing files/default structure, compilation, version bumping, etc.

      source_notify('build') do
        pre_build_step
        build_step
        post_build_step
      end

      # this step should download dependencies, run the package test runner(s) (eg. npm test, rake test, kitchen test)
      # REQUIRES @config.engine_cmd_test
      # REQUIRES @config.engine_disable_test
      source_notify('test') do
        pre_test_step
        test_step
        post_test_step
      end

      # this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
      source_notify('package') do
        pre_package_step
        package_step
        post_package_step
      end

      if @runner_is_pullrequest
        # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
        source_notify('release') do
          pre_release_step
          release_step
          post_release_step
        end

        # this step should push the merged, tested and version updated code up to the source code repository
        # this step should also do any source specific releases (github release, asset uploading, etc)
        source_notify('source release') do
          pre_source_release
          source_release
          post_source_release
        end
      end

      # rescue => ex #TODO if you enable this rescue block, hooks stop working.
      # TODO: it shouldnt be required anylonger because source_notify will handle rescueing the failures.
      #   puts ex
      #
      #   self.run_hook :pre_source_process_failure, ex
      #   source_process_failure(ex)
      #   self.run_hook :post_source_process_failure, ex
    end

    # base methods
    def pre_source_configure; puts 'pre_source_configure'; end
    def post_source_configure; puts 'post_source_configure'; end
    def pre_source_process_pull_request_payload; puts 'pre_source_process_pull_request_payload'; end
    def post_source_process_pull_request_payload; puts 'post_source_process_pull_request_payload'; end
    def pre_source_process_push_payload; puts 'pre_source_process_push_payload'; end
    def post_source_process_push_payload; puts 'post_source_process_push_payload'; end
    def pre_source_release; puts 'pre_source_release'; end
    def post_source_release; puts 'post_source_release'; end

    def pre_runner_retrieve_payload; puts 'pre_runner_retrieve_payload'; end
    def post_runner_retrieve_payload; puts 'post_runner_retrieve_payload'; end

    def pre_build_step; puts 'pre_build_step'; end
    def build_step; puts 'build_step'; end
    def post_build_step; puts 'post_build_step'; end
    def pre_test_step; puts 'pre_test_step'; end
    def test_step; puts 'test_step'; end
    def post_test_step; puts 'post_test_step'; end
    def pre_package_step; puts 'pre_package_step'; end
    def package_step; puts 'package_step'; end
    def post_package_step; puts 'post_package_step'; end
    def pre_release_step; puts 'pre_release_step'; end
    def release_step; puts 'release_step'; end
    def post_release_step; puts 'post_release_step'; end

    protected

    # determine which segment of the semvar version to bump/increment. The default (as specified in CapsuleCD::Configuration is :patch)
    def bump_version(current_version)
      next_version = current_version

      if @config.engine_version_bump_type == :major
        next_version.major = next_version.major + 1
        next_version.minor = 0
        next_version.patch = 0
      elsif @config.engine_version_bump_type == :minor
        next_version.minor = next_version.minor + 1
        next_version.patch = 0
      else
        next_version.patch = next_version.patch + 1
      end
      return next_version
    end
  end
end
