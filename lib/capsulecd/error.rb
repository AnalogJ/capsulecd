module CapsuleCD
  # The collection of Minimart specific errors.
  module Error
    class BaseError < StandardError; end

    # Raised when capsule cannot create an authenticated client for the source.
    class SourceAuthenticationFailed < BaseError; end

    # Raised when there is an error parsing the repo payload format.
    class SourcePayloadFormatError < BaseError; end

    # Raised when a source payload is unsupported.
    class SourcePayloadUnsupported < BaseError; end

    # Raised when the user who started the packaging is unauthorized (non-collaborator)
    class SourceUnauthorizedUser < BaseError; end


    # Raised when Minimart encounters a cookbook with a location type that it can't handle
    class UnknownLocationType < BaseError; end

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
