#fix for merging multiple test suites https://github.com/colszowka/simplecov/issues/350
#require 'coveralls'

#using ARGV.join to generate a new identifier for each coverage run (cant use PID because it'll always be 1 in the container)
SimpleCov.start do
  add_filter '/spec/'
  command_name "capsulecd_#{ARGV.join}"
  merge_timeout 360 # 1 hour
end