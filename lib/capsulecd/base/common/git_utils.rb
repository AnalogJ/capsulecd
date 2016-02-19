require 'git'
class GitUtils

  def self.clone(parent_path, repository_name, git_remote)
    repo_path = File.expand_path("#{parent_path}/#{repository_name}/")
    if !File.directory?(repo_path)
      FileUtils.mkdir_p(repo_path)
    else
      fail 'the repository path already exists, this should never happen'
    end

    repo = Git.clone(git_remote, '', path: repo_path)
    repo.config('user.name', 'CapsuleCD')
    repo.config('user.email', 'CapsuleCD@users.noreply.github.com')
    repo.dir.to_s
  end

  def self.fetch(repo_path, remote_ref, local_branch)
    repo = Git.open(repo_path)
    repo.fetch(['origin', "#{remote_ref}:#{local_branch}"])
  end

  def self.checkout(repo_path, branch)
    repo = Git.open(repo_path)
    repo.checkout(branch)
  end

  def self.commit(repo_path, message, _all = true)
    repo = Git.open(repo_path)
    repo.add(all: true)
    repo.commit_all(message)
  end

  def self.tag(repo_path, version)
    repo = Git.open(repo_path)
    repo.add_tag(version)
  end

  def self.push(repo_path, local_branch, remote_branch)
    repo = Git.open(repo_path)
    repo.push('origin', "#{local_branch}:#{remote_branch}", tags: true)
  end

  # gets the HEAD commit on current branch
  def self.head_commit(repo_path)
    repo = Git.open(repo_path)
  end

  def self.generate_changelog(repo_path, base_sha, head_sha, full_name)
    repo = Git.open(repo_path)
    markdown = "Timestamp |  SHA | Message | Author \n"
    markdown += "------------- | ------------- | ------------- | ------------- \n"
    repo.log.between(base_sha, head_sha).each do |commit|
      markdown += "#{commit.date.strftime('%Y-%m-%d %H:%M:%S%z')} | [`#{commit.sha.slice 0..8}`](https://github.com/#{full_name}/commit/#{commit.sha}) | #{commit.message.gsub!('|', '!') || '--'} | #{commit.author.name}) \n"
    end
    markdown
  end
end
