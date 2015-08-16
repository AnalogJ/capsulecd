require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'
require_relative 'chef_utils'


class ChefEngine < Engine
  def build_step()
    super

    #bump up the chef cookbook version
    metadata_str = ChefUtils.read_repo_metadata(@source_git_local_path)
    chef_metadata = ChefUtils.parse_metadata(metadata_str)
    next_version = SemVer.parse(chef_metadata.version)
    next_version.patch = next_version.patch + 1

    new_metadata_str = metadata_str.gsub(/(version\s+['"])[0-9\.]+(['"])/, "\\1#{next_version}\\2")
    ChefUtils.write_repo_metadata(@source_git_local_path, new_metadata_str)

    #TODO: check if this cookbook name and version already exist.

    #check for/create any required missing folders/files
    #Berksfile.lock and Gemfile.lock are not required to be commited, but they should be.
    if !File.exists(@source_git_local_path + '/Rakefile')
      File.open(@source_git_local_path + '/Rakefile', 'w'){ |file| file.write('task :test') }
    end
    if !File.exists(@source_git_local_path + '/Berksfile')
      File.open(@source_git_local_path + '/Berksfile', 'w'){ |file| file.write('site :opscode') }
    end
    if !File.exists(@source_git_local_path + '/Gemfile')
      File.open(@source_git_local_path + '/Gemfile', 'w'){ |file| file.write('source "https://rubygems.org"') }
    end
    if !File.exists(@source_git_local_path + '/spec')
      FileUtils.mkdir(@source_git_local_path + '/spec')
    end
  end

  def test_step()
    super

    #the cookbook has already been downloaded. lets make sure all its dependencies are available.
    Open3.popen3('berks install', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
        raise 'berks install failed. Check cookbook dependencies'
      end
    end

    # lets download all its gem dependencies
    Bundler.with_clean_env do
      Open3.popen3('bundle install', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
          raise 'bundle install failed. Check gem dependencies'
        end
      end

      # run rake test
      Open3.popen3('rake test', :chdir => @source_git_local_path) do |stdin, stdout, stderr, external|
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
          raise 'rake test failed. Check log for exact error'
        end
      end
    end
  end

  def package_step()
    super
  end

  #this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
  def release_step()
    super
  end
end