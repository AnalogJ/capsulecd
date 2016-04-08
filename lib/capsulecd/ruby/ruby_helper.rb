
module CapsuleCD
  module Ruby
    class RubyHelper
      def self.version_filepath(repo_path, gem_name, version_filename = 'version.rb')
        "#{repo_path}/lib/#{gem_name}/#{version_filename}"
      end

      def self.read_version_file(repo_path, gem_name, version_filename = 'version.rb')
        File.read(self.version_filepath(repo_path, gem_name, version_filename))
      end

      def self.write_version_file(repo_path, gem_name, metadata_str, version_filename = 'version.rb')
        File.open(self.version_filepath(repo_path, gem_name, version_filename), 'w') { |file| file.write(metadata_str) }
      end
    end
  end
end
