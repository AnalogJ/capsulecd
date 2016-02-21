require 'yaml'
require 'erb'
require 'base64'
module CapsuleCD
  class Configuration

    # Order of inheritance: file <- environment <- cli options
    # <- means 'overridden by', eg. file overridden by environment vars
    # @param [String] config_path The path to the configuration file
    def initialize(config_path, options)
      @options = options
      @config_path = config_path
      @configuration = parse_config_file
      detect_runner_and_populate
      populate_overrides
    end

    # Cli config, shouldnt be set via environmental variables (will be overridden)
    attr_reader :package_type
    attr_reader :source
    attr_reader :runner
    attr_reader :dry_run

    # General config
    attr_reader :config_path
    attr_reader :configuration

    # Source config
    attr_reader :source_git_parent_path
    attr_reader :source_github_api_endpoint
    attr_reader :source_github_web_endpoint
    attr_reader :source_github_access_token

    # Runner config
    attr_reader :runner_pull_request
    attr_reader :runner_sha
    attr_reader :runner_branch
    attr_reader :runner_clone_url
    attr_reader :runner_repo_full_name
    attr_reader :runner_repo_name

    # Engine config
    attr_reader :chef_supermarket_username
    attr_reader :npm_auth_token
    attr_reader :pypi_username
    attr_reader :pypi_password


    def engine_version_bump_type
      @engine_version_bump_type ||= :patch #can be :major, :minor, :patch
    end


    def chef_supermarket_key
      @chef_supermarket_key.to_s.empty? ? nil : Base64.strict_decode64(@chef_supermarket_key)
    end

    def chef_supermarket_type
      @chef_supermarket_type ||= 'Other'
    end


    private

    # The raw parsed configuration file

    def parse_config_file
      unless File.exists?(@config_path)
        raise 'The configuration file could not be found. Using defaults'
      end

      file = File.open(@config_path).read
      unserialize(file)
    end

    def detect_runner_and_populate
      if !ENV['CIRCLECI'].to_s.empty?
        @runner = :circleci
      end
      populate_runner
    end

    def populate_runner()
      if(@runner == :circleci)
        @runner_pull_request ||= ENV['CI_PULL_REQUEST']
        @runner_sha ||= ENV['CIRCLE_SHA1']
        @runner_branch ||= ENV['CIRCLE_BRANCH']
        @runner_clone_url ||= 'https://github.com/' + ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'] + '.git'
        @runner_repo_name ||= ENV['CIRCLE_PROJECT_REPONAME']
        @runner_repo_full_name ||= ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME']
      end
    end

    def populate_overrides
      #override config file with env variables.
      ENV.each{|key,value|
        config_key = key.dup
        if config_key.start_with?('CAPSULE_') && !value.to_s.empty?
          config_key.slice!('CAPSULE_')
          config_key.downcase!

          #override instance variable
          instance_variable_set('@'+config_key, value)
        end
      }

      #then override with cli options
      @options.each{|key,value|
        instance_variable_set('@'+key.to_s, value)
      }

      #set types if missing
      @engine_version_bump_type = @engine_version_bump_type.to_sym if @engine_version_bump_type.is_a? String

    end


    def unserialize(string)
      obj = YAML.load(string)
      obj.keys.each do |key|
        instance_variable_set('@'+key, obj[key])
      end
      obj
    end

  end
end
