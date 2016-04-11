require 'spec_helper'

describe CapsuleCD::Engine do

  describe '::new' do
    describe 'without a source specified' do
      it 'should throw an error' do
        expect{CapsuleCD::Engine.new({})}.to raise_error(CapsuleCD::Error::SourceUnspecifiedError)
      end
    end
  end

  describe '::bump_version' do
    subject{
      CapsuleCD::Engine.new({:source => :github})
    }
    it 'should default to bumping patch segement' do
      version = SemVer.parse('1.0.2')
      new_version = subject.send(:bump_version, version)
      expect(new_version.to_s).to eql('1.0.3')
    end

    describe 'when engine_version_bump_type is :minor' do
      subject{
        CapsuleCD::Engine.new({
                                  :source => :github,
                                  :engine_version_bump_type => :minor
                              })
      }
      it 'should correctly bump minor and clear patch segement' do
        version = SemVer.parse('1.0.2')
        new_version = subject.send(:bump_version, version)
        expect(new_version.to_s).to eql('1.1.0')
      end
    end

    describe 'when engine_version_bump_type is :major' do
      subject{
        CapsuleCD::Engine.new({
                                  :source => :github,
                                  :engine_version_bump_type => :major
                              })
      }
      it 'should correctly bump major and clear minor and patch segements' do
        version = SemVer.parse('1.0.2')
        new_version = subject.send(:bump_version, version)
        expect(new_version.to_s).to eql('2.0.0')
      end
    end
  end
end
