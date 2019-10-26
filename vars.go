package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-redis/redis"
)

var (
	flagListenAddr       = flag.String("listen", ":6800", "the address to bind to")
	flagRedisAddr        = flag.String("redis", "redis://:somepass@localhost:6379/1", "the redis server address")
	flagSyncInterval     = flag.Int64("sync", 15, "time in seconds between each sync operation")
	flagCacheDir         = flag.String("dir", ".scrapyd-go", "the directory to use for local caching")
	flagMaxWorkers       = flag.Int("workers", runtime.NumCPU(), "the maximum workers count")
	flagMaxToKeep        = flag.Int64("max2keep", 1000000, "the maximum jobs/logs to keep in memory")
	flagPollInterval     = flag.Int64("poll", 10, "time in millisecond between each poll operation from queue(s)")
	flagDefaultPythonBin = flag.String("python", "python3", "the python binary to use")
)

var (
	redisConn *redis.Client

	nodeName, nodeNameErr = os.Hostname()
	nodeIP                = resolveHostIP()
	cmds                  = map[string]*exec.Cmd{}
)

const (
	redisNodesDiscovery   = "scrapyd-go:nodes"
	redisRunningQueueKey  = "scrapyd-go:queue:running"
	redisPendingQueueKey  = "scrapyd-go:queue:pending"
	redisFinishedQueueKey = "scrapyd-go:queue:finished"
	redisProjectsKey      = "scrapyd-go:projects"
	redisEggsPrefix       = "scrapyd-go:eggs:"
	redisVersionsPrefix   = "scrapyd-go:versions:"
	redisJobsKey          = "scrapyd-go:jobs"
	redisLogsPrefix       = "scrapyd-go:logs:"
	redisCancelsKey       = "scrapyd-go:cancels"
)

func init() {
	flag.Parse()

	opt, err := redis.ParseURL(*flagRedisAddr)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := exec.LookPath(*flagDefaultPythonBin); err != nil {
		log.Fatal(err.Error())
	}

	redisConn = redis.NewClient(opt)
	if _, err := redisConn.Ping().Result(); err != nil {
		log.Fatal(err.Error())
	}

	redisConn.HSet(redisNodesDiscovery, nodeName, nodeIP)

	runWatchers()
}
