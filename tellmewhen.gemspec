require 'rubygems'


SPEC = Gem::Specification.new do |s|
  s.name = "tellmewhen"
  s.version = "1.0.1"
  s.author = "Kyle Burton"
  s.email = "kyle.burton@gmail.com"
  s.platform = Gem::Platform::RUBY
  s.description = <<DESC
Notifys you when another command completes (via email).

    ./tellmewhen 'sleep 3; ls'       # await a command to complete
    ./tellmewhen -p 12345            # await a pid to exit
    ./tellmewhen -e some-file.txt    # await existance
    ./tellmewhen -m some-file.txt    # await update

Tells you when, how long and sends you the output (for commands it runs).

DESC
  s.summary = "Notifys you when another command completes (via email)."
  # s.rubyforge_project = "typrtail"
  s.homepage = "http://github.com/kyleburton/tellmewhen"
  s.files = Dir.glob("**/*")
  s.executables << "tellmewhen"
  s.require_path = "bin"
  s.has_rdoc = false
end
