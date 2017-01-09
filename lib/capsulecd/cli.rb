require 'thor'
require 'capsulecd'
require 'pp'

module CapsuleCD
  # The command line interface for CapsuleCD.
  class Cli < Thor

    desc 'version', 'Print out the CapsuleCD version'
    def version
      require 'capsulecd/version'
      puts CapsuleCD::VERSION
    end

    ##
    # Run
    ##
    desc 'start', 'Start a new CapsuleCD package pipeline '
    option :runner,
           type: :string,
           default: 'default', # can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
           desc: 'The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)'

    option :source,
           type: :string,
           default: 'default',
           desc: 'The source for the code, used to determine which git endpoint to clone from, and create releases on'

    option :package_type,
           type: :string,
           default: 'default',
           desc: 'The type of package being built.'

    option :dry_run,
           type: :boolean,
           default: false,
           desc: 'Specifies that no changes should be pushed to source and no package will be released'

    option :config_file,
           type: :string,
           default: nil,
           desc: 'Specifies the location of the config file'

    # Begin processing
    def start
      # parse runner from env
      engine_opts = {}

      engine_opts[:runner] = options[:runner].to_sym # TODO: we cant modify the hash sent by Thor, so we'll duplicate it
      engine_opts[:source] = options[:source].to_sym
      engine_opts[:package_type] = options[:package_type].to_sym
      engine_opts[:dry_run] = options[:dry_run]
      puts '###########################################################################################'
      puts '# Configuration '
      puts '###########################################################################################'
      pp engine_opts

      if engine_opts[:package_type] == :default

        engine = CapsuleCD::Engine.new(engine_opts)
      else
        package_type = engine_opts[:package_type].to_s
        require_relative "#{package_type}/#{package_type}_engine"
        engine = CapsuleCD.const_get(package_type.capitalize).const_get("#{package_type.capitalize}Engine").new(engine_opts)
      end

      engine.start
    end
  end
end
