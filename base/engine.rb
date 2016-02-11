require_relative 'source/github'
require_relative 'runner/circleci'
require_relative 'runner/default'
require 'hooks'
class Engine
  include Hooks
  define_hooks :before_source_configure, :after_source_configure,
               :before_source_process_pull_request_payload, :after_source_process_pull_request_payload,
               :before_source_process_push_payload, :after_source_process_push_payload,
               :before_runner_retrieve_payload, :after_runner_retrieve_payload,
               :before_build_step, :after_build_step,
               :before_test_step, :after_test_step,
               :before_package_step, :after_package_step,
               :before_source_release, :after_source_release,
               :before_release_step, :after_release_step

  #empty hooks
  before_source_configure {
    puts 'before_source_configure'
  }

  after_source_configure {
    puts 'after_source_configure'
  }


  def initialize(options)
    @options = options
    if options[:source] == :github
      self.class.send(:include, GithubSource)
    else
      raise 'No source defined.'
    end

    if options[:runner] == :circleci
      self.class.send(:include, CircleciRunner)
    end

  end

  def start()

    #start the source, and whatever work needs to be done there.
    self.run_hook :before_source_configure
    source_configure()
    self.run_hook :after_source_configure

    #runner must determine if this is a pull request or a push.
    #if it's a pull request the runner must retrieve the pull request payload and return it
    #if its a push, the runner must retrieve the push payload and return it
    #the variable @runner_is_pullrequest MUST be set if a pull request was created.
    self.run_hook :before_runner_retrieve_payload
    payload = runner_retrieve_payload(@options)
    self.run_hook :after_runner_retrieve_payload

    if @runner_is_pullrequest
      # all capsule CD processing will be kicked off via a payload. In this case the payload is the pull request data.
      # should check if the pull request opener even has permissions to create a release.
      # all sources should process the payload by downloading a git repository that contains the master branch merged with the test branch
      # MUST set source_git_local_path
      self.run_hook :before_source_process_pull_request_payload
      source_process_pull_request_payload(payload)
      self.run_hook :after_source_process_pull_request_payload
    else
      #start processing the payload, which should result in a local git repository that we
      # can begin to test. This step should bump up the package version. Since this is a push, no packaging is required
      # MUST set source_git_local_path
      self.run_hook :before_source_process_push_payload
      source_process_push_payload(payload)
      self.run_hook :after_source_process_push_payload
    end


    # now that the payload has been processed we can begin by building the code.
    # this may be compilation, dependency downloading, etc.
    self.run_hook :before_build_step
    build_step()
    self.run_hook :after_build_step

    # this step should run the package test runner(s) (eg. npm test, rake test, kitchen test)
    self.run_hook :before_test_step
    test_step()
    self.run_hook :after_test_step

    # this step should commit any local changes and create a git tag. Nothing should be pushed to remote repository
    self.run_hook :before_package_step
    package_step()
    self.run_hook :after_package_step

    if(@runner_is_pullrequest)
      #this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      self.run_hook :before_release_step
      release_step()
      self.run_hook :after_release_step

      # this step should push the merged, tested and version updated code up to the source code repository
      # this step should also do any source specific releases (github release, asset uploading, etc)
      self.run_hook :before_source_release
      source_release()
      self.run_hook :after_source_release
    end



  # rescue => ex #TODO if you enable this rescue block, hooks stop working.
  #   puts ex
  #
  #   self.run_hook :before_source_process_failure, ex
  #   source_process_failure(ex)
  #   self.run_hook :after_source_process_failure, ex

  end

  def build_step()
    puts 'build_step'
  end

  def test_step()
    puts 'test_step'
  end

  #the package_step should always set the @source_release_commit and optionally set add-to/set the @source_release_artifacts array
  def package_step()
    puts 'package_step'
  end

  def release_step()
    puts 'release_step'
  end

end