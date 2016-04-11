require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'

module CapsuleCD
  module Node
    class NodeEngine < Engine
      def build_step
        super
        # validate that the chef metadata.rb file exists
        unless File.exist?(@source_git_local_path + '/package.json')
          fail CapsuleCD::Error::BuildPackageInvalid, 'package.json file is required to process Npm package'
        end

        # no need to bump up the version here. It will automatically be bumped up via the npm version patch command.
        # however we need to read the version from the package.json file and check if a npm module already exists.

        # TODO: check if this module name and version already exist.

        # check for/create any required missing folders/files
        unless File.exist?(@source_git_local_path + '/test')
          FileUtils.mkdir(@source_git_local_path + '/test')
        end
        unless File.exist?(@source_git_local_path + '/.gitignore')
          CapsuleCD::GitUtils.create_gitignore(@source_git_local_path, ['Node'])
        end
      end

      def test_step
        super

        # the module has already been downloaded. lets make sure all its dependencies are available.
        Open3.popen3('npm install', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
          { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
            Thread.new do
              until (line = stream_buffer.gets).nil?
                puts "#{name} -> #{line}"
              end
            end
          end
          # wait for process
          external.join
          unless external.value.success?
            fail CapsuleCD::Error::TestDependenciesError, 'npm install failed. Check module dependencies'
          end
        end

        # create a shrinkwrap file.
        Open3.popen3('npm shrinkwrap', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
          { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
            Thread.new do
              until (line = stream_buffer.gets).nil?
                puts "#{name} -> #{line}"
              end
            end
          end
          # wait for process
          external.join
          unless external.value.success?
            fail CapsuleCD::Error::TestDependenciesError, 'npm shrinkwrap failed. Check log for exact error'
          end
        end

        # run test command
        test_cmd = @config.engine_cmd_test || 'npm test'
        Open3.popen3(ENV, test_cmd, chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
          { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
            Thread.new do
              until (line = stream_buffer.gets).nil?
                puts "#{name} -> #{line}"
              end
            end
          end
          # wait for process
          external.join
          unless external.value.success?
            fail CapsuleCD::Error::TestRunnerError, test_cmd + ' failed. Check log for exact error'
          end
        end unless @config.engine_disable_test
      end

      # run npm publish
      def package_step
        super

        # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
        # as they will be included in the commmit and any release artifacts)
        CapsuleCD::GitUtils.commit(@source_git_local_path, 'Committing automated changes before packaging.') rescue puts 'No changes to commit..'

        # run npm publish
        Open3.popen3("npm version #{@config.engine_version_bump_type} -m '(v%s) Automated packaging of release by CapsuleCD'", chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
          { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
            Thread.new do
              until (line = stream_buffer.gets).nil?
                puts "#{name} -> #{line}"
              end
            end
          end
          # wait for process
          external.join
          fail 'npm version bump failed' unless external.value.success?
        end

        @source_release_commit = CapsuleCD::GitUtils.get_latest_tag_commit(@source_git_local_path)
      end

      # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      def release_step
        super
        npmrc_path = File.expand_path('~/.npmrc')

        unless @config.npm_auth_token
          fail CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy page to npm, credentials missing'
        end

        # write the knife.rb config file.
        File.open(npmrc_path, 'w+') do |file|
          file.write("//registry.npmjs.org/:_authToken=#{@config.npm_auth_token}")
        end

        # run npm publish
        Open3.popen3('npm publish .', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
          { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
            Thread.new do
              until (line = stream_buffer.gets).nil?
                puts "#{name} -> #{line}"
              end
            end
          end
          # wait for process
          external.join
          unless external.value.success?
            fail CapsuleCD::Error::ReleasePackageError, 'npm publish failed. Check log for exact error'
          end
        end
      end
    end
  end
end
