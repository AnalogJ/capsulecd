#!/usr/bin/env ruby

require 'yaml'
require 'optparse'
require 'erb'
require 'pp'
require_relative 'node/node_engine'
require_relative 'chef/chef_engine'

# This script will kick off the pull request CD process.

# CLI Parser
class OptsParse

  def self.parse(args)
    options = {
        :runner => :none, # can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
        :source => :github,
        :type => :general, #can be node, chef, ruby, python, crate, puppet,  general, (more to come)
    }

    opts = OptionParser.new do |opts|
      opts.banner = ' Usage: start.rb [options]'
      opts.separator ''
      opts.separator 'Specific options'



      opts.on('--runner runner',
                'The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)') do |runner|
        options[:runner] = runner.to_sym
      end

      opts.on('--source source',
              'The source for the code, used to determine which git endpoint to clone from, and create releases on') do |source|
        options[:source] = source.to_sym
      end

      opts.on('--type cd_type',
              'The type of package being built.') do |cd_type|
        options[:type] = cd_type.to_sym
      end
    end
    opts.parse!(args)
    options
  end
end

options = OptsParse.parse(ARGV)
puts '###########################################################################################'
puts '# Configuration '
puts '###########################################################################################'
pp options

if options[:type] == :node
  engine = NodeEngine.new(options)
elsif options[:type] == :chef
  engine = ChefEngine.new(options)
end

engine.start