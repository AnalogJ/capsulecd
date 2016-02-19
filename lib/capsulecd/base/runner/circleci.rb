require 'uri'
module CircleciRunner

  def runner_retrieve_payload(options)
    puts 'circleci runner_retrieve_payload'

    #circleci only works with github, no need to parse @options[:source]
    # here are the environmental variables we need to handle:
    # https://circleci.com/docs/environment-variables

    if ENV['CI_PULL_REQUEST'].to_s.empty?
      puts 'This is not a pull request. No automatic continuous deployment processing required. Continuous Integration testing will continue.'
      @runner_is_pullrequest = false
      # make this as similar to the pull request payload as possible.
      payload = {
          'head'=> {
              'repo' => {
                  'clone_url' => 'https://github.com/'+ ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'] + '.git',
                  'name' => ENV['CIRCLE_PROJECT_REPONAME'],
                  'full_name' => ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'],
                  'branch' => ENV['CIRCLE_BRANCH'],
                  'sha' => ENV['CIRCLE_SHA1']
              }
          }
      }

      payload
    else
      @runner_is_pullrequest = true
      # parse the PR# from the environment variable, eg. https://github.com/AnalogJ/cookbook_analogj_test/pull/9
      pull_request_number =  File.basename(URI.parse(ENV['CI_PULL_REQUEST']).path).to_i # => baz
      @source_client.pull_request(ENV['CIRCLE_PROJECT_USERNAME'] + '/' + ENV['CIRCLE_PROJECT_REPONAME'], pull_request_number)
    end


  end

end