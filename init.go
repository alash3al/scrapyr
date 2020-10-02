package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	config    *Config
	redisConn *redis.Client

	fifoQueueName   = "scrapyr.queue::fifo"
	lifoQueueName   = "scrapyr.queue::lifo"
	weightQueueName = "scrapyr.queue::weight"
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

	// workers
	{
		for workerName, workerConfig := range config.Workers {
			go (func(workerName string, workerConfig WorkerConfig) {
				for i := 0; i < workerConfig.MaxProcs; i++ {
					go (func() {
						for {
							var items []string

							if workerConfig.Method == WorkerMethodFIFO {
								items = redisConn.BLPop(context.Background(), 0, fifoQueueName).Val()
							} else if workerConfig.Method == WorkerMethodLIFO {
								items = redisConn.BRPop(context.Background(), 0, lifoQueueName).Val()
							} else if workerConfig.Method == WorkerMethodWeight {
								val := redisConn.BZPopMax(context.Background(), 0, weightQueueName).Val()
								items = []string{val.Member.(string)}
							}

							if len(items) < 1 {
								continue
							}

							job, err := UnserializeJob(items[0])
							if err != nil {
								catchErr(err)
								continue
							}

							if _, err := job.Dispatch(); err != nil {
								catchErr(err)
							}
						}
					})()
				}
			})(workerName, workerConfig)
		}
	}
}
