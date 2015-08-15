require_relative '../base/engine'

class ChefEngine < Engine
  def build_step()
    #bump up the chef cookbook version

    #create any base missing folders/files

    #the cookbook has already been downloaded. lets make sure all its dependencies are available.
    Open3.popen3('berks install', :chdir => @source_git_clone_path) do |stdin, stdout, stderr, external|
      {:stdout => stdout, :stderr => stderr}. each do |name, stream_buffer|
        Thread.new do
          until (line = stream_buffer.gets).nil? do
            puts "#{name} -> #{line}"
          end
        end
      end
      #wait for process
      external.join
      if !external.value.success?
        raise 'berks install failed. Check dependencies'
      end
    end

    # lets download all its gem dependencies



  end

  def test_step()
    # run rake test
  end

  def package_step()
  end

  def release_step()
  end
end