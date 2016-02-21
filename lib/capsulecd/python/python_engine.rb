require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'

module CapsuleCD
  module Python
    class PythonEngine < Engine
      def build_step
        super

        # check for/create required VERSION file
        unless File.exist?(@source_git_local_path + '/VERSION')
          File.open(@source_git_local_path + '/VERSION', 'w') { |file| file.write('0.0.0') }
        end

        # bump up the version here.
        # since there's no standardized way to bump up the version in the setup.py file, we're going to assume that the version
        # is specified in a VERSION file in the root of the source repository
        # this is option #4 in the python packaging guide:
        # https://packaging.python.org/en/latest/single_source_version/#single-sourcing-the-version
        #
        # additional packaging structures, like those listed below, may also be supported in the future.
        # http://stackoverflow.com/a/7071358/1157633

        version = File.read(@source_git_local_path + '/VERSION').strip
        next_version = SemVer.parse(version)
        next_version.patch = next_version.patch + 1
        File.open(@source_git_local_path + '/VERSION', 'w') do |file|
          file.write(next_version)
        end

        # TODO: check if this module name and version already exist.

        # check for/create any required missing folders/files
        unless File.exist?(@source_git_local_path + '/requirements.txt')
          File.open(@source_git_local_path + '/requirements.txt', 'w') { |file| file.write('') }
        end

        unless File.exist?(@source_git_local_path + '/tests')
          FileUtils.mkdir(@source_git_local_path + '/tests')
        end
        unless File.exist?(@source_git_local_path + '/tests/__init__.py')
          File.open(@source_git_local_path + '/tests/__init__.py', 'w') { |file| file.write('') }
        end
      end

      def test_step
        super

        Open3.popen3('pip install -r requirements.txt', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            raise CapsuleCD::Error::TestDependenciesError, 'pip install package requirements failed. Check module dependencies'
          end
        end

        # the module has already been downloaded. lets make sure all its dependencies are available.
        # https://packaging.python.org/en/latest/distributing/
        Open3.popen3('pip install -e .', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            raise CapsuleCD::Error::TestDependenciesError, 'pip install package in development mode failed.'
          end
        end

        # run tests
        Open3.popen3('python setup.py test', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            raise CapsuleCD::Error::TestDependenciesError, 'pip install package in development mode failed.'
          end
        end
      end

      # run npm publish
      def package_step
        super

        # commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
        # as they will be included in the commmit and any release artifacts)
        version = File.read(@source_git_local_path + '/VERSION').strip
        next_version = SemVer.parse(version)
        CapsuleCD::GitUtils.commit(@source_git_local_path, "(v#{next_version}) Automated packaging of release by CapsuleCD")
        @source_release_commit = CapsuleCD::GitUtils.tag(@source_git_local_path, "v#{next_version}")
      end

      # this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
      def release_step
        super
        pypirc_path = File.expand_path('~/.pypirc')

        unless ENV['CAPSULE_PYPI_USERNAME'] || ENV['CAPSULE_PYPI_PASSWORD']
          raise CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy package to pip, credentials missing'
          return
        end

        # write the knife.rb config file.
        File.open(pypirc_path, 'w+') do |file|
          file.write(<<-EOT.gsub(/^\s+/, '')
            [distutils]
            index-servers=pypi

            [pypi]
            repository = https://pypi.python.org/pypi
            username = #{ENV['CAPSULE_PYPI_USERNAME']}
            password = #{ENV['CAPSULE_PYPI_PASSWORD']}
          EOT
                    )
        end

        # run python setup.py sdist upload
        # TODO: use twine instead (it supports HTTPS.)https://python-packaging-user-guide.readthedocs.org/en/latest/distributing/#uploading-your-project-to-pypi
        Open3.popen3('python setup.py sdist upload', chdir: @source_git_local_path) do |_stdin, stdout, stderr, external|
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
            raise CapsuleCD::Error::ReleasePackageError, 'python setup.py upload failed. Check log for exact error'
          end
        end
      end
    end
  end
end