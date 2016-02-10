require_relative 'source/github'
require 'hooks'
class Engine
  include Hooks
  define_hooks :before_source_configure, :after_source_configure,
               :before_source_process_payload, :after_source_process_payload,
               :before_runner_retrieve_payload, :after_runner_retrieve_payload,
               :before_build_step, :after_build_step,
               :before_test_step, :after_test_step,
               :before_package_step, :after_package_step,
               :before_source_release, :after_source_release,
               :before_release_step, :after_release_step


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
    run_hook :before_source_configure
    source_configure()
    run_hook :after_source_configure

    run_hook :before_runner_retrieve_payload
    @payload = runner_retrieve_payload(@options)
    run_hook :after_runner_retrieve_payload


    #start processing the payload, which should result in a local merged git repository that we
    # can begin to test. Processing the payload should also verify if the payload creator has correct access/permissions
    # to kick off a new release.
    run_hook :before_source_process_payload
    source_process_payload(@payload)
    run_hook :after_source_process_payload

    # now that the payload has been processed we can begin by building the code.
    # this may be compilation, dependency downloading, etc.
    run_hook :before_build_step
    build_step()
    run_hook :after_build_step

    # this step should run the package test runner(s) (eg. npm test, rake test, kitchen test)
    run_hook :before_test_step
    test_step()
    run_hook :after_test_step

    #
    run_hook :before_package_step
    package_step()
    run_hook :after_package_step

    run_hook :before_release_step
    release_step()
    run_hook :after_release_step

    run_hook :before_source_release
    source_release()
    run_hook :after_source_release
  rescue => ex

    run_hook :before_source_process_failure, ex
    source_process_failure(ex)
    run_hook :after_source_process_failure, ex

  end


  def build_step()
  end

  def test_step()
  end

  #the package_step should always set the @source_release_commit and optionally set add-to/set the @source_release_artifacts array
  def package_step()

  end

  def release_step()
  end
end