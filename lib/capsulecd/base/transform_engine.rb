require 'yaml'
module CapsuleCD
  module EngineExtension
  end

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
      populate_engine_extension
      register_extension
    end

    def populate_engine_extension()
      # lets loop though all the keys, and raise an error if the hooks specified are not available.
      if @type == :repo
        @config.keys.each do |key|
           fail CapsuleCD::Error::EngineTransformUnavailableStep, key + ' cannot be overridden by repo capsule.yml file.' if %w(source_configure source_process_pull_request_payload source_process_push_payload runner_retrieve_payload).include?(key)
        end
      end

      # yeah yeah, metaprogramming is evil. But actually its really just a tool, and like any tool, you can use it incorrectly.
      # In general metaprogramming is bad because it makes it hard to reason about your code. In this case we're using it
      # to allow other developers to override our engine steps, and/or attach to our hooks.

      # http://www.monkeyandcrow.com/blog/building_classes_dynamically/
      # https://rubymonk.com/learning/books/5-metaprogramming-ruby-ascent/chapters/24-eval/lessons/68-class-eval
      # https://www.ruby-forum.com/topic/207350
      @config.each do |step, value|
        next unless %w(source_configure source_process_pull_request_payload source_process_push_payload runner_retrieve_payload build_step test_step package_step source_release release_step).include?(step)

        value.each do |prefix, method_script|
          EngineExtension.class_eval(<<-METHOD
          def #{prefix == 'override' ? '' : prefix+ '_'}#{step};
            #{method_script};
          end
METHOD
)
        end
      end
    end

    # at this point the EngineExtension module should be populated with all the hooks and methods.
    # now we need to add them to the engine.
    def register_extension
      @engine.class.prepend(CapsuleCD::EngineExtension)
    end
  end
end
