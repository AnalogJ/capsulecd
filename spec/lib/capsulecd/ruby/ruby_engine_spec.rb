require 'spec_helper'

describe 'CapsuleCD::Ruby::RubyEngine', :ruby do
  describe '#build_step' do
    describe 'when building an empty package' do
      let(:engine) do
        require 'capsulecd/ruby/ruby_engine'
        CapsuleCD::Ruby::RubyEngine.new(source: :github,
                                            package_type: :ruby)
      end
      it 'should raise an error' do
        engine.instance_variable_set(:@source_git_local_path, test_directory )

        expect { engine.build_step }.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

    describe 'when building a simple package ' do
      let(:engine) do
        require 'capsulecd/ruby/ruby_engine'
        CapsuleCD::Ruby::RubyEngine.new(source: :github,
                                            package_type: :ruby)
      end
      it 'should create a .gitignore file and spec folder' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)

        engine.instance_variable_set(:@source_git_local_path, test_directory)

        VCR.use_cassette('gem_build_step',:tag => :ruby) do
          engine.build_step
        end

        expect(File.exist?(test_directory+'/.gitignore')).to eql(true)
      end

      it 'should raise an error if version.rb is missing' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
        FileUtils.rm(test_directory + '/lib/gem_analogj_test/version.rb')
        engine.instance_variable_set(:@source_git_local_path, test_directory)

        VCR.use_cassette('gem_build_step_without_version.rb',:tag => :ruby) do
          expect{engine.build_step}.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
        end

      end
    end
  end

  describe '#test_step' do
    before(:each) do
      FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
      FileUtils.cp('spec/fixtures/ruby/gem_analogj_test-0.1.4.gem', test_directory)

    end
    let(:engine) do
      require 'capsulecd/ruby/ruby_engine'
      CapsuleCD::Ruby::RubyEngine.new(source: :github,
                                          package_type: :ruby)
    end
    let(:config_double) { CapsuleCD::Configuration.new }
    describe 'when testing ruby package' do
      it 'should run install dependencies' do
        allow(Open3).to receive(:popen3).and_return(false)
        allow(config_double).to receive(:engine_cmd_test).and_call_original
        allow(config_double).to receive(:engine_disable_test).and_call_original

        engine.instance_variable_set(:@source_git_local_path, test_directory)
        engine.instance_variable_set(:@config, config_double)

        engine.test_step
      end
    end
  end

  describe '#release_step' do
    let(:engine) do
      require 'capsulecd/ruby/ruby_engine'
      CapsuleCD::Ruby::RubyEngine.new(source: :github,
                                      package_type: :ruby)
    end
    let(:config_double) { CapsuleCD::Configuration.new }
    describe 'when no rubygems_api_key provided' do
      it 'should raise an error' do
        engine.instance_variable_set(:@config, config_double)

        expect{engine.release_step}.to raise_error(CapsuleCD::Error::ReleaseCredentialsMissing)
      end
    end
  end

  describe 'integration tests' do
    let(:engine) do
      require 'capsulecd/ruby/ruby_engine'
      CapsuleCD::Ruby::RubyEngine.new(source: :github,
                                          runner: :default,
                                          package_type: :ruby,
                                          config_file: 'spec/fixtures/sample_ruby_configuration.yml'
      # config_file: 'spec/fixtures/live_ruby_configuration.yml'
      )
    end
    let(:git_commit_double) { instance_double(Git::Object::Commit) }
    describe 'when testing ruby package' do
      it 'should complete successfully' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)

        VCR.use_cassette('integration_ruby',:tag => :ruby) do
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
          allow(CapsuleCD::GitUtils).to receive(:create_gitignore).with(source_git_local_path, ['Ruby']).and_return(true)

          #stub methods in package_step
          allow(CapsuleCD::GitUtils).to receive(:commit).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:tag).with(source_git_local_path,'v0.1.4').and_return(git_commit_double)
          allow(git_commit_double).to receive(:sha).and_return('7d10007c0e1c6262d5a93cc2d3225c1c651fa13a')
          allow(git_commit_double).to receive(:name).and_return('v0.1.4')

          #stub methods in release_step
          allow(Open3).to receive(:popen3).with('gem push gem_analogj_test-0.1.4.gem',{:chdir=>source_git_local_path}).and_return(true)
          allow(FileUtils).to receive(:mkdir_p).with(File.expand_path('~/.gem')).and_return(true)
          allow(File).to receive(:open).with(File.expand_path('~/.gem/credentials'), 'w+').and_return(true)

          #stub methods in source_release
          allow(CapsuleCD::GitUtils).to receive(:push).and_return(true)
          allow(CapsuleCD::GitUtils).to receive(:generate_changelog).and_return('')

          engine.start

        end

      end
    end
  end

end
