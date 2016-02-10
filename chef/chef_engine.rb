require 'semverly'
require 'open3'
require 'bundler'
require_relative '../base/engine'
require_relative 'chef_utils'
require 'base64'
require 'fileutils'

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
    if !File.exists?(@source_git_local_path + '/Rakefile')
      File.open(@source_git_local_path + '/Rakefile', 'w'){ |file| file.write('task :test') }
    end
    if !File.exists?(@source_git_local_path + '/Berksfile')
      File.open(@source_git_local_path + '/Berksfile', 'w'){ |file| file.write('site :opscode') }
    end
    if !File.exists?(@source_git_local_path + '/Gemfile')
      File.open(@source_git_local_path + '/Gemfile', 'w'){ |file| file.write('source "https://rubygems.org"') }
    end
    if !File.exists?(@source_git_local_path + '/spec')
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
    metadata_str = ChefUtils.read_repo_metadata(@source_git_local_path)
    chef_metadata = ChefUtils.parse_metadata(metadata_str)
    next_version = SemVer.parse(chef_metadata.version)
    #commit changes to the cookbook. (test run occurs before this, and it should clean up any instrumentation files, created,
    # as they will be included in the commmit and any release artifacts)
    GitUtils.commit(@source_git_local_path, "(v#{next_version.to_s}) Automated packaging of release by CapsuleCD")
    @source_release_commit = GitUtils.tag(@source_git_local_path, "v#{next_version.to_s}")

  end

  #this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
  def release_step()
    super
    puts @source_git_parent_path
    pem_path = File.join(@source_git_parent_path, 'client.pem')
    knife_path = File.join(@source_git_parent_path, 'knife.rb')

    if !(ENV['CAPSULE_CHEF_SUPERMARKET_USERNAME'] || ENV['CAPSULE_CHEF_SUPERMARKET_KEY'])
      #TODO: make this a warning
      puts 'cannot deploy cookbook to supermarket, credentials missing'
      return
    end

    #write the knife.rb config file.
    File.open(knife_path, 'w+') { |file|
      file.write(<<EOT.gsub(/^\s+/, '')
  node_name "#{ENV['CAPSULE_CHEF_SUPERMARKET_USERNAME']}" # Replace with the login name you use to login to the Supermarket.
  client_key "#{pem_path}" # Define the path to wherever your client.pem file lives.  This is the key you generated when you signed up for a Chef account.
  cookbook_path [ '#{@source_git_parent_path}' ] # Directory where the cookbook you're uploading resides.
EOT
      )
    }

    File.open(pem_path, 'w+') { |file|
      key = Base64.strict_decode64(ENV['CAPSULE_CHEF_SUPERMARKET_KEY'])
      file.write(key)
    }

    metadata_str = ChefUtils.read_repo_metadata(@source_git_local_path)
    chef_metadata = ChefUtils.parse_metadata(metadata_str)

    command = "knife cookbook site share #{chef_metadata.name} #{ENV['CAPSULE_CHEF_SUPERMARKET_TYPE'] || 'Other'}  -c #{knife_path}"
    Open3.popen3(command) do |stdin, stdout, stderr, external|
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
        raise 'knife cookbook upload to supermarket failed'
      end
    end

  end
end