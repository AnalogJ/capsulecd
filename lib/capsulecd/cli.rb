require 'thor'
require 'capsulecd'
require 'pp'

module CapsuleCD
  # The command line interface for CapsuleCD.
  class Cli < Thor

    DEFAULT_INVENTORY_CONFIG    = './inventory.yml'
    DEFAULT_INVENTORY_DIRECTORY = './inventory'
    DEFAULT_WEB_DIRECTORY       = './web'

    ##
    # Run
    ##
    desc 'start', 'Start a new CapsuleCD package pipeline '
    option :runner,
      :type => :string,
      :default => 'none', #can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
      :desc => 'The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)'

    option :source,
           :type => :string,
           :default => 'github',
           :desc => 'The source for the code, used to determine which git endpoint to clone from, and create releases on'

    option :package_type,
           :type => :string,
           :default => 'general',
           :desc => 'The type of package being built.'

    # Begin processing
    def start()
      # parse runner from env
      engine_opts = {}
      engine_opts[:runner] = :circleci if ENV['CIRCLECI']

      engine_opts[:runner] = options[:runner].to_sym
      engine_opts[:source] = options[:source].to_sym
      engine_opts[:package_type] = options[:package_type].to_sym


      puts '###########################################################################################'
      puts '# Configuration '
      puts '###########################################################################################'
      pp engine_opts

      if engine_opts[:package_type] == :node
        require_relative 'node/node_engine'
        engine = NodeEngine.new(engine_opts)
      elsif engine_opts[:package_type] == :chef
        require_relative 'chef/chef_engine'
        engine = ChefEngine.new(engine_opts)
      elsif engine_opts[:package_type] == :python
        require_relative 'python/python_engine'
        engine = PythonEngine.new(engine_opts)
      else
        engine = Engine.new(engine_opts)
      end

      engine.start
    end

  end
end
