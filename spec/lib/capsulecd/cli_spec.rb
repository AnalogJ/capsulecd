require 'spec_helper'
require 'rspec/support/spec/in_sub_process'
require 'capsulecd/cli'
require 'capsulecd/base/engine'

describe CapsuleCD::Cli do
  describe '#start' do
    describe 'with default_engine' do
      let(:engine_double) { instance_double(Engine, start: true) }
      it 'should call the default start command with the proper options' do
        expect(Engine).to receive(:new).with(
          runner: :default, source: :default, package_type: :default,:dry_run=>false).and_return engine_double

        CapsuleCD::Cli.start %w(start)
      end
    end

    describe 'with node_engine', :node do
      require 'capsulecd/node/node_engine'
      let(:engine_double) { instance_double(NodeEngine, start: true) }
      it 'should call the node start command with the proper options' do
        expect(NodeEngine).to receive(:new).with(
          runner: :default, source: :default, package_type: :node,:dry_run=>false).and_return engine_double

        CapsuleCD::Cli.start %w(start--package_type node)
      end
    end

    describe 'with chef_engine', :chef do
      let(:engine_double) do
        require 'capsulecd/chef/chef_engine'
        instance_double(ChefEngine, start: true)
      end
      it 'should call the chef start command with the proper options' do
        require 'capsulecd/chef/chef_engine'
        expect(ChefEngine).to receive(:new).with(
          runner: :default, source: :default, package_type: :chef,:dry_run=>false).and_return engine_double

        CapsuleCD::Cli.start %w(start --package_type chef)
      end
    end

    describe 'with python_engine', :python do
      let(:engine_double) do
        require 'capsulecd/python/python_engine'
        instance_double(PythonEngine, start: true)
      end
      it 'should call the python start command with the proper options' do
        require 'capsulecd/python/python_engine'
        expect(PythonEngine).to receive(:new).with(
                                  runner: :default, source: :default, package_type: :python,:dry_run=>false).and_return engine_double

        CapsuleCD::Cli.start %w(start --package_type python)
      end
    end
  end
end
