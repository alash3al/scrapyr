scrapyd-go
===========
> an drop-in replacement for [scrapyd](https://github.com/scrapy/scrapyd) that is more easy to be scalable and distributed on any number of commodity machines with no hassle, each `scrapyd-go` instance is a stateless microservice, all instances must be connected to the same `redis` server, `redis` is used as a ceneralized registry system for all instances, so each instance se what others see.

Why
===
[scrapyd](https://github.com/scrapy/scrapyd) isn't bad, but it is very stateful, it isn't that easy to deploy in a destributed environment like `k8s`, also I wanted to add more features, so I started as a drop-in replacement of `scrapyd` but writing in modern & scalable environment like `go` for restful server and `redis` as centeralized registry.

Implementation
==============
- [x] `schedule.json` 
- [x] `cancel.json` 
- [x] `addversion.json`
- [x] `listprojects.json`
- [x] `listversions.json`
- [x] `listspiders.json`
- [x] `delproject.json`
- [x] `delversion.json`
- [x] `listjobs.json`
- [x] `daemonstatus.json`
- [x] `logs/{jobid}`, *new*: realtime output of the job log

Configurations
===============
> `scrapyd-go` configs are just simple command line `flags`

```bash
  -dir string
        the directory to use for local caching (default ".scrapyd-go")
  -listen string
        the address to bind to (default ":6800")
  -max2keep int
        the maximum jobs/logs to keep in memory (default 1000000)
  -poll int
        time in millisecond between each poll operation from queue(s) (default 10)
  -redis string
        the redis server address (default "redis://:somepass@localhost:6379/1")
  -sync int
        time in seconds between each sync operation (default 15)
  -workers int
        the maximum workers count (default 8)
```

Installation
=============
- *binary* : go to [releases page](https://github.com/alash3al/scrapyd-go) and download your os based release
- *docker*: `$ docker pull alash3al/scrapyd-go`
- *source*: `$ go get github.com/alash3al/scrapyd-go`

Running
========
- *binary*: `$ ./scrapyd_bin_file -redis redis://localhost:6379/1`
- *docker*: `$ docker run --link SomeRedisServerContainer -p 6800:6800 alash3al/scrapyd-go -redis redis://SomeRedisServerContainer:6379/1`
- *source*: `$ scrapyd-go -redis redis://localhost:6379/1`

Contributing
=============
- Fork the repo
- Create a feature branch
- Push your changes
- Create a pull request

License
========
Apache License v2.0

Author
=======
- Mohamed Al Ashaal
- Software Engineer
- m7medalash3al@gmail.com
