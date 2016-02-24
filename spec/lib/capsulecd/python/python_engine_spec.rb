require 'spec_helper'
require 'capsulecd/python/python_engine'

describe CapsuleCD::Python::PythonEngine, :python do
  describe '#build_step' do
    let(:engine) do
      CapsuleCD::Python::PythonEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :python)
    end
    describe 'when building an empty package ' do
      it 'should create a VERSION file, requirements.txt file and tests folder' do
        engine.instance_variable_set(:@source_git_local_path, 'spec/fixtures/empty_package')

        engine.build_step

        File.exist?('spec/fixtures/empty_package/VERSION')
        File.exist?('spec/fixtures/empty_package/requirements.txt')

        FileUtils.rm_rf(Dir['spec/fixtures/empty_package/[^.]*'])
      end
    end
  end

  describe '#test_step' do
    let(:engine) do
      CapsuleCD::Python::PythonEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :python)
    end
    describe 'when testing python package' do
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
      CapsuleCD::Python::PythonEngine.new(source: :github,
                                          runner: :circleci,
                                          package_type: :python,
                                          config_file: 'spec/fixtures/live_python_configuration.yml'
      )
    end
    let(:git_commit_double) { instance_double(Git::Object::Commit) }
    describe 'when testing python package' do
      it 'should complete successfully' do

        VCR.use_cassette('integration_circleci_python',:tag => :python) do
          #stub methods in source_process_pull_request_payload
          allow(CapsuleCD::GitUtils).to receive(:clone).and_return(engine.config.source_git_parent_path + '/pip_analogj_test')
          allow(CapsuleCD::GitUtils).to receive(:fetch).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:checkout).and_return(true)

          #stub methods in build_step
          allow(File).to receive(:open).with('spec/fixtures/python/pip_analogj_test/VERSION','w').and_call_original

          #stub methods in test_step
          allow(Open3).to receive(:popen3).with('pip install -r requirements.txt', {:chdir=>'spec/fixtures/python/pip_analogj_test'}).and_call_original
          allow(Open3).to receive(:popen3).with('pip install -e .', {:chdir=>'spec/fixtures/python/pip_analogj_test'}).and_call_original
          allow(Open3).to receive(:popen3).with('python setup.py test', {:chdir=>'spec/fixtures/python/pip_analogj_test'}).and_call_original


          #stub methods in package_step
          allow(CapsuleCD::GitUtils).to receive(:commit).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:tag).and_return(git_commit_double)
          allow(git_commit_double).to receive(:sha).and_return('0a5948802a2bba02e019fd13bf3db3c5329faae6')
          allow(git_commit_double).to receive(:name).and_return('v1.0.0')

          #stub methods in release_step
          allow(Open3).to receive(:popen3).with('python setup.py sdist upload',{:chdir=>'spec/fixtures/python/pip_analogj_test'}).and_return(true)
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
