require 'spec_helper'
require 'capsulecd/python/python_engine'

describe PythonEngine, :python do
  describe '#build_step' do
    let(:engine) {
      PythonEngine.new({
        source: :github,
        runner: :circleci,
        package_type: :python
      })
    }
    describe 'when building an empty package ' do
      it 'should create a VERSION file, requirements.txt file and tests folder' do

        engine.instance_variable_set(:@source_git_local_path, 'spec/fixtures/empty_package')

        engine.build_step

        File.exists?('spec/fixtures/empty_package/VERSION')
        File.exists?('spec/fixtures/empty_package/requirements.txt')

        FileUtils.rm_rf(Dir["spec/fixtures/empty_package/[^.]*"])
      end
    end
  end

  describe '#test_step' do
    let(:engine) {
      PythonEngine.new({
                           source: :github,
                           runner: :circleci,
                           package_type: :python
                       })
    }
    describe 'when testing python package' do
      it 'should run install dependencies' do

        allow(Open3).to receive(:popen3).and_return(false)

        engine.instance_variable_set(:@source_git_local_path, 'spec/fixtures/empty_package')

        engine.test_step

        File.exists?('spec/fixtures/empty_package/VERSION')
        File.exists?('spec/fixtures/empty_package/requirements.txt')

        FileUtils.rm_rf(Dir["spec/fixtures/empty_package/[^.]*"])
      end
    end
  end

end
