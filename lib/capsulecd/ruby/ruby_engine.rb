require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'
require_relative 'ruby_helper'

module CapsuleCD
  module Ruby
    class RubyEngine < Engine
      def build_step
        super
        gemspec_path = CapsuleCD::Ruby::RubyHelper.get_gemspec_path(@source_git_local_path)

        # check for required VERSION file
        gemspec_data = CapsuleCD::Ruby::RubyHelper.get_gemspec_data(@source_git_local_path)

        if !File.exist?(CapsuleCD::Ruby::RubyHelper.version_filepath(@source_git_local_path, gemspec_data.name))
          fail CapsuleCD::Error::BuildPackageInvalid, 'version.rb file is required to process Ruby gem'
        end

        # bump up the version here.
        # since there's no standardized way to bump up the version in the *.gemspec file, we're going to assume that the version
        # is specified in a version file in the lib/<gem_name>/ directory, similar to how the bundler gem does it.
        # ie. bundle gem <gem_name> will create a file: my_gem/lib/my_gem/version.rb with the following contents
        # module MyGem
        #   VERSION = "0.1.0"
        # end
        #
        # Jeweler and Hoe both do something similar.
        # http://yehudakatz.com/2010/04/02/using-gemspecs-as-intended/
        # http://timelessrepo.com/making-ruby-gems
        # http://guides.rubygems.org/make-your-own-gem/

        version_str = CapsuleCD::Ruby::RubyHelper.read_version_file(@source_git_local_path, gemspec_data.name)
        next_version = bump_version(SemVer.parse(gemspec_data.version.to_s))

        new_version_str = version_str.gsub(/(VERSION\s*=\s*['"])[0-9\.]+(['"])/, "\\1#{next_version}\\2")
        CapsuleCD::Ruby::RubyHelper.write_version_file(@source_git_local_path, gemspec_data.name, new_version_str)

        # check for/create any required missing folders/files
        unless File.exist?(@source_git_local_path + '/Gemfile')
          File.open(@source_git_local_path + '/Gemfile', 'w') { |file|
            file.puts("source 'https://rubygems.org'")
            file.puts('gemspec')
          }
        end
        unless File.exist?(@source_git_local_path + '/Rakefile')
          File.open(@source_git_local_path + '/Rakefile', 'w') { |file| file.write('task :default => :spec') }
        end
        unless File.exist?(@source_git_local_path + '/spec')
          FileUtils.mkdir(@source_git_local_path + '/spec')
        end
        unless File.exist?(@source_git_local_path + '/.gitignore')
          CapsuleCD::GitUtils.create_gitignore(@source_git_local_path, ['Ruby'])
        end

        # package the gem, make sure it builds correctly
        Open3.popen3('gem build '+ File.basename(gemspec_path), chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            fail CapsuleCD::Error::BuildPackageFailed, 'gem build failed. Check gemspec file and dependencies'
          end
          unless File.exist?(@source_git_local_path + "/#{gemspec_data.name}-#{next_version.to_s}.gem")
            fail CapsuleCD::Error::BuildPackageFailed, "gem build failed. #{gemspec_data.name}-#{next_version.to_s}.gem not found"
          end
        end
      end

      def test_step
        super

        gems = Dir.glob(@source_git_local_path + '/*.gem')
        if gems.empty?
          fail CapsuleCD::Error::TestDependenciesError, 'Ruby gem file could not be found'
        end
        gem_path = gems.first

        # lets install the gem, and any dependencies
        # http://guides.rubygems.org/make-your-own-gem/
        Bundler.with_clean_env do
          Open3.popen3('gem install ./' + File.basename(gem_path) + ' --ignore-dependencies', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
              fail CapsuleCD::Error::TestDependenciesError, 'gem install failed. Check gemspec and gem dependencies'
            end
          end

          Open3.popen3('bundle install', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
              fail CapsuleCD::Error::TestDependenciesError, 'bundle install failed. Check Gemfile'
            end
          end

          # run test command
          test_cmd = @config.engine_cmd_test || 'rake spec'
          Open3.popen3(test_cmd, chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
      end

      # run npm publish
      def package_step
        super

        # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
        # as they will be included in the commmit and any release artifacts)
        gemspec_data = CapsuleCD::Ruby::RubyHelper.get_gemspec_data(@source_git_local_path)
        next_version = SemVer.parse(gemspec_data.version.to_s)
        CapsuleCD::GitUtils.commit(@source_git_local_path, "(v#{next_version}) Automated packaging of release by CapsuleCD")
        @source_release_commit = CapsuleCD::GitUtils.tag(@source_git_local_path, "v#{next_version}")
      end

      # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      def release_step
        super

        unless @config.rubygems_api_key
          fail CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy package to rubygems, credentials missing'
          return
        end

        # write the config file.
        rubygems_cred_path = File.expand_path('~/.gem')

        FileUtils.mkdir_p(rubygems_cred_path)
        File.open(rubygems_cred_path + '/credentials', 'w+', 0600) do |file|
          file.write(<<-EOT.gsub(/^\s+/, '')
            ---
            :rubygems_api_key: #{@config.rubygems_api_key}
            EOT
          )
        end

        # run gem push *.gem
        gems = Dir.glob(@source_git_local_path + '/*.gem')
        if gems.empty?
          fail CapsuleCD::Error::TestDependenciesError, 'Ruby gem file could not be found'
        end
        gem_path = gems.first
        Open3.popen3('gem push ' + File.basename(gem_path), chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            fail CapsuleCD::Error::ReleasePackageError, 'Pushing gem to RubyGems.org using `gem push` failed. Check log for exact error'
          end
        end
      end
    end
  end
end
