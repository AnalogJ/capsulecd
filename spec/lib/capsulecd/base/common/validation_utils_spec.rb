require 'spec_helper'

describe 'CapsuleCD::ValidationUtils' do
  subject{
    CapsuleCD::ValidationUtils
  }
  describe '#validate_repo_payload' do
    let(:payload){
      {
          'sha' => '0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c',
          'ref' => 'mybranch',
          'repo' => {
              'clone_url' => 'https://github.com/analogj/capsulecd.git',
              'name' => 'capsulecd'
          }
      }
    }
    it 'should run successfully when parsing correctly structured payload' do
      expect(subject.validate_repo_payload(payload)).to eql(nil)
    end

    it 'should raise an error when payload is missing sha' do
      payload.delete('sha')
      expect{subject.validate_repo_payload(payload)}.to raise_error(CapsuleCD::Error::SourcePayloadFormatError)
    end

    it 'should raise an error when payload is missing ref' do
      payload.delete('ref')
      expect{subject.validate_repo_payload(payload)}.to raise_error(CapsuleCD::Error::SourcePayloadFormatError)
    end

    it 'should raise an error when payload is missing clone_url' do
      payload['repo'].delete('clone_url')
      expect{subject.validate_repo_payload(payload)}.to raise_error(CapsuleCD::Error::SourcePayloadFormatError)
    end

    it 'should raise an error when payload is missing name' do
      payload['repo'].delete('name')
      expect{subject.validate_repo_payload(payload)}.to raise_error(CapsuleCD::Error::SourcePayloadFormatError)
    end
  end

end
