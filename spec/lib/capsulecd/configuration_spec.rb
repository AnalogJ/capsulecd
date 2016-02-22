require 'spec_helper'

describe CapsuleCD::Configuration do
  describe '::new' do
    describe 'with a sample configuration file' do
      let(:config_file_path) { 'spec/fixtures/sample_configuration.yml' }
      subject { CapsuleCD::Configuration.new(config_file:config_file_path, runner: :circleci, source: :github, package_type: :node) }

      it 'should populate github_access_token' do
        expect(subject.source_github_access_token).to eql('sample_test_token')
      end

      it 'should populate engine_npm_auth_token' do
        expect(subject.npm_auth_token).to eql('sample_auth_token')
      end

      it 'should use cli specified source' do
        expect(subject.source).to eql(:github)
      end

      it 'should have correct defaults' do
        expect(subject.engine_version_bump_type).to eql(:patch)
        expect(subject.chef_supermarket_type).to eql('Other')
        expect(subject.chef_supermarket_key).to eql(nil)
      end
    end

    describe 'with an incorrect configuration file' do
      let(:config_file_path) { 'spec/fixtures/sample_configuration.yml' }
      subject { CapsuleCD::Configuration.new(config_file:config_file_path, runner: :circleci, source: :github, package_type: :node) }

      it 'should populate github_access_token' do
        expect(subject.source_github_access_token).to eql('sample_test_token')
      end

      it 'should populate engine_npm_auth_token' do
        expect(subject.npm_auth_token).to eql('sample_auth_token')
      end

      it 'should use cli specified source' do
        expect(subject.source).to eql(:github)
      end
    end

    describe 'when overriding sample configuration file' do
      let(:config_file_path) { 'spec/fixtures/sample_configuration.yml' }

      it 'should use CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN instead of config file' do
        allow(ENV).to receive(:each).and_yield('CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN', 'override_test_token')

        config = CapsuleCD::Configuration.new(config_file: config_file_path, runner: :circleci, source: :github, package_type: :node)

        expect(config.source_github_access_token).to eql('override_test_token')
      end
    end
  end
end
