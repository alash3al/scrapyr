package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	config    *Config
	redisConn *redis.Client
)

const (
	fifoQueueName   = "scrapyr.queue::fifo"
	lifoQueueName   = "scrapyr.queue::lifo"
	weightQueueName = "scrapyr.queue::weight"

	runningCounterName  = "scrapyr.running.counter"
	FinishedCounterName = "scrapyr.finished.counter"
	totalTimeCounter    = "scrapyr.elapsed_time.total"
)

func init() {

	var err error

	{
		config, err = ParseConfigFile("scrapyr.hcl")
		if err != nil {
			log.Fatal(err.Error())
		}

	}

	// prepare redis connection
	{
		opts, err := redis.ParseURL(config.RedisDSN)
		if err != nil {
			log.Fatal(err)
		}

		redisConn = redis.NewClient(opts)

		if err := redisConn.Ping(context.Background()).Err(); err != nil {
			log.Fatal(err)
		}
	}

	// mount settings.py if needed
	{
		if strings.TrimSpace(config.SettingsPy) != "" {
			config.SettingsPy = "# Scrapyr Mounted File \n#--------------------------------------\n" + config.SettingsPy
			filename := path.Join(config.Scrapy.ProjectDir, path.Base(config.Scrapy.ProjectDir), "settings.py")
			if err := ioutil.WriteFile(filename, []byte(config.SettingsPy), 0644); err != nil {
				log.Fatal(err.Error())
			}
		}
	}

	// workers
	{
		for workerName, workerConfig := range config.Workers {
			go (func(workerName string, workerConfig WorkerConfig) {
				for i := 0; i < workerConfig.MaxProcs; i++ {
					go (func() {
						for {
							var item string

							if workerConfig.Method == WorkerMethodFIFO {
								item = redisConn.BLPop(context.Background(), 0, fifoQueueName).Val()[1]
							} else if workerConfig.Method == WorkerMethodLIFO {
								item = redisConn.BRPop(context.Background(), 0, lifoQueueName).Val()[1]
							} else if workerConfig.Method == WorkerMethodWeight {
								val := redisConn.BZPopMax(context.Background(), 0, weightQueueName).Val()
								if nil == val {
									continue
								}
								item = val.Member.(string)
							}

							job, err := UnserializeJob(item)
							if err != nil {
								catchErr(fmt.Errorf("unserializeJOB:: %s", err.Error()))
								continue
							}

							incrRunning(1)
							startedAt := time.Now()
							if _, err := job.Dispatch(); err != nil {
								catchErr(fmt.Errorf("dispatch:: %s", err.Error()))
							}
							incrFinished(1)
							incrRunning(-1)
							incrElapsedTime(int64(time.Now().Sub(startedAt).Seconds()))
						}
					})()
				}
			})(workerName, workerConfig)
		}
	}
}
