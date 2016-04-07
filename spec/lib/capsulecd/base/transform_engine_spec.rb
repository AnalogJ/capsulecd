require 'spec_helper'

describe CapsuleCD::TransformEngine do
  before(:each) do
    CapsuleCD.send(:remove_const, 'EngineExtension')
    load 'lib/capsulecd/base/transform_engine.rb'
  end
  describe '::new' do
    describe 'with a global configuration file' do

      describe 'with an empty engine' do
        before(:each) do
          stub_const 'EmptyEngine', Class.new
        end

        let(:config_file_path) { 'spec/fixtures/sample_global_configuration.yml' }
        subject { CapsuleCD::TransformEngine.new() }

        it 'should raise an error if we specify that its a repo configuration' do
          expect{
            subject.transform(nil, config_file_path)
          }.to raise_error(CapsuleCD::Error::EngineTransformUnavailableStep)
        end

        it 'should register the EngineExtension methods after transform' do
          engine = EmptyEngine.new
          expect(engine.methods - Object.methods).to be_empty
          subject.transform(engine, config_file_path, :global)
          expect((engine.methods - Object.methods).sort).to eql([:post_build_step, :post_source_configure, :pre_source_configure, :source_configure])
          expect(engine.source_configure).to eql('override source_configure')
        end
      end

      describe 'with an base engine' do
        before(:each) do
          stub_const 'BaseEngine', CapsuleCD::Engine
        end

        let(:config_file_path) { 'spec/fixtures/sample_global_configuration.yml' }
        subject { CapsuleCD::TransformEngine.new() }


        it 'should register the EngineExtension methods after transform, and override any existing methods' do
          engine = BaseEngine.new(source: :github)
          engine.instance_variable_set(:@source_git_local_path, test_directory)
          subject.transform(engine, config_file_path, :global)
          expect(engine.source_configure).to eql('override source_configure')
          expect(engine.post_build_step).to eql('override post_build_step'+test_directory)
        end
      end

    end

  end
end
