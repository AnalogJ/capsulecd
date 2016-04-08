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
        gemspecs = Dir.glob(@source_git_local_path + '/*.gemspec')
        if gemspecs.empty?
          fail CapsuleCD::Error::BuildPackageInvalid, '*.gemspec file is required to process Ruby gem'
        end
        gemspec_path = gemspecs.first


        # check for/create required VERSION file
        require 'rubygems'
        gemspec_data = Gem::Specification::load(gemspec_path)

        unless File.exist?(CapsuleCD::Ruby::RubyHelper.version_filepath(@source_git_local_path, gemspec_data.name))
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
        next_version = bump_version(SemVer.parse(gemspec_data.version))

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
          Open3.popen3('gem install ./' + gem_path, chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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

          # run test command
          test_cmd = @config.engine_cmd_test || 'rake test'
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

      # # run npm publish
      # def package_step
      #   super
      #
      #   # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
      #   # as they will be included in the commmit and any release artifacts)
      #   version = File.read(@source_git_local_path + '/VERSION').strip
      #   next_version = SemVer.parse(version)
      #   CapsuleCD::GitUtils.commit(@source_git_local_path, "(v#{next_version}) Automated packaging of release by CapsuleCD")
      #   @source_release_commit = CapsuleCD::GitUtils.tag(@source_git_local_path, "v#{next_version}")
      # end
      #
      # # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      # def release_step
      #   super
      #   pypirc_path = File.expand_path('~/.pypirc')
      #
      #   unless @config.pypi_username || @config.pypi_password
      #     fail CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy package to pip, credentials missing'
      #     return
      #   end
      #
      #   # write the knife.rb config file.
      #   File.open(pypirc_path, 'w+') do |file|
      #     file.write(<<-EOT.gsub(/^\s+/, '')
      #       [distutils]
      #       index-servers=pypi
      #
      #       [pypi]
      #       repository = https://pypi.python.org/pypi
      #       username = #{@config.pypi_username}
      #       password = #{@config.pypi_password}
      #     EOT
      #     )
      #   end
      #
      #   # run python setup.py sdist upload
      #   # TODO: use twine instead (it supports HTTPS.)https://python-packaging-user-guide.readthedocs.org/en/latest/distributing/#uploading-your-project-to-pypi
      #   Open3.popen3('python setup.py sdist upload', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
      #     { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
      #       Thread.new do
      #         until (line = stream_buffer.gets).nil?
      #           puts "#{name} -> #{line}"
      #         end
      #       end
      #     end
      #     # wait for process
      #     external.join
      #     unless external.value.success?
      #       fail CapsuleCD::Error::ReleasePackageError, 'python setup.py upload failed. Check log for exact error'
      #     end
      #   end
      # end
    end
  end
end
