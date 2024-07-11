Debugging
---------

* We have chosen delve for this task.

1. Install: `go get -u github.com/go-delve/delve/cmd/dlv`
2. Init debugger: `go build nebulant.go && ../../go/bin/dlv exec ./nebulant` (use your own $GOPATH/bin/dlv)
3. Set breakpoint: `b Control.Start`
4. Set breakpoint at file:line: `b nebulant.go:11`
5. Navigate: Use common `continue`, `next`, `step` (or alias `c`, `n`, `s`)
6. View vars/mem/stack: `print varname`, `locals`, `args`, `stack`
7. More commands: `help` ;)

Run while hack
--------------

- set `AWS_PROFILE` to the profile you configured in config and credentials
- add ssh key to agent as needed: `ssh-add ../../.ssh/nebulant-keypair-fortesting.pem`
- run `export AWS_PROFILE=nebulant-cli-tests; make run ARGS="examples/dummy.json"`
- run as server mode `make run ARGS="-d"`


Profiling
---------

For security reasons we have decided to disable profiling in the official version, however we still maintain this feature commented in the code. Uncomment the indicated parts of asdf.go to start profiling.

1. Uncomment profiling code and compile:

```go
	// hey hacker:
	// uncomment for profiling
	// _ "net/http/pprof"
```

```go
	// if config.PROFILING {
	// 	go func() {
	// 		cast.LogInfo("Starting profiling at localhost:6060", nil)
	// 		http.ListenAndServe("localhost:6060", nil)
	// 	}()
	// 	grmon.Start()
	// }
```

2. add `export NEBULANT_PROFILING=true` to open an http profiler in port 6060. The -d option also applies to port 15678.
3. Use go tool and pprof for profiling. 
 
Examples:


- Trace during 5s and open web profiler:
```
$ wget -O trace.out http://localhost:6060/debug/pprof/trace\?seconds\=5
$ go tool trace trace.out
```

- Profile heap:
```shell
$ go tool pprof http://localhost:6060/debug/pprof/heap
Fetching profile over HTTP from http://localhost:6060/debug/pprof/heap
Saved profile in /Users/user/pprof/pprof.alloc_objects.alloc_space.inuse_objects.inuse_space.005.pb.gz
Type: inuse_space
Time: Jan 6, 2021 at 8:55pm (CET)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) inuse_objects
(pprof) top
Showing nodes accounting for 3641, 100% of 3641 total
      flat  flat%   sum%        cum   cum%
      3641   100%   100%       3641   100%  github.com/aws/aws-sdk-go/aws/endpoints.init
         0     0%   100%       3641   100%  runtime.doInit
         0     0%   100%       3641   100%  runtime.main
(pprof) q
```

You can read more about pprof in the package documentation [https://golang.org/pkg/net/http/pprof/](https://golang.org/pkg/net/http/pprof/). And also you can read this helpful article: [https://www.freecodecamp.org/news/how-i-investigated-memory-leaks-in-go-using-pprof-on-a-large-codebase-4bec4325e192/](https://www.freecodecamp.org/news/how-i-investigated-memory-leaks-in-go-using-pprof-on-a-large-codebase-4bec4325e192/)


- Watch goroutines in live mode:
```
$ go get -u github.com/bcicen/grmon # only the first time (doc: https://github.com/bcicen/grmon)
~./go/bin/grmo
```

```
Keybindings (press pause to activate up/down and open trace):
| Key                 | Action                          |
| ------------------- | ------------------------------- |
| r                   | manually refresh                |
| p                   | pause/unpause automatic updates |
| s                   | toggle sort column and refresh  |
| f                   | filter by keyword               |
| \<up\>,\<down\>,j,k | move cursor position            |
| \<enter\>,o         | expand trace under cursor       |
| t                   | open trace in full screen mode  |
| q                   | exit grmon                      |
```
