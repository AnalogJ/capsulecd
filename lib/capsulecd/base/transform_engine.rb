require 'yaml'
module CapsuleCD
  class TransformEngine
    def initialize()
    end

    def transform(engine, config_file, type = :repo) #type can only be :repo or :global
      @engine = engine
      @type = type

      unless File.exists?(config_file)
        puts 'no configuration file found, no engine hooks'
        return
      end

      #parse the config file and generate a module file that we can use as part of our engine
      @config = YAML.load(File.open(config_file).read)
      generate_module
    end

    private
    def generate_module()
      # lets loop though all the keys, and raise an error if the hooks specified are not available.
      if @type == :repo
        @config.keys.each do |key|
           fail CapsuleCD::Error::EngineTransformUnavailableStep, key + ' cannot be overridden by repo capsule.yml file.' if %w(source_configure source_process_pull_request_payload source_process_push_payload runner_retrieve_payload).include?(key)
        end
      end

    end
  end
end
