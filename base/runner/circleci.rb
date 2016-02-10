module CircleciRunner

  def runner_retrieve_payload(options)
    print 'circleci runner_retrieve_payload'

    #circleci only works with github, no need to parse @options[:source]
    # here are the environmental variables we need to handle:
    # https://circleci.com/docs/environment-variables

    unless ENV['CI_PULL_REQUEST']
      raise 'This is not a pull request. No automatic continuous deployment processing required. exiting..'
    end

    @source_client.pull_request(ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'], ENV['CI_PULL_REQUEST'].to_i)
  end

end