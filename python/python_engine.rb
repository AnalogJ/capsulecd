require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'


class PythonEngine < Engine
  def build_step()
    super

    #check for/create required VERSION file
    if !File.exists?(@source_git_local_path + '/VERSION')
      File.open(@source_git_local_path + '/VERSION', 'w'){ |file| file.write('0.0.0') }
    end

    #bump up the version here.
    #since there's no standardized way to bump up the version in the setup.py file, we're going to assume that the version
    #is specified in a VERSION file in the root of the source repository
    #this is option #4 in the python packaging guide:
    # https://packaging.python.org/en/latest/single_source_version/#single-sourcing-the-version
    #
    #additional packaging structures, like those listed below, may also be supported in the future.
    # http://stackoverflow.com/a/7071358/1157633

    version = File.read(@source_git_local_path + '/VERSION').strip
    next_version = SemVer.parse(version)
    next_version.patch = next_version.patch + 1
    File.open(@source_git_local_path + '/VERSION', 'w') { |file|
      file.write(next_version)
    }


    #TODO: check if this module name and version already exist.

    #check for/create any required missing folders/files
    if !File.exists?(@source_git_local_path + '/requirements.txt')
      File.open(@source_git_local_path + '/requirements.txt', 'w'){ |file| file.write('') }
    end

    if !File.exists?(@source_git_local_path + '/tests')
      FileUtils.mkdir(@source_git_local_path + '/tests')
    end

  end

  def test_step()
    super

    #the module has already been downloaded. lets make sure all its dependencies are available.
    #https://packaging.python.org/en/latest/distributing/
    Open3.popen3('pip install -e .', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
      {:stdout => stdout, :stderr => stderr}. each do |name, stream_buffer|
        Thread.new do
          until (line = stream_buffer.gets).nil? do
            puts "#{name} -> #{line}"
          end
        end
      end
      #wait for process
      external.join
      if !external.value.success?
        raise 'pip install failed. Check module dependencies'
      end
    end

    # there's no standardized method to start tests in python.
    # TODO: check for Makefile?
  end

  # run npm publish
  def package_step()
    super

    # #commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
    # # as they will be included in the commmit and any release artifacts)
    # GitUtils.commit(@source_git_local_path, 'Committing automated changes before packaging.') rescue puts 'Could not commit changes locally..'
    #
    # # run npm publish
    # Open3.popen3('npm version patch -m "(v%s) Automated packaging of release by CapsuleCD"', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
    #   {:stdout => stdout, :stderr => stderr}. each do |name, stream_buffer|
    #     Thread.new do
    #       until (line = stream_buffer.gets).nil? do
    #         puts "#{name} -> #{line}"
    #       end
    #     end
    #   end
    #   #wait for process
    #   external.join
    #   if !external.value.success?
    #     raise 'npm version bump failed'
    #   end
    # end
    #
    # @source_release_commit = GitUtils.head_commit(@source_git_local_path)

  end

  #this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
  def release_step()
    super
    # npmrc_path = File.join(@source_git_local_path, '.npmrc')
    #
    # if !ENV['CAPSULE_NODE_AUTH_TOKEN']
    #   #TODO: make this a warning
    #   puts 'cannot deploy page to npm, credentials missing'
    #   return
    # end
    #
    # #write the knife.rb config file.
    # File.open(npmrc_path, 'w+') { |file|
    #   file.write("//registry.npmjs.org/:_authToken=#{ENV['CAPSULE_NODE_AUTH_TOKEN']}")
    # }

    # run python setup.py sdist upload
    Open3.popen3('python setup.py sdist upload', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
      {:stdout => stdout, :stderr => stderr}. each do |name, stream_buffer|
        Thread.new do
          until (line = stream_buffer.gets).nil? do
            puts "#{name} -> #{line}"
          end
        end
      end
      #wait for process
      external.join
      if !external.value.success?
        raise 'python setup.py upload failed. Check log for exact error'
      end
    end

  end

end