# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'gem_analogj_test/version'

Gem::Specification.new do |spec|
  spec.name          = "gem_analogj_test"
  spec.version       = GemTest::VERSION
  spec.authors       = ["Jason Kulatunga"]
  spec.email         = ["jk17@ualberta.ca"]

  spec.summary       = 'this is my test summary'
  spec.description   = 'this is my test description'
  spec.homepage      = "http://www.github.com/Analogj/gem_analogj_test"
  spec.license       = "MIT"

  spec.files         = `git ls-files -z`.split("\x0").reject { |f| f.match(%r{^(test|spec|features)/}) }
  spec.bindir        = "exe"
  spec.executables   = spec.files.grep(%r{^exe/}) { |f| File.basename(f) }
  spec.require_paths = ["lib"]

  spec.add_development_dependency "bundler", "~> 1.11"
  spec.add_development_dependency "rake", "~> 10.0"
  spec.add_development_dependency "rspec", "~> 3.0"
end
