# tellmewhen

Tell me when "something" has happened (eg: another program has finished).

I created this utility because I had long running database scripts that I wanted a completion notification for.  I had done this various ways in the past by using bash and other shell utilities.  This program captures all of the features and behavior I wanted without having to script it from scratch again every time I needed it.

* can run a command, make an http call
* success/fail based on exit code
* watch an already running process (by pid)

# Installation

```bash
go install github.com/kyleburton/tellmewhen
```

# Usage

```bash
./tellmewhen ...
```

# Examples

```bash
####################
# when a process exits
tellmewhen process-exits \
  --command="sleep 2; date" \
  --notify-by-running="zenity --info --text='all the things are done' --title='Status'"

####################
# when a PID exits, in a terminal, run:
vim nothing-to-see-here.txt

# then, in another terminal
tellmewhen \
  --notify-by-running="zenity --info 'done'" \
  pid-exits \
  --pid="$(ps aux | grep [n]othing-to-see-here | awk '{print $2}')"

# go and exit vim, [if you can :)](https://stackoverflow.com/questions/11828270/how-do-i-exit-vim)

####################
# when a process succeeds
tellmewhen  \
  --notify-by-running="zenity \
  --info 'done'" process-succeeds \
  --pid="$(ps aux | grep [n]othing-to-see-here | awk '{print $2}')"
```

# Contributors

Kyle Burton <kyle.burton@gmail.com>

# License

"MIT":http://www.opensource.org/licenses/mit-license.php

# Features I'd like to See

* Socket related conditions: notify when we can connect to a socket, notify when a socket connection timesout, notify when a connection refused to a port (i.e. a service stops listening)
* HTTP Status related conditions: notify when a specific HTTP status code is returned for a GET or HEAD (404 - when something goes missing, or 200/201/202 when something exists.
* Alternate Notfication Options: HTTP POST, send an email (yes I know these can be implemented by short commands or shell scripts)
* Periodic Notifications: for very long running things we are waiting on, getting an intermediate notfication (i.e. "still running, started at XYZ, running for ELAPSED so far...", followed by a "Final" notification.

# Changes

### 1.0.0 2024-09-24T02:32:57Z

Re-implementation in Go (converted from Ruby)
