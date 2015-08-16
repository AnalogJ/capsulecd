require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'
require_relative 'chef_utils'


class ChefEngine < Engine
  def build_step()
    super

    #no need to bump up the version here. It will automatically be bumped up via the npm version patch command.
    #however we need to read the version from the package.json file and check if a npm module already exists.

    #TODO: check if this module name and version already exist.

    #check for/create any required missing folders/files
    if !File.exists(@source_git_local_path + '/test')
      FileUtils.mkdir(@source_git_local_path + '/test')
    end
  end

  def test_step()
    super

    #the module has already been downloaded. lets make sure all its dependencies are available.
    Open3.popen3('npm install', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
        raise 'npm install failed. Check module dependencies'
      end
    end

    # create a shrinkwrap file.
    if !File.exists(@source_git_local_path + '/npm-shrinkwrap.json')
      Open3.popen3('npm shrinkwrap', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
          raise 'npm shrinkwrap failed. Check log for exact error'
        end
      end
    end


    # run npm test
    Open3.popen3('npm test', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
        raise 'npm test failed. Check log for exact error'
      end
    end
  end

  # run npm publish
  def package_step()
    super

    #TODO: create ~/.npmrc file with credential token, email and username
    #_auth = EDIT: HIDDEN
    #email = npm
    #username = npm

    # run npm publish
    Open3.popen3('npm publish .', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
        raise 'npm test failed. Check log for exact error'
      end
    end
  end

end