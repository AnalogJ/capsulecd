require 'uri'
module CircleciRunner

  def runner_retrieve_payload(options)
    puts 'circleci runner_retrieve_payload'

    #circleci only works with github, no need to parse @options[:source]
    # here are the environmental variables we need to handle:
    # https://circleci.com/docs/environment-variables

    if ENV['CI_PULL_REQUEST'].to_s.empty?
      raise 'This is not a pull request. No automatic continuous deployment processing required. exiting..'
    end

    # parse the PR# from the environment variable, eg. https://github.com/AnalogJ/cookbook_analogj_test/pull/9
    pull_request_number =  File.basename(URI.parse(ENV['CI_PULL_REQUEST']).path).to_i # => baz

    @source_client.pull_request(ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'], pull_request_number)
  end

end