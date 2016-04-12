# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'capsulecd/version'

Gem::Specification.new do |spec|
  spec.name          = 'capsulecd'
  spec.version       = CapsuleCD::VERSION
  spec.summary       = 'CapsuleCD is a library for automating package releases (npm, cookbooks, gems, pip, jars, etc)'
  spec.description   = 'CapsuleCD is a library for automating package releases (npm, cookbooks, gems, pip, jars, etc)'
  spec.authors       = ['Jason Kulatunga (AnalogJ)']
  spec.homepage      = 'http://www.github.com/AnalogJ/capsulecd'
  spec.license       = 'MIT'

  spec.required_ruby_version = '~> 2.0'

  spec.files         = `git ls-files -z`.split("\x0")
  spec.executables   = spec.files.grep(%r{^bin/}) { |f| File.basename(f) }
  spec.test_files    = spec.files.grep(%r{^(test|spec|features)/})
  spec.require_paths = ['lib']

  spec.add_dependency 'thor'
  spec.add_dependency 'json'
  spec.add_dependency 'git'
  spec.add_dependency 'semverly'
  spec.add_dependency 'mkgitignore'
end
