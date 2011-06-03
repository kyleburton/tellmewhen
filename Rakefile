require 'rubygems'
require 'rake'
require 'rake/gempackagetask'
require 'spec/rake/spectask'
load 'tellmewhen.gemspec'
 
 
Rake::GemPackageTask.new(SPEC) do |t|
  t.need_tar = true
end

desc "push the gem"
task :push do
  gemspec = "tellmewhen.gemspec"

  gemspec_file = File.basename(gemspec)
  gemfile_basename = File.basename(gemspec_file,'.gemspec')
  system "gem push pkg/#{gemfile_basename}-*.gem" 
end
