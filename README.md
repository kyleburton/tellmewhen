# tellmewhen

Tell me when "something" has happened (eg: another program has finished).  Tell me when it started, tell when it finished.  Tell me how long it took.  Tell me the output and if it failed.

I created this utility because I had long running database scripts that I wanted a completion notification for - including a summary of the time and exit code.  I had done this various ways in the past by using bash and other shell utilities.  This program captures all of the features and behavior I wanted without having to script it from scratch again every time I needed it.

* json Based Configuration from a local file
* sensible default configuration
* can run a command, make an http call, send an email
* includes start/stop/elapsed timing
* success/fail based on exit code
* watch an already running process (by pid)

# Installation

```bash
go install github.com/kyleburton/tellmewhen
```

# Usage

```
./tellmewhen ...
```

# Examples

```bash
# TODO: update for the golang version
./tellmewhen 'sleep 3; ls'
./tellmewhen -p 12345
./tellmewhen -e some-file.txt    # await existence
./tellmewhen -m some-file.txt    # await update
```

You can also have 'pending' notifications so that you know things are still going:

```
    ./tellmewhen -t 2 'sleep 5; ls'
```

Produces 2 emails.  The pending mail is:

```
Subject: When! [NOT] I'm _still_ waiting for sleep...
Body:

Just wanted to let you know that:

   sleep 5; ls

Is _STILL_ running on your-host.com, it has not exited.  You did want me to let you know didn't you?

I started the darn thing at Tue Jan 25 23:25:01 -0500 2011 (1296015901) and it has taken a total of 3 seconds so far.

Just thought you'd like to know.  I'll continue to keep watching what you asked me to. (boor-ring!)

Cordially,

Tellmewhen

P.S. stderr says:
--

--

P.S. stdout says:
--

--
```


And a final mail when it completed:

```
Subject: When! SUCCESS for sleep...
Body:

Just wanted to let you know that:

     sleep 5; ls

completed on your-host.com, with an exit code of: 0

It started at Tue Jan 25 23:25:01 -0500 2011 (1296015901), finished at Tue Jan 25 23:25:06 -0500 2011 (1296015906) and took a total of 5 seconds.

May your day continue to be full of win.

Sincerely,

Tellmewhen

P.S. stderr says:
--

--

P.S. stdout says:
--
README.textile
Rakefile
bin
foo
tellmewhen.gemspec

--
```

# Contributors

Kyle Burton <kyle.burton@gmail.com>

# License

"MIT":http://www.opensource.org/licenses/mit-license.php


# Changes

### 1.0.0 2024-08-17T19:13:19Z

Re-implementation in Go (from Ruby)
