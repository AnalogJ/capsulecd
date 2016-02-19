require 'spec_helper'
require 'capsulecd/cli'
require 'capsulecd/base/engine'
require 'capsulecd/node/node_engine'

describe CapsuleCD::Cli do

  # before(:each) do
  #   @pwd = Dir.pwd
  #   Dir.chdir(test_directory)
  # end

  describe '#start' do
    let(:command_double) { instance_double(NodeEngine, start: true) }
    it 'should call the node start command with the proper options' do
      expect(NodeEngine).to receive(:new).with(
        {:runner=>:none, :source=>:github, :package_type=>:node}).and_return command_double

      CapsuleCD::Cli.start %w[start --package_type node]
    end

    it 'should call the default start command with the proper options' do
      expect(Engine).to receive(:new).with(
                                {:runner=>:none, :source=>:github, :package_type=>:general}).and_return command_double

      CapsuleCD::Cli.start %w[start]
    end
  end
end