require 'rubygems'
require 'thread'
require 'bundler'
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

      def self.get_gemspec_path(repo_path)
        gemspecs = Dir.glob(repo_path + '/*.gemspec')
        if gemspecs.empty?
          fail CapsuleCD::Error::BuildPackageInvalid, '*.gemspec file is required to process Ruby gem'
        end
        gemspecs.first
      end

      def self.get_gemspec_data(repo_path)
        self.load_gemspec_data(self.get_gemspec_path(repo_path))
      end

      ##################################################################################################################
      # protected/private methods.

      # since the Gem::Specification class is basically eval'ing the gemspec file, and the gemspec file is doing a require
      # to load the version.rb file, the version.rb file is cached in memory. We're going to try to get around that issue
      # by parsing the gemspec file in a forked process.
      # https://stackoverflow.com/questions/1076257/returning-data-from-forked-processes
      def self.execute_in_child
        read, write = IO.pipe

        pid = fork do
          read.close
          result = yield
          Marshal.dump(result, write)
          exit!(0) # skips exit handlers.
        end

        write.close
        result = read.read
        Process.wait(pid)
        raise "child failed" if result.empty?
        Marshal.load(result)
      end

      def self.load_gemspec_data(gemspec_path)
        gemspec_data = nil
        Bundler.with_clean_env do
          gemspec_data = self.execute_in_child do
            Gem::Specification::load(gemspec_path)
          end
        end
        if !gemspec_data
          fail CapsuleCD::Error::BuildPackageInvalid, '*.gemspec could not be parsed.'
        end
        return gemspec_data
      end


    end
  end
end
