require 'capsulecd'
class ValidationUtils
  def self.validate_repo_payload(repo_payload)
    unless repo_payload['repo']
      raise CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "repo" key'
    end
    unless repo_payload['repo']['clone_url']
      raise CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "clone_url" key'
    end
    unless repo_payload['repo']['branch']
      raise CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "branch" key'
    end
    unless repo_payload['repo']['name']
      raise CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "name" key'
    end
    unless repo_payload['repo']['sha1']
      raise CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "sha1" key'
    end
  end
end