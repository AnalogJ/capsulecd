require 'capsulecd'
module CapsuleCD
  class ValidationUtils
    # TODO: validation almost needs to be source specific (or inherit from this base function), because source methods
    # may require additional attributes, while these base payload keys are required for general step functions.
    def self.validate_repo_payload(repo_payload)
      unless repo_payload['sha']
        fail CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "sha" key'
      end
      unless repo_payload['ref']
        fail CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "ref" key'
      end
      unless repo_payload['repo']
        fail CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "repo" key'
      end
      unless repo_payload['repo']['clone_url']
        fail CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "clone_url" key'
      end
      unless repo_payload['repo']['name']
        fail CapsuleCD::Error::SourcePayloadFormatError, 'Incorrectly formatted payload, missing "name" key'
      end
    end
  end
end
