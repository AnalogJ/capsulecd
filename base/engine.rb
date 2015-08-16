require_relative 'source/github'
require 'hooks'
class Engine
  include Hooks
  define_hooks :before_source_configure, :after_source_configure,
               :before_source_process_payload, :after_source_process_payload,
               :before_build_step, :after_build_step,
               :before_test_step, :after_test_step,
               :before_package_step, :after_package_step,
               :before_release_step, :after_release_step


  def initialize(source)
    if source == 'github'
      self.class.send(:include, GithubSource)
    else
      raise 'No source defined.'
    end
  end

  def start(source_payload)

    #start the source, and whatever work needs to be done there.
    run_hook :before_source_configure
    source_configure()
    run_hook :after_source_configure

    #start processing the payload, which should result in a local merged git repository that we
    # can begin to test. Processing the payload should also verify if the payload creator has correct access/permissions
    # to kick off a new release.
    run_hook :before_source_process_payload
    source_process_payload(source_payload)
    run_hook :after_source_process_payload

    # now that the payload has been processed we can begin by building the code.
    # this may be compilation, dependency downloading, etc.
    run_hook :before_build_step
    build_step()
    run_hook :after_build_step

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

    run_hook :before_source_release
    source_release()
    run_hook :after_source_release

    run_hook :before_release_step
    release_step()
    run_hook :after_release_step
  end


  def build_step()
  end

  def test_step()
  end

  #the package_step should always set the @source_release_commit and optionally set add-to/set the @source_release_artifacts array
  def package_step()
    #commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
    # as they will be included in the commmit and any release artifacts)
    GitUtils.commit(@source_git_local_path, "(v#{next_version.to_s}) Automated packaging of release by Capsule-CD")
    @source_release_commit = GitUtils.tag(@source_git_local_path, "v#{next_version.to_s}")

  end

  def release_step()
  end
end