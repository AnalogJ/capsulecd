require 'git'
class GitUtils

  def self.clone_repository(parent_path,repository_name, git_remote)
    repo_path = File.expand_path("#{parent_path}/#{repository_name}/")
    if (!File.directory?(repo_path))
      FileUtils.mkdir_p(repo_path)
    else
      raise 'the repository path already exists, this should never happen'
    end

    repo = Git.clone(git_remote, '', :path => repo_path)
    repo.dir.to_s
  end

  def self.fetch(repo_path,remote_ref, local_branch)
    repo = Git.open(repo_path, remote_ref, local_branch)
    repo.fetch(['origin', "#{remote_ref}:#{local_branch}"])

  end

  def self.checkout(repo_path,branch)
    repo = Git.open(repo_path)
    repo.checkout(branch)

  end

end