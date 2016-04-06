require_relative 'runner/default'
require_relative 'configuration'
require 'pp'
require 'hooks'
module CapsuleCD
  class Engine
    attr_reader :config

    include Hooks
    define_hooks :pre_source_configure, :post_source_configure,
                 :pre_source_process_pull_request_payload, :post_source_process_pull_request_payload,
                 :pre_source_process_push_payload, :post_source_process_push_payload,
                 :pre_runner_retrieve_payload, :post_runner_retrieve_payload,
                 :pre_build_step, :post_build_step,
                 :pre_test_step, :post_test_step,
                 :pre_package_step, :post_package_step,
                 :pre_source_release, :post_source_release,
                 :pre_release_step, :post_release_step

    # empty hooks
    pre_source_configure do
      puts 'pre_source_configure'
    end

    post_source_configure do
      puts 'post_source_configure'
    end

    def initialize(options)
      @config = CapsuleCD::Configuration.new(options)
      if @config.source == :github
        require_relative 'source/github'
        self.class.send(:include, CapsuleCD::Source::Github)
      else
        fail 'No source defined.'
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
      run_hook :pre_source_configure
      source_configure
      run_hook :post_source_configure

      # runner must determine if this is a pull request or a push.
      # if it's a pull request the runner must retrieve the pull request payload and return it
      # if its a push, the runner must retrieve the push payload and return it
      # the variable @runner_is_pullrequest will be true if a pull request was created.
      # MUST set runner_is_pullrequest
      # REQUIRES source_client
      run_hook :pre_runner_retrieve_payload
      payload = runner_retrieve_payload(@options)
      run_hook :post_runner_retrieve_payload

      if @runner_is_pullrequest
        # all capsule CD processing will be kicked off via a payload. In this case the payload is the pull request data.
        # should check if the pull request opener even has permissions to create a release.
        # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
        # MUST set source_git_local_path
        # MUST set source_git_local_branch
        # MUST set source_git_base_info
        # MUST set source_git_head_info
        # REQUIRES source_client
        run_hook :pre_source_process_pull_request_payload
        source_process_pull_request_payload(payload)
        run_hook :post_source_process_pull_request_payload
      else
        # start processing the payload, which should result in a local git repository that we
        # can begin to test. Since this is a push, no packaging is required
        # MUST set source_git_local_path
        # MUST set source_git_local_branch
        # MUST set source_git_head_info
        # REQUIRES source_client
        run_hook :pre_source_process_push_payload
        source_process_push_payload(payload)
        run_hook :post_source_process_push_payload
      end

      # now that the payload has been processed we can begin by building the code.
      # this may be creating missing files/default structure, compilation, version bumping, etc.
      run_hook :pre_build_step
      source_notify('build') do build_step end
      run_hook :post_build_step

      # this step should download dependencies, run the package test runner(s) (eg. npm test, rake test, kitchen test)
      # REQUIRES @config.engine_cmd_test
      # REQUIRES @config.engine_disable_test
      run_hook :pre_test_step
      source_notify('test') do test_step end
      run_hook :post_test_step

      # this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
      run_hook :pre_package_step
      source_notify('package') do package_step end
      run_hook :post_package_step

      if @runner_is_pullrequest
        # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
        run_hook :pre_release_step
        source_notify('release') do release_step end
        run_hook :post_release_step

        # this step should push the merged, tested and version updated code up to the source code repository
        # this step should also do any source specific releases (github release, asset uploading, etc)
        run_hook :pre_source_release
        source_notify('source release') do source_release end
        run_hook :post_source_release
      end

      # rescue => ex #TODO if you enable this rescue block, hooks stop working.
      # TODO: it shouldnt be required anylonger because source_notify will handle rescueing the failures.
      #   puts ex
      #
      #   self.run_hook :pre_source_process_failure, ex
      #   source_process_failure(ex)
      #   self.run_hook :post_source_process_failure, ex
    end

    def build_step
      puts 'build_step'
    end

    def test_step
      puts 'test_step'
    end

    # the package_step should always set the @source_release_commit and optionally set add-to/set the @source_release_artifacts array
    def package_step
      puts 'package_step'
    end

    def release_step
      puts 'release_step'
    end

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

    # metaprogramming method to
    def modules_include

    end


  end
end
