require 'spec_helper'

describe CapsuleCD::Configuration do
  describe '::new' do
    describe 'with a sample configuration file' do
      let(:config_file_path) { 'spec/fixtures/sample_configuration.yml' }
      subject { CapsuleCD::Configuration.new(config_file:config_file_path, source: :github, package_type: :node) }

      it 'should populate github_access_token' do
        expect(subject.source_github_access_token).to eql('sample_test_token')
      end

      it 'should populate engine_npm_auth_token' do
        expect(subject.npm_auth_token).to eql('sample_auth_token')
      end

      it 'should use cli specified source' do
        expect(subject.source).to eql(:github)
      end

      it 'should correctly base64 decode chef supermarket key' do
        expect(subject.chef_supermarket_key).to eql("-----BEGIN RSA PRIVATE KEY-----\nsample_supermarket_key\n-----END RSA PRIVATE KEY-----\n")
      end

      it 'should have correct defaults' do
        expect(subject.engine_version_bump_type).to eql(:patch)
        expect(subject.chef_supermarket_type).to eql('Other')
      end
    end

    describe 'with an incorrect configuration file' do
      let(:config_file_path) { 'spec/fixtures/incorrect_configuration.yml' }
      subject { CapsuleCD::Configuration.new(config_file:config_file_path, source: :github, package_type: :node) }

      it 'should populate github_access_token' do
        expect(subject.source_github_access_token).to eql('sample_test_token')
      end

      it 'should populate engine_npm_auth_token' do
        expect(subject.npm_auth_token).to eql('sample_auth_token')
      end

      it 'should use cli specified source' do
        expect(subject.source).to eql(:github)
      end

      it 'should have a nil chef supermarket key' do
        expect(subject.chef_supermarket_key).to eql(nil)
      end
    end

    describe 'when overriding sample configuration file' do
      let(:config_file_path) { 'spec/fixtures/sample_configuration.yml' }

      describe 'with ENV variables' do
        it 'should use CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN instead of config file' do
          allow(ENV).to receive(:each).and_yield('CAPSULE_SOURCE_GITHUB_ACCESS_TOKEN', 'override_test_token')

          config = CapsuleCD::Configuration.new(config_file: config_file_path, source: :github, package_type: :node)

          expect(config.source_github_access_token).to eql('override_test_token')
        end

        it 'should use CAPSULE_ENGINE_VERSION_BUMP_TYPE instead of default and return symbol' do
          allow(ENV).to receive(:each).and_yield('CAPSULE_ENGINE_VERSION_BUMP_TYPE', 'patch')

          config = CapsuleCD::Configuration.new(config_file: config_file_path, source: :github, package_type: :node)

          expect(config.engine_version_bump_type).to eql(:patch)
        end
      end
    end

  end
end
