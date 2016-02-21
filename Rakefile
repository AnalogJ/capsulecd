begin
  require 'rspec/core/rake_task'
  RSpec::Core::RakeTask.new(:node) do |t|
    t.rspec_opts = '--tag node'
  end

  RSpec::Core::RakeTask.new(:chef) do |t|
    t.rspec_opts = '--tag chef'
  end

  RSpec::Core::RakeTask.new(:python) do |t|
    t.rspec_opts = '--tag python'
  end

  RSpec::Core::RakeTask.new(:test) do |t|
    t.rspec_opts = '--tag ~chef --tag ~node --tag ~python'
  end

  task default: :test
rescue LoadError
end
