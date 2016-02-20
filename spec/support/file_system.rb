module CapsuleCD
  module RSpecSupport
    module FileSystem
      def test_directory
        'spec/tmp'
      end

      def test_directory_contents
        Dir.glob(File.join(test_directory, '*/**'))
      end

      def make_test_directory
        FileUtils.mkdir_p(test_directory)
      end

      def remove_test_directory
        FileUtils.remove_dir(test_directory) if Dir.exist?(test_directory)
      end
    end
  end
end
