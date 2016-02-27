require 'spec_helper'
require 'capsulecd/node/node_engine'

describe CapsuleCD::Node::NodeEngine, :node do
  describe '#build_step' do
    let(:engine) do
      CapsuleCD::Node::NodeEngine.new(source: :github,
                                      runner: :circleci,
                                      package_type: :node)
    end
    describe 'when building an empty package ' do
      it 'should create a .gitignore file and tests folder' do
        engine.instance_variable_set(:@source_git_local_path, 'spec/fixtures/empty_package')

        engine.build_step

        File.exist?('spec/fixtures/empty_package/.gitignore')

        FileUtils.rm_rf(Dir['spec/fixtures/empty_package/[^.]*'])
      end
    end
  end

  describe '#test_step' do
    let(:engine) do
      CapsuleCD::Node::NodeEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :node)
    end
    describe 'when testing node package' do
      it 'should run install dependencies' do
        allow(Open3).to receive(:popen3).and_return(false)

        engine.instance_variable_set(:@source_git_local_path, 'spec/fixtures/empty_package')

        engine.test_step

        File.exist?('spec/fixtures/empty_package/VERSION')
        File.exist?('spec/fixtures/empty_package/requirements.txt')

        FileUtils.rm_rf(Dir['spec/fixtures/empty_package/[^.]*'])
      end
    end
  end

  describe 'integration tests' do
    let(:engine) do
      CapsuleCD::Node::NodeEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :node,
                                          config_file: 'spec/fixtures/live_node_configuration.yml'
      )
    end
    let(:git_commit_double) { instance_double(Git::Object::Commit) }
    describe 'when testing node package' do
      it 'should complete successfully' do

        VCR.use_cassette('integration_node',:tag => :node) do
          #stub methods in source_process_pull_request_payload
          allow(CapsuleCD::GitUtils).to receive(:clone).and_return(engine.config.source_git_parent_path + '/npm_analogj_test')
          allow(CapsuleCD::GitUtils).to receive(:fetch).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:checkout).and_return(true)

          #stub methods in build_step
          allow(CapsuleCD::GitUtils).to receive(:create_gitignore).with(engine.config.source_git_parent_path + '/npm_analogj_test', ['Node']).and_return(true)

          #stub methods in test_step
          allow(Open3).to receive(:popen3).with('npm install', {:chdir=>'spec/fixtures/node/npm_analogj_test'}).and_call_original
          allow(Open3).to receive(:popen3).with('npm shrinkwrap', {:chdir=>'spec/fixtures/node/npm_analogj_test'}).and_call_original
          allow(Open3).to receive(:popen3).with(ENV, 'npm test', {:chdir=>'spec/fixtures/node/npm_analogj_test'}).and_call_original


          #stub methods in package_step
          allow(CapsuleCD::GitUtils).to receive(:commit).and_return(true)
          allow(Open3).to receive(:popen3).with('npm version patch -m "(v%s) Automated packaging of release by CapsuleCD"', {:chdir=>'spec/fixtures/node/npm_analogj_test'}).and_call_original

          allow(CapsuleCD::GitUtils).to receive(:get_latest_tag_commit).and_return(git_commit_double)
          allow(git_commit_double).to receive(:sha).and_return('0a5948802a2bba02e019fd13bf3db3c5329faae6')
          allow(git_commit_double).to receive(:name).and_return('v1.0.8')

          #stub methods in release_step
          allow(Open3).to receive(:popen3).with('npm publish .',{:chdir=>'spec/fixtures/node/npm_analogj_test'}).and_return(true)
          allow(File).to receive(:open).with(File.expand_path('~/.npmrc'), 'w+').and_return(true)

          #stub methods in source_release
          allow(CapsuleCD::GitUtils).to receive(:push).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:generate_changelog).and_return('')

          engine.start

        end

      end
    end
  end

end
