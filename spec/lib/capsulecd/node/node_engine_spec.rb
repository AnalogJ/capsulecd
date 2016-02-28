require 'spec_helper'

describe 'CapsuleCD::Node::NodeEngine', :node do
  describe '#build_step' do
    describe 'when building an empty package' do
      let(:engine) do
        require 'capsulecd/node/node_engine'
        CapsuleCD::Node::NodeEngine.new(source: :github,
                                        runner: :circleci,
                                        package_type: :node)
      end
      it 'should raise an error' do
        engine.instance_variable_set(:@source_git_local_path, test_directory )

        expect { engine.build_step }.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

    describe 'when building a simple package ' do
      let(:engine) do
        require 'capsulecd/node/node_engine'
        CapsuleCD::Node::NodeEngine.new(source: :github,
                                        runner: :circleci,
                                        package_type: :node)
      end
      it 'should create a .gitignore file and tests folder' do
        FileUtils.copy_entry('spec/fixtures/node/npm_analogj_test', test_directory)
        engine.instance_variable_set(:@source_git_local_path, test_directory)

        VCR.use_cassette('node_build_step',:tag => :chef) do
          engine.build_step
        end

        File.exist?(test_directory+'/.gitignore')
      end
    end
  end

  describe '#test_step' do
    let(:engine) do
      require 'capsulecd/node/node_engine'
      CapsuleCD::Node::NodeEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :node)
    end
    describe 'when testing node package' do
      it 'should run install dependencies' do
        FileUtils.copy_entry('spec/fixtures/node/npm_analogj_test', test_directory)
        allow(Open3).to receive(:popen3).and_return(false)
        engine.instance_variable_set(:@source_git_local_path, test_directory)

        engine.test_step

        File.exist?(test_directory+'/npm-shrinkwrap.json')
      end
    end
  end

  describe 'integration tests' do
    let(:engine) do
      require 'capsulecd/node/node_engine'
      CapsuleCD::Node::NodeEngine.new(source: :github,
                                          runner: :default,
                                          package_type: :node,
                                          config_file: 'spec/fixtures/sample_node_configuration.yml'
                                          # config_file: 'spec/fixtures/live_node_configuration.yml'
      )
    end
    let(:git_commit_double) { instance_double(Git::Object::Commit) }
    describe 'when testing node package' do
      it 'should complete successfully' do
        FileUtils.copy_entry('spec/fixtures/node/npm_analogj_test', test_directory)

        VCR.use_cassette('integration_node',:tag => :node) do
          #set defaults for stubbed classes
          source_git_local_path = test_directory
          allow(File).to receive(:exist?).and_call_original
          allow(File).to receive(:open).and_call_original
          allow(Open3).to receive(:popen3).and_call_original

          #stub methods in source_process_pull_request_payload
          allow(CapsuleCD::GitUtils).to receive(:clone).and_return(source_git_local_path)
          allow(CapsuleCD::GitUtils).to receive(:fetch).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:checkout).and_return(true)

          #stub methods in build_step
          allow(CapsuleCD::GitUtils).to receive(:create_gitignore).with(source_git_local_path, ['Node']).and_return(true)

          #stub methods in package_step
          allow(CapsuleCD::GitUtils).to receive(:commit).and_return(true)

          allow(CapsuleCD::GitUtils).to receive(:get_latest_tag_commit).and_return(git_commit_double)
          allow(git_commit_double).to receive(:sha).and_return('0a5948802a2bba02e019fd13bf3db3c5329faae6')
          allow(git_commit_double).to receive(:name).and_return('v1.0.8')

          #stub methods in release_step
          allow(Open3).to receive(:popen3).with('npm publish .',{:chdir=>source_git_local_path}).and_return(true)
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
