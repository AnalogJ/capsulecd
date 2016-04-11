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

    describe 'when modifying gemspec file' do
      it 'should not keep old constant from version.rb file in memory' do
        FileUtils.copy_entry('spec/fixtures/ruby/gem_analogj_test', test_directory)

        gemspec_data = CapsuleCD::Ruby::RubyHelper.get_gemspec_data(test_directory)
        expect(gemspec_data.version.to_s).to eql('0.1.3')

        version_str = CapsuleCD::Ruby::RubyHelper.read_version_file(test_directory, gemspec_data.name)
        next_version = CapsuleCD::Engine.new(:source => :github).send(:bump_version, SemVer.parse(gemspec_data.version.to_s))
        expect(next_version.to_s).to eql('0.1.4')

        new_version_str = version_str.gsub(/(VERSION\s*=\s*['"])[0-9\.]+(['"])/, "\\1#{next_version}\\2")
        CapsuleCD::Ruby::RubyHelper.write_version_file(test_directory, gemspec_data.name, new_version_str)

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
          unless File.exist?(test_directory + "/#{gemspec_data.name}-#{next_version.to_s}.gem")
            fail CapsuleCD::Error::BuildPackageFailed, "gem build failed. #{gemspec_data.name}-#{next_version.to_s}.gem not found"
          end
        end

        updated_gemspec_data = CapsuleCD::Ruby::RubyHelper.get_gemspec_data(test_directory)
        expect(updated_gemspec_data.version.to_s).to eql('0.1.4')



      end
    end

  end

end
