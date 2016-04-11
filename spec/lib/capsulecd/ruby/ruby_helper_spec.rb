require 'spec_helper'

describe 'CapsuleCD::Ruby::RubyHelper', :ruby do
  subject{
    CapsuleCD::Ruby::RubyHelper
  }
  describe '#version_filepath' do
    it 'default should generate the correct path to version.rb' do
      expect(subject.version_filepath('/tmp','capsulecd')).to eql('/tmp/lib/capsulecd/version.rb')
    end
    it 'with custom version filename should generate the correct path' do
      expect(subject.version_filepath('/tmp','capsulecd', 'VERSION.rb')).to eql('/tmp/lib/capsulecd/VERSION.rb')
    end
  end

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

  describe '#get_gemspec_data' do
    it 'should parse gemspec data' do
      FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
      gemspec_data = subject.get_gemspec_data(test_directory)
      expect(gemspec_data.name).to eql('gem_analogj_test')
      expect(gemspec_data.version.to_s).to eql('0.1.3')
    end

    describe 'with an invalid gemspec file' do
      it 'should raise an error' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)
        FileUtils.rm(test_directory + '/lib/gem_analogj_test/version.rb')

        expect{subject.get_gemspec_data(test_directory)}.to raise_error(CapsuleCD::Error::BuildPackageInvalid)
      end
    end

  end

end
