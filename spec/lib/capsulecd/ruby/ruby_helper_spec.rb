require 'spec_helper'

describe 'CapsuleCD::Ruby::RubyHelper', :ruby do
  subject{
    CapsuleCD::Ruby::RubyHelper
  }

  describe '#get_gemspec_path' do
    describe 'without a gemspec file' do
      it 'should raise an error' do
        expect{subject.get_gemspec_path(test_directory)}.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

    describe 'with a valid gemspec file' do
      it 'should generate correct gemspec path' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)

        expect(subject.get_gemspec_path(test_directory)).to eql(test_directory + '/gem_analogj_test.gemspec')
      end
    end
  end

  describe '#get_gem_name' do
    it 'should parse gemspec data' do
      FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
      gem_name = subject.get_gem_name(test_directory)
      expect(gem_name).to eql('gem_analogj_test')
    end

    describe 'with an invalid gemspec file' do
      it 'should raise an error' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
        FileUtils.rm(test_directory + '/lib/gem_analogj_test/version.rb')

        expect{subject.get_gem_name(test_directory)}.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end
  end

  describe '#get_version' do

    it 'should parse version.rb file' do
      FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
      gem_version = subject.get_version(test_directory)
      expect(gem_version).to eql('0.1.3')
    end

    describe 'without a version.rb file' do
      it 'should raise an error' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
        FileUtils.rm(test_directory + '/lib/gem_analogj_test/version.rb')

        expect{subject.get_gem_name(test_directory)}.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

    describe 'with too many version.rb files' do
      it 'should fallback to using version_filepath & gem name' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
        FileUtils.mkdir_p(test_directory + '/lib/gem_analogj_test/test/')
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test/lib/gem_analogj_test/version.rb',
                             test_directory + '/lib/gem_analogj_test/test/version.rb')

        gem_version = subject.get_version(test_directory)
        expect(gem_version).to eql('0.1.3')

        #TODO: cant figure out how to verify that a class method was called.
        # expect(CapsuleCD::Ruby::RubyHelper).to receive(:version_filepath).with(test_directory, 'gem_analogj_test')
      end
    end
  end

  describe '#set_version' do
    it 'should correctly update version in version.rb file' do
      FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)

      gem_name = CapsuleCD::Ruby::RubyHelper.get_gem_name(test_directory)
      gem_version = CapsuleCD::Ruby::RubyHelper.get_version(test_directory)
      expect(gem_version).to eql('0.1.3')

      next_version = CapsuleCD::Engine.new(:source => :github).send(:bump_version, SemVer.parse(gem_version))
      expect(next_version.to_s).to eql('0.1.4')

      CapsuleCD::Ruby::RubyHelper.set_version(test_directory, next_version.to_s)

      Open3.popen3('gem build gem_analogj_test.gemspec', chdir: test_directory) do |_stdin, stdout, stderr, external|
        { stdout: stdout, stderr: stderr }. each do |name, stream_buffer|
          Thread.new do
            until (line = stream_buffer.gets).nil?
              puts "#{name} -> #{line}"
            end
          end
        end
        # wait for process
        external.join
        unless external.value.success?
          fail CapsuleCD::Error::BuildPackageFailed, 'gem build failed. Check gemspec file and dependencies'
        end
        unless File.exist?(test_directory + "/#{gem_name}-#{next_version.to_s}.gem")
          fail CapsuleCD::Error::BuildPackageFailed, "gem build failed. #{gem_name}-#{next_version.to_s}.gem not found"
        end
      end

      updated_gem_version = CapsuleCD::Ruby::RubyHelper.get_version(test_directory)
      expect(updated_gem_version).to eql('0.1.4')
    end

  end

end
