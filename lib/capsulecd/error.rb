module CapsuleCD
  # The collection of Minimart specific errors.
  module Error
    class BaseError < StandardError; end

    # Raised when capsule cannot create an authenticated client for the source.
    class SourceAuthenticationFailed < BaseError; end

    # Raised when there is an error parsing the repo payload format.
    class SourcePayloadFormatError < BaseError; end

    # Raised when a source payload is unsupported/action is invalid
    class SourcePayloadUnsupported < BaseError; end

    # Raised when the user who started the packaging is unauthorized (non-collaborator)
    class SourceUnauthorizedUser < BaseError; end

    # Raised when the package is missing certain required files (ie metadata.rb, package.json, setup.py, etc)
    class BuildPackageInvalid < BaseError; end

    # Raised when package dependencies fail to install correctly.
    class TestDependenciesError < BaseError; end

    # Raised when the package test runner fails
    class TestRunnerError < BaseError; end

    # Raised when credentials required to upload/deploy new package are missing.
    class ReleaseCredentialsMissing < BaseError; end

    # Raised when an error occurs while uploading package.
    class ReleasePackageError < BaseError; end

    # Gracefully handle any errors raised by CapsuleCD, and exit with a failure
    # status code.
    # @param [CapsuleCD::Error::BaseError] ex
    def self.handle_exception(ex)
      puts ex.message
      # Configuration.output.puts_red(ex.message)
      exit false
    end
  end
end
