module CapsuleCD
  module RSpecSupport
    module PackageTypes
      def package_types
        Dir.entries('lib/capsulecd').select {|entry|
          File.directory? File.join('lib/capsulecd',entry) and !(entry =='.' || entry == '..' || entry == 'base')
        }
      end
    end
  end
end
