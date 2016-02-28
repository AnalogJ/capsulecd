begin
  require 'rspec/core/rake_task'
  PACKAGE_TYPES = Dir.entries('lib/capsulecd').select {|entry|
    File.directory? File.join('lib/capsulecd',entry) and !(entry =='.' || entry == '..' || entry == 'base')
  }

  namespace :spec do

    namespace :suite do
      #spec:suite tests are language specific. only the 'python', 'javascript', 'chef', etc tests are run.
      PACKAGE_TYPES.each{|type|
        RSpec::Core::RakeTask.new(type.to_sym) do |t|
          t.rspec_opts = '--tag '+type
        end
      }
    end

    #spec tests run the language specific tests as well as the unit tests
    PACKAGE_TYPES.each{|type|
      task type.to_sym => ['spec:unit', 'spec:suite:'+type]
    }

    RSpec::Core::RakeTask.new(:unit) do |t|
      options = ''
      PACKAGE_TYPES.each{|type|
        options +=' --tag ~'+type
      }
      t.rspec_opts = options
    end

  end

  task :test => 'spec:unit'
  task :default => 'spec:unit'

  rescue LoadError
end
