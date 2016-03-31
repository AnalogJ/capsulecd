require 'spec_helper'

describe 'CapsuleCD::Python::PythonEngine', :python do
  describe '#build_step' do
    describe 'when building an empty package' do
      let(:engine) do
        require 'capsulecd/python/python_engine'
        CapsuleCD::Python::PythonEngine.new(source: :github,
                                            runner: :circleci,
                                            package_type: :python)
      end
      it 'should raise an error' do
        engine.instance_variable_set(:@source_git_local_path, test_directory )

        expect { engine.build_step }.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

    describe 'when building a simple package ' do
      let(:engine) do
        require 'capsulecd/python/python_engine'
        CapsuleCD::Python::PythonEngine.new(source: :github,
                                            runner: :circleci,
                                            package_type: :python)
      end
      it 'should create a VERSION file, requirements.txt file and tests folder' do
        FileUtils.copy_entry('spec/fixtures/python/pip_analogj_test', test_directory)
        engine.instance_variable_set(:@source_git_local_path, test_directory)

        VCR.use_cassette('pip_build_step',:tag => :chef) do
          engine.build_step
        end

        File.exist?(test_directory+'/VERSION')
        File.exist?(test_directory+'/requirements.txt')
        File.exist?(test_directory+'/.gitignore')
      end
    end
  end

  describe '#test_step' do
    before(:each) do
      FileUtils.copy_entry('spec/fixtures/python/pip_analogj_test', test_directory)
    end
    let(:engine) do
      require 'capsulecd/python/python_engine'
      CapsuleCD::Python::PythonEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :python)
    end
    let(:config_double) { CapsuleCD::Configuration.new }
    describe 'when testing python package' do
      it 'should run install dependencies' do
        allow(Open3).to receive(:popen3).and_return(false)
        allow(config_double).to receive(:engine_cmd_test).and_call_original
        allow(config_double).to receive(:engine_disable_test).and_call_original

        engine.instance_variable_set(:@source_git_local_path, test_directory)
        engine.instance_variable_set(:@config, config_double)

        engine.test_step

        File.exist?(test_directory+'/VERSION')
        File.exist?(test_directory+'/requirements.txt')
      end
    end
  end

  describe 'integration tests' do
    let(:engine) do
      require 'capsulecd/python/python_engine'
      CapsuleCD::Python::PythonEngine.new(source: :github,
                                          runner: :default,
                                          package_type: :python,
                                          config_file: 'spec/fixtures/sample_python_configuration.yml'
                                          # config_file: 'spec/fixtures/live_python_configuration.yml'
      )
    end
    let(:git_commit_double) { instance_double(Git::Object::Commit) }
    describe 'when testing python package' do
      it 'should complete successfully' do
        FileUtils.copy_entry('spec/fixtures/python/pip_analogj_test', test_directory)

        VCR.use_cassette('integration_python',:tag => :python) do
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
          allow(CapsuleCD::GitUtils).to receive(:create_gitignore).with(source_git_local_path, ['Python']).and_return(true)

          #stub methods in package_step
          allow(CapsuleCD::GitUtils).to receive(:commit).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:tag).with(source_git_local_path,'v1.0.7').and_return(git_commit_double)
          allow(git_commit_double).to receive(:sha).and_return('0a5948802a2bba02e019fd13bf3db3c5329faae6')
          allow(git_commit_double).to receive(:name).and_return('v1.0.7')

          #stub methods in release_step
          allow(Open3).to receive(:popen3).with('python setup.py sdist upload',{:chdir=>source_git_local_path}).and_return(true)
          allow(File).to receive(:open).with(File.expand_path('~/.pypirc'), 'w+').and_return(true)

          #stub methods in source_release
          allow(CapsuleCD::GitUtils).to receive(:push).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:generate_changelog).and_return('')

          engine.start

        end

      end
    end
  end

end
