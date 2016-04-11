require 'uri'
module CapsuleCD
  module Runner
    module Circleci
      # def runner_retrieve_payload(_options)
      #   puts 'circleci runner_retrieve_payload'
      #
      #   # circleci only works with github, no need to parse @options[:source]
      #   # here are the environmental variables we need to handle:
      #   # https://circleci.com/docs/environment-variables
      #
      #   if @config.runner_pull_request.to_s.empty?
      #     puts 'This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.'
      #     @runner_is_pullrequest = false
      #     # make this as similar to the pull request payload as possible.
      #     payload = {
      #       'head' => {
      #         'sha' => @config.runner_sha,
      #         'ref' => @config.runner_branch,
      #         'repo' => {
      #           'clone_url' => @config.runner_clone_url,
      #           'name' => @config.runner_repo_name,
      #           'full_name' => @config.runner_repo_full_name
      #         }
      #       }
      #     }
      #
      #     payload
      #   else
      #     @runner_is_pullrequest = true
      #     pull_request_number =  @config.runner_pull_request.to_i
      #     @source_client.pull_request(@config.runner_repo_full_name, pull_request_number)
      #   end
      # end
    end
  end
end
