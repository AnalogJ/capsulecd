begin
  require 'rspec/core/rake_task'
  RSpec::Core::RakeTask.new(:node_engine) do |t|
    t.rspec_opts = '--tag node_engine'
  end

  RSpec::Core::RakeTask.new(:chef_engine) do |t|
    t.rspec_opts = '--tag chef_engine'
  end

  RSpec::Core::RakeTask.new(:test) do |t|
    t.rspec_opts = '--tag ~chef_engine --tag ~node_engine'
  end

  task default: :test
rescue LoadError
end
