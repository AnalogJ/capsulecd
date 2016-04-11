require 'spec_helper'
require 'capsulecd/base/runner/default'

describe CapsuleCD::Runner::Default do
  describe '#runner_retrieve_payload' do
    let(:default_runner) {
      Class.new { include CapsuleCD::Runner::Default }
    }
    describe 'when no config.runner_pull_request is set' do
      let(:config) { CapsuleCD::Configuration.new({
          :runner_sha => '0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c',
          :runner_branch => 'master',
          :runner_clone_url => 'https://github.com/analogj/capsulecd.git',
          :runner_repo_name => 'capsulecd',
          :runner_repo_full_name => 'AnalogJ/capsulecd'
                                                  })
      }
      it 'should populate a branch release payload' do
        runner = default_runner.new
        runner.instance_variable_set(:@config, config)

        payload = runner.runner_retrieve_payload({})

        expect(runner.instance_variable_get(:@runner_is_pullrequest)).to eql(false)
        expect(payload).to eql({
                                   'head' => {
                                     'sha' => '0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c',
                                     'ref' => 'master',
                                     'repo' => {
                                         'clone_url' => 'https://github.com/analogj/capsulecd.git',
                                         'name' => 'capsulecd',
                                         'full_name' => 'AnalogJ/capsulecd'
                                     }
                                 }
                               })
      end
    end

    describe 'when config.runner_pull_request is set' do
      let(:config) { CapsuleCD::Configuration.new({
                                                      :runner_pull_request => '4',
                                                  })
      }
      let(:source_client_double) { instance_double(Octokit::Client) }

      it 'should retrieve the payload from source' do
        allow(source_client_double).to receive(:pull_request).and_return({:test => :payload})

        runner = default_runner.new
        runner.instance_variable_set(:@config, config)
        runner.instance_variable_set(:@source_client, source_client_double)

        payload = runner.runner_retrieve_payload({})

        expect(runner.instance_variable_get(:@runner_is_pullrequest)).to eql(true)
        expect(payload).to eql({:test => :payload})
      end
    end

  end

end
