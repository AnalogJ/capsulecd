require 'yaml'
require 'erb'
require 'base64'
module CapsuleCD
  class Configuration
    # Order of inheritance: file <- environment <- cli options
    # <- means 'overridden by', eg. file overridden by environment vars
    # @param [String] config_path The path to the configuration file
    def initialize(options={})
      @options = options
      @config_path = @options[:config_file]

      populate_defaults
      populate_system_config_file
      populate_runner_overrides
      populate_env_overrides
      populate_cli_overrides
      standardize_settings
    end

    # Cli config, shouldnt be set via environmental variables (will be overridden)
    attr_reader :package_type
    attr_reader :source
    attr_reader :runner
    attr_reader :dry_run

    # General config
    attr_reader :config_path
    attr_reader :configuration

    # Source config (any credentials added here should also be added to the spec_helper.rb VCR config)
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

    # Package auth/config (any credentials added here should also be added to the spec_helper.rb VCR config)
    attr_reader :chef_supermarket_username
    attr_reader :npm_auth_token
    attr_reader :pypi_username
    attr_reader :pypi_password
    attr_reader :chef_supermarket_type
    attr_reader :rubygems_api_key
    def chef_supermarket_key
      @chef_supermarket_key.to_s.empty? ? nil : Base64.strict_decode64(@chef_supermarket_key)
    end

    # Engine config
    attr_reader :engine_disable_test
    attr_reader :engine_disable_minification
    attr_reader :engine_disable_lint
    attr_reader :engine_disable_coverage
    attr_reader :engine_cmd_test
    attr_reader :engine_cmd_minification
    attr_reader :engine_cmd_lint
    attr_reader :engine_cmd_coverage
    attr_reader :engine_version_bump_type

    def populate_repo_config_file(repo_local_path)
      repo_config_file_path = repo_local_path + '/capsule.yml'
      load_config_file(repo_config_file_path)
      populate_runner_overrides
      populate_env_overrides
      populate_cli_overrides
      standardize_settings
    end

    # The raw parsed configuration file, system level, a repo level configuration file will override settings in this file.
    def populate_system_config_file
      load_config_file(@config_path)
    end

    private
    # These are defaults for engine settings. They can be overridden via configuration files or env variables
    def populate_defaults
      @engine_version_bump_type = :patch # can be :major, :minor, :patch
      @chef_supermarket_type = 'Other'
    end

    def load_config_file(path)
      if !path || !File.exist?(path)
        puts 'The configuration file could not be found. Using defaults'
        return
      end
      puts 'Loading configuration file: ' + path
      file = File.open(path).read
      unserialize(file)
    end


    def populate_runner_overrides
      # @runner = :circleci unless ENV['CIRCLECI'].to_s.empty?
      populate_runner
    end

    def populate_runner
      # if (@runner == :circleci)
      #   # parse the PR# from the environment variable, eg. https://github.com/AnalogJ/cookbook_analogj_test/pull/9
      #   @runner_pull_request ||= File.basename(URI.parse(ENV['CI_PULL_REQUEST']).path).to_i # => baz
      #   @runner_sha ||= ENV['CIRCLE_SHA1']
      #   @runner_branch ||= ENV['CIRCLE_BRANCH']
      #   @runner_clone_url ||= 'https://github.com/' + ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'] + '.git'
      #   @runner_repo_name ||= ENV['CIRCLE_PROJECT_REPONAME']
      #   @runner_repo_full_name ||= ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME']
      # end
    end

    def populate_env_overrides
      # override config file with env variables.
      ENV.each do|key, value|
        config_key = key.dup
        if config_key.start_with?('CAPSULE_') && !value.to_s.empty?
          config_key.slice!('CAPSULE_')
          config_key.downcase!

          # override instance variable
          instance_variable_set('@' + config_key, value)
        end
      end
    end

    def populate_cli_overrides
      # then override with cli options
      @options.each do|key, value|
        instance_variable_set('@' + key.to_s, value)
      end
    end

    # certain settings are symbols, so make sure that any settings that are specified via a string are converted to the correct type.
    def standardize_settings
      # set types if missing
      @engine_version_bump_type = @engine_version_bump_type.to_sym if @engine_version_bump_type.is_a? String
    end

    def unserialize(string)
      obj = YAML.load(string)
      obj.keys.each do |key|
        next if %w(source_configure source_process_pull_request_payload source_process_push_payload runner_retrieve_payload build_step test_step package_step source_release release_step).include?(key)
        instance_variable_set('@' + key, obj[key])
      end
      obj
    end
  end
end
