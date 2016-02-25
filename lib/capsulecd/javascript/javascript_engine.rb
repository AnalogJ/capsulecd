require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'

module CapsuleCD
  module Javascript
    class JavascriptEngine < Engine
      def build_step
        super

        # no need to bump up the version here. It will automatically be bumped up via the npm version patch command.
        # however we need to read the version from the package.json file and check if a npm module already exists.

        # TODO: check if this module name and version already exist.

        # check for/create any required missing folders/files
        unless File.exist?(@source_git_local_path + '/test')
          FileUtils.mkdir(@source_git_local_path + '/test')
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
        unless File.exist?(@source_git_local_path + '/npm-shrinkwrap.json')
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
        end

        # run npm test
        Open3.popen3(ENV, 'npm test', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            fail CapsuleCD::Error::TestRunnerError, 'npm test failed. Check log for exact error'
          end
        end
      end

      # run npm publish
      def package_step
        super

        # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
        # as they will be included in the commmit and any release artifacts)
        CapsuleCD::GitUtils.commit(@source_git_local_path, 'Committing automated changes before packaging.') rescue puts 'Could not commit changes locally..'

        # run npm publish
        Open3.popen3('npm version patch -m "(v%s) Automated packaging of release by CapsuleCD"', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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

        @source_release_commit = CapsuleCD::GitUtils.head_commit(@source_git_local_path)
      end

      # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      def release_step
        super
        npmrc_path = File.join(@source_git_local_path, '.npmrc')

        unless @config.npm_auth_token
          fail CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy page to npm, credentials missing'
          return
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
