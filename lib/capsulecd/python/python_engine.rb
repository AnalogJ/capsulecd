require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'

module CapsuleCD
  module Python
    class PythonEngine < Engine
      def build_step
        super
        unless File.exist?(@source_git_local_path + '/setup.py')
          fail CapsuleCD::Error::BuildPackageInvalid, 'setup.py file is required to process Python package'
        end

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

        # make sure the package testing manager is available.
        # there is a standardized way to test packages (python setup.py tests), however for automation tox is preferred
        # because of virtualenv and its support for multiple interpreters.
        unless File.exist?(@source_git_local_path + '/tox.ini')
          # if a tox.ini file is not present, we'll create a default one and specify 'python setup.py test' as the test
          # runner command, and requirements.txt as the dependencies for this package.
          File.open(@source_git_local_path + '/tox.ini', 'w') { |file|
            file.write(<<-TOX.gsub(/^\s+/, '')
# Tox (http://tox.testrun.org/) is a tool for running tests
# in multiple virtualenvs. This configuration file will run the
# test suite on all supported python versions. To use it, "pip install tox"
# and then run "tox" from this directory.

[tox]
envlist = py27
usedevelop = True

[testenv]
commands = python setup.py test
deps =
  -rrequirements.txt
TOX
            )
          }
        end

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
        unless File.exist?(@source_git_local_path + '/.gitignore')
          CapsuleCD::GitUtils.create_gitignore(@source_git_local_path, ['Python'])
        end
      end

      def test_step
        super


        # download the package dependencies and register it in the virtualenv using tox (which will do pip install -e .)
        # https://packaging.python.org/en/latest/distributing/
        # once that's done, tox will run tests
        # run test command
        test_cmd = @config.engine_cmd_test || 'tox'
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
            fail CapsuleCD::Error::TestDependenciesError, test_cmd + ' failed to test package.'
          end
        end unless @config.engine_disable_test
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

        unless @config.pypi_username || @config.pypi_password
          fail CapsuleCD::Error::ReleaseCredentialsMissing, 'cannot deploy package to pip, credentials missing'
          return
        end

        # write the knife.rb config file.
        File.open(pypirc_path, 'w+') do |file|
          file.write(<<-EOT.gsub(/^\s+/, '')
            [distutils]
            index-servers=pypi

            [pypi]
            repository = https://pypi.python.org/pypi
            username = #{@config.pypi_username}
            password = #{@config.pypi_password}
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
            fail CapsuleCD::Error::ReleasePackageError, 'python setup.py upload failed. Check log for exact error'
          end
        end
      end
    end
  end
end
