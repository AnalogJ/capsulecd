require 'rubygems'
require 'thread'
require 'bundler'
module CapsuleCD
  module Ruby
    class RubyHelper
      PRERELEASE    = ["alpha","beta","rc",nil]
      VERSION_REGEX = /(\d+\.\d+\.\d+(?:-(?:#{PRERELEASE.compact.join('|')}))?)/
      REPLACE_VERSION_REGEX =

      # get gemspec file path (used for `gem build ..` command)
      def self.get_gemspec_path(repo_path)
        gemspecs = Dir.glob(repo_path + '/*.gemspec')
        if gemspecs.empty?
          fail CapsuleCD::Error::BuildPackageInvalid, '*.gemspec file is required to process Ruby gem'
        end
        gemspecs.first
      end

      # get gem name
      def self.get_gem_name(repo_path)
        self.load_gemspec_data(self.get_gemspec_path(repo_path)).name
      end

      # get gem version
      def self.get_version(repo_path)
        gem_version_file = self.find_version_file(repo_path)
        gem_version = File.read(gem_version_file)[VERSION_REGEX]
        if !gem_version
          fail CapsuleCD::Error::BuildPackageInvalid, 'version.rb file is invalid'
        end
        return gem_version
      end

      # set gem version
      def self.set_version(repo_path, next_version)
        gem_version_file = self.find_version_file(repo_path)
        gem_version_file_content = File.read(gem_version_file)

        next_gem_version_file_content = gem_version_file_content.gsub(/(VERSION\s*=\s*['"])[0-9\.]+(['"])/, "\\1#{next_version.to_s}\\2")

        File.open(gem_version_file, 'w') { |file| file.write(next_gem_version_file_content) }
      end




      private
      ##################################################################################################################
      # NEW protected/private methods.
      # based on bump gem methods: https://github.com/gregorym/bump/blob/master/lib/bump.rb

      def self.find_version_file(repo_path)
        files = Dir.glob("#{repo_path}/lib/**/version.rb")
        if files.size == 0
          fail CapsuleCD::Error::BuildPackageInvalid, 'version.rb file is required to process Ruby gem'
        elsif files.size == 1
          return files.first
        else
          # too many version.rb files found, lets try to find the correct version.rb file using gem name in gemspec
          return self.version_filepath(repo_path, self.get_gem_name(repo_path))
        end
      end

      def self.version_filepath(repo_path, gem_name, version_filename = 'version.rb')
        "#{repo_path}/lib/#{gem_name}/#{version_filename}"
      end


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
            # reload the version.rb file if found (fixes dogfooding issue)
            # Dir["#{File.dirname(gemspec_path)}/**/version.rb"].each { |f| load(f) }

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
