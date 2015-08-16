require 'ridley'
require 'berkshelf'

class ChefUtils

  def self.read_repo_metadata(repo_path, metadata_filename = 'metadata.rb')
    return File.read("#{repo_path}/#{metadata_filename}")
  end

  def self.write_repo_metadata(repo_path, metadata_str, metadata_filename = 'metadata.rb')
    File.open("#{repo_path}/#{metadata_filename}",'w'){ |file| f.write(metadata_str)}
  end
  def self.parse_metadata(metadata_str)
    chef_metadata = Ridley::Chef::Cookbook::Metadata.new
    chef_metadata.instance_eval(metadata_str)
    return chef_metadata
  end

  def self.parse_berksfile_lock_from_repo(repo_path, lockfile_filename='Berksfile.lock')
    return Berkshelf::Lockfile.from_file("#{repo_path}/#{lockfile_filename}")
  end

  def self.parse_berksfile_from_repo(repo_path, berksfile_filename='Berksfile')
    return Berkshelf::Lockfile.from_file("#{repo_path}/#{berksfile_filename}")
  end
end