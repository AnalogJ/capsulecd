require 'ridley'
require 'berkshelf'

module CapsuleCD
  module Chef
    class ChefHelper
      def self.read_repo_metadata(repo_path, metadata_filename = 'metadata.rb')
        File.read("#{repo_path}/#{metadata_filename}")
      end

      def self.write_repo_metadata(repo_path, metadata_str, metadata_filename = 'metadata.rb')
        File.open("#{repo_path}/#{metadata_filename}", 'w') { |file| file.write(metadata_str) }
      end
      def self.parse_metadata(metadata_str)
        chef_metadata = Ridley::Chef::Cookbook::Metadata.new
        chef_metadata.instance_eval(metadata_str)
        chef_metadata
      end

      def self.parse_berksfile_lock_from_repo(repo_path, lockfile_filename = 'Berksfile.lock')
        Berkshelf::Lockfile.from_file("#{repo_path}/#{lockfile_filename}")
      end

      def self.parse_berksfile_from_repo(repo_path, berksfile_filename = 'Berksfile')
        Berkshelf::Lockfile.from_file("#{repo_path}/#{berksfile_filename}")
      end
    end
  end
end
