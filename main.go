package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	log.Fatal(initServer())
}

func initServer() error {
	e := echo.New()

	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.Recover())
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return nil
	})

	e.GET("/status", func(c echo.Context) error {
		totalPending := int64(0)

		for m, n := range WorkerMethodQueueNameMap {
			if m == WorkerMethodFIFO || m == WorkerMethodLIFO {
				totalPending += redisConn.LLen(context.Background(), n).Val()
			} else {
				totalPending += redisConn.ZCard(context.Background(), n).Val()
			}
		}

		totalRunning, _ := redisConn.Get(context.Background(), runningCounterName).Int64()
		totalFinished, _ := redisConn.Get(context.Background(), FinishedCounterName).Int64()
		totalElapsedTime, _ := redisConn.Get(context.Background(), totalTimeCounter).Int64()
		avgJobTime := int64(0)

		if totalFinished > 0 {
			avgJobTime = totalElapsedTime / totalFinished
		}

		expectedTimeToFinish := totalPending * avgJobTime

		return c.JSON(200, map[string]interface{}{
			"success": true,
			"payload": map[string]int64{
				"pending":                totalPending,
				"running":                totalRunning,
				"finished":               totalFinished,
				"avg_exec_time:sec":      avgJobTime,
				"expected_end_after:sec": expectedTimeToFinish,
			},
		})
	})

	e.POST("/enqueue", func(c echo.Context) error {
		var input struct {
			Worker          string            `json:"worker"`
			Spider          string            `json:"spider"`
			Args            map[string]string `json:"args"`
			MaxExecDuration string            `json:"max_execution_time"`
			Weight          int               `json:"weight"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		maxExecDur, err := time.ParseDuration(input.MaxExecDuration)
		if err != nil {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		workerConfig, exists := config.Workers[input.Worker]
		if !exists {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		queueName := WorkerMethodQueueNameMap[workerConfig.Method]
		if !exists {
			return c.JSON(400, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		job := NewJob(input.Spider, input.Args, maxExecDur)
		jobAsString, err := job.Serialize()
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		executeNow := false

		if _, exists := c.QueryParams()["force"]; exists {
			executeNow = true
		}

		// queue the execution
		if !executeNow {
			if workerConfig.Method == WorkerMethodWeight {
				if err := redisConn.ZIncrBy(context.Background(), queueName, float64(input.Weight), jobAsString).Err(); err != nil {
					return c.JSON(500, map[string]interface{}{
						"success": false,
						"error":   err.Error(),
					})
				}

				return c.JSON(201, map[string]interface{}{
					"success": true,
					"payload": job,
				})
			}

			if err := redisConn.LPush(context.Background(), queueName, jobAsString).Err(); err != nil {
				return c.JSON(500, map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				})
			}

			return c.JSON(201, map[string]interface{}{
				"success": true,
				"payload": job,
			})
		}

		// execute and return the result now
		// this is useful for debugging purposes.
		result, err := job.Dispatch()
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		}

		return c.JSON(200, map[string]interface{}{
			"success": true,
			"payload": result,
		})
	})

	return e.Start(*&config.ListenAddr)
}
