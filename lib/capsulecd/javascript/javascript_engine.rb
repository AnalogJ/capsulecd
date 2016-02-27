require 'semverly'
require 'open3'
require 'bundler'
require 'json'
require_relative '../base/engine'

module CapsuleCD
  module Javascript
    class JavascriptEngine < Engine
      def build_step
        super

        @_is_bower = File.exist?(@source_git_local_path + '/bower.json')
        @_is_npm = File.exist?(@source_git_local_path + '/package.json')

        # we can't bump the npm version here because the npm version patch command will set it.
        # howerver we need to make sure the bower.json and package.json versions are insync.
        # we'll take the latest version of either the package.json or bower.json and set that as the version of both.
        sync_versions

        # now that the bower and package versions are in sync, lets bump the version of bower.json
        # (because package.json will be bumped automatically.)
        if @_is_bower
          bower_file = File.read(@source_git_local_path + '/bower.json')
          bower_data = JSON.parse(bower_file)
          next_version = SemVer.parse(bower_data['version'])
          next_version.patch = next_version.patch + 1
          bower_data['version'] = next_version.to_s
          File.write(@source_git_local_path + '/bower.json', bower_data.to_json)
        end


        # TODO: check if this module name and version already exist.

        # check for/create any required missing folders/files
        unless File.exist?(@source_git_local_path + '/test')
          FileUtils.mkdir(@source_git_local_path + '/test')
        end
        unless File.exist?(@source_git_local_path + '/.gitignore')
          CapsuleCD::GitHelper.create_gitignore(@source_git_local_path, ['Node','Yeoman'])
        end
      end

      def test_step
        super

        if @_is_npm
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

        if @_is_bower
          # lets make sure all the bower dependencies are available.
          Open3.popen3('bower install', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
        end
      end

      # run npm publish
      def package_step
        super

        # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
        # as they will be included in the commmit and any release artifacts)
        CapsuleCD::GitUtils.commit(@source_git_local_path, 'Committing automated changes before packaging.') rescue puts 'No changes to commit..'
        if @_is_bower && !@_is_npm
          bower_file = File.read(@source_git_local_path + '/bower.json')
          bower_data = JSON.parse(bower_file)
          next_version = SemVer.parse(bower_data['version'])
          @source_release_commit = CapsuleCD::GitUtils.tag(@source_git_local_path, "v#{next_version}")

        else

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
          @source_release_commit = CapsuleCD::GitUtils.get_latest_tag_commit(@source_git_local_path)
        end


      end

      # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      def release_step
        super

        if @_is_npm
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

      private
      def sync_versions
        #this method only needs to run if bower and package json files exist.
        if !@_is_bower || !@_is_npm
          return
        end

        bower_file = File.read(@source_git_local_path + '/bower.json')
        bower_data = JSON.parse(bower_file)
        bower_version = SemVer.parse(bower_data['version'])
        package_file = File.read(@source_git_local_path + '/package.json')
        package_data = JSON.parse(package_file)
        package_version = SemVer.parse(package_data['version'])

        if(bower_version>package_version)
          package_data['version'] = bower_version.to_s
          File.write(@source_git_local_path + '/package.json', package_data.to_json)
        else
          bower_data['version'] = package_version.to_s
          File.write(@source_git_local_path + '/bower.json', bower_data.to_json)
        end
      end
    end
  end
end
